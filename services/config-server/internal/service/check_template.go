package service

import (
	"context"
	"errors"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/domain"
	"github.com/fregataa/aami/config-server/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CheckTemplateService handles business logic for check templates
type CheckTemplateService struct {
	templateRepo repository.CheckTemplateRepository
	instanceRepo repository.CheckInstanceRepository
}

// NewCheckTemplateService creates a new CheckTemplateService
func NewCheckTemplateService(
	templateRepo repository.CheckTemplateRepository,
	instanceRepo repository.CheckInstanceRepository,
) *CheckTemplateService {
	return &CheckTemplateService{
		templateRepo: templateRepo,
		instanceRepo: instanceRepo,
	}
}

// Create creates a new check template
func (s *CheckTemplateService) Create(ctx context.Context, req dto.CreateCheckTemplateRequest) (*domain.CheckTemplate, error) {
	// Validate required fields
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if template name already exists
	existing, err := s.templateRepo.GetByName(ctx, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyExists
	}

	template := &domain.CheckTemplate{
		ID:            uuid.New().String(),
		Name:          req.Name,
		CheckType:     req.CheckType,
		ScriptContent: req.ScriptContent,
		Language:      req.Language,
		DefaultConfig: req.DefaultConfig,
		Description:   req.Description,
		Version:       req.Version,
	}

	// Compute hash from script content
	template.UpdateHash()

	// Validate domain object
	if err := template.Validate(); err != nil {
		return nil, err
	}

	if err := s.templateRepo.Create(ctx, template); err != nil {
		return nil, err
	}

	return template, nil
}

// GetByID retrieves a check template by ID
func (s *CheckTemplateService) GetByID(ctx context.Context, id string) (*domain.CheckTemplate, error) {
	template, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return template, nil
}

// GetByName retrieves a check template by name
func (s *CheckTemplateService) GetByName(ctx context.Context, name string) (*domain.CheckTemplate, error) {
	template, err := s.templateRepo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return template, nil
}

// GetByCheckType retrieves all templates for a specific check type
func (s *CheckTemplateService) GetByCheckType(ctx context.Context, checkType string) ([]domain.CheckTemplate, error) {
	return s.templateRepo.GetByCheckType(ctx, checkType)
}

// Update updates an existing check template
func (s *CheckTemplateService) Update(ctx context.Context, id string, req dto.UpdateCheckTemplateRequest) (*domain.CheckTemplate, error) {
	template, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Update fields if provided
	if req.Description != nil {
		template.Description = *req.Description
	}
	if req.ScriptContent != nil {
		template.ScriptContent = *req.ScriptContent
		// Recalculate hash when script content changes
		template.UpdateHash()
	}
	if req.Language != nil {
		template.Language = *req.Language
	}
	if req.DefaultConfig != nil {
		template.DefaultConfig = req.DefaultConfig
	}
	if req.Version != nil {
		template.Version = *req.Version
	}

	// Validate updated template
	if err := template.Validate(); err != nil {
		return nil, err
	}

	if err := s.templateRepo.Update(ctx, template); err != nil {
		return nil, err
	}

	return template, nil
}

// Delete performs soft delete on a check template
func (s *CheckTemplateService) Delete(ctx context.Context, id string) error {
	// Check if template exists
	_, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	// Check if template is in use by any instances
	instances, err := s.instanceRepo.GetByTemplateID(ctx, id)
	if err != nil {
		return err
	}
	if len(instances) > 0 {
		return ErrInUse
	}

	return s.templateRepo.Delete(ctx, id)
}

// Purge permanently removes a check template (hard delete, admin operation)
func (s *CheckTemplateService) Purge(ctx context.Context, id string) error {
	return s.templateRepo.Purge(ctx, id)
}

// Restore restores a soft-deleted check template
func (s *CheckTemplateService) Restore(ctx context.Context, id string) error {
	return s.templateRepo.Restore(ctx, id)
}

// List retrieves a paginated list of check templates
func (s *CheckTemplateService) List(ctx context.Context, pagination dto.PaginationRequest) ([]domain.CheckTemplate, int, error) {
	pagination.Normalize()
	return s.templateRepo.List(ctx, pagination.Page, pagination.Limit)
}

// ListActive retrieves all active (non-deleted) templates
func (s *CheckTemplateService) ListActive(ctx context.Context) ([]domain.CheckTemplate, error) {
	return s.templateRepo.ListActive(ctx)
}

// VerifyHash checks if the stored hash matches the script content
func (s *CheckTemplateService) VerifyHash(ctx context.Context, id string) (bool, error) {
	template, err := s.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	return template.VerifyHash(), nil
}
