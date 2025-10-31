package update

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/daleiii/podsync-web/pkg/builder"
	"github.com/daleiii/podsync-web/pkg/db"
	"github.com/daleiii/podsync-web/pkg/feed"
	"github.com/daleiii/podsync-web/pkg/fs"
	"github.com/daleiii/podsync-web/pkg/history"
	"github.com/daleiii/podsync-web/pkg/model"
	"github.com/daleiii/podsync-web/pkg/progress"
	"github.com/daleiii/podsync-web/pkg/ytdl"
)

type Downloader interface {
	Download(ctx context.Context, feedConfig *feed.Config, episode *model.Episode) (io.ReadCloser, error)
	PlaylistMetadata(ctx context.Context, url string) (metadata ytdl.PlaylistMetadata, err error)
}

type TokenList []string

type Manager struct {
	hostname        string
	downloader      Downloader
	db              db.Storage
	fs              fs.Storage
	feeds           map[string]*feed.Config
	keys            map[model.Provider]feed.KeyProvider
	progressTracker *progress.Tracker
	historyManager  *history.Manager
}

func NewUpdater(
	feeds map[string]*feed.Config,
	keys map[model.Provider]feed.KeyProvider,
	hostname string,
	downloader Downloader,
	db db.Storage,
	fs fs.Storage,
	historyManager *history.Manager,
) (*Manager, error) {
	return &Manager{
		hostname:        hostname,
		downloader:      downloader,
		db:              db,
		fs:              fs,
		feeds:           feeds,
		keys:            keys,
		progressTracker: progress.New(),
		historyManager:  historyManager,
	}, nil
}

// GetProgressTracker returns the progress tracker for this manager
func (u *Manager) GetProgressTracker() *progress.Tracker {
	return u.progressTracker
}

// GetHistoryManager returns the history manager for this manager
func (u *Manager) GetHistoryManager() *history.Manager {
	return u.historyManager
}

func (u *Manager) Update(ctx context.Context, feedConfig *feed.Config) error {
	log.WithFields(log.Fields{
		"feed_id": feedConfig.ID,
		"format":  feedConfig.Format,
		"quality": feedConfig.Quality,
	}).Infof("-> updating %s", feedConfig.URL)

	started := time.Now()

	// Get feed info for history logging
	feedInfo, _ := u.db.GetFeed(ctx, feedConfig.ID)
	feedTitle := feedConfig.ID
	if feedInfo != nil && feedInfo.Title != "" {
		feedTitle = feedInfo.Title
	}

	// Log history entry start
	historyID, _ := u.historyManager.LogFeedUpdateStart(ctx, feedConfig.ID, feedTitle, model.TriggerScheduled)

	// Track statistics for history
	stats := model.JobStatistics{}
	var updateErr error

	if err := u.updateFeed(ctx, feedConfig); err != nil {
		updateErr = errors.Wrap(err, "update failed")
		u.logHistoryEnd(ctx, historyID, model.JobStatusFailed, stats, updateErr.Error())
		return updateErr
	}

	// Fetch episodes for download
	episodesToDownload, err := u.fetchEpisodes(ctx, feedConfig)
	if err != nil {
		updateErr = errors.Wrap(err, "fetch episodes failed")
		u.logHistoryEnd(ctx, historyID, model.JobStatusFailed, stats, updateErr.Error())
		return updateErr
	}

	stats.EpisodesQueued = len(episodesToDownload)

	// Collect episode IDs for history tracking
	episodeIDs := make([]string, len(episodesToDownload))
	for i, ep := range episodesToDownload {
		episodeIDs[i] = ep.ID
	}

	downloadedCount, failedCount, bytesDownloaded := u.downloadEpisodesWithStats(ctx, feedConfig, episodesToDownload)
	stats.EpisodesDownloaded = downloadedCount
	stats.EpisodesFailed = failedCount
	stats.BytesDownloaded = bytesDownloaded

	if err := u.cleanup(ctx, feedConfig); err != nil {
		log.WithError(err).Error("cleanup failed")
	}

	if err := u.buildXML(ctx, feedConfig); err != nil {
		updateErr = errors.Wrap(err, "xml build failed")
		u.logHistoryEnd(ctx, historyID, model.JobStatusFailed, stats, updateErr.Error())
		return updateErr
	}

	if err := u.buildOPML(ctx); err != nil {
		updateErr = errors.Wrap(err, "opml build failed")
		u.logHistoryEnd(ctx, historyID, model.JobStatusFailed, stats, updateErr.Error())
		return updateErr
	}

	elapsed := time.Since(started)
	log.Infof("successfully updated feed in %s", elapsed)

	// Determine final status
	status := model.JobStatusSuccess
	if stats.EpisodesFailed > 0 && stats.EpisodesDownloaded > 0 {
		status = model.JobStatusPartial
	} else if stats.EpisodesFailed > 0 {
		status = model.JobStatusFailed
	}

	u.logHistoryEndWithEpisodes(ctx, historyID, feedConfig.ID, episodeIDs, status, stats, "")
	return nil
}

