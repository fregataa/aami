package service

import (
	"context"
	"errors"

	"github.com/fregataa/aami/config-server/internal/api/dto"
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
func (s *ExporterService) Create(ctx context.Context, req dto.CreateExporterRequest) (*domain.Exporter, error) {
	// Validate target exists
	_, err := s.targetRepo.GetByID(ctx, req.TargetID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.NewValidationError("target_id", "target not found")
		}
		return nil, err
	}

	// Validate exporter type
	if !req.Type.IsValid() {
		return nil, domainerrors.NewValidationError("type", "invalid exporter type")
	}

	exporter := &domain.Exporter{
		ID:             uuid.New().String(),
		TargetID:       req.TargetID,
		Type:           req.Type,
		Port:           req.Port,
		Enabled:        true,
		MetricsPath:    req.MetricsPath,
		ScrapeInterval: req.ScrapeInterval,
		ScrapeTimeout:  req.ScrapeTimeout,
		Config:         req.Config,
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
		return nil, err
	}

	return exporter, nil
}

// GetByID retrieves an exporter by ID
func (s *ExporterService) GetByID(ctx context.Context, id string) (*domain.Exporter, error) {
	exporter, err := s.exporterRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return exporter, nil
}

// Update updates an existing exporter
func (s *ExporterService) Update(ctx context.Context, id string, req dto.UpdateExporterRequest) (*domain.Exporter, error) {
	exporter, err := s.exporterRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}

	if req.Type != nil {
		if !req.Type.IsValid() {
			return nil, domainerrors.NewValidationError("type", "invalid exporter type")
		}
		exporter.Type = *req.Type
	}
	if req.Port != nil {
		exporter.Port = *req.Port
	}
	if req.Enabled != nil {
		exporter.Enabled = *req.Enabled
	}
	if req.MetricsPath != nil {
		exporter.MetricsPath = *req.MetricsPath
	}
	if req.ScrapeInterval != nil {
		exporter.ScrapeInterval = *req.ScrapeInterval
	}
	if req.ScrapeTimeout != nil {
		exporter.ScrapeTimeout = *req.ScrapeTimeout
	}
	if req.Config != nil {
		exporter.Config = *req.Config
	}

	if err := s.exporterRepo.Update(ctx, exporter); err != nil {
		return nil, err
	}

	return exporter, nil
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
func (s *ExporterService) List(ctx context.Context, pagination dto.PaginationRequest) ([]domain.Exporter, int, error) {
	pagination.Normalize()
	return s.exporterRepo.List(ctx, pagination.Page, pagination.Limit)
}

// GetByTargetID retrieves all exporters for a target
func (s *ExporterService) GetByTargetID(ctx context.Context, targetID string) ([]domain.Exporter, error) {
	return s.exporterRepo.GetByTargetID(ctx, targetID)
}

// GetByType retrieves all exporters of a specific type
func (s *ExporterService) GetByType(ctx context.Context, exporterType domain.ExporterType) ([]domain.Exporter, error) {
	return s.exporterRepo.GetByType(ctx, exporterType)
}
