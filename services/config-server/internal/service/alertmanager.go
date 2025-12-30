package service

import (
	"context"
	"time"

	"github.com/fregataa/aami/config-server/internal/pkg/alertmanager"
)

// ActiveAlert represents an active alert for API response
type ActiveAlert struct {
	Fingerprint  string            `json:"fingerprint"`
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"starts_at"`
	GeneratorURL string            `json:"generator_url"`
}

// ActiveAlertsResult contains the list of active alerts
type ActiveAlertsResult struct {
	Alerts []ActiveAlert `json:"alerts"`
	Total  int           `json:"total"`
}

// AlertmanagerService provides operations for Alertmanager
type AlertmanagerService struct {
	client *alertmanager.AlertmanagerClient
}

// NewAlertmanagerService creates a new AlertmanagerService
func NewAlertmanagerService(client *alertmanager.AlertmanagerClient) *AlertmanagerService {
	return &AlertmanagerService{
		client: client,
	}
}

// GetActiveAlerts retrieves all currently firing alerts
func (s *AlertmanagerService) GetActiveAlerts(ctx context.Context) (*ActiveAlertsResult, error) {
	if s.client == nil {
		// Return empty result if Alertmanager is not configured
		return &ActiveAlertsResult{
			Alerts: []ActiveAlert{},
			Total:  0,
		}, nil
	}

	alerts, err := s.client.GetActiveAlerts(ctx)
	if err != nil {
		return nil, err
	}

	result := &ActiveAlertsResult{
		Alerts: make([]ActiveAlert, 0, len(alerts)),
		Total:  len(alerts),
	}

	for _, alert := range alerts {
		result.Alerts = append(result.Alerts, ActiveAlert{
			Fingerprint:  alert.Fingerprint,
			Status:       alert.Status.State,
			Labels:       alert.Labels,
			Annotations:  alert.Annotations,
			StartsAt:     alert.StartsAt,
			GeneratorURL: alert.GeneratorURL,
		})
	}

	return result, nil
}

// HealthCheck checks if Alertmanager is reachable
func (s *AlertmanagerService) HealthCheck(ctx context.Context) error {
	if s.client == nil {
		return nil
	}
	return s.client.HealthCheck(ctx)
}

// IsConfigured returns true if Alertmanager client is configured
func (s *AlertmanagerService) IsConfigured() bool {
	return s.client != nil
}