// logHistoryEnd is a helper to log history end
func (u *Manager) logHistoryEnd(ctx context.Context, historyID string, status model.JobStatus, stats model.JobStatistics, errMsg string) {
	if u.historyManager != nil {
		_ = u.historyManager.LogFeedUpdateEnd(ctx, historyID, status, stats, errMsg)
	}
}

// logHistoryEndWithEpisodes is a helper to log history end with episode details
func (u *Manager) logHistoryEndWithEpisodes(ctx context.Context, historyID, feedID string, episodeIDs []string, status model.JobStatus, stats model.JobStatistics, errMsg string) {
	if u.historyManager != nil {
		_ = u.historyManager.LogFeedUpdateEndWithEpisodes(ctx, historyID, feedID, episodeIDs, status, stats, errMsg)
	}
}

// updateFeed pulls API for new episodes and saves them to database
func (u *Manager) updateFeed(ctx context.Context, feedConfig *feed.Config) error {
	info, err := builder.ParseURL(feedConfig.URL)
	if err != nil {
		return errors.Wrapf(err, "failed to parse URL: %s", feedConfig.URL)
	}

	keyProvider, ok := u.keys[info.Provider]
	if !ok {
		return errors.Errorf("key provider %q not loaded", info.Provider)
	}

	// Create an updater for this feed type
	provider, err := builder.New(ctx, info.Provider, keyProvider.Get(), u.downloader)
	if err != nil {
		return err
	}

	// Query API to get episodes
	log.Debug("building feed")
	result, err := provider.Build(ctx, feedConfig)
	if err != nil {
		return err
	}

	log.Debugf("received %d episode(s) for %q", len(result.Episodes), result.Title)

	// Build a set of episodes that should be removed
	// (episodes that are new/error but no longer in the feed)
	episodeSet := make(map[string]struct{})
	blockedEpisodes := make(map[string]struct{})
	if err := u.db.WalkEpisodes(ctx, feedConfig.ID, func(episode *model.Episode) error {
		// Track blocked episodes so we don't overwrite them
		if episode.Status == model.EpisodeBlocked {
			blockedEpisodes[episode.ID] = struct{}{}
		} else if episode.Status != model.EpisodeDownloaded && episode.Status != model.EpisodeCleaned {
			episodeSet[episode.ID] = struct{}{}
		}
		return nil
	}); err != nil {
		return err
	}

	// Filter out blocked episodes from the API results before adding to database
	filteredEpisodes := make([]*model.Episode, 0, len(result.Episodes))
	for _, episode := range result.Episodes {
		if _, isBlocked := blockedEpisodes[episode.ID]; !isBlocked {
			filteredEpisodes = append(filteredEpisodes, episode)
		} else {
			log.Debugf("skipping blocked episode %q", episode.ID)
		}
	}
	result.Episodes = filteredEpisodes

	if err := u.db.AddFeed(ctx, feedConfig.ID, result); err != nil {
		return err
	}

	for _, episode := range result.Episodes {
		delete(episodeSet, episode.ID)
	}

	// removing episodes that are no longer available in the feed and not downloaded or cleaned
	for id := range episodeSet {
		log.Infof("removing episode %q", id)
		err := u.db.DeleteEpisode(feedConfig.ID, id)
		if err != nil {
			return err
		}
	}

	log.Debug("successfully saved updates to storage")
	return nil
}

