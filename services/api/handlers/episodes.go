package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/daleiii/podsync-web/pkg/db"
	"github.com/daleiii/podsync-web/pkg/feed"
	"github.com/daleiii/podsync-web/pkg/model"
	"github.com/daleiii/podsync-web/services/api/models"
	log "github.com/sirupsen/logrus"
)

// EpisodesHandler handles episode-related API endpoints
type EpisodesHandler struct {
	feeds    map[string]*feed.Config
	database db.Storage
	hostname string
	updater  UpdateManager
}

// NewEpisodesHandler creates a new episodes handler
func NewEpisodesHandler(feeds map[string]*feed.Config, database db.Storage, hostname string, updater UpdateManager) *EpisodesHandler {
	return &EpisodesHandler{
		feeds:    feeds,
		database: database,
		hostname: hostname,
		updater:  updater,
	}
}

// ListEpisodes returns paginated episodes with filtering
func (h *EpisodesHandler) ListEpisodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	feedID := query.Get("feed_id")
	status := query.Get("status")
	search := strings.ToLower(query.Get("search"))
	showIgnored := query.Get("show_ignored") == "true"
	dateFilter := query.Get("date_filter") // today, yesterday, week, month, year, all
	dateStart := query.Get("date_start")   // custom start date (RFC3339 format)
	dateEnd := query.Get("date_end")       // custom end date (RFC3339 format)

	ctx := r.Context()
	var allEpisodes []models.EpisodeResponse

	// Walk through all feeds
	err := h.database.WalkFeeds(ctx, func(f *model.Feed) error {
		// Filter by feed ID if specified
		if feedID != "" && f.ID != feedID {
			return nil
		}

		// Walk through episodes in this feed
		return h.database.WalkEpisodes(ctx, f.ID, func(episode *model.Episode) error {
			// Filter out ignored episodes by default unless showIgnored is true
			if !showIgnored && episode.Status == model.EpisodeIgnored {
				return nil
			}

			// Filter by status if specified
			if status != "" && string(episode.Status) != status {
				return nil
			}

			// Filter by search term if specified
			if search != "" {
				titleLower := strings.ToLower(episode.Title)
				descLower := strings.ToLower(episode.Description)
				if !strings.Contains(titleLower, search) && !strings.Contains(descLower, search) {
					return nil
				}
			}

			// Filter by custom date range if specified
			if dateStart != "" || dateEnd != "" {
				var startTime, endTime time.Time
				var err error

				if dateStart != "" {
					startTime, err = time.Parse(time.RFC3339, dateStart)
					if err != nil {
						// Try parsing as date only (YYYY-MM-DD)
						startTime, err = time.Parse("2006-01-02", dateStart)
						if err == nil {
							// Set to start of day
							startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())
						}
					}
				}

				if dateEnd != "" {
					endTime, err = time.Parse(time.RFC3339, dateEnd)
					if err != nil {
						// Try parsing as date only (YYYY-MM-DD)
						endTime, err = time.Parse("2006-01-02", dateEnd)
						if err == nil {
							// Set to end of day
							endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 999999999, endTime.Location())
						}
					}
				}

				// Filter by date range
				if !startTime.IsZero() && episode.PubDate.Before(startTime) {
					return nil
				}
				if !endTime.IsZero() && episode.PubDate.After(endTime) {
					return nil
				}
			} else if dateFilter != "" && dateFilter != "all" {
				// Filter by preset date filter
				now := time.Now()
				var cutoffTime time.Time

				switch dateFilter {
				case "today":
					cutoffTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				case "yesterday":
					yesterday := now.AddDate(0, 0, -1)
					cutoffTime = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, now.Location())
				case "week":
					cutoffTime = now.AddDate(0, 0, -7)
				case "month":
					cutoffTime = now.AddDate(0, -1, 0)
				case "year":
					cutoffTime = now.AddDate(-1, 0, 0)
				default:
					// Invalid filter, skip filtering
				}

				if !cutoffTime.IsZero() && episode.PubDate.Before(cutoffTime) {
					return nil
				}
			}

			episodeResp := models.FromModelEpisode(episode, f.ID, f.Title, h.hostname, f.Format)
			allEpisodes = append(allEpisodes, episodeResp)
			return nil
		})
	})

	if err != nil {
		log.WithError(err).Error("failed to list episodes")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Sort episodes by published date (newest first)
	sort.Slice(allEpisodes, func(i, j int) bool {
		return allEpisodes[i].PubDate.After(allEpisodes[j].PubDate)
	})

	// Calculate pagination
	total := len(allEpisodes)
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedEpisodes := allEpisodes[start:end]
	if paginatedEpisodes == nil {
		paginatedEpisodes = []models.EpisodeResponse{}
	}

	response := models.EpisodeListResponse{
		Episodes:   paginatedEpisodes,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.WithError(err).Error("failed to encode episodes response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// DeleteEpisode deletes a specific episode (database entry and media file)
func (h *EpisodesHandler) DeleteEpisode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract episode ID from URL path: /api/v1/episodes/:feedID/:episodeID
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 5 {
		http.Error(w, "Feed ID and Episode ID required", http.StatusBadRequest)
		return
	}
	feedID := pathParts[3]
	episodeID := pathParts[4]

	ctx := r.Context()

	// Delete both database entry and media file
	if err := h.updater.DeleteEpisode(ctx, feedID, episodeID); err != nil {
		log.WithError(err).Errorf("failed to delete episode %s/%s", feedID, episodeID)
		http.Error(w, fmt.Sprintf("Failed to delete episode: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// BlockEpisode blocks an episode from being re-downloaded
func (h *EpisodesHandler) BlockEpisode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract episode ID from URL path: /api/v1/episodes/:feedID/:episodeID/block
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 5 {
		http.Error(w, "Feed ID and Episode ID required", http.StatusBadRequest)
		return
	}
	feedID := pathParts[3]
	episodeID := pathParts[4]

	ctx := r.Context()

	// Block the episode
	if err := h.updater.BlockEpisode(ctx, feedID, episodeID); err != nil {
		log.WithError(err).Errorf("failed to block episode %s/%s", feedID, episodeID)
		http.Error(w, fmt.Sprintf("Failed to block episode: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Episode blocked successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.WithError(err).Error("failed to encode block response")
	}
}

// RetryEpisode retries downloading a failed episode
func (h *EpisodesHandler) RetryEpisode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract episode ID from URL path: /api/v1/episodes/:feedID/:episodeID/retry
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 5 {
		http.Error(w, "Feed ID and Episode ID required", http.StatusBadRequest)
		return
	}
	feedID := pathParts[3]
	episodeID := pathParts[4]

	ctx := r.Context()

	// Trigger the retry
	if err := h.updater.RetryEpisode(ctx, feedID, episodeID); err != nil {
		log.WithError(err).Errorf("failed to retry episode %s/%s", feedID, episodeID)
		http.Error(w, fmt.Sprintf("Failed to retry episode: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Episode retry initiated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.WithError(err).Error("failed to encode retry response")
	}
}
