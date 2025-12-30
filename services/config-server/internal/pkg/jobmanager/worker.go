package jobmanager

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// jobItem represents a job in the queue
type jobItem struct {
	job *Job
	fn  JobFunc
}

// workerPool manages a pool of workers that process jobs
type workerPool struct {
	store       JobStore
	queue       chan *jobItem
	workerCount int
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc

	// Track cancellable jobs
	mu          sync.RWMutex
	cancellable map[string]context.CancelFunc
}

// newWorkerPool creates a new worker pool
func newWorkerPool(store JobStore, workerCount int, queueSize int) *workerPool {
	ctx, cancel := context.WithCancel(context.Background())

	wp := &workerPool{
		store:       store,
		queue:       make(chan *jobItem, queueSize),
		workerCount: workerCount,
		ctx:         ctx,
		cancel:      cancel,
		cancellable: make(map[string]context.CancelFunc),
	}

	// Start workers
	for i := 0; i < workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	return wp
}

// submit adds a job to the queue
func (wp *workerPool) submit(job *Job, fn JobFunc) error {
	item := &jobItem{
		job: job,
		fn:  fn,
	}

	select {
	case wp.queue <- item:
		return nil
	case <-wp.ctx.Done():
		return ErrManagerClosed
	default:
		return ErrQueueFull
	}
}

// cancelJob cancels a running job
func (wp *workerPool) cancelJob(jobID string) bool {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if cancel, exists := wp.cancellable[jobID]; exists {
		cancel()
		delete(wp.cancellable, jobID)
		return true
	}
	return false
}

// registerCancellable registers a job's cancel function for later cancellation
func (wp *workerPool) registerCancellable(jobID string, cancel context.CancelFunc) {
	wp.mu.Lock()
	wp.cancellable[jobID] = cancel
	wp.mu.Unlock()
}

// unregisterCancellable removes a job's cancel function from tracking
func (wp *workerPool) unregisterCancellable(jobID string) {
	wp.mu.Lock()
	delete(wp.cancellable, jobID)
	wp.mu.Unlock()
}

// worker processes jobs from the queue
func (wp *workerPool) worker(id int) {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case item, ok := <-wp.queue:
			if !ok {
				return
			}
			wp.processJob(id, item)
		}
	}
}

// processJob executes a single job
func (wp *workerPool) processJob(workerID int, item *jobItem) {
	job := item.job

	// Create job-specific context that can be cancelled
	jobCtx, jobCancel := context.WithCancel(wp.ctx)
	defer jobCancel()

	wp.registerCancellable(job.ID, jobCancel)
	defer wp.unregisterCancellable(job.ID)

	if err := wp.markJobRunning(jobCtx, job, workerID); err != nil {
		return
	}

	result, err := item.fn(jobCtx, wp.createProgressCallback(job))
	wp.finalizeJob(jobCtx, job, result, err, workerID)
}

// markJobRunning updates job status to running and logs the start
func (wp *workerPool) markJobRunning(ctx context.Context, job *Job, workerID int) error {
	now := time.Now()
	job.Status = JobStatusRunning
	job.StartedAt = &now

	if err := wp.store.Update(ctx, job); err != nil {
		slog.Error("Failed to update job status to running",
			"worker_id", workerID,
			"job_id", job.ID,
			"error", err)
		return err
	}

	slog.Info("Job started",
		"worker_id", workerID,
		"job_id", job.ID,
		"job_type", job.Type)

	return nil
}

// createProgressCallback returns a function for reporting job progress
func (wp *workerPool) createProgressCallback(job *Job) func(int) {
	return func(progress int) {
		progress = clamp(progress, 0, 100)
		job.Progress = progress

		if err := wp.store.Update(context.Background(), job); err != nil {
			slog.Warn("Failed to update job progress",
				"job_id", job.ID,
				"progress", progress,
				"error", err)
		}
	}
}

// finalizeJob updates job status based on execution result and persists final state
func (wp *workerPool) finalizeJob(ctx context.Context, job *Job, result any, execErr error, workerID int) {
	endTime := time.Now()
	job.EndedAt = &endTime
	job.Progress = 100

	switch {
	case ctx.Err() == context.Canceled:
		job.Status = JobStatusCancelled
		job.Error = "job was cancelled"
		slog.Info("Job cancelled",
			"worker_id", workerID,
			"job_id", job.ID,
			"job_type", job.Type,
			"duration", job.Duration())

	case execErr != nil:
		job.Status = JobStatusFailed
		job.Error = execErr.Error()
		slog.Error("Job failed",
			"worker_id", workerID,
			"job_id", job.ID,
			"job_type", job.Type,
			"error", execErr,
			"duration", job.Duration())

	default:
		job.Status = JobStatusCompleted
		job.Result = result
		slog.Info("Job completed",
			"worker_id", workerID,
			"job_id", job.ID,
			"job_type", job.Type,
			"duration", job.Duration())
	}

	if err := wp.store.Update(context.Background(), job); err != nil {
		slog.Error("Failed to update job final status",
			"worker_id", workerID,
			"job_id", job.ID,
			"error", err)
	}
}

// clamp restricts a value to be within the specified range
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// shutdown gracefully shuts down the worker pool
func (wp *workerPool) shutdown(ctx context.Context) error {
	// Signal workers to stop
	wp.cancel()

	// Close the queue
	close(wp.queue)

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// queueLength returns the current number of jobs in the queue
func (wp *workerPool) queueLength() int {
	return len(wp.queue)
}
