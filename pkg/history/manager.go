package history

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/mxpv/podsync/pkg/db"
	"github.com/mxpv/podsync/pkg/model"
)

// Manager handles creation and updates of history entries
type Manager struct {
	storage db.Storage
	enabled bool
}

// NewManager creates a new history manager
func NewManager(storage db.Storage, enabled bool) *Manager {
	return &Manager{
		storage: storage,
		enabled: enabled,
	}
}

// LogFeedUpdateStart creates a new history entry for a feed update
// Returns the entry ID for later updates
func (m *Manager) LogFeedUpdateStart(ctx context.Context, feedID, feedTitle string, triggerType model.TriggerType) (string, error) {
	if !m.enabled {
		return "", nil
	}

	// Generate a unique ID with timestamp prefix for chronological sorting
	timestamp := time.Now().Unix()
	entryID := fmt.Sprintf("%d-%s", timestamp, uuid.New().String())

	entry := &model.HistoryEntry{
		ID:          entryID,
		JobType:     model.JobTypeFeedUpdate,
		FeedID:      feedID,
		FeedTitle:   feedTitle,
		StartTime:   time.Now(),
		Status:      model.JobStatusRunning,
		TriggerType: triggerType,
		Statistics:  model.JobStatistics{},
	}

	if err := m.storage.AddHistory(ctx, entry); err != nil {
		log.WithError(err).Warnf("failed to create history entry for feed %s", feedID)
		return "", err
	}

	log.Debugf("created history entry %s for feed %s update", entryID, feedID)
	return entryID, nil
}

// LogFeedUpdateEnd updates a feed update history entry with final status and statistics
func (m *Manager) LogFeedUpdateEnd(ctx context.Context, entryID string, status model.JobStatus, stats model.JobStatistics, errMsg string) error {
	if !m.enabled || entryID == "" {
		return nil
	}

	err := m.storage.UpdateHistory(ctx, entryID, func(entry *model.HistoryEntry) error {
		now := time.Now()
		entry.EndTime = &now
		entry.Duration = now.Sub(entry.StartTime)
		entry.Status = status
		entry.Statistics = stats
		entry.Error = errMsg
		return nil
	})

	if err != nil {
		log.WithError(err).Warnf("failed to update history entry %s", entryID)
		return err
	}

	log.Debugf("updated history entry %s with status %s", entryID, status)
	return nil
}

// LogFeedUpdateEndWithEpisodes updates a feed update history entry with final status, statistics, and episode details
// The episodeIDs parameter should contain the IDs of episodes that were processed during this job
func (m *Manager) LogFeedUpdateEndWithEpisodes(ctx context.Context, entryID, feedID string, episodeIDs []string, status model.JobStatus, stats model.JobStatistics, errMsg string) error {
	if !m.enabled || entryID == "" {
		return nil
	}

	// Collect episode details only for the episodes that were processed during this job
	episodeDetails := []model.EpisodeDetail{}
	for _, episodeID := range episodeIDs {
		episode, err := m.storage.GetEpisode(ctx, feedID, episodeID)
		if err != nil {
			log.WithError(err).Warnf("failed to get episode %s for history entry %s", episodeID, entryID)
			continue
		}

		detail := model.EpisodeDetail{
			ID:       episode.ID,
			Title:    episode.Title,
			Status:   string(episode.Status),
			Error:    episode.Error,
			Size:     episode.Size,
			Duration: episode.Duration,
		}
		episodeDetails = append(episodeDetails, detail)
	}

	// Update statistics with episode details
	stats.EpisodeDetails = episodeDetails

	err := m.storage.UpdateHistory(ctx, entryID, func(entry *model.HistoryEntry) error {
		now := time.Now()
		entry.EndTime = &now
		entry.Duration = now.Sub(entry.StartTime)
		entry.Status = status
		entry.Statistics = stats
		entry.Error = errMsg
		return nil
	})

	if err != nil {
		log.WithError(err).Warnf("failed to update history entry %s", entryID)
		return err
	}

	log.Debugf("updated history entry %s with status %s and %d episode details", entryID, status, len(episodeDetails))
	return nil
}

