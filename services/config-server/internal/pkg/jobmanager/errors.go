package jobmanager

import "errors"

var (
	// ErrJobNotFound is returned when a job is not found
	ErrJobNotFound = errors.New("job not found")

	// ErrJobAlreadyExists is returned when trying to save a job with an existing ID
	ErrJobAlreadyExists = errors.New("job already exists")

	// ErrJobNotCancellable is returned when trying to cancel a job that cannot be cancelled
	ErrJobNotCancellable = errors.New("job cannot be cancelled")

	// ErrManagerClosed is returned when operations are attempted on a closed manager
	ErrManagerClosed = errors.New("job manager is closed")

	// ErrQueueFull is returned when the job queue is full
	ErrQueueFull = errors.New("job queue is full")
)
