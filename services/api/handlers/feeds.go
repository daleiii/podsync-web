package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/db"
	"github.com/mxpv/podsync/pkg/feed"
	"github.com/mxpv/podsync/pkg/history"
	"github.com/mxpv/podsync/pkg/model"
	"github.com/mxpv/podsync/pkg/progress"
	"github.com/mxpv/podsync/services/api/models"
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// UpdateManager interface for triggering feed updates and episode retries
type UpdateManager interface {
	Update(ctx context.Context, feedConfig *feed.Config) error
	RetryEpisode(ctx context.Context, feedID, episodeID string) error
	DeleteEpisode(ctx context.Context, feedID, episodeID string) error
	BlockEpisode(ctx context.Context, feedID, episodeID string) error
	GetProgressTracker() *progress.Tracker
	GetHistoryManager() *history.Manager
}

// FeedsHandler handles feed-related API endpoints
type FeedsHandler struct {
	feeds      map[string]*feed.Config
	database   db.Storage
	configPath string
	writer     *config.Writer
	updater    UpdateManager
}

// NewFeedsHandler creates a new feeds handler
func NewFeedsHandler(feeds map[string]*feed.Config, database db.Storage, configPath string, updater UpdateManager) *FeedsHandler {
	return &FeedsHandler{
		feeds:      feeds,
		database:   database,
		configPath: configPath,
		writer:     config.NewWriter(configPath),
		updater:    updater,
	}
}

