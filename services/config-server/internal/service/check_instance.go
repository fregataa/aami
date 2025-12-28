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

// CheckInstanceService handles business logic for check instances
type CheckInstanceService struct {
	instanceRepo  repository.CheckInstanceRepository
	templateRepo  repository.CheckTemplateRepository
	namespaceRepo repository.NamespaceRepository
	groupRepo     repository.GroupRepository
	targetRepo    repository.TargetRepository
}

// NewCheckInstanceService creates a new CheckInstanceService
func NewCheckInstanceService(
	instanceRepo repository.CheckInstanceRepository,
	templateRepo repository.CheckTemplateRepository,
	namespaceRepo repository.NamespaceRepository,
	groupRepo repository.GroupRepository,
	targetRepo repository.TargetRepository,
) *CheckInstanceService {
	return &CheckInstanceService{
		instanceRepo:  instanceRepo,
		templateRepo:  templateRepo,
		namespaceRepo: namespaceRepo,
		groupRepo:     groupRepo,
		targetRepo:    targetRepo,
	}
}

// Create creates a new check instance
func (s *CheckInstanceService) Create(ctx context.Context, req dto.CreateCheckInstanceRequest) (*domain.CheckInstance, error) {
	// Validate required fields
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Verify namespace exists if namespace-level or group-level
	if req.NamespaceID != nil {
		if _, err := s.namespaceRepo.GetByID(ctx, *req.NamespaceID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrForeignKeyViolation
			}
			return nil, err
		}
	}

	// Verify group exists if group-level
	if req.GroupID != nil {
		if _, err := s.groupRepo.GetByID(ctx, *req.GroupID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrForeignKeyViolation
			}
			return nil, err
		}
	}

	var instance *domain.CheckInstance

	// Two creation modes: from template or direct
	if req.TemplateID != nil {
		// Option 1: Create from template
		template, err := s.templateRepo.GetByID(ctx, *req.TemplateID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrForeignKeyViolation
			}
			return nil, err
		}

		// Use domain constructor to deep copy template fields
		instance = domain.NewCheckInstanceFromTemplate(
			template,
			req.Scope,
			req.NamespaceID,
			req.GroupID,
			req.Config,
		)
	} else {
		// Option 2: Direct creation (all fields provided in request)
		instance = &domain.CheckInstance{
			Name:          *req.Name,
			CheckType:     *req.CheckType,
			ScriptContent: *req.ScriptContent,
			Language:      *req.Language,
			DefaultConfig: *req.DefaultConfig,
			Description:   *req.Description,
			Version:       *req.Version,
			Scope:         req.Scope,
			NamespaceID:   req.NamespaceID,
			GroupID:       req.GroupID,
			Config:        req.Config,
		}
	}

	// Set ID and override priority/is_active if provided
	instance.ID = uuid.New().String()
	if req.Priority != 0 {
		instance.Priority = req.Priority
	}
	instance.IsActive = req.IsActive

	// Validate domain object (checks scope consistency)
	if err := instance.Validate(); err != nil {
		return nil, err
	}

	if err := s.instanceRepo.Create(ctx, instance); err != nil {
		return nil, err
	}

	return instance, nil
}

// GetByID retrieves a check instance by ID
func (s *CheckInstanceService) GetByID(ctx context.Context, id string) (*domain.CheckInstance, error) {
	instance, err := s.instanceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return instance, nil
}

// GetByTemplateID retrieves all instances for a specific template
func (s *CheckInstanceService) GetByTemplateID(ctx context.Context, templateID string) ([]domain.CheckInstance, error) {
	return s.instanceRepo.GetByTemplateID(ctx, templateID)
}

// GetGlobalInstances retrieves all global-scope instances
func (s *CheckInstanceService) GetGlobalInstances(ctx context.Context) ([]domain.CheckInstance, error) {
	return s.instanceRepo.GetGlobalInstances(ctx)
}

// GetByNamespaceID retrieves all namespace-level instances for a specific namespace
func (s *CheckInstanceService) GetByNamespaceID(ctx context.Context, namespaceID string) ([]domain.CheckInstance, error) {
	return s.instanceRepo.GetByNamespaceID(ctx, namespaceID)
}

// GetByGroupID retrieves all group-level instances for a specific group
func (s *CheckInstanceService) GetByGroupID(ctx context.Context, groupID string) ([]domain.CheckInstance, error) {
	return s.instanceRepo.GetByGroupID(ctx, groupID)
}

