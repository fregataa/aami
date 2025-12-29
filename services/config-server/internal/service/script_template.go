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

// ScriptTemplateService handles business logic for script templates
type ScriptTemplateService struct {
	templateRepo repository.ScriptTemplateRepository
	policyRepo   repository.ScriptPolicyRepository
}

// NewScriptTemplateService creates a new ScriptTemplateService
func NewScriptTemplateService(
	templateRepo repository.ScriptTemplateRepository,
	policyRepo repository.ScriptPolicyRepository,
) *ScriptTemplateService {
	return &ScriptTemplateService{
		templateRepo: templateRepo,
		policyRepo:   policyRepo,
	}
}

// Create creates a new script template
func (s *ScriptTemplateService) Create(ctx context.Context, act action.CreateScriptTemplate) (action.ScriptTemplateResult, error) {
	// Check if template name already exists
	existing, err := s.templateRepo.GetByName(ctx, act.Name)
	if err != nil && !errors.Is(err, domainerrors.ErrNotFound) {
		return action.ScriptTemplateResult{}, err
	}
	if existing != nil {
		return action.ScriptTemplateResult{}, domainerrors.ErrAlreadyExists
	}

	template := &domain.ScriptTemplate{
		ID:            uuid.New().String(),
		Name:          act.Name,
		ScriptType:    act.ScriptType,
		ScriptContent: act.ScriptContent,
		Language:      act.Language,
		DefaultConfig: act.DefaultConfig,
		Description:   act.Description,
		Version:       act.Version,
	}

	// Compute hash from script content
	template.UpdateHash()

	// Validate domain object
	if err := template.Validate(); err != nil {
		return action.ScriptTemplateResult{}, err
	}

	if err := s.templateRepo.Create(ctx, template); err != nil {
		return action.ScriptTemplateResult{}, err
	}

	return action.NewScriptTemplateResult(template), nil
}

// GetByID retrieves a script template by ID
func (s *ScriptTemplateService) GetByID(ctx context.Context, id string) (action.ScriptTemplateResult, error) {
	template, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.ScriptTemplateResult{}, domainerrors.ErrNotFound
		}
		return action.ScriptTemplateResult{}, err
	}
	return action.NewScriptTemplateResult(template), nil
}

// GetByName retrieves a script template by name
func (s *ScriptTemplateService) GetByName(ctx context.Context, name string) (action.ScriptTemplateResult, error) {
	template, err := s.templateRepo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.ScriptTemplateResult{}, domainerrors.ErrNotFound
		}
		return action.ScriptTemplateResult{}, err
	}
	return action.NewScriptTemplateResult(template), nil
}

// GetByScriptType retrieves all templates for a specific script type
func (s *ScriptTemplateService) GetByScriptType(ctx context.Context, scriptType string) ([]action.ScriptTemplateResult, error) {
	templates, err := s.templateRepo.GetByScriptType(ctx, scriptType)
	if err != nil {
		return nil, err
	}
	return action.NewScriptTemplateResultList(templates), nil
}

// Update updates an existing script template
func (s *ScriptTemplateService) Update(ctx context.Context, id string, act action.UpdateScriptTemplate) (action.ScriptTemplateResult, error) {
	template, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.ScriptTemplateResult{}, domainerrors.ErrNotFound
		}
		return action.ScriptTemplateResult{}, err
	}

	// Update fields if provided
	if act.Name != nil {
		template.Name = *act.Name
	}
	if act.ScriptType != nil {
		template.ScriptType = *act.ScriptType
	}
	if act.Description != nil {
		template.Description = *act.Description
	}
	if act.ScriptContent != nil {
		template.ScriptContent = *act.ScriptContent
		// Recalculate hash when script content changes
		template.UpdateHash()
	}
	if act.Language != nil {
		template.Language = *act.Language
	}
	if act.DefaultConfig != nil {
		template.DefaultConfig = act.DefaultConfig
	}
	if act.Version != nil {
		template.Version = *act.Version
	}

	// Validate updated template
	if err := template.Validate(); err != nil {
		return action.ScriptTemplateResult{}, err
	}

	if err := s.templateRepo.Update(ctx, template); err != nil {
		return action.ScriptTemplateResult{}, err
	}

	return action.NewScriptTemplateResult(template), nil
}

// Delete performs soft delete on a script template
func (s *ScriptTemplateService) Delete(ctx context.Context, id string) error {
	// Check if template exists
	_, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return domainerrors.ErrNotFound
		}
		return err
	}

	// Check if template is in use by any policies
	policies, err := s.policyRepo.GetByTemplateID(ctx, id)
	if err != nil {
		return err
	}
	if len(policies) > 0 {
		return domainerrors.ErrInUse
	}

	return s.templateRepo.Delete(ctx, id)
}

// Purge permanently removes a script template (hard delete, admin operation)
func (s *ScriptTemplateService) Purge(ctx context.Context, id string) error {
	return s.templateRepo.Purge(ctx, id)
}

// Restore restores a soft-deleted script template
func (s *ScriptTemplateService) Restore(ctx context.Context, id string) error {
	return s.templateRepo.Restore(ctx, id)
}

// List retrieves a paginated list of script templates
func (s *ScriptTemplateService) List(ctx context.Context, pagination action.Pagination) (action.ListResult[action.ScriptTemplateResult], error) {
	templates, total, err := s.templateRepo.List(ctx, pagination.Page, pagination.Limit)
	if err != nil {
		return action.ListResult[action.ScriptTemplateResult]{}, err
	}

	results := action.NewScriptTemplateResultList(templates)
	return action.NewListResult(results, pagination, total), nil
}

// ListActive retrieves all active (non-deleted) templates
func (s *ScriptTemplateService) ListActive(ctx context.Context) ([]action.ScriptTemplateResult, error) {
	templates, err := s.templateRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	return action.NewScriptTemplateResultList(templates), nil
}

// VerifyHash checks if the stored hash matches the script content
func (s *ScriptTemplateService) VerifyHash(ctx context.Context, id string) (bool, error) {
	template, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return false, domainerrors.ErrNotFound
		}
		return false, err
	}
	return template.VerifyHash(), nil
}
