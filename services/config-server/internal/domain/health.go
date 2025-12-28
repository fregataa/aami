package domain

import "time"

// HealthStatus represents the overall health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentHealth represents the health status of a component
type ComponentHealth struct {
	Name      string                 `json:"name"`
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// HealthCheckResponse represents the complete health check response
type HealthCheckResponse struct {
	Status     HealthStatus               `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Version    string                     `json:"version"`
	Components map[string]ComponentHealth `json:"components"`
	Uptime     float64                    `json:"uptime_seconds"`
}

// IsHealthy returns true if all components are healthy
func (h *HealthCheckResponse) IsHealthy() bool {
	return h.Status == HealthStatusHealthy
}

// IsReady returns true if critical components are healthy
func (h *HealthCheckResponse) IsReady() bool {
	// Check critical components (database)
	for _, component := range h.Components {
		if component.Name == "database" {
			if component.Status == HealthStatusUnhealthy {
				return false
			}
		}
	}
	return true
}
