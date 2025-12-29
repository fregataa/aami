package action

import (
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
)

// ============================================================================
// Actions (Input)
// ============================================================================

// CreateExporter represents the action to create an exporter
type CreateExporter struct {
	TargetID       string
	Type           domain.ExporterType
	Port           int
	Enabled        bool
	MetricsPath    string
	ScrapeInterval string
	ScrapeTimeout  string
	Config         domain.ExporterConfig
}

// UpdateExporter represents the action to update an exporter
// nil fields mean "do not update"
type UpdateExporter struct {
	Enabled        *bool
	MetricsPath    *string
	ScrapeInterval *string
	ScrapeTimeout  *string
	Config         *domain.ExporterConfig
}

// ============================================================================
// Action Results (Output)
// ============================================================================

// ExporterResult represents the result of exporter operations
type ExporterResult struct {
	ID             string
	TargetID       string
	Type           domain.ExporterType
	Port           int
	Enabled        bool
	MetricsPath    string
	ScrapeInterval string
	ScrapeTimeout  string
	Config         domain.ExporterConfig
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// FromDomain converts domain.Exporter to ExporterResult
func (r *ExporterResult) FromDomain(e *domain.Exporter) {
	r.ID = e.ID
	r.TargetID = e.TargetID
	r.Type = e.Type
	r.Port = e.Port
	r.Enabled = e.Enabled
	r.MetricsPath = e.MetricsPath
	r.ScrapeInterval = e.ScrapeInterval
	r.ScrapeTimeout = e.ScrapeTimeout
	r.Config = e.Config
	r.CreatedAt = e.CreatedAt
	r.UpdatedAt = e.UpdatedAt
}

// NewExporterResult creates ExporterResult from domain.Exporter
func NewExporterResult(e *domain.Exporter) ExporterResult {
	var result ExporterResult
	result.FromDomain(e)
	return result
}

// NewExporterResultList creates []ExporterResult from []domain.Exporter
func NewExporterResultList(exporters []domain.Exporter) []ExporterResult {
	results := make([]ExporterResult, len(exporters))
	for i, e := range exporters {
		results[i] = NewExporterResult(&e)
	}
	return results
}