func (u *Manager) fetchEpisodes(ctx context.Context, feedConfig *feed.Config) ([]*model.Episode, error) {
	var (
		feedID       = feedConfig.ID
		downloadList []*model.Episode
		pageSize     = feedConfig.PageSize
	)

	log.WithField("page_size", pageSize).Info("fetching episodes for download")

	// Build the list of files to download
	err := u.db.WalkEpisodes(ctx, feedID, func(episode *model.Episode) error {
		var (
			logger = log.WithFields(log.Fields{"episode_id": episode.ID})
		)
		if episode.Status == model.EpisodeBlocked {
			// Episode is blocked
			logger.Debug("skipping blocked episode")
			return nil
		}
		if episode.Status != model.EpisodeNew && episode.Status != model.EpisodeError {
			// File already downloaded or cleaned
			logger.Infof("skipping due to already downloaded")
			return nil
		}

		if !matchFilters(episode, &feedConfig.Filters) {
			// Mark episode as ignored in database if it doesn't match filters
			if episode.Status == model.EpisodeNew {
				if err := u.db.UpdateEpisode(feedID, episode.ID, func(ep *model.Episode) error {
					ep.Status = model.EpisodeIgnored
					return nil
				}); err != nil {
					logger.WithError(err).Warn("failed to mark episode as ignored")
				}
			}
			return nil
		}

		// Limit the number of episodes downloaded at once
		pageSize--
		if pageSize < 0 {
			return nil
		}

		log.Debugf("adding %s (%q) to queue", episode.ID, episode.Title)
		downloadList = append(downloadList, episode)
		return nil
	})

	if err != nil {
		return nil, errors.Wrapf(err, "failed to build update list")
	}

	return downloadList, nil
}

// downloadEpisodesWithStats wraps downloadEpisodes and returns statistics
func (u *Manager) downloadEpisodesWithStats(ctx context.Context, feedConfig *feed.Config, downloadList []*model.Episode) (downloaded, failed int, bytesDownloaded int64) {
	// Track stats before download
	initialStats := u.collectEpisodeStats(ctx, feedConfig.ID, downloadList)

	// Perform the download
	_ = u.downloadEpisodes(ctx, feedConfig, downloadList)

	// Collect stats after download
	finalStats := u.collectEpisodeStats(ctx, feedConfig.ID, downloadList)

	return finalStats.downloaded - initialStats.downloaded,
		finalStats.failed - initialStats.failed,
		finalStats.bytesDownloaded - initialStats.bytesDownloaded
}

type episodeStats struct {
	downloaded      int
	failed          int
	bytesDownloaded int64
}

func (u *Manager) collectEpisodeStats(ctx context.Context, feedID string, episodes []*model.Episode) episodeStats {
	stats := episodeStats{}
	for _, ep := range episodes {
		current, err := u.db.GetEpisode(ctx, feedID, ep.ID)
		if err != nil {
			continue
		}
		switch current.Status {
		case model.EpisodeDownloaded:
			stats.downloaded++
			stats.bytesDownloaded += current.Size
		case model.EpisodeError:
			stats.failed++
		}
	}
	return stats
}

