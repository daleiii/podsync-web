package db

import (
	"context"

	"github.com/daleiii/podsync-web/pkg/model"
)

type Version int

const (
	CurrentVersion = 1
)

type Storage interface {
	Close() error
	Version() (int, error)

	// AddFeed will:
	// - Insert or update feed info
	// - Append new episodes to the existing list of episodes (existing episodes are not overwritten!)
	AddFeed(ctx context.Context, feedID string, feed *model.Feed) error

	// GetFeed gets a feed by ID
	GetFeed(ctx context.Context, feedID string) (*model.Feed, error)

	// WalkFeeds iterates over feeds saved to database
	WalkFeeds(ctx context.Context, cb func(feed *model.Feed) error) error

	// DeleteFeed deletes feed and all related data from database
	DeleteFeed(ctx context.Context, feedID string) error

	// GetEpisode gets episode by identifier
	GetEpisode(ctx context.Context, feedID string, episodeID string) (*model.Episode, error)

	// UpdateEpisode updates episode fields
	UpdateEpisode(feedID string, episodeID string, cb func(episode *model.Episode) error) error

	// DeleteEpisode deletes an episode
	DeleteEpisode(feedID string, episodeID string) error

	// WalkEpisodes iterates over episodes that belong to the given feed ID
	WalkEpisodes(ctx context.Context, feedID string, cb func(episode *model.Episode) error) error

	// AddHistory adds a new history entry
	AddHistory(ctx context.Context, entry *model.HistoryEntry) error

	// GetHistory gets a history entry by ID
	GetHistory(ctx context.Context, id string) (*model.HistoryEntry, error)

	// ListHistory returns a paginated list of history entries with filters
	ListHistory(ctx context.Context, filters model.HistoryFilters, page, pageSize int) ([]*model.HistoryEntry, int, error)

	// UpdateHistory updates a history entry
	UpdateHistory(ctx context.Context, id string, cb func(entry *model.HistoryEntry) error) error

	// DeleteHistory deletes a history entry by ID
	DeleteHistory(ctx context.Context, id string) error

	// CleanupHistory removes old history entries based on retention policy
	CleanupHistory(ctx context.Context, retentionDays int, maxEntries int) error

	// GetHistoryStats returns statistics about the history
	GetHistoryStats(ctx context.Context) (count int, oldestEntry *model.HistoryEntry, err error)
}
