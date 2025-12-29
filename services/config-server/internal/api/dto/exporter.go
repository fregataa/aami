package dto

import (
	"github.com/fregataa/aami/config-server/internal/domain"
)

// CreateExporterRequest represents a request to create a new exporter
type CreateExporterRequest struct {
	TargetID       string              `json:"target_id" binding:"required,uuid"`
	Type           domain.ExporterType `json:"type" binding:"required"`
	Port           int                 `json:"port" binding:"required,min=1,max=65535"`
	Enabled        bool                `json:"enabled"`
	MetricsPath    string              `json:"metrics_path" binding:"omitempty,max=255"`
	ScrapeInterval string              `json:"scrape_interval" binding:"omitempty"`
	ScrapeTimeout  string              `json:"scrape_timeout" binding:"omitempty"`
	Config         domain.ExporterConfig `json:"config,omitempty"`
}

// UpdateExporterRequest represents a request to update an existing exporter
type UpdateExporterRequest struct {
	Type           *domain.ExporterType    `json:"type,omitempty"`
	Port           *int                    `json:"port,omitempty" binding:"omitempty,min=1,max=65535"`
	Enabled        *bool                   `json:"enabled,omitempty"`
	MetricsPath    *string                 `json:"metrics_path,omitempty" binding:"omitempty,max=255"`
	ScrapeInterval *string                 `json:"scrape_interval,omitempty"`
	ScrapeTimeout  *string                 `json:"scrape_timeout,omitempty"`
	Config         *domain.ExporterConfig  `json:"config,omitempty"`
}

// ExporterResponse represents an exporter in API responses
type ExporterResponse struct {
	ID             string              `json:"id"`
	TargetID       string              `json:"target_id"`
	Type           domain.ExporterType `json:"type"`
	Port           int                 `json:"port"`
	Enabled        bool                `json:"enabled"`
	MetricsPath    string              `json:"metrics_path"`
	ScrapeInterval string              `json:"scrape_interval"`
	ScrapeTimeout  string              `json:"scrape_timeout"`
	Config         domain.ExporterConfig `json:"config"`
	TimestampResponse
}

// ToExporterResponse converts a domain.Exporter to ExporterResponse
func ToExporterResponse(exporter *domain.Exporter) ExporterResponse {
	return ExporterResponse{
		ID:             exporter.ID,
		TargetID:       exporter.TargetID,
		Type:           exporter.Type,
		Port:           exporter.Port,
		Enabled:        exporter.Enabled,
		MetricsPath:    exporter.MetricsPath,
		ScrapeInterval: exporter.ScrapeInterval,
		ScrapeTimeout:  exporter.ScrapeTimeout,
		Config:         exporter.Config,
		TimestampResponse: TimestampResponse{
			CreatedAt: exporter.CreatedAt,
			UpdatedAt: exporter.UpdatedAt,
		},
	}
}

// ToExporterResponseList converts a slice of domain.Exporter to slice of ExporterResponse
func ToExporterResponseList(exporters []domain.Exporter) []ExporterResponse {
	responses := make([]ExporterResponse, len(exporters))
	for i, exporter := range exporters {
		responses[i] = ToExporterResponse(&exporter)
	}
	return responses
}
