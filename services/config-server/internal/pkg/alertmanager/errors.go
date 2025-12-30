package alertmanager

import "errors"

// Sentinel errors for Alertmanager client operations
var (
	// ErrConnectionFailed is returned when connection to Alertmanager fails
	ErrConnectionFailed = errors.New("failed to connect to Alertmanager")

	// ErrHealthCheckFailed is returned when Alertmanager health check fails
	ErrHealthCheckFailed = errors.New("Alertmanager health check failed")

	// ErrFetchAlertsFailed is returned when fetching alerts fails
	ErrFetchAlertsFailed = errors.New("failed to fetch alerts from Alertmanager")

	// ErrRequestFailed is returned when creating HTTP request fails
	ErrRequestFailed = errors.New("failed to create request")

	// ErrResponseDecodeFailed is returned when decoding response fails
	ErrResponseDecodeFailed = errors.New("failed to decode response")

	// ErrStatusFailed is returned when getting status fails
	ErrStatusFailed = errors.New("failed to get status")
)
