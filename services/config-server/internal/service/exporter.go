package service

import (
	"context"
	"errors"

	"github.com/fregataa/aami/config-server/internal/action"
	"github.com/fregataa/aami/config-server/internal/domain"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
	"github.com/fregataa/aami/config-server/internal/repository"
	"github.com/google/uuid"
)

// ExporterService handles business logic for exporters
type ExporterService struct {
	exporterRepo repository.ExporterRepository
	targetRepo   repository.TargetRepository
}

// NewExporterService creates a new ExporterService
func NewExporterService(
	exporterRepo repository.ExporterRepository,
	targetRepo repository.TargetRepository,
) *ExporterService {
	return &ExporterService{
		exporterRepo: exporterRepo,
		targetRepo:   targetRepo,
	}
}

// Create creates a new exporter
func (s *ExporterService) Create(ctx context.Context, act action.CreateExporter) (action.ExporterResult, error) {
	// Validate target exists
	_, err := s.targetRepo.GetByID(ctx, act.TargetID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.ExporterResult{}, domainerrors.NewValidationError("target_id", "target not found")
		}
		return action.ExporterResult{}, err
	}

	// Validate exporter type
	if !act.Type.IsValid() {
		return action.ExporterResult{}, domainerrors.NewValidationError("type", "invalid exporter type")
	}

	exporter := &domain.Exporter{
		ID:             uuid.New().String(),
		TargetID:       act.TargetID,
		Type:           act.Type,
		Port:           act.Port,
		Enabled:        true,
		MetricsPath:    act.MetricsPath,
		ScrapeInterval: act.ScrapeInterval,
		ScrapeTimeout:  act.ScrapeTimeout,
		Config:         act.Config,
	}

	// Apply defaults
	if exporter.MetricsPath == "" {
		exporter.MetricsPath = "/metrics"
	}
	if exporter.ScrapeInterval == "" {
		exporter.ScrapeInterval = "15s"
	}
	if exporter.ScrapeTimeout == "" {
		exporter.ScrapeTimeout = "10s"
	}

	if err := s.exporterRepo.Create(ctx, exporter); err != nil {
		return action.ExporterResult{}, err
	}

	return action.NewExporterResult(exporter), nil
}

// GetByID retrieves an exporter by ID
func (s *ExporterService) GetByID(ctx context.Context, id string) (action.ExporterResult, error) {
	exporter, err := s.exporterRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.ExporterResult{}, domainerrors.ErrNotFound
		}
		return action.ExporterResult{}, err
	}
	return action.NewExporterResult(exporter), nil
}

// Update updates an existing exporter
func (s *ExporterService) Update(ctx context.Context, id string, act action.UpdateExporter) (action.ExporterResult, error) {
	exporter, err := s.exporterRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.ExporterResult{}, domainerrors.ErrNotFound
		}
		return action.ExporterResult{}, err
	}

	if act.Type != nil {
		if !act.Type.IsValid() {
			return action.ExporterResult{}, domainerrors.NewValidationError("type", "invalid exporter type")
		}
		exporter.Type = *act.Type
	}
	if act.Port != nil {
		exporter.Port = *act.Port
	}
	if act.Enabled != nil {
		exporter.Enabled = *act.Enabled
	}
	if act.MetricsPath != nil {
		exporter.MetricsPath = *act.MetricsPath
	}
	if act.ScrapeInterval != nil {
		exporter.ScrapeInterval = *act.ScrapeInterval
	}
	if act.ScrapeTimeout != nil {
		exporter.ScrapeTimeout = *act.ScrapeTimeout
	}
	if act.Config != nil {
		exporter.Config = *act.Config
	}

	if err := s.exporterRepo.Update(ctx, exporter); err != nil {
		return action.ExporterResult{}, err
	}

	return action.NewExporterResult(exporter), nil
}

// Delete performs soft delete on an exporter
func (s *ExporterService) Delete(ctx context.Context, id string) error {
	_, err := s.exporterRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return domainerrors.ErrNotFound
		}
		return err
	}

	return s.exporterRepo.Delete(ctx, id)
}

// Purge permanently removes an exporter (hard delete, admin operation)
func (s *ExporterService) Purge(ctx context.Context, id string) error {
	return s.exporterRepo.Purge(ctx, id)
}

// Restore restores a soft-deleted exporter
func (s *ExporterService) Restore(ctx context.Context, id string) error {
	return s.exporterRepo.Restore(ctx, id)
}

// List retrieves a paginated list of exporters
func (s *ExporterService) List(ctx context.Context, pagination action.Pagination) (action.ListResult[action.ExporterResult], error) {
	exporters, total, err := s.exporterRepo.List(ctx, pagination.Page, pagination.Limit)
	if err != nil {
		return action.ListResult[action.ExporterResult]{}, err
	}

	results := action.NewExporterResultList(exporters)
	return action.NewListResult(results, pagination, total), nil
}

// GetByTargetID retrieves all exporters for a target
func (s *ExporterService) GetByTargetID(ctx context.Context, targetID string) ([]action.ExporterResult, error) {
	exporters, err := s.exporterRepo.GetByTargetID(ctx, targetID)
	if err != nil {
		return nil, err
	}
	return action.NewExporterResultList(exporters), nil
}

// GetByType retrieves all exporters of a specific type
func (s *ExporterService) GetByType(ctx context.Context, exporterType domain.ExporterType) ([]action.ExporterResult, error) {
	exporters, err := s.exporterRepo.GetByType(ctx, exporterType)
	if err != nil {
		return nil, err
	}
	return action.NewExporterResultList(exporters), nil
}
