package dto

import "github.com/fregataa/aami/config-server/internal/domain"

// PrometheusSDTargetResponse represents a Prometheus SD target in API response
type PrometheusSDTargetResponse struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

// ToPrometheusSDTargetResponse converts domain.PrometheusSDTarget to DTO
func ToPrometheusSDTargetResponse(target *domain.PrometheusSDTarget) PrometheusSDTargetResponse {
	return PrometheusSDTargetResponse{
		Targets: target.Targets,
		Labels:  target.Labels,
	}
}

// ToPrometheusSDTargetResponseList converts a list of domain.PrometheusSDTarget to DTO list
func ToPrometheusSDTargetResponseList(targets []domain.PrometheusSDTarget) []PrometheusSDTargetResponse {
	responses := make([]PrometheusSDTargetResponse, len(targets))
	for i, target := range targets {
		responses[i] = ToPrometheusSDTargetResponse(&target)
	}
	return responses
}

// GroupTargetsRequest represents query parameters for group-specific target requests
type GroupTargetsRequest struct {
	EnabledOnly bool `form:"enabled_only"`
}

// ServiceDiscoveryFilterRequest represents filter parameters for service discovery
type ServiceDiscoveryFilterRequest struct {
	Status       string            `form:"status"`
	ExporterType string            `form:"exporter_type"`
	GroupID      string            `form:"group_id"`
	EnabledOnly  bool              `form:"enabled_only"`
	Labels       map[string]string `form:"labels"`
}

// ToDomainFilter converts DTO to domain.ServiceDiscoveryFilter
func (r *ServiceDiscoveryFilterRequest) ToDomainFilter() *domain.ServiceDiscoveryFilter {
	filter := &domain.ServiceDiscoveryFilter{
		EnabledOnly: r.EnabledOnly,
		Labels:      r.Labels,
	}

	if r.Status != "" {
		status := domain.TargetStatus(r.Status)
		filter.Status = &status
	}

	if r.ExporterType != "" {
		exporterType := domain.ExporterType(r.ExporterType)
		filter.ExporterType = &exporterType
	}

	if r.GroupID != "" {
		filter.GroupID = &r.GroupID
	}

	return filter
}
