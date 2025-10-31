package model

import (
	"time"
)

// JobType represents the type of job being tracked
type JobType string

const (
	JobTypeFeedUpdate    = JobType("feed_update")
	JobTypeEpisodeRetry  = JobType("episode_retry")
	JobTypeEpisodeDelete = JobType("episode_delete")
	JobTypeEpisodeBlock  = JobType("episode_block")
)

// JobStatus represents the current status of a job
type JobStatus string

const (
	JobStatusRunning = JobStatus("running")
	JobStatusSuccess = JobStatus("success")
	JobStatusFailed  = JobStatus("failed")
	JobStatusPartial = JobStatus("partial") // Some episodes succeeded, some failed
)

// TriggerType represents how a job was initiated
type TriggerType string

const (
	TriggerScheduled = TriggerType("scheduled") // Cron schedule
	TriggerManual    = TriggerType("manual")    // User-initiated from UI/API
)

// HistoryEntry represents a single entry in the job history
type HistoryEntry struct {
	ID           string        `json:"id"`
	JobType      JobType       `json:"job_type"`
	FeedID       string        `json:"feed_id"`
	FeedTitle    string        `json:"feed_title"`
	EpisodeID    string        `json:"episode_id,omitempty"`    // For episode-specific operations
	EpisodeTitle string        `json:"episode_title,omitempty"` // For episode-specific operations
	StartTime    time.Time     `json:"start_time"`
	EndTime      *time.Time    `json:"end_time,omitempty"`
	Duration     time.Duration `json:"duration"` // In nanoseconds
	Status       JobStatus     `json:"status"`
	TriggerType  TriggerType   `json:"trigger_type"`
	Statistics   JobStatistics `json:"statistics"`
	Error        string        `json:"error,omitempty"` // Error message if status is failed
}

// JobStatistics contains metrics about a job execution
type JobStatistics struct {
	EpisodesQueued     int             `json:"episodes_queued"`
	EpisodesDownloaded int             `json:"episodes_downloaded"`
	EpisodesFailed     int             `json:"episodes_failed"`
	EpisodesIgnored    int             `json:"episodes_ignored"`
	BytesDownloaded    int64           `json:"bytes_downloaded"`
	EpisodeDetails     []EpisodeDetail `json:"episode_details,omitempty"` // Detailed list of episodes
}

// EpisodeDetail contains information about an individual episode in a job
type EpisodeDetail struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"` // downloaded, failed, ignored, etc.
	Error    string `json:"error,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Duration int64  `json:"duration,omitempty"`
}

// HistoryFilters represents query filters for history entries
type HistoryFilters struct {
	FeedID    string    `json:"feed_id"`
	JobType   JobType   `json:"job_type"`
	Status    JobStatus `json:"status"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Search    string    `json:"search"` // Search in episode titles
}
