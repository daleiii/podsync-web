package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mxpv/podsync/pkg/db"
	"github.com/mxpv/podsync/pkg/history"
	"github.com/mxpv/podsync/pkg/model"
	log "github.com/sirupsen/logrus"
)

// HistoryHandler handles history-related API endpoints
type HistoryHandler struct {
	database       db.Storage
	historyManager *history.Manager
	retentionDays  int
	maxEntries     int
}

// NewHistoryHandler creates a new history handler
func NewHistoryHandler(database db.Storage, historyManager *history.Manager, retentionDays, maxEntries int) *HistoryHandler {
	return &HistoryHandler{
		database:       database,
		historyManager: historyManager,
		retentionDays:  retentionDays,
		maxEntries:     maxEntries,
	}
}

// HistoryListResponse represents the paginated response for history entries
type HistoryListResponse struct {
	Entries    []*model.HistoryEntry `json:"entries"`
	Total      int                   `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

// HistoryStatsResponse represents history statistics
type HistoryStatsResponse struct {
	Count       int                 `json:"count"`
	OldestEntry *model.HistoryEntry `json:"oldest_entry,omitempty"`
}

// ListHistory returns paginated history entries with filters
func (h *HistoryHandler) ListHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	query := r.URL.Query()

	// Parse pagination parameters
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// Parse filters
	filters := model.HistoryFilters{
		FeedID:  query.Get("feed_id"),
		JobType: model.JobType(query.Get("job_type")),
		Status:  model.JobStatus(query.Get("status")),
		Search:  query.Get("search"),
	}

	// Parse date range
	if startDateStr := query.Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filters.StartDate = startDate
		}
	}

	if endDateStr := query.Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filters.EndDate = endDate
		}
	}

	// Fetch history entries
	entries, total, err := h.database.ListHistory(ctx, filters, page, pageSize)
	if err != nil {
		log.WithError(err).Error("failed to list history")
		http.Error(w, "Failed to fetch history", http.StatusInternalServerError)
		return
	}

	totalPages := (total + pageSize - 1) / pageSize

	response := HistoryListResponse{
		Entries:    entries,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.WithError(err).Error("failed to encode history response")
	}
}

// GetHistory returns a single history entry by ID
func (h *HistoryHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/history/")
	if id == "" {
		http.Error(w, "History ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	entry, err := h.database.GetHistory(ctx, id)
	if err != nil {
		if err == model.ErrNotFound {
			http.Error(w, "History entry not found", http.StatusNotFound)
			return
		}
		log.WithError(err).Error("failed to get history entry")
		http.Error(w, "Failed to fetch history entry", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(entry); err != nil {
		log.WithError(err).Error("failed to encode history entry")
	}
}

// DeleteHistory deletes a history entry by ID
func (h *HistoryHandler) DeleteHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/history/")
	if id == "" {
		http.Error(w, "History ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.database.DeleteHistory(ctx, id); err != nil {
		log.WithError(err).Error("failed to delete history entry")
		http.Error(w, "Failed to delete history entry", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "History entry deleted successfully",
	})
}

// DeleteAllHistory deletes all history entries
func (h *HistoryHandler) DeleteAllHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Delete all by setting retention to 0 days
	if err := h.database.CleanupHistory(ctx, 0, 0); err != nil {
		log.WithError(err).Error("failed to delete all history")
		http.Error(w, "Failed to delete all history", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "All history entries deleted successfully",
	})
}

// GetHistoryStats returns statistics about history
func (h *HistoryHandler) GetHistoryStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	count, oldestEntry, err := h.database.GetHistoryStats(ctx)
	if err != nil {
		log.WithError(err).Error("failed to get history stats")
		http.Error(w, "Failed to fetch history statistics", http.StatusInternalServerError)
		return
	}

	response := HistoryStatsResponse{
		Count:       count,
		OldestEntry: oldestEntry,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.WithError(err).Error("failed to encode history stats")
	}
}

// CleanupHistory triggers history cleanup based on retention policy
func (h *HistoryHandler) CleanupHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	if err := h.historyManager.CleanupOldEntries(ctx, h.retentionDays, h.maxEntries); err != nil {
		log.WithError(err).Error("failed to cleanup history")
		http.Error(w, "Failed to cleanup history", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "History cleanup completed successfully",
	})
}