func (u *Manager) downloadEpisodes(ctx context.Context, feedConfig *feed.Config, downloadList []*model.Episode) error {
	var (
		downloadCount = len(downloadList)
		downloaded    = 0
		feedID        = feedConfig.ID
	)

	if downloadCount > 0 {
		log.Infof("download count: %d", downloadCount)
	} else {
		log.Info("no episodes to download")
		return nil
	}

	// Initialize progress tracking for this feed
	u.progressTracker.InitFeedProgress(feedID, downloadCount)
	defer u.progressTracker.ClearFeed(feedID)

	// Mark all episodes as queued and update their status in the database
	for _, episode := range downloadList {
		if err := u.db.UpdateEpisode(feedID, episode.ID, func(ep *model.Episode) error {
			ep.Status = model.EpisodeQueued
			return nil
		}); err != nil {
			log.WithError(err).Warnf("failed to update episode %s status to queued", episode.ID)
		}
	}
	u.progressTracker.QueueEpisodes(feedID, downloadCount)

	// Download pending episodes

	for idx, episode := range downloadList {
		var (
			logger      = log.WithFields(log.Fields{"index": idx, "episode_id": episode.ID})
			episodeName = feed.EpisodeName(feedConfig, episode)
		)

		// Check whether episode already exists
		size, err := u.fs.Size(ctx, fmt.Sprintf("%s/%s", feedID, episodeName))
		if err == nil {
			logger.Infof("episode %q already exists on disk", episode.ID)

			// File already exists, update file status and disk size
			if err := u.db.UpdateEpisode(feedID, episode.ID, func(episode *model.Episode) error {
				episode.Size = size
				episode.Status = model.EpisodeDownloaded
				return nil
			}); err != nil {
				logger.WithError(err).Error("failed to update file info")
				return err
			}

			continue
		} else if os.IsNotExist(err) {
			// Will download, do nothing here
		} else {
			logger.WithError(err).Error("failed to stat file")
			return err
		}

		// Download episode to disk
		// We download the episode to a temp directory first to avoid downloading this file by clients
		// while still being processed by youtube-dl (e.g. a file is being downloaded from YT or encoding in progress)

		// Update episode status to downloading and start progress tracking
		if err := u.db.UpdateEpisode(feedID, episode.ID, func(ep *model.Episode) error {
			ep.Status = model.EpisodeDownloading
			return nil
		}); err != nil {
			logger.WithError(err).Warn("failed to update episode status to downloading")
		}
		u.progressTracker.StartEpisode(feedID, episode.ID, episode.Title)

		// Set up progress callback for ytdl if it supports it
		if ytdlDownloader, ok := u.downloader.(*ytdl.YoutubeDl); ok {
			ytdlDownloader.SetProgressCallback(func(stage string, percent float64, downloaded, total int64, speed string) {
				u.progressTracker.UpdateEpisode(feedID, episode.ID, stage, percent, downloaded, total, speed)
			})
		}

		logger.Infof("! downloading episode %s", episode.VideoURL)
		tempFile, err := u.downloader.Download(ctx, feedConfig, episode)
		if err != nil {
			// YouTube might block host with HTTP Error 429: Too Many Requests
			// We still need to generate XML, so just stop sending download requests and
			// retry next time
			if err == ytdl.ErrTooManyRequests {
				logger.Warn("server responded with a 'Too Many Requests' error")
				break
			}

			logger.WithError(err).Error("failed to download episode")
			if err := u.db.UpdateEpisode(feedID, episode.ID, func(episode *model.Episode) error {
				episode.Status = model.EpisodeError
				episode.Error = err.Error()
				return nil
			}); err != nil {
				return err
			}

			continue
		}

		logger.Debug("copying file")
		fileSize, err := u.fs.Create(ctx, fmt.Sprintf("%s/%s", feedID, episodeName), tempFile)
		tempFile.Close()
		if err != nil {
			logger.WithError(err).Error("failed to copy file")
			return err
		}

		// Execute post episode download hooks
		if len(feedConfig.PostEpisodeDownload) > 0 {
			env := []string{
				"EPISODE_FILE=" + fmt.Sprintf("%s/%s", feedID, episodeName),
				"FEED_NAME=" + feedID,
				"EPISODE_TITLE=" + episode.Title,
			}

			for i, hook := range feedConfig.PostEpisodeDownload {
				if err := hook.Invoke(env); err != nil {
					logger.Errorf("failed to execute post episode download hook %d: %v", i+1, err)
				} else {
					logger.Infof("post episode download hook %d executed successfully", i+1)
				}
			}
		}

		// Update file status in database

		logger.Infof("successfully downloaded file %q", episode.ID)
		if err := u.db.UpdateEpisode(feedID, episode.ID, func(episode *model.Episode) error {
			episode.Size = fileSize
			episode.Status = model.EpisodeDownloaded
			return nil
		}); err != nil {
			return err
		}

		// Mark episode as complete in progress tracker
		u.progressTracker.CompleteEpisode(feedID, episode.ID)

		downloaded++
	}

	log.Infof("downloaded %d episode(s)", downloaded)
	return nil
}

