package alertmanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// Custom errors for Alertmanager client operations
var (
	ErrConnectionFailed  = errors.New("failed to connect to Alertmanager")
	ErrHealthCheckFailed = errors.New("Alertmanager health check failed")
	ErrFetchAlertsFailed = errors.New("failed to fetch alerts from Alertmanager")
)

// AlertmanagerClient manages interactions with Alertmanager server
type AlertmanagerClient struct {
	baseURL string
	client  *http.Client
	timeout time.Duration
	logger  *slog.Logger
}

// AlertmanagerClientConfig holds configuration for AlertmanagerClient
type AlertmanagerClientConfig struct {
	BaseURL string
	Timeout time.Duration
}

// AlertStatus represents the status of an alert
type AlertStatus struct {
	State       string   `json:"state"`
	SilencedBy  []string `json:"silencedBy"`
	InhibitedBy []string `json:"inhibitedBy"`
}

// Alert represents an alert from Alertmanager API v2
type Alert struct {
	Fingerprint  string            `json:"fingerprint"`
	Status       AlertStatus       `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	UpdatedAt    time.Time         `json:"updatedAt"`
}

// NewAlertmanagerClient creates a new AlertmanagerClient instance
func NewAlertmanagerClient(config AlertmanagerClientConfig, logger *slog.Logger) *AlertmanagerClient {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &AlertmanagerClient{
		baseURL: config.BaseURL,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		timeout: config.Timeout,
		logger:  logger,
	}
}

// GetAlerts retrieves all alerts from Alertmanager
func (c *AlertmanagerClient) GetAlerts(ctx context.Context) ([]Alert, error) {
	alertsURL := c.baseURL + "/api/v2/alerts"

	req, err := http.NewRequestWithContext(ctx, "GET", alertsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create alerts request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status=%d, body=%s", ErrFetchAlertsFailed, resp.StatusCode, string(body))
	}

	var alerts []Alert
	if err := json.NewDecoder(resp.Body).Decode(&alerts); err != nil {
		return nil, fmt.Errorf("failed to decode alerts response: %w", err)
	}

	c.logger.Debug("Retrieved alerts from Alertmanager", "count", len(alerts))
	return alerts, nil
}

// GetActiveAlerts retrieves only firing (active) alerts from Alertmanager
func (c *AlertmanagerClient) GetActiveAlerts(ctx context.Context) ([]Alert, error) {
	// Use filter parameter to get only active alerts
	alertsURL := c.baseURL + "/api/v2/alerts?active=true&silenced=false&inhibited=false"

	req, err := http.NewRequestWithContext(ctx, "GET", alertsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create active alerts request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status=%d, body=%s", ErrFetchAlertsFailed, resp.StatusCode, string(body))
	}

	var alerts []Alert
	if err := json.NewDecoder(resp.Body).Decode(&alerts); err != nil {
		return nil, fmt.Errorf("failed to decode active alerts response: %w", err)
	}

	c.logger.Debug("Retrieved active alerts from Alertmanager", "count", len(alerts))
	return alerts, nil
}

// HealthCheck checks if Alertmanager is healthy
func (c *AlertmanagerClient) HealthCheck(ctx context.Context) error {
	healthURL := c.baseURL + "/-/healthy"

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: status=%d, body=%s", ErrHealthCheckFailed, resp.StatusCode, string(body))
	}

	c.logger.Debug("Alertmanager health check passed")
	return nil
}

// IsReachable checks if Alertmanager server is reachable
func (c *AlertmanagerClient) IsReachable(ctx context.Context) bool {
	err := c.HealthCheck(ctx)
	return err == nil
}

// GetStatus retrieves Alertmanager status information
func (c *AlertmanagerClient) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	statusURL := c.baseURL + "/api/v2/status"

	req, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create status request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get status: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var status map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode status response: %w", err)
	}

	c.logger.Debug("Retrieved Alertmanager status")
	return status, nil
}
