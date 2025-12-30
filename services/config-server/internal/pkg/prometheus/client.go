package prometheus

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// PrometheusClient manages interactions with Prometheus server
type PrometheusClient struct {
	baseURL string
	client  *http.Client
	timeout time.Duration
	logger  *slog.Logger

	// Retry configuration
	maxRetries      int
	retryDelay      time.Duration
	retryMultiplier float64
}

// PrometheusClientConfig holds configuration for PrometheusClient
type PrometheusClientConfig struct {
	BaseURL         string
	Timeout         time.Duration
	MaxRetries      int           // Default: 3
	RetryDelay      time.Duration // Default: 1s
	RetryMultiplier float64       // Default: 2.0 (exponential backoff)
}

// NewPrometheusClient creates a new PrometheusClient instance
func NewPrometheusClient(config PrometheusClientConfig, logger *slog.Logger) *PrometheusClient {
	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}
	if config.RetryMultiplier == 0 {
		config.RetryMultiplier = 2.0
	}

	return &PrometheusClient{
		baseURL: config.BaseURL,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		timeout:         config.Timeout,
		logger:          logger,
		maxRetries:      config.MaxRetries,
		retryDelay:      config.RetryDelay,
		retryMultiplier: config.RetryMultiplier,
	}
}

// Reload triggers Prometheus to reload its configuration
func (c *PrometheusClient) Reload(ctx context.Context) error {
	c.logger.Info("Triggering Prometheus reload")

	// Build reload URL
	reloadURL := c.baseURL + "/-/reload"

	// Execute with retry logic
	err := c.executeWithRetry(ctx, func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(ctx, "POST", reloadURL, nil)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrRequestFailed, err)
		}

		resp, err := c.client.Do(req)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
		}
		defer resp.Body.Close()

		// Read response body for error details
		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("%w: status=%d, body=%s", ErrReloadFailed, resp.StatusCode, string(body))
		}

		c.logger.Info("Prometheus reload triggered successfully")
		return nil
	})

	if err != nil {
		return err
	}

	// Verify Prometheus is healthy after reload
	if err := c.HealthCheck(ctx); err != nil {
		c.logger.Warn("Health check failed after reload", "error", err)
		return fmt.Errorf("%w: reload succeeded but health check failed: %v", ErrHealthCheckFailed, err)
	}

	return nil
}

// HealthCheck checks if Prometheus is healthy and ready
func (c *PrometheusClient) HealthCheck(ctx context.Context) error {
	// Check readiness endpoint
	readyURL := c.baseURL + "/-/ready"

	req, err := http.NewRequestWithContext(ctx, "GET", readyURL, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: status=%d, body=%s", ErrHealthCheckFailed, resp.StatusCode, string(body))
	}

	c.logger.Debug("Prometheus health check passed")
	return nil
}

// ValidateConfig validates Prometheus configuration without reloading
func (c *PrometheusClient) ValidateConfig(ctx context.Context) error {
	// Prometheus doesn't have a direct config validation endpoint
	// The best we can do is check if it's healthy
	// In practice, config validation happens during reload
	return c.HealthCheck(ctx)
}

// GetStatus retrieves Prometheus runtime status
func (c *PrometheusClient) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	statusURL := c.baseURL + "/api/v1/status/runtimeinfo"

	req, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status=%d, body=%s", ErrStatusFailed, resp.StatusCode, string(body))
	}

	// In a real implementation, we would parse the JSON response
	// For now, just return an empty map to indicate success
	c.logger.Debug("Retrieved Prometheus status")
	return make(map[string]interface{}), nil
}

// executeWithRetry executes a function with exponential backoff retry logic
func (c *PrometheusClient) executeWithRetry(ctx context.Context, fn func(context.Context) error) error {
	var lastErr error
	delay := c.retryDelay

	for attempt := 0; attempt < c.maxRetries; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the function
		err := fn(ctx)
		if err == nil {
			return nil
		}

		lastErr = err
		c.logger.Warn("Operation failed, will retry",
			"attempt", attempt+1,
			"max_retries", c.maxRetries,
			"error", err,
			"next_delay", delay)

		// Don't sleep after the last attempt
		if attempt < c.maxRetries-1 {
			select {
			case <-time.After(delay):
				// Calculate next delay with exponential backoff
				delay = time.Duration(float64(delay) * c.retryMultiplier)
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("%w: after %d attempts: %v", ErrRetryExhausted, c.maxRetries, lastErr)
}

// Ping checks basic connectivity to Prometheus
func (c *PrometheusClient) Ping(ctx context.Context) error {
	// Use the health check endpoint for ping
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/-/healthy", nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status=%d", ErrPingFailed, resp.StatusCode)
	}

	return nil
}

// IsReachable checks if Prometheus server is reachable
func (c *PrometheusClient) IsReachable(ctx context.Context) bool {
	err := c.Ping(ctx)
	return err == nil
}
