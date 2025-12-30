package jobmanager

import (
	"context"
	"time"
)

// JobStatus represents the current state of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// IsTerminal returns true if the job status is a terminal state
func (s JobStatus) IsTerminal() bool {
	return s == JobStatusCompleted || s == JobStatusFailed || s == JobStatusCancelled
}

// Job represents an asynchronous job
type Job struct {
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	Status    JobStatus  `json:"status"`
	Result    any        `json:"result,omitempty"`
	Error     string     `json:"error,omitempty"`
	Progress  int        `json:"progress,omitempty"` // 0-100
	CreatedAt time.Time  `json:"created_at"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

// Duration returns the duration of the job execution
// Returns 0 if the job hasn't started or ended yet
func (j *Job) Duration() time.Duration {
	if j.StartedAt == nil {
		return 0
	}
	if j.EndedAt == nil {
		return time.Since(*j.StartedAt)
	}
	return j.EndedAt.Sub(*j.StartedAt)
}

// JobFunc is the function signature for job execution
// The progress callback can be used to report progress (0-100)
type JobFunc func(ctx context.Context, progress func(int)) (any, error)

// Config holds configuration for the job manager
type Config struct {
	// MaxWorkers is the maximum number of concurrent workers
	// Default: 5
	MaxWorkers int

	// JobTTL is how long completed/failed jobs are kept in storage
	// Default: 1 hour
	JobTTL time.Duration

	// CleanupInterval is how often to run the cleanup routine
	// Default: 5 minutes
	CleanupInterval time.Duration

	// QueueSize is the size of the job queue buffer
	// Default: 100
	QueueSize int
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		MaxWorkers:      5,
		JobTTL:          time.Hour,
		CleanupInterval: 5 * time.Minute,
		QueueSize:       100,
	}
}

// ListOptions contains options for listing jobs
type ListOptions struct {
	// Filter by job type
	Type string

	// Filter by status
	Status []JobStatus

	// Pagination
	Limit  int
	Offset int

	// Sort order (default: newest first)
	OldestFirst bool
}

// SubmitResult is returned when a job is submitted
type SubmitResult struct {
	JobID    string `json:"job_id"`
	IsNew    bool   `json:"is_new"`    // false if reusing existing job (SubmitUnique)
	PollURL  string `json:"poll_url"`
}
