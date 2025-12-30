package jobmanager

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Manager handles job submission and lifecycle management
type Manager struct {
	store   JobStore
	pool    *workerPool
	config  Config
	closed  bool
	mu      sync.RWMutex
	cleanup *time.Ticker
	done    chan struct{}
}

// Option is a functional option for configuring the Manager
type Option func(*Manager)

// WithStore sets a custom JobStore implementation
func WithStore(store JobStore) Option {
	return func(m *Manager) {
		m.store = store
	}
}

// WithConfig sets a custom configuration
func WithConfig(config Config) Option {
	return func(m *Manager) {
		m.config = config
	}
}

// WithMaxWorkers sets the maximum number of concurrent workers
func WithMaxWorkers(n int) Option {
	return func(m *Manager) {
		m.config.MaxWorkers = n
	}
}

// WithJobTTL sets how long completed jobs are kept
func WithJobTTL(d time.Duration) Option {
	return func(m *Manager) {
		m.config.JobTTL = d
	}
}

// NewManager creates a new job manager with the given options
func NewManager(opts ...Option) *Manager {
	m := &Manager{
		config: DefaultConfig(),
		done:   make(chan struct{}),
	}

	// Apply options
	for _, opt := range opts {
		opt(m)
	}

	// Use default memory store if none provided
	if m.store == nil {
		m.store = NewMemoryStore()
	}

	// Create worker pool
	m.pool = newWorkerPool(m.store, m.config.MaxWorkers, m.config.QueueSize)

	// Start cleanup routine
	m.startCleanup()

	slog.Info("Job manager started",
		"max_workers", m.config.MaxWorkers,
		"job_ttl", m.config.JobTTL,
		"cleanup_interval", m.config.CleanupInterval)

	return m
}

// Submit submits a new job for execution
// Returns the job ID immediately
func (m *Manager) Submit(ctx context.Context, jobType string, fn JobFunc) (string, error) {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return "", ErrManagerClosed
	}
	m.mu.RUnlock()

	job := &Job{
		ID:        uuid.New().String(),
		Type:      jobType,
		Status:    JobStatusPending,
		CreatedAt: time.Now(),
	}

	// Save to store
	if err := m.store.Save(ctx, job); err != nil {
		return "", fmt.Errorf("failed to save job: %w", err)
	}

	// Submit to worker pool
	if err := m.pool.submit(job, fn); err != nil {
		// Try to update job status to failed
		job.Status = JobStatusFailed
		job.Error = err.Error()
		now := time.Now()
		job.EndedAt = &now
		_ = m.store.Update(ctx, job)
		return "", fmt.Errorf("failed to submit job: %w", err)
	}

	slog.Info("Job submitted",
		"job_id", job.ID,
		"job_type", jobType)

	return job.ID, nil
}

// SubmitUnique submits a job, but returns an existing job if one of the same type is already running
// Returns (jobID, isNew, error)
func (m *Manager) SubmitUnique(ctx context.Context, jobType string, fn JobFunc) (string, bool, error) {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return "", false, ErrManagerClosed
	}
	m.mu.RUnlock()

	// Check for existing running job of the same type
	existing, err := m.store.GetRunningByType(ctx, jobType)
	if err != nil {
		return "", false, fmt.Errorf("failed to check existing jobs: %w", err)
	}

	if existing != nil {
		slog.Info("Returning existing job",
			"job_id", existing.ID,
			"job_type", jobType,
			"status", existing.Status)
		return existing.ID, false, nil
	}

	// No existing job, submit new one
	jobID, err := m.Submit(ctx, jobType, fn)
	if err != nil {
		return "", false, err
	}

	return jobID, true, nil
}

// Get retrieves a job by ID
func (m *Manager) Get(ctx context.Context, id string) (*Job, error) {
	return m.store.Get(ctx, id)
}

// Cancel cancels a running or pending job
func (m *Manager) Cancel(ctx context.Context, id string) error {
	job, err := m.store.Get(ctx, id)
	if err != nil {
		return err
	}

	if job.Status.IsTerminal() {
		return ErrJobNotCancellable
	}

	// Cancel in worker pool
	if m.pool.cancelJob(id) {
		// Job will be updated by worker
		return nil
	}

	// If not in worker pool (still in queue), update directly
	job.Status = JobStatusCancelled
	job.Error = "job was cancelled before execution"
	now := time.Now()
	job.EndedAt = &now

	return m.store.Update(ctx, job)
}

// List returns jobs matching the given options
func (m *Manager) List(ctx context.Context, opts ListOptions) ([]*Job, error) {
	return m.store.List(ctx, opts)
}

// Shutdown gracefully shuts down the manager
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil
	}
	m.closed = true
	m.mu.Unlock()

	slog.Info("Shutting down job manager")

	// Stop cleanup routine
	if m.cleanup != nil {
		m.cleanup.Stop()
	}
	close(m.done)

	// Shutdown worker pool
	if err := m.pool.shutdown(ctx); err != nil {
		slog.Error("Error shutting down worker pool", "error", err)
	}

	// Close store
	if err := m.store.Close(); err != nil {
		slog.Error("Error closing job store", "error", err)
	}

	slog.Info("Job manager shutdown complete")
	return nil
}

// Stats returns current manager statistics
func (m *Manager) Stats(ctx context.Context) (*ManagerStats, error) {
	count, err := m.store.Count(ctx)
	if err != nil {
		return nil, err
	}

	return &ManagerStats{
		TotalJobs:   count,
		QueueLength: m.pool.queueLength(),
		MaxWorkers:  m.config.MaxWorkers,
	}, nil
}

// startCleanup starts the periodic cleanup routine
func (m *Manager) startCleanup() {
	m.cleanup = time.NewTicker(m.config.CleanupInterval)

	go func() {
		for {
			select {
			case <-m.done:
				return
			case <-m.cleanup.C:
				m.runCleanup()
			}
		}
	}()
}

// runCleanup removes expired jobs
func (m *Manager) runCleanup() {
	cutoff := time.Now().Add(-m.config.JobTTL)
	deleted, err := m.store.DeleteExpired(context.Background(), cutoff)
	if err != nil {
		slog.Error("Failed to cleanup expired jobs", "error", err)
		return
	}

	if deleted > 0 {
		slog.Info("Cleaned up expired jobs", "count", deleted)
	}
}

// ManagerStats contains statistics about the manager
type ManagerStats struct {
	TotalJobs   int `json:"total_jobs"`
	QueueLength int `json:"queue_length"`
	MaxWorkers  int `json:"max_workers"`
}
