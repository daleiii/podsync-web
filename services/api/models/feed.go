package models

import (
	"time"

	"github.com/mxpv/podsync/pkg/feed"
	"github.com/mxpv/podsync/pkg/model"
)

// FeedResponse represents a feed in API responses
type FeedResponse struct {
	ID            string     `json:"id"`
	URL           string     `json:"url"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	EpisodeCount  int        `json:"episode_count"`
	LastUpdate    time.Time  `json:"last_update"`
	Status        string     `json:"status"`
	Configuration FeedConfig `json:"configuration"`
	Author        string     `json:"author"`
	CoverArt      string     `json:"cover_art"`
	Provider      string     `json:"provider"`
	Format        string     `json:"format"`
	Quality       string     `json:"quality"`
}

// FeedConfig represents feed configuration in API
type FeedConfig struct {
	UpdatePeriod string        `json:"update_period"`
	CronSchedule string        `json:"cron_schedule"`
	Quality      string        `json:"quality"`
	Format       string        `json:"format"`
	PageSize     int           `json:"page_size"`
	MaxHeight    int           `json:"max_height"`
	CleanupKeep  int           `json:"cleanup_keep"`
	PlaylistSort string        `json:"playlist_sort"`
	PrivateFeed  bool          `json:"private_feed"`
	OPML         bool          `json:"opml"`
	CustomFormat *CustomFormat `json:"custom_format,omitempty"`
	Filters      Filters       `json:"filters"`
	Custom       Custom        `json:"custom"`
}

// CustomFormat represents custom format settings
type CustomFormat struct {
	YouTubeDLFormat string `json:"youtube_dl_format,omitempty"`
	Extension       string `json:"extension,omitempty"`
}

// Filters represents episode filtering options
type Filters struct {
	Title          string `json:"title,omitempty"`
	NotTitle       string `json:"not_title,omitempty"`
	Description    string `json:"description,omitempty"`
	NotDescription string `json:"not_description,omitempty"`
	MinDuration    int64  `json:"min_duration,omitempty"`
	MaxDuration    int64  `json:"max_duration,omitempty"`
	MaxAge         int    `json:"max_age,omitempty"`
	MinAge         int    `json:"min_age,omitempty"`
}

// Custom represents custom feed metadata
type Custom struct {
	CoverArt        string   `json:"cover_art,omitempty"`
	CoverArtQuality string   `json:"cover_art_quality,omitempty"`
	Category        string   `json:"category,omitempty"`
	Subcategories   []string `json:"subcategories,omitempty"`
	Explicit        bool     `json:"explicit"`
	Language        string   `json:"lang,omitempty"`
	Author          string   `json:"author,omitempty"`
	Title           string   `json:"title,omitempty"`
	Description     string   `json:"description,omitempty"`
	OwnerName       string   `json:"owner_name,omitempty"`
	OwnerEmail      string   `json:"owner_email,omitempty"`
	Link            string   `json:"link,omitempty"`
}

// CreateFeedRequest represents a request to create a new feed
type CreateFeedRequest struct {
	ID     string     `json:"id"`
	URL    string     `json:"url"`
	Config FeedConfig `json:"config"`
}

// UpdateFeedRequest represents a request to update a feed
type UpdateFeedRequest struct {
	Config FeedConfig `json:"config"`
}

// FromModelFeed converts model.Feed to FeedResponse
func FromModelFeed(f *model.Feed, cfg *feed.Config, episodeCount int) FeedResponse {
	cleanupKeep := 0
	if cfg.Clean != nil {
		cleanupKeep = cfg.Clean.KeepLast
	}

	var customFormat *CustomFormat
	if cfg.CustomFormat.YouTubeDLFormat != "" || cfg.CustomFormat.Extension != "" {
		customFormat = &CustomFormat{
			YouTubeDLFormat: cfg.CustomFormat.YouTubeDLFormat,
			Extension:       cfg.CustomFormat.Extension,
		}
	}

	return FeedResponse{
		ID:           f.ID,
		URL:          f.ItemURL,
		Title:        f.Title,
		Description:  f.Description,
		EpisodeCount: episodeCount,
		LastUpdate:   f.UpdatedAt,
		Status:       "active",
		Author:       f.Author,
		CoverArt:     f.CoverArt,
		Provider:     string(f.Provider),
		Format:       string(f.Format),
		Quality:      string(f.Quality),
		Configuration: FeedConfig{
			UpdatePeriod: cfg.UpdatePeriod.String(),
			CronSchedule: cfg.CronSchedule,
			Quality:      string(cfg.Quality),
			Format:       string(cfg.Format),
			PageSize:     cfg.PageSize,
			MaxHeight:    cfg.MaxHeight,
			CleanupKeep:  cleanupKeep,
			PlaylistSort: string(cfg.PlaylistSort),
			PrivateFeed:  cfg.PrivateFeed,
			OPML:         cfg.OPML,
			CustomFormat: customFormat,
			Filters: Filters{
				Title:          cfg.Filters.Title,
				NotTitle:       cfg.Filters.NotTitle,
				Description:    cfg.Filters.Description,
				NotDescription: cfg.Filters.NotDescription,
				MinDuration:    cfg.Filters.MinDuration,
				MaxDuration:    cfg.Filters.MaxDuration,
				MaxAge:         cfg.Filters.MaxAge,
				MinAge:         cfg.Filters.MinAge,
			},
			Custom: Custom{
				CoverArt:        cfg.Custom.CoverArt,
				CoverArtQuality: string(cfg.Custom.CoverArtQuality),
				Category:        cfg.Custom.Category,
				Subcategories:   cfg.Custom.Subcategories,
				Explicit:        cfg.Custom.Explicit,
				Language:        cfg.Custom.Language,
				Author:          cfg.Custom.Author,
				Title:           cfg.Custom.Title,
				Description:     cfg.Custom.Description,
				OwnerName:       cfg.Custom.OwnerName,
				OwnerEmail:      cfg.Custom.OwnerEmail,
				Link:            cfg.Custom.Link,
			},
		},
	}
}
