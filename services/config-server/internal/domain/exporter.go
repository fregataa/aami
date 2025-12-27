package domain

import (
	"fmt"
	"time"
)

// ExporterType represents the type of metrics exporter
type ExporterType string

const (
	// ExporterTypeNodeExporter for system metrics (CPU, memory, disk, network)
	ExporterTypeNodeExporter ExporterType = "node_exporter"

	// ExporterTypeDCGMExporter for NVIDIA GPU metrics
	ExporterTypeDCGMExporter ExporterType = "dcgm_exporter"

	// ExporterTypeCustom for custom exporters
	ExporterTypeCustom ExporterType = "custom"
)

// IsValid checks if the exporter type is one of the allowed values
func (t ExporterType) IsValid() bool {
	switch t {
	case ExporterTypeNodeExporter, ExporterTypeDCGMExporter, ExporterTypeCustom:
		return true
	default:
		return false
	}
}

// Exporter represents a metrics exporter configuration for a target
type Exporter struct {
	ID             string                 `json:"id"`
	TargetID       string                 `json:"target_id"`
	Type           ExporterType           `json:"type"`
	Port           int                    `json:"port"`
	Enabled        bool                   `json:"enabled"`
	MetricsPath    string                 `json:"metrics_path"`
	ScrapeInterval string                 `json:"scrape_interval"`
	ScrapeTimeout  string                 `json:"scrape_timeout"`
	Config         map[string]interface{} `json:"config"`
	DeletedAt      *time.Time             `json:"deleted_at,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// GetEndpoint returns the full metrics endpoint URL for this exporter
func (e *Exporter) GetEndpoint(target Target) string {
	return fmt.Sprintf("http://%s:%d%s", target.IPAddress, e.Port, e.MetricsPath)
}

// Validate performs validation on exporter fields
func (e *Exporter) Validate() error {
	if !e.Type.IsValid() {
		return fmt.Errorf("invalid exporter type: %s", e.Type)
	}
	if e.Port <= 0 || e.Port > 65535 {
		return fmt.Errorf("invalid port: %d", e.Port)
	}
	if e.MetricsPath == "" {
		e.MetricsPath = "/metrics"
	}
	if e.ScrapeInterval == "" {
		e.ScrapeInterval = "15s"
	}
	if e.ScrapeTimeout == "" {
		e.ScrapeTimeout = "10s"
	}
	return nil
}

// DefaultPortForType returns the default port for a given exporter type
func DefaultPortForType(exporterType ExporterType) int {
	switch exporterType {
	case ExporterTypeNodeExporter:
		return 9100
	case ExporterTypeDCGMExporter:
		return 9400
	case ExporterTypeCustom:
		return 9090
	default:
		return 9090
	}
}
