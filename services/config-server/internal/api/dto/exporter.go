package dto

import (
	"github.com/fregataa/aami/config-server/internal/action"
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

// ToAction converts CreateExporterRequest to action.CreateExporter
func (r *CreateExporterRequest) ToAction() action.CreateExporter {
	return action.CreateExporter{
		TargetID:       r.TargetID,
		Type:           r.Type,
		Port:           r.Port,
		Enabled:        r.Enabled,
		MetricsPath:    r.MetricsPath,
		ScrapeInterval: r.ScrapeInterval,
		ScrapeTimeout:  r.ScrapeTimeout,
		Config:         r.Config,
	}
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

// ToAction converts UpdateExporterRequest to action.UpdateExporter
func (r *UpdateExporterRequest) ToAction() action.UpdateExporter {
	return action.UpdateExporter{
		Type:           r.Type,
		Port:           r.Port,
		Enabled:        r.Enabled,
		MetricsPath:    r.MetricsPath,
		ScrapeInterval: r.ScrapeInterval,
		ScrapeTimeout:  r.ScrapeTimeout,
		Config:         r.Config,
	}
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

// ToExporterResponse converts action.ExporterResult to ExporterResponse
func ToExporterResponse(result action.ExporterResult) ExporterResponse {
	return ExporterResponse{
		ID:             result.ID,
		TargetID:       result.TargetID,
		Type:           result.Type,
		Port:           result.Port,
		Enabled:        result.Enabled,
		MetricsPath:    result.MetricsPath,
		ScrapeInterval: result.ScrapeInterval,
		ScrapeTimeout:  result.ScrapeTimeout,
		Config:         result.Config,
		TimestampResponse: TimestampResponse{
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
		},
	}
}

// ToExporterResponseList converts a slice of action.ExporterResult to slice of ExporterResponse
func ToExporterResponseList(results []action.ExporterResult) []ExporterResponse {
	responses := make([]ExporterResponse, len(results))
	for i, result := range results {
		responses[i] = ToExporterResponse(result)
	}
	return responses
}
