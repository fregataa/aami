package jobmanager_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fregataa/aami/config-server/internal/pkg/jobmanager"
)

func TestManager_Submit(t *testing.T) {
	manager := jobmanager.NewManager(
		jobmanager.WithMaxWorkers(2),
		jobmanager.WithJobTTL(time.Hour),
	)
	defer manager.Shutdown(context.Background())

	ctx := context.Background()
	executed := make(chan struct{})

	jobFn := func(ctx context.Context, progress func(int)) (any, error) {
		close(executed)
		return "success", nil
	}

	jobID, err := manager.Submit(ctx, "test-job", jobFn)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}
	if jobID == "" {
		t.Error("Expected non-empty job ID")
	}

	// Wait for job execution
	select {
	case <-executed:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Job did not execute in time")
	}

	// Wait a bit for job to complete
	time.Sleep(100 * time.Millisecond)

	// Check job status
	job, err := manager.Get(ctx, jobID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if job.Status != jobmanager.JobStatusCompleted {
		t.Errorf("Expected completed status, got %s", job.Status)
	}
	if job.Result != "success" {
		t.Errorf("Expected 'success' result, got %v", job.Result)
	}
}

func TestManager_SubmitUnique(t *testing.T) {
	manager := jobmanager.NewManager(
		jobmanager.WithMaxWorkers(1),
	)
	defer manager.Shutdown(context.Background())

	ctx := context.Background()
	started := make(chan struct{})
	done := make(chan struct{})

	// Long-running job
	jobFn := func(ctx context.Context, progress func(int)) (any, error) {
		close(started)
		<-done
		return "completed", nil
	}

	// Submit first job
	jobID1, isNew1, err := manager.SubmitUnique(ctx, "unique-job", jobFn)
	if err != nil {
		t.Fatalf("First SubmitUnique failed: %v", err)
	}
	if !isNew1 {
		t.Error("First job should be new")
	}

	// Wait for job to start
	<-started

	// Submit second job of same type - should return existing
	jobID2, isNew2, err := manager.SubmitUnique(ctx, "unique-job", jobFn)
	if err != nil {
		t.Fatalf("Second SubmitUnique failed: %v", err)
	}
	if isNew2 {
		t.Error("Second job should not be new")
	}
	if jobID2 != jobID1 {
		t.Errorf("Expected same job ID, got %s and %s", jobID1, jobID2)
	}

	// Let job complete
	close(done)
}