// DeleteEpisode deletes both the database entry and media file for an episode
func (u *Manager) DeleteEpisode(ctx context.Context, feedID, episodeID string) error {
	feedConfig, ok := u.feeds[feedID]
	if !ok {
		return errors.Errorf("feed %q not found", feedID)
	}

	logger := log.WithFields(log.Fields{"feed_id": feedID, "episode_id": episodeID})

	// Get the episode from the database
	episode, err := u.db.GetEpisode(ctx, feedID, episodeID)
	if err != nil {
		_ = u.historyManager.LogEpisodeDelete(ctx, feedID, getFeedTitle(ctx, u.db, feedID), episodeID, "", false, err.Error())
		return errors.Wrapf(err, "failed to get episode %s/%s", feedID, episodeID)
	}

	episodeTitle := episode.Title

	// Delete the media file if it exists
	episodeName := feed.EpisodeName(feedConfig, episode)
	path := fmt.Sprintf("%s/%s", feedID, episodeName)
	if err := u.fs.Delete(ctx, path); err != nil {
		if !os.IsNotExist(err) {
			logger.WithError(err).Warn("failed to delete media file")
		} else {
			logger.Debug("media file does not exist, skipping deletion")
		}
	} else {
		logger.Info("deleted media file")
	}

	// Delete the database entry
	if err := u.db.DeleteEpisode(feedID, episodeID); err != nil {
		_ = u.historyManager.LogEpisodeDelete(ctx, feedID, getFeedTitle(ctx, u.db, feedID), episodeID, episodeTitle, false, err.Error())
		return errors.Wrapf(err, "failed to delete episode from database %s/%s", feedID, episodeID)
	}

	logger.Info("successfully deleted episode")
	_ = u.historyManager.LogEpisodeDelete(ctx, feedID, getFeedTitle(ctx, u.db, feedID), episodeID, episodeTitle, true, "")
	return nil
}

// BlockEpisode marks an episode as blocked, preventing it from being re-downloaded
func (u *Manager) BlockEpisode(ctx context.Context, feedID, episodeID string) error {
	feedConfig, ok := u.feeds[feedID]
	if !ok {
		return errors.Errorf("feed %q not found", feedID)
	}

	logger := log.WithFields(log.Fields{"feed_id": feedID, "episode_id": episodeID})

	episodeTitle := ""

	// Get the episode from the database (or create if it doesn't exist)
	episode, err := u.db.GetEpisode(ctx, feedID, episodeID)
	if err != nil {
		if err == model.ErrNotFound {
			// Episode doesn't exist, create a blocked entry
			logger.Info("episode not in database, creating blocked entry")
			episode = &model.Episode{
				ID:     episodeID,
				Status: model.EpisodeBlocked,
			}
			// Add to feed
			tempFeed := &model.Feed{
				ID:       feedID,
				Episodes: []*model.Episode{episode},
			}
			if err := u.db.AddFeed(ctx, feedID, tempFeed); err != nil {
				_ = u.historyManager.LogEpisodeBlock(ctx, feedID, getFeedTitle(ctx, u.db, feedID), episodeID, episodeTitle, false, err.Error())
				return errors.Wrapf(err, "failed to create blocked episode %s/%s", feedID, episodeID)
			}
		} else {
			_ = u.historyManager.LogEpisodeBlock(ctx, feedID, getFeedTitle(ctx, u.db, feedID), episodeID, episodeTitle, false, err.Error())
			return errors.Wrapf(err, "failed to get episode %s/%s", feedID, episodeID)
		}
	} else {
		episodeTitle = episode.Title
		// Episode exists, update status to blocked
		if err := u.db.UpdateEpisode(feedID, episodeID, func(ep *model.Episode) error {
			ep.Status = model.EpisodeBlocked
			return nil
		}); err != nil {
			_ = u.historyManager.LogEpisodeBlock(ctx, feedID, getFeedTitle(ctx, u.db, feedID), episodeID, episodeTitle, false, err.Error())
			return errors.Wrapf(err, "failed to block episode %s/%s", feedID, episodeID)
		}
	}

	// Delete the media file if it exists
	episodeName := feed.EpisodeName(feedConfig, episode)
	path := fmt.Sprintf("%s/%s", feedID, episodeName)
	if err := u.fs.Delete(ctx, path); err != nil {
		if !os.IsNotExist(err) {
			logger.WithError(err).Warn("failed to delete media file")
		} else {
			logger.Debug("media file does not exist")
		}
	} else {
		logger.Info("deleted media file")
	}

	logger.Info("successfully blocked episode")
	_ = u.historyManager.LogEpisodeBlock(ctx, feedID, getFeedTitle(ctx, u.db, feedID), episodeID, episodeTitle, true, "")
	return nil
}

