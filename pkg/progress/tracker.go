package progress

import (
	"sync"
	"time"
)

// EpisodeProgress represents the download progress for a single episode
type EpisodeProgress struct {
	FeedID         string    `json:"feed_id"`
	EpisodeID      string    `json:"episode_id"`
	EpisodeTitle   string    `json:"episode_title"`
	Stage          string    `json:"stage"`       // "downloading", "encoding", "saving"
	Percent        float64   `json:"percent"`     // 0-100
	Downloaded     int64     `json:"downloaded"`  // bytes downloaded
	Total          int64     `json:"total"`       // total size in bytes (estimate)
	Speed          string    `json:"speed"`       // e.g. "1.2MiB/s"
	StartTime      time.Time `json:"start_time"`  // when download started
	LastUpdateTime time.Time `json:"last_update"` // last progress update
}

// FeedProgress represents the overall progress for a feed update
type FeedProgress struct {
	FeedID           string    `json:"feed_id"`
	TotalEpisodes    int       `json:"total_episodes"`
	CompletedCount   int       `json:"completed_count"`
	DownloadingCount int       `json:"downloading_count"`
	QueuedCount      int       `json:"queued_count"`
	OverallPercent   float64   `json:"overall_percent"` // 0-100
	StartTime        time.Time `json:"start_time"`
}

// Tracker manages download progress tracking in memory
type Tracker struct {
	mu              sync.RWMutex
	feedProgress    map[string]*FeedProgress    // feedID -> progress
	episodeProgress map[string]*EpisodeProgress // "feedID/episodeID" -> progress
}

// New creates a new progress tracker
func New() *Tracker {
	return &Tracker{
		feedProgress:    make(map[string]*FeedProgress),
		episodeProgress: make(map[string]*EpisodeProgress),
	}
}

// InitFeedProgress initializes progress tracking for a feed update
func (t *Tracker) InitFeedProgress(feedID string, totalEpisodes int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.feedProgress[feedID] = &FeedProgress{
		FeedID:        feedID,
		TotalEpisodes: totalEpisodes,
		StartTime:     time.Now(),
	}
}

// StartEpisode marks an episode as starting download
func (t *Tracker) StartEpisode(feedID, episodeID, episodeTitle string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := feedID + "/" + episodeID
	t.episodeProgress[key] = &EpisodeProgress{
		FeedID:         feedID,
		EpisodeID:      episodeID,
		EpisodeTitle:   episodeTitle,
		Stage:          "downloading",
		StartTime:      time.Now(),
		LastUpdateTime: time.Now(),
	}

	// Update feed progress
	if fp, ok := t.feedProgress[feedID]; ok {
		fp.DownloadingCount++
		if fp.QueuedCount > 0 {
			fp.QueuedCount--
		}
		t.updateFeedPercent(fp)
	}
}

// UpdateEpisode updates progress for an episode
func (t *Tracker) UpdateEpisode(feedID, episodeID string, stage string, percent float64, downloaded, total int64, speed string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := feedID + "/" + episodeID
	ep, ok := t.episodeProgress[key]
	if !ok {
		// Episode wasn't started properly, create it
		ep = &EpisodeProgress{
			FeedID:    feedID,
			EpisodeID: episodeID,
			StartTime: time.Now(),
		}
		t.episodeProgress[key] = ep
	}

	ep.Stage = stage
	ep.Percent = percent
	ep.Downloaded = downloaded
	ep.Total = total
	ep.Speed = speed
	ep.LastUpdateTime = time.Now()
}

// CompleteEpisode marks an episode as completed
func (t *Tracker) CompleteEpisode(feedID, episodeID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := feedID + "/" + episodeID
	delete(t.episodeProgress, key)

	// Update feed progress
	if fp, ok := t.feedProgress[feedID]; ok {
		if fp.DownloadingCount > 0 {
			fp.DownloadingCount--
		}
		fp.CompletedCount++
		t.updateFeedPercent(fp)
	}
}

// QueueEpisodes marks episodes as queued
func (t *Tracker) QueueEpisodes(feedID string, count int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if fp, ok := t.feedProgress[feedID]; ok {
		fp.QueuedCount += count
		t.updateFeedPercent(fp)
	}
}

// ClearFeed removes all progress tracking for a feed
func (t *Tracker) ClearFeed(feedID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.feedProgress, feedID)

	// Remove all episodes for this feed
	for key := range t.episodeProgress {
		if ep := t.episodeProgress[key]; ep.FeedID == feedID {
			delete(t.episodeProgress, key)
		}
	}
}

// GetFeedProgress returns progress for a specific feed
func (t *Tracker) GetFeedProgress(feedID string) (*FeedProgress, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	fp, ok := t.feedProgress[feedID]
	if !ok {
		return nil, false
	}

	// Return a copy
	fpCopy := *fp
	return &fpCopy, true
}

// GetAllFeedProgress returns progress for all feeds
func (t *Tracker) GetAllFeedProgress() map[string]*FeedProgress {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make(map[string]*FeedProgress)
	for feedID, fp := range t.feedProgress {
		fpCopy := *fp
		result[feedID] = &fpCopy
	}
	return result
}

// GetEpisodeProgress returns progress for a specific episode
func (t *Tracker) GetEpisodeProgress(feedID, episodeID string) (*EpisodeProgress, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	key := feedID + "/" + episodeID
	ep, ok := t.episodeProgress[key]
	if !ok {
		return nil, false
	}

	// Return a copy
	epCopy := *ep
	return &epCopy, true
}

// GetAllEpisodeProgress returns progress for all episodes
func (t *Tracker) GetAllEpisodeProgress() []*EpisodeProgress {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]*EpisodeProgress, 0, len(t.episodeProgress))
	for _, ep := range t.episodeProgress {
		epCopy := *ep
		result = append(result, &epCopy)
	}
	return result
}

// GetEpisodesForFeed returns all episode progress for a specific feed
func (t *Tracker) GetEpisodesForFeed(feedID string) []*EpisodeProgress {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]*EpisodeProgress, 0)
	for _, ep := range t.episodeProgress {
		if ep.FeedID == feedID {
			epCopy := *ep
			result = append(result, &epCopy)
		}
	}
	return result
}

// HasActiveDownloads returns true if there are any active downloads
func (t *Tracker) HasActiveDownloads() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return len(t.feedProgress) > 0 || len(t.episodeProgress) > 0
}

// updateFeedPercent recalculates the overall percent for a feed (must be called with lock held)
func (t *Tracker) updateFeedPercent(fp *FeedProgress) {
	if fp.TotalEpisodes == 0 {
		fp.OverallPercent = 0
		return
	}

	// Calculate based on completed + partial progress from downloading episodes
	completed := float64(fp.CompletedCount)

	// Add partial progress from currently downloading episodes
	for _, ep := range t.episodeProgress {
		if ep.FeedID == fp.FeedID {
			completed += ep.Percent / 100.0
		}
	}

	fp.OverallPercent = (completed / float64(fp.TotalEpisodes)) * 100
}