func TestManager_JobFailure(t *testing.T) {
	manager := jobmanager.NewManager(
		jobmanager.WithMaxWorkers(2),
	)
	defer manager.Shutdown(context.Background())

	ctx := context.Background()
	expectedErr := errors.New("job failed")

	jobFn := func(ctx context.Context, progress func(int)) (any, error) {
		return nil, expectedErr
	}

	jobID, err := manager.Submit(ctx, "failing-job", jobFn)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	// Wait for job to complete
	time.Sleep(200 * time.Millisecond)

	job, err := manager.Get(ctx, jobID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if job.Status != jobmanager.JobStatusFailed {
		t.Errorf("Expected failed status, got %s", job.Status)
	}
	if job.Error != expectedErr.Error() {
		t.Errorf("Expected error '%s', got '%s'", expectedErr.Error(), job.Error)
	}
}

func TestManager_Progress(t *testing.T) {
	manager := jobmanager.NewManager(
		jobmanager.WithMaxWorkers(2),
	)
	defer manager.Shutdown(context.Background())

	ctx := context.Background()
	progressValues := make([]int, 0)
	done := make(chan struct{})

	jobFn := func(ctx context.Context, progress func(int)) (any, error) {
		for i := 0; i <= 100; i += 25 {
			progress(i)
			progressValues = append(progressValues, i)
		}
		close(done)
		return nil, nil
	}

	_, err := manager.Submit(ctx, "progress-job", jobFn)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	<-done

	// Verify progress was reported
	if len(progressValues) != 5 {
		t.Errorf("Expected 5 progress values, got %d", len(progressValues))
	}
}

func TestManager_Cancel(t *testing.T) {
	manager := jobmanager.NewManager(
		jobmanager.WithMaxWorkers(1),
	)
	defer manager.Shutdown(context.Background())

	ctx := context.Background()
	started := make(chan struct{})
	cancelled := make(chan struct{})

	jobFn := func(ctx context.Context, progress func(int)) (any, error) {
		close(started)
		<-ctx.Done()
		close(cancelled)
		return nil, ctx.Err()
	}

	jobID, err := manager.Submit(ctx, "cancellable-job", jobFn)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	// Wait for job to start
	<-started

	// Cancel the job
	err = manager.Cancel(ctx, jobID)
	if err != nil {
		t.Fatalf("Cancel failed: %v", err)
	}

	// Wait for cancellation
	select {
	case <-cancelled:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Job was not cancelled in time")
	}

	// Wait a bit for status update
	time.Sleep(100 * time.Millisecond)

	// Verify status
	job, err := manager.Get(ctx, jobID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if job.Status != jobmanager.JobStatusCancelled {
		t.Errorf("Expected cancelled status, got %s", job.Status)
	}
}

func TestManager_ConcurrentJobs(t *testing.T) {
	manager := jobmanager.NewManager(
		jobmanager.WithMaxWorkers(5),
	)
	defer manager.Shutdown(context.Background())

	ctx := context.Background()
	jobCount := 10
	var completed int32

	for i := 0; i < jobCount; i++ {
		jobFn := func(ctx context.Context, progress func(int)) (any, error) {
			time.Sleep(50 * time.Millisecond)
			atomic.AddInt32(&completed, 1)
			return nil, nil
		}

		_, err := manager.Submit(ctx, "concurrent-job", jobFn)
		if err != nil {
			t.Fatalf("Submit failed: %v", err)
		}
	}

	// Wait for all jobs to complete
	time.Sleep(500 * time.Millisecond)

	if atomic.LoadInt32(&completed) != int32(jobCount) {
		t.Errorf("Expected %d completed jobs, got %d", jobCount, completed)
	}
}

func TestManager_List(t *testing.T) {
	manager := jobmanager.NewManager(
		jobmanager.WithMaxWorkers(2),
	)
	defer manager.Shutdown(context.Background())

	ctx := context.Background()

	// Submit several jobs
	for i := 0; i < 5; i++ {
		jobFn := func(ctx context.Context, progress func(int)) (any, error) {
			return nil, nil
		}
		_, _ = manager.Submit(ctx, "list-test-job", jobFn)
	}

	// Wait for completion
	time.Sleep(200 * time.Millisecond)

	// List jobs
	jobs, err := manager.List(ctx, jobmanager.ListOptions{Type: "list-test-job"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(jobs) != 5 {
		t.Errorf("Expected 5 jobs, got %d", len(jobs))
	}
}

func TestManager_Stats(t *testing.T) {
	manager := jobmanager.NewManager(
		jobmanager.WithMaxWorkers(3),
	)
	defer manager.Shutdown(context.Background())

	ctx := context.Background()

	// Submit a few jobs
	for i := 0; i < 3; i++ {
		jobFn := func(ctx context.Context, progress func(int)) (any, error) {
			return nil, nil
		}
		_, _ = manager.Submit(ctx, "stats-test", jobFn)
	}

	// Wait for completion
	time.Sleep(200 * time.Millisecond)

	stats, err := manager.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}
	if stats.TotalJobs != 3 {
		t.Errorf("Expected 3 total jobs, got %d", stats.TotalJobs)
	}
	if stats.MaxWorkers != 3 {
		t.Errorf("Expected 3 max workers, got %d", stats.MaxWorkers)
	}
}

func TestManager_Shutdown(t *testing.T) {
	manager := jobmanager.NewManager(
		jobmanager.WithMaxWorkers(2),
	)

	ctx := context.Background()

	// Submit a long-running job
	started := make(chan struct{})
	jobFn := func(ctx context.Context, progress func(int)) (any, error) {
		close(started)
		time.Sleep(5 * time.Second)
		return nil, nil
	}

	_, _ = manager.Submit(ctx, "shutdown-test", jobFn)
	<-started

	// Shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	err := manager.Shutdown(shutdownCtx)
	// Shutdown may return context error if workers don't finish in time
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Unexpected shutdown error: %v", err)
	}

	// Verify manager is closed
	_, err = manager.Submit(ctx, "after-shutdown", jobFn)
	if err != jobmanager.ErrManagerClosed {
		t.Errorf("Expected ErrManagerClosed, got %v", err)
	}
}

func TestManager_WithCustomStore(t *testing.T) {
	customStore := jobmanager.NewMemoryStore()
	manager := jobmanager.NewManager(
		jobmanager.WithStore(customStore),
		jobmanager.WithMaxWorkers(1),
	)
	defer manager.Shutdown(context.Background())

	ctx := context.Background()

	jobFn := func(ctx context.Context, progress func(int)) (any, error) {
		return "custom-store-result", nil
	}

	jobID, err := manager.Submit(ctx, "custom-store-job", jobFn)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	// Wait for completion
	time.Sleep(100 * time.Millisecond)

	// Verify job is in custom store
	job, err := customStore.Get(ctx, jobID)
	if err != nil {
		t.Fatalf("Get from custom store failed: %v", err)
	}
	if job.Result != "custom-store-result" {
		t.Errorf("Unexpected result: %v", job.Result)
	}
}
