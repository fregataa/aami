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

// ScriptPolicyService handles business logic for script policys
type ScriptPolicyService struct {
	policyRepo  repository.ScriptPolicyRepository
	scriptRepo    repository.MonitoringScriptRepository
	namespaceRepo repository.NamespaceRepository
	groupRepo     repository.GroupRepository
	targetRepo    repository.TargetRepository
}

// NewScriptPolicyService creates a new ScriptPolicyService
func NewScriptPolicyService(
	policyRepo repository.ScriptPolicyRepository,
	scriptRepo repository.MonitoringScriptRepository,
	namespaceRepo repository.NamespaceRepository,
	groupRepo repository.GroupRepository,
	targetRepo repository.TargetRepository,
) *ScriptPolicyService {
	return &ScriptPolicyService{
		policyRepo:  policyRepo,
		scriptRepo:    scriptRepo,
		namespaceRepo: namespaceRepo,
		groupRepo:     groupRepo,
		targetRepo:    targetRepo,
	}
}

// CreateFromTemplate creates a new script policy from a template
func (s *ScriptPolicyService) CreateFromTemplate(ctx context.Context, req dto.CreateScriptPolicyFromTemplateRequest) (*domain.ScriptPolicy, error) {
	// Validate scope consistency
	if err := validateScopeConsistency(req.Scope, req.NamespaceID, req.GroupID); err != nil {
		return nil, err
	}

	// Verify namespace exists if namespace-level or group-level
	if req.NamespaceID != nil {
		if _, err := s.namespaceRepo.GetByID(ctx, *req.NamespaceID); err != nil {
			if errors.Is(err, domainerrors.ErrNotFound) {
				return nil, domainerrors.ErrForeignKeyViolation
			}
			return nil, err
		}
	}

	// Verify group exists if group-level
	if req.GroupID != nil {
		if _, err := s.groupRepo.GetByID(ctx, *req.GroupID); err != nil {
			if errors.Is(err, domainerrors.ErrNotFound) {
				return nil, domainerrors.ErrForeignKeyViolation
			}
			return nil, err
		}
	}

	// Get template
	script, err := s.scriptRepo.GetByID(ctx, req.TemplateID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrForeignKeyViolation
		}
		return nil, err
	}

	// Set defaults
	config := req.Config
	if config == nil {
		config = make(map[string]interface{})
	}

	priority := req.Priority
	if priority == 0 {
		priority = 100
	}

	// Use domain constructor
	instance := domain.NewScriptPolicyFromTemplate(script, req.Scope, req.NamespaceID, req.GroupID, config)
	instance.ID = uuid.New().String()
	instance.Priority = priority
	instance.IsActive = req.IsActive

	if err := s.policyRepo.Create(ctx, instance); err != nil {
		return nil, err
	}

	return s.policyRepo.GetByID(ctx, instance.ID)
}

// CreateDirect creates a new script policy directly without a template
func (s *ScriptPolicyService) CreateDirect(ctx context.Context, req dto.CreateScriptPolicyDirectRequest) (*domain.ScriptPolicy, error) {
	// Validate scope consistency
	if err := validateScopeConsistency(req.Scope, req.NamespaceID, req.GroupID); err != nil {
		return nil, err
	}

	// Verify namespace exists if namespace-level or group-level
	if req.NamespaceID != nil {
		if _, err := s.namespaceRepo.GetByID(ctx, *req.NamespaceID); err != nil {
			if errors.Is(err, domainerrors.ErrNotFound) {
				return nil, domainerrors.ErrForeignKeyViolation
			}
			return nil, err
		}
	}

	// Verify group exists if group-level
	if req.GroupID != nil {
		if _, err := s.groupRepo.GetByID(ctx, *req.GroupID); err != nil {
			if errors.Is(err, domainerrors.ErrNotFound) {
				return nil, domainerrors.ErrForeignKeyViolation
			}
			return nil, err
		}
	}

	// Set defaults
	config := req.Config
	if config == nil {
		config = make(map[string]interface{})
	}

	defaultConfig := req.DefaultConfig
	if defaultConfig == nil {
		defaultConfig = make(map[string]interface{})
	}

	priority := req.Priority
	if priority == 0 {
		priority = 100
	}

	// Create instance directly
	instance := &domain.ScriptPolicy{
		ID:            uuid.New().String(),
		Name:          req.Name,
		ScriptType:    req.ScriptType,
		ScriptContent: req.ScriptContent,
		Language:      req.Language,
		DefaultConfig: defaultConfig,
		Description:   req.Description,
		Version:       req.Version,
		Scope:         req.Scope,
		NamespaceID:   req.NamespaceID,
		GroupID:       req.GroupID,
		Config:        config,
		Priority:      priority,
		IsActive:      req.IsActive,
	}

	if err := s.policyRepo.Create(ctx, instance); err != nil {
		return nil, err
	}

	return s.policyRepo.GetByID(ctx, instance.ID)
}

// GetByID retrieves a script policy by ID
func (s *ScriptPolicyService) GetByID(ctx context.Context, id string) (*domain.ScriptPolicy, error) {
	instance, err := s.policyRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return instance, nil
}

// GetByTemplateID retrieves all instances for a specific template
func (s *ScriptPolicyService) GetByTemplateID(ctx context.Context, templateID string) ([]domain.ScriptPolicy, error) {
	return s.policyRepo.GetByTemplateID(ctx, templateID)
}