// RetryEpisode retries downloading a single episode
func (u *Manager) RetryEpisode(ctx context.Context, feedID, episodeID string) error {
	feedConfig, ok := u.feeds[feedID]
	if !ok {
		return errors.Errorf("feed %q not found", feedID)
	}

	// Get the episode from the database
	episode, err := u.db.GetEpisode(ctx, feedID, episodeID)
	if err != nil {
		_ = u.historyManager.LogEpisodeRetry(ctx, feedID, getFeedTitle(ctx, u.db, feedID), episodeID, "", false, err.Error())
		return errors.Wrapf(err, "failed to get episode %s/%s", feedID, episodeID)
	}

	episodeTitle := episode.Title
	logger := log.WithFields(log.Fields{"feed_id": feedID, "episode_id": episodeID})
	episodeName := feed.EpisodeName(feedConfig, episode)

	// Reset episode status to new and clear any error message
	if err := u.db.UpdateEpisode(feedID, episodeID, func(ep *model.Episode) error {
		ep.Status = model.EpisodeNew
		ep.Error = ""
		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to reset episode status")
	}

	// Check whether episode already exists on disk
	size, err := u.fs.Size(ctx, fmt.Sprintf("%s/%s", feedID, episodeName))
	if err == nil {
		logger.Infof("episode %q already exists on disk", episodeID)

		// File already exists, update file status and disk size
		if err := u.db.UpdateEpisode(feedID, episodeID, func(ep *model.Episode) error {
			ep.Size = size
			ep.Status = model.EpisodeDownloaded
			return nil
		}); err != nil {
			logger.WithError(err).Error("failed to update file info")
			return err
		}

		return nil
	} else if !os.IsNotExist(err) {
		logger.WithError(err).Error("failed to stat file")
		return err
	}

	// Download episode to disk
	logger.Infof("downloading episode %s", episode.VideoURL)
	tempFile, err := u.downloader.Download(ctx, feedConfig, episode)
	if err != nil {
		// Update episode status to error with the error message
		updateErr := u.db.UpdateEpisode(feedID, episodeID, func(ep *model.Episode) error {
			ep.Status = model.EpisodeError
			ep.Error = err.Error()
			return nil
		})
		if updateErr != nil {
			logger.WithError(updateErr).Error("failed to update episode error status")
		}
		_ = u.historyManager.LogEpisodeRetry(ctx, feedID, getFeedTitle(ctx, u.db, feedID), episodeID, episodeTitle, false, err.Error())
		return errors.Wrap(err, "download failed")
	}

	logger.Debug("copying file")
	fileSize, err := u.fs.Create(ctx, fmt.Sprintf("%s/%s", feedID, episodeName), tempFile)
	tempFile.Close()
	if err != nil {
		logger.WithError(err).Error("failed to copy file")
		// Update episode status to error
		updateErr := u.db.UpdateEpisode(feedID, episodeID, func(ep *model.Episode) error {
			ep.Status = model.EpisodeError
			ep.Error = fmt.Sprintf("failed to copy file: %v", err)
			return nil
		})
		if updateErr != nil {
			logger.WithError(updateErr).Error("failed to update episode error status")
		}
		return err
	}

	// Execute post episode download hooks
	if len(feedConfig.PostEpisodeDownload) > 0 {
		env := []string{
			"EPISODE_FILE=" + fmt.Sprintf("%s/%s", feedID, episodeName),
			"FEED_NAME=" + feedID,
			"EPISODE_TITLE=" + episode.Title,
		}

		for i, hook := range feedConfig.PostEpisodeDownload {
			if err := hook.Invoke(env); err != nil {
				logger.Errorf("failed to execute post episode download hook %d: %v", i+1, err)
			} else {
				logger.Infof("post episode download hook %d executed successfully", i+1)
			}
		}
	}

	// Update file status in database
	logger.Infof("successfully downloaded file %q", episodeID)
	if err := u.db.UpdateEpisode(feedID, episodeID, func(ep *model.Episode) error {
		ep.Size = fileSize
		ep.Status = model.EpisodeDownloaded
		ep.Error = ""
		return nil
	}); err != nil {
		return err
	}

	// Rebuild XML feed to include the newly downloaded episode
	if err := u.buildXML(ctx, feedConfig); err != nil {
		logger.WithError(err).Warn("failed to rebuild XML feed after episode download")
	}

	_ = u.historyManager.LogEpisodeRetry(ctx, feedID, getFeedTitle(ctx, u.db, feedID), episodeID, episodeTitle, true, "")
	return nil
}

func (u *Manager) buildXML(ctx context.Context, feedConfig *feed.Config) error {
	f, err := u.db.GetFeed(ctx, feedConfig.ID)
	if err != nil {
		return err
	}

	// Build iTunes XML feed with data received from builder
	log.Debug("building iTunes podcast feed")
	podcast, err := feed.Build(ctx, f, feedConfig, u.hostname)
	if err != nil {
		return err
	}

	var (
		reader  = bytes.NewReader([]byte(podcast.String()))
		xmlName = fmt.Sprintf("%s.xml", feedConfig.ID)
	)

	if _, err := u.fs.Create(ctx, xmlName, reader); err != nil {
		return errors.Wrap(err, "failed to upload new XML feed")
	}

	return nil
}

func (u *Manager) buildOPML(ctx context.Context) error {
	// Build OPML with data received from builder
	log.Debug("building podcast OPML")
	opml, err := feed.BuildOPML(ctx, u.feeds, u.db, u.hostname)
	if err != nil {
		return err
	}

	var (
		reader  = bytes.NewReader([]byte(opml))
		xmlName = fmt.Sprintf("%s.opml", "podsync")
	)

	if _, err := u.fs.Create(ctx, xmlName, reader); err != nil {
		return errors.Wrap(err, "failed to upload OPML")
	}

	return nil
}

func (u *Manager) cleanup(ctx context.Context, feedConfig *feed.Config) error {
	var (
		feedID = feedConfig.ID
		logger = log.WithField("feed_id", feedID)
		list   []*model.Episode
		result *multierror.Error
	)

	if feedConfig.Clean == nil {
		logger.Debug("no cleanup policy configured")
		return nil
	}

	count := feedConfig.Clean.KeepLast
	if count < 1 {
		logger.Info("nothing to clean")
		return nil
	}

	logger.WithField("count", count).Info("running cleaner")
	if err := u.db.WalkEpisodes(ctx, feedConfig.ID, func(episode *model.Episode) error {
		if episode.Status == model.EpisodeDownloaded {
			list = append(list, episode)
		}
		return nil
	}); err != nil {
		return err
	}

	if count > len(list) {
		return nil
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].PubDate.After(list[j].PubDate)
	})

	for _, episode := range list[count:] {
		logger.WithField("episode_id", episode.ID).Infof("deleting %q", episode.Title)

		var (
			episodeName = feed.EpisodeName(feedConfig, episode)
			path        = fmt.Sprintf("%s/%s", feedConfig.ID, episodeName)
		)

		err := u.fs.Delete(ctx, path)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				logger.WithError(err).Errorf("failed to delete episode file: %s", episode.ID)
				result = multierror.Append(result, errors.Wrapf(err, "failed to delete episode: %s", episode.ID))
				continue
			}

			logger.WithField("episode_id", episode.ID).Info("episode was not found - file does not exist")
		}

		if err := u.db.UpdateEpisode(feedID, episode.ID, func(episode *model.Episode) error {
			episode.Status = model.EpisodeCleaned
			episode.Title = ""
			episode.Description = ""
			return nil
		}); err != nil {
			result = multierror.Append(result, errors.Wrapf(err, "failed to set state for cleaned episode: %s", episode.ID))
			continue
		}
	}

	return result.ErrorOrNil()
}

// getFeedTitle is a helper function to get feed title from database
func getFeedTitle(ctx context.Context, storage db.Storage, feedID string) string {
	feed, err := storage.GetFeed(ctx, feedID)
	if err != nil || feed == nil {
		return feedID
	}
	if feed.Title != "" {
		return feed.Title
	}
	return feedID
}