// ListFeeds returns all feeds
func (h *FeedsHandler) ListFeeds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	var feeds []models.FeedResponse

	// Walk through all feeds in database
	err := h.database.WalkFeeds(ctx, func(f *model.Feed) error {
		cfg, ok := h.feeds[f.ID]
		if !ok {
			return nil
		}

		// Count episodes for this feed (excluding ignored episodes)
		episodeCount := 0
		totalCount := 0
		_ = h.database.WalkEpisodes(ctx, f.ID, func(episode *model.Episode) error {
			totalCount++
			if episode.Status != model.EpisodeIgnored {
				episodeCount++
			}
			return nil
		})
		log.Infof("Feed %s: total=%d, non-ignored=%d, ignored=%d", f.ID, totalCount, episodeCount, totalCount-episodeCount)

		feedResp := models.FromModelFeed(f, cfg, episodeCount)
		feeds = append(feeds, feedResp)
		return nil
	})

	if err != nil {
		log.WithError(err).Error("failed to list feeds")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if feeds == nil {
		feeds = []models.FeedResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(feeds); err != nil {
		log.WithError(err).Error("failed to encode feeds response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetFeed returns a single feed by ID
func (h *FeedsHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract feed ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Feed ID required", http.StatusBadRequest)
		return
	}
	feedID := pathParts[3]

	ctx := r.Context()
	f, err := h.database.GetFeed(ctx, feedID)
	if err != nil {
		log.WithError(err).Errorf("failed to get feed %s", feedID)
		http.Error(w, "Feed not found", http.StatusNotFound)
		return
	}

	cfg, ok := h.feeds[feedID]
	if !ok {
		http.Error(w, "Feed configuration not found", http.StatusNotFound)
		return
	}

	// Count episodes for this feed (excluding ignored episodes)
	episodeCount := 0
	_ = h.database.WalkEpisodes(ctx, feedID, func(episode *model.Episode) error {
		if episode.Status != model.EpisodeIgnored {
			episodeCount++
		}
		return nil
	})

	feedResp := models.FromModelFeed(f, cfg, episodeCount)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(feedResp); err != nil {
		log.WithError(err).Error("failed to encode feed response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// CreateFeed creates a new feed
func (h *FeedsHandler) CreateFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateFeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Error("failed to decode create feed request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ID == "" || req.URL == "" {
		http.Error(w, "ID and URL are required", http.StatusBadRequest)
		return
	}

	// Check if feed already exists
	if _, ok := h.feeds[req.ID]; ok {
		http.Error(w, "Feed already exists", http.StatusConflict)
		return
	}

	// Add feed to config.toml
	err := h.writer.UpdatePartial(func(tree *toml.Tree) error {
		var feedsTree *toml.Tree
		if tree.Get("feeds") != nil {
			feedsTree = tree.Get("feeds").(*toml.Tree)
		}
		if feedsTree == nil {
			feedsTree, _ = toml.TreeFromMap(make(map[string]interface{}))
			tree.Set("feeds", feedsTree)
		}

		// Create new feed configuration
		feedConfig := map[string]interface{}{
			"url": req.URL,
		}

		// Add config fields if provided
		if req.Config.Format != "" {
			feedConfig["format"] = req.Config.Format
		}
		if req.Config.Quality != "" {
			feedConfig["quality"] = req.Config.Quality
		}
		if req.Config.MaxHeight > 0 {
			feedConfig["max_height"] = int64(req.Config.MaxHeight)
		}
		if req.Config.PageSize > 0 {
			feedConfig["page_size"] = int64(req.Config.PageSize)
		}
		if req.Config.UpdatePeriod != "" {
			feedConfig["update_period"] = req.Config.UpdatePeriod
		}
		if req.Config.CronSchedule != "" {
			feedConfig["cron_schedule"] = req.Config.CronSchedule
		}
		if req.Config.PlaylistSort != "" {
			feedConfig["playlist_sort"] = req.Config.PlaylistSort
		}
		feedConfig["opml"] = req.Config.OPML
		feedConfig["private_feed"] = req.Config.PrivateFeed

		// Add custom format if provided
		if req.Config.CustomFormat != nil && (req.Config.CustomFormat.YouTubeDLFormat != "" || req.Config.CustomFormat.Extension != "") {
			customFormatConfig := map[string]interface{}{}
			if req.Config.CustomFormat.YouTubeDLFormat != "" {
				customFormatConfig["youtube_dl_format"] = req.Config.CustomFormat.YouTubeDLFormat
			}
			if req.Config.CustomFormat.Extension != "" {
				customFormatConfig["extension"] = req.Config.CustomFormat.Extension
			}
			feedConfig["custom_format"] = customFormatConfig
		}

		// Add cleanup configuration
		if req.Config.CleanupKeep > 0 {
			cleanConfig := map[string]interface{}{
				"keep_last": int64(req.Config.CleanupKeep),
			}
			feedConfig["clean"] = cleanConfig
		}

		// Add filters if any are provided
		hasFilters := req.Config.Filters.Title != "" ||
			req.Config.Filters.NotTitle != "" ||
			req.Config.Filters.Description != "" ||
			req.Config.Filters.NotDescription != "" ||
			req.Config.Filters.MinDuration > 0 ||
			req.Config.Filters.MaxDuration > 0 ||
			req.Config.Filters.MinAge > 0 ||
			req.Config.Filters.MaxAge > 0

		if hasFilters {
			filters := make(map[string]interface{})
			if req.Config.Filters.Title != "" {
				filters["title"] = req.Config.Filters.Title
			}
			if req.Config.Filters.NotTitle != "" {
				filters["not_title"] = req.Config.Filters.NotTitle
			}
			if req.Config.Filters.Description != "" {
				filters["description"] = req.Config.Filters.Description
			}
			if req.Config.Filters.NotDescription != "" {
				filters["not_description"] = req.Config.Filters.NotDescription
			}
			if req.Config.Filters.MinDuration > 0 {
				filters["min_duration"] = req.Config.Filters.MinDuration
			}
			if req.Config.Filters.MaxDuration > 0 {
				filters["max_duration"] = req.Config.Filters.MaxDuration
			}
			if req.Config.Filters.MinAge > 0 {
				filters["min_age"] = int64(req.Config.Filters.MinAge)
			}
			if req.Config.Filters.MaxAge > 0 {
				filters["max_age"] = int64(req.Config.Filters.MaxAge)
			}
			feedConfig["filters"] = filters
		}

		// Add custom metadata if any are provided
		hasCustom := req.Config.Custom.CoverArt != "" ||
			req.Config.Custom.CoverArtQuality != "" ||
			req.Config.Custom.Category != "" ||
			len(req.Config.Custom.Subcategories) > 0 ||
			req.Config.Custom.Language != "" ||
			req.Config.Custom.Author != "" ||
			req.Config.Custom.Title != "" ||
			req.Config.Custom.Description != "" ||
			req.Config.Custom.OwnerName != "" ||
			req.Config.Custom.OwnerEmail != "" ||
			req.Config.Custom.Link != ""

		if hasCustom {
			custom := make(map[string]interface{})
			if req.Config.Custom.CoverArt != "" {
				custom["cover_art"] = req.Config.Custom.CoverArt
			}
			if req.Config.Custom.CoverArtQuality != "" {
				custom["cover_art_quality"] = req.Config.Custom.CoverArtQuality
			}
			if req.Config.Custom.Category != "" {
				custom["category"] = req.Config.Custom.Category
			}
			if len(req.Config.Custom.Subcategories) > 0 {
				custom["subcategories"] = req.Config.Custom.Subcategories
			}
			custom["explicit"] = req.Config.Custom.Explicit
			if req.Config.Custom.Language != "" {
				custom["lang"] = req.Config.Custom.Language
			}
			if req.Config.Custom.Author != "" {
				custom["author"] = req.Config.Custom.Author
			}
			if req.Config.Custom.Title != "" {
				custom["title"] = req.Config.Custom.Title
			}
			if req.Config.Custom.Description != "" {
				custom["description"] = req.Config.Custom.Description
			}
			if req.Config.Custom.OwnerName != "" {
				custom["ownerName"] = req.Config.Custom.OwnerName
			}
			if req.Config.Custom.OwnerEmail != "" {
				custom["ownerEmail"] = req.Config.Custom.OwnerEmail
			}
			if req.Config.Custom.Link != "" {
				custom["link"] = req.Config.Custom.Link
			}
			feedConfig["custom"] = custom
		}

		feedTree, _ := toml.TreeFromMap(feedConfig)
		feedsTree.Set(req.ID, feedTree)

		return nil
	})

	if err != nil {
		log.WithError(err).Error("failed to create feed")
		http.Error(w, "Failed to create feed", http.StatusInternalServerError)
		return
	}

	// Reload the config to get the newly created feed configuration
	tree, err := toml.LoadFile(h.configPath)
	if err != nil {
		log.WithError(err).Error("failed to reload config after feed creation")
		http.Error(w, "Failed to reload config", http.StatusInternalServerError)
		return
	}

	// Parse the new feed configuration
	feedsTree := tree.Get("feeds").(*toml.Tree)
	if feedsTree != nil {
		feedTree := feedsTree.Get(req.ID)
		if feedTree != nil {
			var newFeedConfig feed.Config
			if err := feedTree.(*toml.Tree).Unmarshal(&newFeedConfig); err != nil {
				log.WithError(err).Error("failed to unmarshal new feed config")
			} else {
				newFeedConfig.ID = req.ID
				// Add to in-memory feeds map
				h.feeds[req.ID] = &newFeedConfig
				log.WithField("feed_id", req.ID).Info("feed added to in-memory configuration")
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Feed created successfully",
		"id":      req.ID,
	})
}

// UpdateFeed updates an existing feed
func (h *FeedsHandler) UpdateFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract feed ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Feed ID required", http.StatusBadRequest)
		return
	}
	feedID := pathParts[3]

	var req models.UpdateFeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Error("failed to decode update feed request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if feed exists
	if _, ok := h.feeds[feedID]; !ok {
		http.Error(w, "Feed not found", http.StatusNotFound)
		return
	}

	// Update feed in config.toml
	err := h.writer.UpdatePartial(func(tree *toml.Tree) error {
		var feedsTree *toml.Tree
		if tree.Get("feeds") != nil {
			feedsTree = tree.Get("feeds").(*toml.Tree)
		}
		if feedsTree == nil {
			return errors.New("feeds section not found in config")
		}

		var feedTree *toml.Tree
		if feedsTree.Get(feedID) != nil {
			feedTree = feedsTree.Get(feedID).(*toml.Tree)
		}
		if feedTree == nil {
			return errors.New("feed not found in config")
		}

		// Update fields if provided
		if req.Config.Format != "" {
			feedTree.Set("format", req.Config.Format)
		}
		if req.Config.Quality != "" {
			feedTree.Set("quality", req.Config.Quality)
		}
		if req.Config.MaxHeight > 0 {
			feedTree.Set("max_height", int64(req.Config.MaxHeight))
		}
		if req.Config.PageSize > 0 {
			feedTree.Set("page_size", int64(req.Config.PageSize))
		}
		if req.Config.UpdatePeriod != "" {
			feedTree.Set("update_period", req.Config.UpdatePeriod)
		}
		if req.Config.CronSchedule != "" {
			feedTree.Set("cron_schedule", req.Config.CronSchedule)
		}
		if req.Config.PlaylistSort != "" {
			feedTree.Set("playlist_sort", req.Config.PlaylistSort)
		}
		feedTree.Set("opml", req.Config.OPML)
		feedTree.Set("private_feed", req.Config.PrivateFeed)

		// Update cleanup configuration
		if req.Config.CleanupKeep > 0 {
			cleanConfig := map[string]interface{}{
				"keep_last": int64(req.Config.CleanupKeep),
			}
			cleanTree, _ := toml.TreeFromMap(cleanConfig)
			feedTree.Set("clean", cleanTree)
		}

		// Update custom format if provided
		if req.Config.CustomFormat != nil && (req.Config.CustomFormat.YouTubeDLFormat != "" || req.Config.CustomFormat.Extension != "") {
			customFormatConfig := map[string]interface{}{}
			if req.Config.CustomFormat.YouTubeDLFormat != "" {
				customFormatConfig["youtube_dl_format"] = req.Config.CustomFormat.YouTubeDLFormat
			}
			if req.Config.CustomFormat.Extension != "" {
				customFormatConfig["extension"] = req.Config.CustomFormat.Extension
			}
			customFormatTree, _ := toml.TreeFromMap(customFormatConfig)
			feedTree.Set("custom_format", customFormatTree)
		}

		// Update filters if any are provided
		hasFilters := req.Config.Filters.Title != "" ||
			req.Config.Filters.NotTitle != "" ||
			req.Config.Filters.Description != "" ||
			req.Config.Filters.NotDescription != "" ||
			req.Config.Filters.MinDuration > 0 ||
			req.Config.Filters.MaxDuration > 0 ||
			req.Config.Filters.MinAge > 0 ||
			req.Config.Filters.MaxAge > 0

		if hasFilters {
			filters := make(map[string]interface{})
			if req.Config.Filters.Title != "" {
				filters["title"] = req.Config.Filters.Title
			}
			if req.Config.Filters.NotTitle != "" {
				filters["not_title"] = req.Config.Filters.NotTitle
			}
			if req.Config.Filters.Description != "" {
				filters["description"] = req.Config.Filters.Description
			}
			if req.Config.Filters.NotDescription != "" {
				filters["not_description"] = req.Config.Filters.NotDescription
			}
			if req.Config.Filters.MinDuration > 0 {
				filters["min_duration"] = req.Config.Filters.MinDuration
			}
			if req.Config.Filters.MaxDuration > 0 {
				filters["max_duration"] = req.Config.Filters.MaxDuration
			}
			if req.Config.Filters.MinAge > 0 {
				filters["min_age"] = int64(req.Config.Filters.MinAge)
			}
			if req.Config.Filters.MaxAge > 0 {
				filters["max_age"] = int64(req.Config.Filters.MaxAge)
			}
			filtersTree, _ := toml.TreeFromMap(filters)
			feedTree.Set("filters", filtersTree)
		}

		// Update custom metadata if any are provided
		hasCustom := req.Config.Custom.CoverArt != "" ||
			req.Config.Custom.CoverArtQuality != "" ||
			req.Config.Custom.Category != "" ||
			len(req.Config.Custom.Subcategories) > 0 ||
			req.Config.Custom.Language != "" ||
			req.Config.Custom.Author != "" ||
			req.Config.Custom.Title != "" ||
			req.Config.Custom.Description != "" ||
			req.Config.Custom.OwnerName != "" ||
			req.Config.Custom.OwnerEmail != "" ||
			req.Config.Custom.Link != ""

		if hasCustom {
			custom := make(map[string]interface{})
			if req.Config.Custom.CoverArt != "" {
				custom["cover_art"] = req.Config.Custom.CoverArt
			}
			if req.Config.Custom.CoverArtQuality != "" {
				custom["cover_art_quality"] = req.Config.Custom.CoverArtQuality
			}
			if req.Config.Custom.Category != "" {
				custom["category"] = req.Config.Custom.Category
			}
			if len(req.Config.Custom.Subcategories) > 0 {
				custom["subcategories"] = req.Config.Custom.Subcategories
			}
			custom["explicit"] = req.Config.Custom.Explicit
			if req.Config.Custom.Language != "" {
				custom["lang"] = req.Config.Custom.Language
			}
			if req.Config.Custom.Author != "" {
				custom["author"] = req.Config.Custom.Author
			}
			if req.Config.Custom.Title != "" {
				custom["title"] = req.Config.Custom.Title
			}
			if req.Config.Custom.Description != "" {
				custom["description"] = req.Config.Custom.Description
			}
			if req.Config.Custom.OwnerName != "" {
				custom["ownerName"] = req.Config.Custom.OwnerName
			}
			if req.Config.Custom.OwnerEmail != "" {
				custom["ownerEmail"] = req.Config.Custom.OwnerEmail
			}
			if req.Config.Custom.Link != "" {
				custom["link"] = req.Config.Custom.Link
			}
			customTree, _ := toml.TreeFromMap(custom)
			feedTree.Set("custom", customTree)
		}

		return nil
	})

	if err != nil {
		log.WithError(err).Errorf("failed to update feed %s", feedID)
		http.Error(w, "Failed to update feed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Feed updated successfully. Restart required for changes to take effect.",
		"id":      feedID,
	})
}

// DeleteFeed deletes a feed and its episodes
func (h *FeedsHandler) DeleteFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract feed ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Feed ID required", http.StatusBadRequest)
		return
	}
	feedID := pathParts[3]

	ctx := r.Context()

	// Delete from database (includes episodes)
	if err := h.database.DeleteFeed(ctx, feedID); err != nil {
		log.WithError(err).Errorf("failed to delete feed from database %s", feedID)
		http.Error(w, "Failed to delete feed", http.StatusInternalServerError)
		return
	}

	// Remove from config file
	err := h.writer.UpdatePartial(func(tree *toml.Tree) error {
		var feedsTree *toml.Tree
		if tree.Get("feeds") != nil {
			feedsTree = tree.Get("feeds").(*toml.Tree)
		}
		if feedsTree != nil {
			feedsTree.Delete(feedID)
		}
		return nil
	})
	if err != nil {
		log.WithError(err).Errorf("failed to remove feed from config %s", feedID)
		http.Error(w, "Failed to update config", http.StatusInternalServerError)
		return
	}

	// Remove from in-memory map
	delete(h.feeds, feedID)
	log.WithField("feed_id", feedID).Info("feed deleted successfully")

	w.WriteHeader(http.StatusNoContent)
}

// RefreshFeed triggers an immediate update for a specific feed
func (h *FeedsHandler) RefreshFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract feed ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		http.Error(w, "Feed ID required", http.StatusBadRequest)
		return
	}
	feedID := pathParts[3]

	// Check if updater is available (might be nil if no feeds configured at startup)
	if h.updater == nil {
		http.Error(w, "Update manager not available", http.StatusServiceUnavailable)
		return
	}

	// Get feed configuration
	feedConfig, ok := h.feeds[feedID]
	if !ok {
		http.Error(w, "Feed not found", http.StatusNotFound)
		return
	}

	// Trigger update in background with a detached context
	go func() {
		// Use context.Background() instead of request context so it doesn't get canceled
		ctx := context.Background()
		log.WithField("feed_id", feedID).Info("triggering manual feed refresh")
		if err := h.updater.Update(ctx, feedConfig); err != nil {
			log.WithError(err).Errorf("failed to refresh feed %s", feedID)
		} else {
			log.WithField("feed_id", feedID).Info("feed refresh completed successfully")
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Feed refresh triggered successfully",
		"id":      feedID,
	})
}
