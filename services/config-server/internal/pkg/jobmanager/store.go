package jobmanager

import (
	"context"
	"time"
)

// JobStore defines the interface for job storage
// Implementations can use in-memory, Redis, PostgreSQL, etc.
type JobStore interface {
	// Save stores a new job
	// Returns error if job with same ID already exists
	Save(ctx context.Context, job *Job) error

	// Get retrieves a job by ID
	// Returns nil, ErrJobNotFound if not found
	Get(ctx context.Context, id string) (*Job, error)

	// Update updates an existing job
	// Returns ErrJobNotFound if job doesn't exist
	Update(ctx context.Context, job *Job) error

	// Delete removes a job by ID
	// Returns ErrJobNotFound if job doesn't exist
	Delete(ctx context.Context, id string) error

	// List returns jobs matching the given options
	List(ctx context.Context, opts ListOptions) ([]*Job, error)

	// GetByType returns jobs of a specific type with optional status filter
	// Used for duplicate detection in SubmitUnique
	GetByType(ctx context.Context, jobType string, status ...JobStatus) ([]*Job, error)

	// GetRunningByType returns the first running or pending job of the given type
	// Returns nil, nil if no such job exists
	GetRunningByType(ctx context.Context, jobType string) (*Job, error)

	// DeleteExpired removes jobs that ended before the given time
	// Returns the number of deleted jobs
	DeleteExpired(ctx context.Context, before time.Time) (int, error)

	// Count returns the total number of jobs in the store
	Count(ctx context.Context) (int, error)

	// Close releases any resources held by the store
	Close() error
}