// GetGlobalInstances retrieves all global-scope instances
func (s *ScriptPolicyService) GetGlobalInstances(ctx context.Context) ([]domain.ScriptPolicy, error) {
	return s.policyRepo.GetGlobalInstances(ctx)
}

// GetByNamespaceID retrieves all namespace-level instances for a specific namespace
func (s *ScriptPolicyService) GetByNamespaceID(ctx context.Context, namespaceID string) ([]domain.ScriptPolicy, error) {
	return s.policyRepo.GetByNamespaceID(ctx, namespaceID)
}

// GetByGroupID retrieves all group-level instances for a specific group
func (s *ScriptPolicyService) GetByGroupID(ctx context.Context, groupID string) ([]domain.ScriptPolicy, error) {
	return s.policyRepo.GetByGroupID(ctx, groupID)
}

// GetEffectiveInstance finds the most specific active instance for a template
// Priority: Group > Namespace > Global
func (s *ScriptPolicyService) GetEffectiveInstance(ctx context.Context, templateID, namespaceID, groupID string) (*domain.ScriptPolicy, error) {
	instance, err := s.policyRepo.GetEffectiveInstance(ctx, templateID, namespaceID, groupID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return instance, nil
}

// GetEffectiveChecksByTargetID retrieves all effective checks for a specific target
// This is the main method used by nodes to get their check configurations
func (s *ScriptPolicyService) GetEffectiveChecksByTargetID(ctx context.Context, targetID string) ([]domain.EffectiveCheck, error) {
	// Get effective script policys from repository (handles priority resolution)
	result, err := s.targetRepo.GetEffectivePolicies(ctx, targetID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
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
			ScriptType:     instance.ScriptType,
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
func (s *ScriptPolicyService) GetEffectiveChecksByNamespace(ctx context.Context, namespaceID string) ([]domain.ScriptPolicy, error) {
	// Verify namespace exists
	if _, err := s.namespaceRepo.GetByID(ctx, namespaceID); err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}

	return s.policyRepo.GetEffectiveInstancesByNamespace(ctx, namespaceID)
}

// GetEffectiveChecksByGroup retrieves all effective checks for a group
func (s *ScriptPolicyService) GetEffectiveChecksByGroup(ctx context.Context, namespaceID, groupID string) ([]domain.ScriptPolicy, error) {
	// Verify group exists
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}

	// Verify group belongs to namespace
	if group.NamespaceID != namespaceID {
		return nil, domainerrors.NewValidationError("namespace_id", "group does not belong to the specified namespace")
	}

	return s.policyRepo.GetEffectiveInstancesByGroup(ctx, namespaceID, groupID)
}

// Update updates an existing script policy
func (s *ScriptPolicyService) Update(ctx context.Context, id string, req dto.UpdateScriptPolicyRequest) (*domain.ScriptPolicy, error) {
	instance, err := s.policyRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
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

	if err := s.policyRepo.Update(ctx, instance); err != nil {
		return nil, err
	}

	return instance, nil
}

// Delete performs soft delete on a script policy
func (s *ScriptPolicyService) Delete(ctx context.Context, id string) error {
	// Check if instance exists
	_, err := s.policyRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return domainerrors.ErrNotFound
		}
		return err
	}

	return s.policyRepo.Delete(ctx, id)
}

// Purge permanently removes a script policy (hard delete, admin operation)
func (s *ScriptPolicyService) Purge(ctx context.Context, id string) error {
	return s.policyRepo.Purge(ctx, id)
}

// Restore restores a soft-deleted script policy
func (s *ScriptPolicyService) Restore(ctx context.Context, id string) error {
	return s.policyRepo.Restore(ctx, id)
}

// List retrieves a paginated list of script policys
func (s *ScriptPolicyService) List(ctx context.Context, pagination dto.PaginationRequest) ([]domain.ScriptPolicy, int, error) {
	pagination.Normalize()
	return s.policyRepo.List(ctx, pagination.Page, pagination.Limit)
}

// ListActive retrieves all active (non-deleted) instances
func (s *ScriptPolicyService) ListActive(ctx context.Context) ([]domain.ScriptPolicy, error) {
	return s.policyRepo.ListActive(ctx)
}

// validateScopeConsistency validates that scope and IDs are consistent
func validateScopeConsistency(scope domain.PolicyScope, namespaceID, groupID *string) error {
	switch scope {
	case domain.ScopeGlobal:
		if namespaceID != nil || groupID != nil {
			return domainerrors.NewValidationError("scope", "global scope must not have namespace_id or group_id")
		}
	case domain.ScopeNamespace:
		if namespaceID == nil {
			return domainerrors.NewValidationError("namespace_id", "namespace_id is required for namespace scope")
		}
		if groupID != nil {
			return domainerrors.NewValidationError("group_id", "namespace scope must not have group_id")
		}
	case domain.ScopeGroup:
		if groupID == nil {
			return domainerrors.NewValidationError("group_id", "group_id is required for group scope")
		}
		if namespaceID == nil {
			return domainerrors.NewValidationError("namespace_id", "namespace_id is required for group scope")
		}
	default:
		return domainerrors.NewValidationError("scope", "invalid scope value")
	}
	return nil
}