// GetEffectiveInstance finds the most specific active instance for a template
// Priority: Group > Namespace > Global
func (s *CheckInstanceService) GetEffectiveInstance(ctx context.Context, templateID, namespaceID, groupID string) (*domain.CheckInstance, error) {
	instance, err := s.instanceRepo.GetEffectiveInstance(ctx, templateID, namespaceID, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return instance, nil
}

// GetEffectiveChecksByTargetID retrieves all effective checks for a specific target
// This is the main method used by nodes to get their check configurations
func (s *CheckInstanceService) GetEffectiveChecksByTargetID(ctx context.Context, targetID string) ([]domain.EffectiveCheck, error) {
	// Get effective check instances from repository (handles priority resolution)
	result, err := s.targetRepo.GetEffectiveCheckInstances(ctx, targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Combine namespace and group instances
	allInstances := append(result.NamespaceInstances, result.GroupInstances...)

	// Convert to EffectiveCheck with merged configs
	effectiveChecks := make([]domain.EffectiveCheck, len(allInstances))
	for i, instance := range allInstances {
		// Instance contains all template fields, just merge configs
		mergedConfig := instance.MergeConfig()

		effectiveChecks[i] = domain.EffectiveCheck{
			Name:          instance.Name,
			CheckType:     instance.CheckType,
			ScriptContent: instance.ScriptContent,
			Language:      instance.Language,
			Config:        mergedConfig,
			Version:       instance.Version,
			Hash:          instance.Hash,
			InstanceID:    instance.ID,
		}
	}

	return effectiveChecks, nil
}

// GetEffectiveChecksByNamespace retrieves all effective checks for a namespace
func (s *CheckInstanceService) GetEffectiveChecksByNamespace(ctx context.Context, namespaceID string) ([]domain.CheckInstance, error) {
	// Verify namespace exists
	if _, err := s.namespaceRepo.GetByID(ctx, namespaceID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return s.instanceRepo.GetEffectiveInstancesByNamespace(ctx, namespaceID)
}

// GetEffectiveChecksByGroup retrieves all effective checks for a group
func (s *CheckInstanceService) GetEffectiveChecksByGroup(ctx context.Context, namespaceID, groupID string) ([]domain.CheckInstance, error) {
	// Verify group exists
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Verify group belongs to namespace
	if group.NamespaceID != namespaceID {
		return nil, NewValidationError("namespace_id", "group does not belong to the specified namespace")
	}

	return s.instanceRepo.GetEffectiveInstancesByGroup(ctx, namespaceID, groupID)
}

// Update updates an existing check instance
func (s *CheckInstanceService) Update(ctx context.Context, id string, req dto.UpdateCheckInstanceRequest) (*domain.CheckInstance, error) {
	instance, err := s.instanceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Update fields if provided
	if req.Config != nil {
		instance.Config = req.Config
	}
	if req.Priority != nil {
		instance.Priority = *req.Priority
	}
	if req.IsActive != nil {
		instance.IsActive = *req.IsActive
	}

	// Validate updated instance
	if err := instance.Validate(); err != nil {
		return nil, err
	}

	if err := s.instanceRepo.Update(ctx, instance); err != nil {
		return nil, err
	}

	return instance, nil
}

// Delete performs soft delete on a check instance
func (s *CheckInstanceService) Delete(ctx context.Context, id string) error {
	// Check if instance exists
	_, err := s.instanceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	return s.instanceRepo.Delete(ctx, id)
}

// Purge permanently removes a check instance (hard delete, admin operation)
func (s *CheckInstanceService) Purge(ctx context.Context, id string) error {
	return s.instanceRepo.Purge(ctx, id)
}

// Restore restores a soft-deleted check instance
func (s *CheckInstanceService) Restore(ctx context.Context, id string) error {
	return s.instanceRepo.Restore(ctx, id)
}

// List retrieves a paginated list of check instances
func (s *CheckInstanceService) List(ctx context.Context, pagination dto.PaginationRequest) ([]domain.CheckInstance, int, error) {
	pagination.Normalize()
	return s.instanceRepo.List(ctx, pagination.Page, pagination.Limit)
}

// ListActive retrieves all active (non-deleted) instances
func (s *CheckInstanceService) ListActive(ctx context.Context) ([]domain.CheckInstance, error) {
	return s.instanceRepo.ListActive(ctx)
}
