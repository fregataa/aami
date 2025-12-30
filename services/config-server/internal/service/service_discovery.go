package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/fregataa/aami/config-server/internal/domain"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
	"github.com/fregataa/aami/config-server/internal/repository"
)

// ServiceDiscoveryService handles service discovery operations
type ServiceDiscoveryService struct {
	targetRepo repository.TargetRepository
}

// NewServiceDiscoveryService creates a new ServiceDiscoveryService instance
func NewServiceDiscoveryService(targetRepo repository.TargetRepository) *ServiceDiscoveryService {
	return &ServiceDiscoveryService{
		targetRepo: targetRepo,
	}
}

// GetPrometheusTargets returns all targets in Prometheus HTTP SD format
func (s *ServiceDiscoveryService) GetPrometheusTargets(ctx context.Context, filter *domain.ServiceDiscoveryFilter) ([]domain.PrometheusSDTarget, error) {
	// Get all targets with exporters (use a large limit for now)
	targets, _, err := s.targetRepo.List(ctx, 1, 10000)
	if err != nil {
		return nil, err
	}

	var sdTargets []domain.PrometheusSDTarget

	for _, target := range targets {
		// Apply status filter
		if filter.Status != nil && target.Status != *filter.Status {
			continue
		}

		// Apply group filter
		if filter.GroupID != nil && !target.HasGroup(*filter.GroupID) {
			continue
		}

		// Apply label filters
		if len(filter.Labels) > 0 {
			matchesAll := true
			for k, v := range filter.Labels {
				if targetVal, ok := target.Labels[k]; !ok || targetVal != v {
					matchesAll = false
					break
				}
			}
			if !matchesAll {
				continue
			}
		}

		// Process each exporter
		for _, exporter := range target.Exporters {
			// Apply enabled filter
			if filter.EnabledOnly && !exporter.Enabled {
				continue
			}

			// Apply exporter type filter
			if filter.ExporterType != nil && exporter.Type != *filter.ExporterType {
				continue
			}

			// Create SD target
			sdTarget := domain.NewPrometheusSDTarget(&target, &exporter)
			sdTargets = append(sdTargets, *sdTarget)
		}
	}

	return sdTargets, nil
}

// GetPrometheusTargetsForGroup returns targets for a specific group
func (s *ServiceDiscoveryService) GetPrometheusTargetsForGroup(ctx context.Context, groupID string, enabledOnly bool) ([]domain.PrometheusSDTarget, error) {
	filter := &domain.ServiceDiscoveryFilter{
		GroupID:     &groupID,
		EnabledOnly: enabledOnly,
	}
	return s.GetPrometheusTargets(ctx, filter)
}

// GetActivePrometheusTargets returns only active targets with enabled exporters
func (s *ServiceDiscoveryService) GetActivePrometheusTargets(ctx context.Context) ([]domain.PrometheusSDTarget, error) {
	status := domain.TargetStatusActive
	filter := &domain.ServiceDiscoveryFilter{
		Status:      &status,
		EnabledOnly: true,
	}
	return s.GetPrometheusTargets(ctx, filter)
}

// GenerateFileSD generates a Prometheus file SD JSON file
func (s *ServiceDiscoveryService) GenerateFileSD(ctx context.Context, outputPath string, filter *domain.ServiceDiscoveryFilter) error {
	// Get targets
	targets, err := s.GetPrometheusTargets(ctx, filter)
	if err != nil {
		return domainerrors.Wrap(err, "failed to get targets")
	}

	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return domainerrors.Wrap(err, "failed to create output directory")
	}

	// Convert to JSON
	data, err := json.MarshalIndent(targets, "", "  ")
	if err != nil {
		return domainerrors.Wrap(err, "failed to marshal targets")
	}

	// Write to temporary file first (atomic write)
	tempPath := outputPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return domainerrors.Wrap(err, "failed to write temp file")
	}

	// Rename to final file (atomic operation)
	if err := os.Rename(tempPath, outputPath); err != nil {
		os.Remove(tempPath) // Clean up temp file on error
		return domainerrors.Wrap(err, "failed to rename file")
	}

	return nil
}

// GenerateActiveFileSD generates a file SD JSON containing only active targets
func (s *ServiceDiscoveryService) GenerateActiveFileSD(ctx context.Context, outputPath string) error {
	status := domain.TargetStatusActive
	filter := &domain.ServiceDiscoveryFilter{
		Status:      &status,
		EnabledOnly: true,
	}
	return s.GenerateFileSD(ctx, outputPath, filter)
}

// GenerateGroupFileSD generates a file SD JSON for a specific group
func (s *ServiceDiscoveryService) GenerateGroupFileSD(ctx context.Context, groupID, outputPath string, enabledOnly bool) error {
	filter := &domain.ServiceDiscoveryFilter{
		GroupID:     &groupID,
		EnabledOnly: enabledOnly,
	}
	return s.GenerateFileSD(ctx, outputPath, filter)
}
