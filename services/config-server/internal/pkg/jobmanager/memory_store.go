package jobmanager

import (
	"context"
	"sort"
	"sync"
	"time"
)

// MemoryStore is an in-memory implementation of JobStore
type MemoryStore struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

// NewMemoryStore creates a new in-memory job store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		jobs: make(map[string]*Job),
	}
}

// Save stores a new job
func (s *MemoryStore) Save(ctx context.Context, job *Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[job.ID]; exists {
		return ErrJobAlreadyExists
	}

	// Store a copy to prevent external modification
	s.jobs[job.ID] = s.copyJob(job)
	return nil
}

// Get retrieves a job by ID
func (s *MemoryStore) Get(ctx context.Context, id string) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[id]
	if !exists {
		return nil, ErrJobNotFound
	}

	return s.copyJob(job), nil
}

// Update updates an existing job
func (s *MemoryStore) Update(ctx context.Context, job *Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[job.ID]; !exists {
		return ErrJobNotFound
	}

	s.jobs[job.ID] = s.copyJob(job)
	return nil
}

// Delete removes a job by ID
func (s *MemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[id]; !exists {
		return ErrJobNotFound
	}

	delete(s.jobs, id)
	return nil
}

// List returns jobs matching the given options
func (s *MemoryStore) List(ctx context.Context, opts ListOptions) ([]*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Job

	for _, job := range s.jobs {
		// Filter by type
		if opts.Type != "" && job.Type != opts.Type {
			continue
		}

		// Filter by status
		if len(opts.Status) > 0 && !s.statusMatches(job.Status, opts.Status) {
			continue
		}

		result = append(result, s.copyJob(job))
	}

	// Sort by creation time
	sort.Slice(result, func(i, j int) bool {
		if opts.OldestFirst {
			return result[i].CreatedAt.Before(result[j].CreatedAt)
		}
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	// Apply pagination
	if opts.Offset > 0 {
		if opts.Offset >= len(result) {
			return []*Job{}, nil
		}
		result = result[opts.Offset:]
	}

	if opts.Limit > 0 && opts.Limit < len(result) {
		result = result[:opts.Limit]
	}

	return result, nil
}

// GetByType returns jobs of a specific type with optional status filter
func (s *MemoryStore) GetByType(ctx context.Context, jobType string, status ...JobStatus) ([]*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Job

	for _, job := range s.jobs {
		if job.Type != jobType {
			continue
		}

		if len(status) > 0 && !s.statusMatches(job.Status, status) {
			continue
		}

		result = append(result, s.copyJob(job))
	}

	return result, nil
}

// GetRunningByType returns the first running or pending job of the given type
func (s *MemoryStore) GetRunningByType(ctx context.Context, jobType string) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, job := range s.jobs {
		if job.Type == jobType && (job.Status == JobStatusPending || job.Status == JobStatusRunning) {
			return s.copyJob(job), nil
		}
	}

	return nil, nil
}

// DeleteExpired removes jobs that ended before the given time
func (s *MemoryStore) DeleteExpired(ctx context.Context, before time.Time) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for id, job := range s.jobs {
		// Only delete terminal jobs
		if !job.Status.IsTerminal() {
			continue
		}

		// Check if job ended before the cutoff time
		if job.EndedAt != nil && job.EndedAt.Before(before) {
			delete(s.jobs, id)
			count++
		}
	}

	return count, nil
}

// Count returns the total number of jobs in the store
func (s *MemoryStore) Count(ctx context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.jobs), nil
}

// Close releases any resources held by the store
func (s *MemoryStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs = make(map[string]*Job)
	return nil
}

// statusMatches checks if the given status is in the list
func (s *MemoryStore) statusMatches(status JobStatus, list []JobStatus) bool {
	for _, st := range list {
		if status == st {
			return true
		}
	}
	return false
}

// copyJob creates a deep copy of a job to prevent external modification
func (s *MemoryStore) copyJob(job *Job) *Job {
	if job == nil {
		return nil
	}

	copied := &Job{
		ID:        job.ID,
		Type:      job.Type,
		Status:    job.Status,
		Result:    job.Result, // Note: Result is not deep copied
		Error:     job.Error,
		Progress:  job.Progress,
		CreatedAt: job.CreatedAt,
	}

	if job.StartedAt != nil {
		t := *job.StartedAt
		copied.StartedAt = &t
	}

	if job.EndedAt != nil {
		t := *job.EndedAt
		copied.EndedAt = &t
	}

	return copied
}
