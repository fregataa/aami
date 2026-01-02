package ssh

import (
	"context"
	"strings"
	"time"
)

// RunWithRetry executes a command with retry logic
func (e *Executor) RunWithRetry(ctx context.Context, node Node, command string) Result {
	var lastResult Result

	for attempt := 0; attempt < e.config.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := e.calculateBackoff(attempt)
			select {
			case <-ctx.Done():
				return Result{Node: node.Name, Error: ctx.Err()}
			case <-time.After(backoff):
			}
		}

		lastResult = e.Run(ctx, node, command)
		if lastResult.Error == nil {
			return lastResult
		}

		// Don't retry certain types of errors
		if isNonRetryableError(lastResult.Error) {
			return lastResult
		}
	}

	return lastResult
}

// calculateBackoff calculates the backoff duration for a retry attempt
func (e *Executor) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: base * 2^attempt
	backoff := e.config.BackoffBase * time.Duration(1<<uint(attempt))
	if backoff > e.config.BackoffMax {
		backoff = e.config.BackoffMax
	}
	return backoff
}

// isNonRetryableError checks if an error should not be retried
func isNonRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Authentication errors should not be retried
	authErrors := []string{
		"unable to authenticate",
		"no supported methods remain",
		"permission denied",
		"authentication failed",
	}

	for _, authErr := range authErrors {
		if strings.Contains(strings.ToLower(errStr), authErr) {
			return true
		}
	}

	// Command execution errors (non-zero exit) should not be retried
	if strings.Contains(errStr, "Process exited with status") {
		return true
	}

	return false
}
