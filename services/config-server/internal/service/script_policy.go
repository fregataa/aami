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

// ScriptPolicyService handles business logic for script policies
type ScriptPolicyService struct {
	policyRepo    repository.ScriptPolicyRepository
	templateRepo  repository.ScriptTemplateRepository
	namespaceRepo repository.NamespaceRepository
	groupRepo     repository.GroupRepository
	targetRepo    repository.TargetRepository
}

// NewScriptPolicyService creates a new ScriptPolicyService
func NewScriptPolicyService(
	policyRepo repository.ScriptPolicyRepository,
	templateRepo repository.ScriptTemplateRepository,
	namespaceRepo repository.NamespaceRepository,
	groupRepo repository.GroupRepository,
	targetRepo repository.TargetRepository,
) *ScriptPolicyService {
	return &ScriptPolicyService{
		policyRepo:    policyRepo,
		templateRepo:  templateRepo,
		namespaceRepo: namespaceRepo,
		groupRepo:     groupRepo,
		targetRepo:    targetRepo,
	}
}

// CreateFromTemplate creates a new script policy from a template
func (s *ScriptPolicyService) CreateFromTemplate(ctx context.Context, act action.CreateScriptPolicyFromTemplate) (action.ScriptPolicyResult, error) {
	// Validate scope consistency
	if err := validateScopeConsistency(act.Scope, act.NamespaceID, act.GroupID); err != nil {
		return action.ScriptPolicyResult{}, err
	}

	// Verify namespace exists if namespace-level or group-level
	if act.NamespaceID != nil {
		if _, err := s.namespaceRepo.GetByID(ctx, *act.NamespaceID); err != nil {
			if errors.Is(err, domainerrors.ErrNotFound) {
				return action.ScriptPolicyResult{}, domainerrors.ErrForeignKeyViolation
			}
			return action.ScriptPolicyResult{}, err
		}
	}

	// Verify group exists if group-level
	if act.GroupID != nil {
		if _, err := s.groupRepo.GetByID(ctx, *act.GroupID); err != nil {
			if errors.Is(err, domainerrors.ErrNotFound) {
				return action.ScriptPolicyResult{}, domainerrors.ErrForeignKeyViolation
			}
			return action.ScriptPolicyResult{}, err
		}
	}

	// Get template
	script, err := s.templateRepo.GetByID(ctx, act.TemplateID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.ScriptPolicyResult{}, domainerrors.ErrForeignKeyViolation
		}
		return action.ScriptPolicyResult{}, err
	}

	// Set defaults
	config := act.Config
	if config == nil {
		config = make(map[string]interface{})
	}

	priority := act.Priority
	if priority == 0 {
		priority = 100
	}

	// Use domain constructor
	instance := domain.NewScriptPolicyFromTemplate(script, act.Scope, act.NamespaceID, act.GroupID, config)
	instance.ID = uuid.New().String()
	instance.Priority = priority
	instance.IsActive = act.IsActive

	if err := s.policyRepo.Create(ctx, instance); err != nil {
		return action.ScriptPolicyResult{}, err
	}

	instance, err = s.policyRepo.GetByID(ctx, instance.ID)
	if err != nil {
		return action.ScriptPolicyResult{}, err
	}
	return action.NewScriptPolicyResult(instance), nil
}

// CreateDirect creates a new script policy directly without a template
func (s *ScriptPolicyService) CreateDirect(ctx context.Context, act action.CreateScriptPolicyDirect) (action.ScriptPolicyResult, error) {
	// Validate scope consistency
	if err := validateScopeConsistency(act.Scope, act.NamespaceID, act.GroupID); err != nil {
		return action.ScriptPolicyResult{}, err
	}

	// Verify namespace exists if namespace-level or group-level
	if act.NamespaceID != nil {
		if _, err := s.namespaceRepo.GetByID(ctx, *act.NamespaceID); err != nil {
			if errors.Is(err, domainerrors.ErrNotFound) {
				return action.ScriptPolicyResult{}, domainerrors.ErrForeignKeyViolation
			}
			return action.ScriptPolicyResult{}, err
		}
	}

	// Verify group exists if group-level
	if act.GroupID != nil {
		if _, err := s.groupRepo.GetByID(ctx, *act.GroupID); err != nil {
			if errors.Is(err, domainerrors.ErrNotFound) {
				return action.ScriptPolicyResult{}, domainerrors.ErrForeignKeyViolation
			}
			return action.ScriptPolicyResult{}, err
		}
	}

	// Set defaults
	config := act.Config
	if config == nil {
		config = make(map[string]interface{})
	}

	defaultConfig := act.DefaultConfig
	if defaultConfig == nil {
		defaultConfig = make(map[string]interface{})
	}

	priority := act.Priority
	if priority == 0 {
		priority = 100
	}

	// Create instance directly
	instance := &domain.ScriptPolicy{
		ID:            uuid.New().String(),
		Name:          act.Name,
		ScriptType:    act.ScriptType,
		ScriptContent: act.ScriptContent,
		Language:      act.Language,
		DefaultConfig: defaultConfig,
		Description:   act.Description,
		Version:       act.Version,
		Scope:         act.Scope,
		NamespaceID:   act.NamespaceID,
		GroupID:       act.GroupID,
		Config:        config,
		Priority:      priority,
		IsActive:      act.IsActive,
	}

	if err := s.policyRepo.Create(ctx, instance); err != nil {
		return action.ScriptPolicyResult{}, err
	}

	instance, err := s.policyRepo.GetByID(ctx, instance.ID)
	if err != nil {
		return action.ScriptPolicyResult{}, err
	}
	return action.NewScriptPolicyResult(instance), nil
}

