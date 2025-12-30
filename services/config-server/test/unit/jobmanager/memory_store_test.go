package jobmanager_test

import (
	"context"
	"testing"
	"time"

	"github.com/fregataa/aami/config-server/internal/pkg/jobmanager"
)

func TestMemoryStore_SaveAndGet(t *testing.T) {
	store := jobmanager.NewMemoryStore()
	ctx := context.Background()

	job := &jobmanager.Job{
		ID:        "test-job-1",
		Type:      "test-type",
		Status:    jobmanager.JobStatusPending,
		CreatedAt: time.Now(),
	}

	// Save job
	err := store.Save(ctx, job)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Get job
	retrieved, err := store.Get(ctx, job.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.ID != job.ID {
		t.Errorf("ID mismatch: got %s, want %s", retrieved.ID, job.ID)
	}
	if retrieved.Type != job.Type {
		t.Errorf("Type mismatch: got %s, want %s", retrieved.Type, job.Type)
	}
	if retrieved.Status != job.Status {
		t.Errorf("Status mismatch: got %s, want %s", retrieved.Status, job.Status)
	}
}

func TestMemoryStore_SaveDuplicate(t *testing.T) {
	store := jobmanager.NewMemoryStore()
	ctx := context.Background()

	job := &jobmanager.Job{
		ID:        "test-job-1",
		Type:      "test-type",
		Status:    jobmanager.JobStatusPending,
		CreatedAt: time.Now(),
	}

	// First save should succeed
	err := store.Save(ctx, job)
	if err != nil {
		t.Fatalf("First Save failed: %v", err)
	}

	// Second save should fail
	err = store.Save(ctx, job)
	if err != jobmanager.ErrJobAlreadyExists {
		t.Errorf("Expected ErrJobAlreadyExists, got %v", err)
	}
}

func TestMemoryStore_GetNotFound(t *testing.T) {
	store := jobmanager.NewMemoryStore()
	ctx := context.Background()

	_, err := store.Get(ctx, "non-existent")
	if err != jobmanager.ErrJobNotFound {
		t.Errorf("Expected ErrJobNotFound, got %v", err)
	}
}

func TestMemoryStore_Update(t *testing.T) {
	store := jobmanager.NewMemoryStore()
	ctx := context.Background()

	job := &jobmanager.Job{
		ID:        "test-job-1",
		Type:      "test-type",
		Status:    jobmanager.JobStatusPending,
		CreatedAt: time.Now(),
	}

	// Save job
	err := store.Save(ctx, job)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Update job
	job.Status = jobmanager.JobStatusRunning
	now := time.Now()
	job.StartedAt = &now

	err = store.Update(ctx, job)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	retrieved, err := store.Get(ctx, job.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Status != jobmanager.JobStatusRunning {
		t.Errorf("Status not updated: got %s, want %s", retrieved.Status, jobmanager.JobStatusRunning)
	}
	if retrieved.StartedAt == nil {
		t.Error("StartedAt should not be nil")
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	store := jobmanager.NewMemoryStore()
	ctx := context.Background()

	job := &jobmanager.Job{
		ID:        "test-job-1",
		Type:      "test-type",
		Status:    jobmanager.JobStatusPending,
		CreatedAt: time.Now(),
	}

	// Save and delete
	_ = store.Save(ctx, job)
	err := store.Delete(ctx, job.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = store.Get(ctx, job.ID)
	if err != jobmanager.ErrJobNotFound {
		t.Errorf("Expected ErrJobNotFound after deletion, got %v", err)
	}
}

func TestMemoryStore_List(t *testing.T) {
	store := jobmanager.NewMemoryStore()
	ctx := context.Background()

	// Create jobs with different types and statuses
	jobs := []*jobmanager.Job{
		{ID: "job-1", Type: "type-a", Status: jobmanager.JobStatusPending, CreatedAt: time.Now().Add(-3 * time.Second)},
		{ID: "job-2", Type: "type-a", Status: jobmanager.JobStatusRunning, CreatedAt: time.Now().Add(-2 * time.Second)},
		{ID: "job-3", Type: "type-b", Status: jobmanager.JobStatusCompleted, CreatedAt: time.Now().Add(-1 * time.Second)},
	}

	for _, job := range jobs {
		_ = store.Save(ctx, job)
	}

	// List all jobs
	result, err := store.List(ctx, jobmanager.ListOptions{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("Expected 3 jobs, got %d", len(result))
	}

	// List by type
	result, err = store.List(ctx, jobmanager.ListOptions{Type: "type-a"})
	if err != nil {
		t.Fatalf("List by type failed: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 jobs of type-a, got %d", len(result))
	}

	// List by status
	result, err = store.List(ctx, jobmanager.ListOptions{Status: []jobmanager.JobStatus{jobmanager.JobStatusPending, jobmanager.JobStatusRunning}})
	if err != nil {
		t.Fatalf("List by status failed: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 pending/running jobs, got %d", len(result))
	}

	// List with pagination
	result, err = store.List(ctx, jobmanager.ListOptions{Limit: 2})
	if err != nil {
		t.Fatalf("List with limit failed: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 jobs with limit, got %d", len(result))
	}
}

func TestMemoryStore_GetRunningByType(t *testing.T) {
	store := jobmanager.NewMemoryStore()
	ctx := context.Background()

	// Create jobs
	jobs := []*jobmanager.Job{
		{ID: "job-1", Type: "type-a", Status: jobmanager.JobStatusCompleted, CreatedAt: time.Now()},
		{ID: "job-2", Type: "type-a", Status: jobmanager.JobStatusRunning, CreatedAt: time.Now()},
		{ID: "job-3", Type: "type-b", Status: jobmanager.JobStatusPending, CreatedAt: time.Now()},
	}

	for _, job := range jobs {
		_ = store.Save(ctx, job)
	}

	// Get running job of type-a
	running, err := store.GetRunningByType(ctx, "type-a")
	if err != nil {
		t.Fatalf("GetRunningByType failed: %v", err)
	}
	if running == nil {
		t.Fatal("Expected to find running job")
	}
	if running.ID != "job-2" {
		t.Errorf("Expected job-2, got %s", running.ID)
	}

	// Get running job of non-existent type
	running, err = store.GetRunningByType(ctx, "type-c")
	if err != nil {
		t.Fatalf("GetRunningByType failed: %v", err)
	}
	if running != nil {
		t.Error("Expected nil for non-existent type")
	}
}

func TestMemoryStore_DeleteExpired(t *testing.T) {
	store := jobmanager.NewMemoryStore()
	ctx := context.Background()

	now := time.Now()
	oldTime := now.Add(-2 * time.Hour)
	recentTime := now.Add(-30 * time.Minute)

	// Create jobs with different end times
	jobs := []*jobmanager.Job{
		{ID: "job-1", Type: "test", Status: jobmanager.JobStatusCompleted, CreatedAt: now, EndedAt: &oldTime},
		{ID: "job-2", Type: "test", Status: jobmanager.JobStatusCompleted, CreatedAt: now, EndedAt: &recentTime},
		{ID: "job-3", Type: "test", Status: jobmanager.JobStatusRunning, CreatedAt: now}, // No EndedAt
	}

	for _, job := range jobs {
		_ = store.Save(ctx, job)
	}

	// Delete jobs ended before 1 hour ago
	cutoff := now.Add(-1 * time.Hour)
	deleted, err := store.DeleteExpired(ctx, cutoff)
	if err != nil {
		t.Fatalf("DeleteExpired failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("Expected 1 deleted, got %d", deleted)
	}

	// Verify remaining jobs
	count, _ := store.Count(ctx)
	if count != 2 {
		t.Errorf("Expected 2 remaining jobs, got %d", count)
	}
}

func TestMemoryStore_IsolatesJobCopies(t *testing.T) {
	store := jobmanager.NewMemoryStore()
	ctx := context.Background()

	job := &jobmanager.Job{
		ID:        "test-job-1",
		Type:      "test-type",
		Status:    jobmanager.JobStatusPending,
		CreatedAt: time.Now(),
	}

	// Save job
	_ = store.Save(ctx, job)

	// Modify original
	job.Status = jobmanager.JobStatusRunning

	// Get from store - should still be pending
	retrieved, _ := store.Get(ctx, job.ID)
	if retrieved.Status != jobmanager.JobStatusPending {
		t.Error("Store did not isolate job copy")
	}

	// Modify retrieved
	retrieved.Status = jobmanager.JobStatusFailed

	// Get again - should still be pending
	retrieved2, _ := store.Get(ctx, job.ID)
	if retrieved2.Status != jobmanager.JobStatusPending {
		t.Error("Store did not isolate retrieved copy")
	}
}