// LogEpisodeRetry logs an episode retry operation
func (m *Manager) LogEpisodeRetry(ctx context.Context, feedID, feedTitle, episodeID, episodeTitle string, success bool, errMsg string) error {
	if !m.enabled {
		return nil
	}

	timestamp := time.Now().Unix()
	entryID := fmt.Sprintf("%d-%s", timestamp, uuid.New().String())

	status := model.JobStatusSuccess
	if !success {
		status = model.JobStatusFailed
	}

	now := time.Now()
	entry := &model.HistoryEntry{
		ID:           entryID,
		JobType:      model.JobTypeEpisodeRetry,
		FeedID:       feedID,
		FeedTitle:    feedTitle,
		EpisodeID:    episodeID,
		EpisodeTitle: episodeTitle,
		StartTime:    now,
		EndTime:      &now,
		Duration:     0,
		Status:       status,
		TriggerType:  model.TriggerManual,
		Statistics:   model.JobStatistics{},
		Error:        errMsg,
	}

	if err := m.storage.AddHistory(ctx, entry); err != nil {
		log.WithError(err).Warnf("failed to create history entry for episode retry %s/%s", feedID, episodeID)
		return err
	}

	log.Debugf("logged episode retry %s for feed %s", episodeID, feedID)
	return nil
}

// LogEpisodeDelete logs an episode deletion operation
func (m *Manager) LogEpisodeDelete(ctx context.Context, feedID, feedTitle, episodeID, episodeTitle string, success bool, errMsg string) error {
	if !m.enabled {
		return nil
	}

	timestamp := time.Now().Unix()
	entryID := fmt.Sprintf("%d-%s", timestamp, uuid.New().String())

	status := model.JobStatusSuccess
	if !success {
		status = model.JobStatusFailed
	}

	now := time.Now()
	entry := &model.HistoryEntry{
		ID:           entryID,
		JobType:      model.JobTypeEpisodeDelete,
		FeedID:       feedID,
		FeedTitle:    feedTitle,
		EpisodeID:    episodeID,
		EpisodeTitle: episodeTitle,
		StartTime:    now,
		EndTime:      &now,
		Duration:     0,
		Status:       status,
		TriggerType:  model.TriggerManual,
		Statistics:   model.JobStatistics{},
		Error:        errMsg,
	}

	if err := m.storage.AddHistory(ctx, entry); err != nil {
		log.WithError(err).Warnf("failed to create history entry for episode delete %s/%s", feedID, episodeID)
		return err
	}

	log.Debugf("logged episode delete %s for feed %s", episodeID, feedID)
	return nil
}

// LogEpisodeBlock logs an episode block operation
func (m *Manager) LogEpisodeBlock(ctx context.Context, feedID, feedTitle, episodeID, episodeTitle string, success bool, errMsg string) error {
	if !m.enabled {
		return nil
	}

	timestamp := time.Now().Unix()
	entryID := fmt.Sprintf("%d-%s", timestamp, uuid.New().String())

	status := model.JobStatusSuccess
	if !success {
		status = model.JobStatusFailed
	}

	now := time.Now()
	entry := &model.HistoryEntry{
		ID:           entryID,
		JobType:      model.JobTypeEpisodeBlock,
		FeedID:       feedID,
		FeedTitle:    feedTitle,
		EpisodeID:    episodeID,
		EpisodeTitle: episodeTitle,
		StartTime:    now,
		EndTime:      &now,
		Duration:     0,
		Status:       status,
		TriggerType:  model.TriggerManual,
		Statistics:   model.JobStatistics{},
		Error:        errMsg,
	}

	if err := m.storage.AddHistory(ctx, entry); err != nil {
		log.WithError(err).Warnf("failed to create history entry for episode block %s/%s", feedID, episodeID)
		return err
	}

	log.Debugf("logged episode block %s for feed %s", episodeID, feedID)
	return nil
}

// CleanupOldEntries removes history entries based on retention policy
func (m *Manager) CleanupOldEntries(ctx context.Context, retentionDays, maxEntries int) error {
	if !m.enabled {
		return nil
	}

	log.Infof("cleaning up history entries older than %d days or exceeding %d entries", retentionDays, maxEntries)

	if err := m.storage.CleanupHistory(ctx, retentionDays, maxEntries); err != nil {
		log.WithError(err).Error("failed to cleanup history")
		return err
	}

	log.Debug("history cleanup completed successfully")
	return nil
}