// GetByID retrieves a script policy by ID
func (s *ScriptPolicyService) GetByID(ctx context.Context, id string) (action.ScriptPolicyResult, error) {
	instance, err := s.policyRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.ScriptPolicyResult{}, domainerrors.ErrNotFound
		}
		return action.ScriptPolicyResult{}, err
	}
	return action.NewScriptPolicyResult(instance), nil
}

// GetByTemplateID retrieves all instances for a specific template
func (s *ScriptPolicyService) GetByTemplateID(ctx context.Context, templateID string) ([]action.ScriptPolicyResult, error) {
	instances, err := s.policyRepo.GetByTemplateID(ctx, templateID)
	if err != nil {
		return nil, err
	}
	return action.NewScriptPolicyResultList(instances), nil
}

// GetGlobalInstances retrieves all global-scope instances
func (s *ScriptPolicyService) GetGlobalInstances(ctx context.Context) ([]action.ScriptPolicyResult, error) {
	instances, err := s.policyRepo.GetGlobalInstances(ctx)
	if err != nil {
		return nil, err
	}
	return action.NewScriptPolicyResultList(instances), nil
}

// GetByNamespaceID retrieves all namespace-level instances for a specific namespace
func (s *ScriptPolicyService) GetByNamespaceID(ctx context.Context, namespaceID string) ([]action.ScriptPolicyResult, error) {
	instances, err := s.policyRepo.GetByNamespaceID(ctx, namespaceID)
	if err != nil {
		return nil, err
	}
	return action.NewScriptPolicyResultList(instances), nil
}

// GetByGroupID retrieves all group-level instances for a specific group
func (s *ScriptPolicyService) GetByGroupID(ctx context.Context, groupID string) ([]action.ScriptPolicyResult, error) {
	instances, err := s.policyRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return action.NewScriptPolicyResultList(instances), nil
}

// GetEffectiveInstance finds the most specific active instance for a template
// Priority: Group > Namespace > Global
func (s *ScriptPolicyService) GetEffectiveInstance(ctx context.Context, templateID, namespaceID, groupID string) (action.ScriptPolicyResult, error) {
	instance, err := s.policyRepo.GetEffectiveInstance(ctx, templateID, namespaceID, groupID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.ScriptPolicyResult{}, domainerrors.ErrNotFound
		}
		return action.ScriptPolicyResult{}, err
	}
	return action.NewScriptPolicyResult(instance), nil
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
func (s *ScriptPolicyService) GetEffectiveChecksByNamespace(ctx context.Context, namespaceID string) ([]action.ScriptPolicyResult, error) {
	// Verify namespace exists
	if _, err := s.namespaceRepo.GetByID(ctx, namespaceID); err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}

	instances, err := s.policyRepo.GetEffectiveInstancesByNamespace(ctx, namespaceID)
	if err != nil {
		return nil, err
	}
	return action.NewScriptPolicyResultList(instances), nil
}

// GetEffectiveChecksByGroup retrieves all effective checks for a group
func (s *ScriptPolicyService) GetEffectiveChecksByGroup(ctx context.Context, namespaceID, groupID string) ([]action.ScriptPolicyResult, error) {
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

	instances, err := s.policyRepo.GetEffectiveInstancesByGroup(ctx, namespaceID, groupID)
	if err != nil {
		return nil, err
	}
	return action.NewScriptPolicyResultList(instances), nil
}

// Update updates an existing script policy
func (s *ScriptPolicyService) Update(ctx context.Context, id string, act action.UpdateScriptPolicy) (action.ScriptPolicyResult, error) {
	instance, err := s.policyRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.ScriptPolicyResult{}, domainerrors.ErrNotFound
		}
		return action.ScriptPolicyResult{}, err
	}

	// Update fields if provided
	if act.Config != nil {
		instance.Config = act.Config
	}
	if act.Priority != nil {
		instance.Priority = *act.Priority
	}
	if act.IsActive != nil {
		instance.IsActive = *act.IsActive
	}

	// Validate updated instance
	if err := instance.Validate(); err != nil {
		return action.ScriptPolicyResult{}, err
	}

	if err := s.policyRepo.Update(ctx, instance); err != nil {
		return action.ScriptPolicyResult{}, err
	}

	return action.NewScriptPolicyResult(instance), nil
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
func (s *ScriptPolicyService) List(ctx context.Context, pagination action.Pagination) (action.ListResult[action.ScriptPolicyResult], error) {
	instances, total, err := s.policyRepo.List(ctx, pagination.Page, pagination.Limit)
	if err != nil {
		return action.ListResult[action.ScriptPolicyResult]{}, err
	}

	results := action.NewScriptPolicyResultList(instances)
	return action.NewListResult(results, pagination, total), nil
}

// ListActive retrieves all active (non-deleted) instances
func (s *ScriptPolicyService) ListActive(ctx context.Context) ([]action.ScriptPolicyResult, error) {
	instances, err := s.policyRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	return action.NewScriptPolicyResultList(instances), nil
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
