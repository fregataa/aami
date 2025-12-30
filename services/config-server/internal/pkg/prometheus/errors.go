package prometheus

import "errors"

// Client errors
var (
	// ErrConnectionFailed is returned when connection to Prometheus fails
	ErrConnectionFailed = errors.New("failed to connect to Prometheus")

	// ErrReloadFailed is returned when Prometheus reload fails
	ErrReloadFailed = errors.New("failed to reload Prometheus")

	// ErrHealthCheckFailed is returned when Prometheus health check fails
	ErrHealthCheckFailed = errors.New("Prometheus health check failed")

	// ErrConfigInvalid is returned when Prometheus configuration is invalid
	ErrConfigInvalid = errors.New("Prometheus configuration is invalid")

	// ErrRequestFailed is returned when creating HTTP request fails
	ErrRequestFailed = errors.New("failed to create request")

	// ErrRetryExhausted is returned when all retry attempts are exhausted
	ErrRetryExhausted = errors.New("operation failed after retries")

	// ErrStatusFailed is returned when getting status fails
	ErrStatusFailed = errors.New("failed to get status")

	// ErrPingFailed is returned when ping fails
	ErrPingFailed = errors.New("ping failed")
)

// File manager errors
var (
	// ErrDirectoryNotFound is returned when rules directory is not found
	ErrDirectoryNotFound = errors.New("rules directory not found")

	// ErrPermissionDenied is returned when there are insufficient permissions
	ErrPermissionDenied = errors.New("insufficient permissions")

	// ErrValidationFailed is returned when rule validation fails
	ErrValidationFailed = errors.New("rule validation failed")

	// ErrAtomicWriteFailed is returned when atomic write operation fails
	ErrAtomicWriteFailed = errors.New("atomic write failed")

	// ErrBackupFailed is returned when backup operation fails
	ErrBackupFailed = errors.New("backup operation failed")

	// ErrRestoreFailed is returned when restore operation fails
	ErrRestoreFailed = errors.New("restore operation failed")

	// ErrDeleteFailed is returned when delete operation fails
	ErrDeleteFailed = errors.New("failed to delete rule file")

	// ErrListFailed is returned when listing files fails
	ErrListFailed = errors.New("failed to list files")
)
