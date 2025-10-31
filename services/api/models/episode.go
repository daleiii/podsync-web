package models

import (
	"time"

	"github.com/daleiii/podsync-web/pkg/model"
)

// EpisodeResponse represents an episode in API responses
type EpisodeResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Duration    int64     `json:"duration"`
	Size        int64     `json:"size"`
	Status      string    `json:"status"`
	PubDate     time.Time `json:"pub_date"`
	FileURL     string    `json:"file_url"`
	Thumbnail   string    `json:"thumbnail"`
	FeedID      string    `json:"feed_id"`
	FeedTitle   string    `json:"feed_title"`
	VideoURL    string    `json:"video_url"`
	Error       string    `json:"error"`
}

// EpisodeListResponse represents paginated episode list
type EpisodeListResponse struct {
	Episodes   []EpisodeResponse `json:"episodes"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// EpisodeFilters represents filtering options for episodes
type EpisodeFilters struct {
	FeedID string `json:"feed_id"`
	Status string `json:"status"`
	Search string `json:"search"`
}

// FromModelEpisode converts a model.Episode to EpisodeResponse
func FromModelEpisode(episode *model.Episode, feedID, feedTitle, hostname string, format model.Format) EpisodeResponse {
	fileURL := ""
	if episode.Status == model.EpisodeDownloaded {
		ext := getExtensionFromFormat(format)
		fileURL = hostname + "/" + feedID + "/" + episode.ID + ext
	}

	return EpisodeResponse{
		ID:          episode.ID,
		Title:       episode.Title,
		Description: episode.Description,
		Duration:    episode.Duration,
		Size:        episode.Size,
		Status:      string(episode.Status),
		PubDate:     episode.PubDate,
		FileURL:     fileURL,
		Thumbnail:   episode.Thumbnail,
		FeedID:      feedID,
		FeedTitle:   feedTitle,
		VideoURL:    episode.VideoURL,
		Error:       episode.Error,
	}
}

func getExtensionFromFormat(format model.Format) string {
	switch format {
	case model.FormatAudio:
		return ".mp3"
	case model.FormatVideo:
		return ".mp4"
	default:
		return ".mp4"
	}
}
