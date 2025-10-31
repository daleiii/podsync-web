package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/daleiii/podsync-web/pkg/progress"
	log "github.com/sirupsen/logrus"
)

// ProgressHandler handles progress-related API endpoints
type ProgressHandler struct {
	progressTracker *progress.Tracker
}

// NewProgressHandler creates a new progress handler
func NewProgressHandler(progressTracker *progress.Tracker) *ProgressHandler {
	return &ProgressHandler{
		progressTracker: progressTracker,
	}
}

// ProgressResponse represents the current progress state
type ProgressResponse struct {
	Feeds    map[string]*progress.FeedProgress `json:"feeds"`
	Episodes []*progress.EpisodeProgress       `json:"episodes"`
}

// GetProgress returns a snapshot of current progress
func (h *ProgressHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Optional filter by feed ID
	feedID := r.URL.Query().Get("feedID")

	feeds := h.progressTracker.GetAllFeedProgress()
	var episodes []*progress.EpisodeProgress

	if feedID != "" {
		// Filter by specific feed
		episodes = h.progressTracker.GetEpisodesForFeed(feedID)
		// Filter feeds map to only include requested feed
		if fp, ok := feeds[feedID]; ok {
			feeds = map[string]*progress.FeedProgress{feedID: fp}
		} else {
			feeds = make(map[string]*progress.FeedProgress)
		}
	} else {
		// Get all episodes
		episodes = h.progressTracker.GetAllEpisodeProgress()
	}

	response := ProgressResponse{
		Feeds:    feeds,
		Episodes: episodes,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.WithError(err).Error("failed to encode progress response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// StreamProgress streams progress updates via Server-Sent Events (SSE)
func (h *ProgressHandler) StreamProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set up SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Optional filter by feed ID
	feedID := r.URL.Query().Get("feedID")

	// Get flusher to send data immediately
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	log.Info("SSE client connected to progress stream")

	// Send initial data immediately
	h.sendProgressEvent(w, flusher, feedID)

	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			log.Info("SSE client disconnected from progress stream")
			return
		case <-ticker.C:
			// Send progress update every 500ms
			if !h.sendProgressEvent(w, flusher, feedID) {
				// Write failed, client likely disconnected
				return
			}
		}
	}
}

// sendProgressEvent sends a single progress event via SSE
// Returns false if write failed (client disconnected)
func (h *ProgressHandler) sendProgressEvent(w http.ResponseWriter, flusher http.Flusher, feedID string) bool {
	var feeds map[string]*progress.FeedProgress
	var episodes []*progress.EpisodeProgress

	if feedID != "" {
		// Filter by specific feed
		if fp, ok := h.progressTracker.GetFeedProgress(feedID); ok {
			feeds = map[string]*progress.FeedProgress{feedID: fp}
		} else {
			feeds = make(map[string]*progress.FeedProgress)
		}
		episodes = h.progressTracker.GetEpisodesForFeed(feedID)
	} else {
		// Get all progress
		feeds = h.progressTracker.GetAllFeedProgress()
		episodes = h.progressTracker.GetAllEpisodeProgress()
	}

	response := ProgressResponse{
		Feeds:    feeds,
		Episodes: episodes,
	}

	// Marshal to JSON
	data, err := json.Marshal(response)
	if err != nil {
		log.WithError(err).Error("failed to marshal progress data")
		return true // Continue even if marshaling fails
	}

	// Send SSE event
	// Format: data: {...}\n\n
	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(data)); err != nil {
		log.WithError(err).Debug("failed to write SSE event, client likely disconnected")
		return false
	}

	flusher.Flush()
	return true
}

// ProgressManager interface for accessing progress tracker from update manager
type ProgressManager interface {
	GetProgressTracker() *progress.Tracker
}
