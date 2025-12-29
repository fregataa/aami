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

// MonitoringScriptService handles business logic for monitoring scripts
type MonitoringScriptService struct {
	scriptRepo   repository.MonitoringScriptRepository
	instanceRepo repository.ScriptPolicyRepository
}

// NewMonitoringScriptService creates a new MonitoringScriptService
func NewMonitoringScriptService(
	scriptRepo repository.MonitoringScriptRepository,
	instanceRepo repository.ScriptPolicyRepository,
) *MonitoringScriptService {
	return &MonitoringScriptService{
		scriptRepo:   scriptRepo,
		instanceRepo: instanceRepo,
	}
}

// Create creates a new monitoring script
func (s *MonitoringScriptService) Create(ctx context.Context, act action.CreateMonitoringScript) (action.MonitoringScriptResult, error) {
	// Check if script name already exists
	existing, err := s.scriptRepo.GetByName(ctx, act.Name)
	if err != nil && !errors.Is(err, domainerrors.ErrNotFound) {
		return action.MonitoringScriptResult{}, err
	}
	if existing != nil {
		return action.MonitoringScriptResult{}, domainerrors.ErrAlreadyExists
	}

	script := &domain.MonitoringScript{
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
	script.UpdateHash()

	// Validate domain object
	if err := script.Validate(); err != nil {
		return action.MonitoringScriptResult{}, err
	}

	if err := s.scriptRepo.Create(ctx, script); err != nil {
		return action.MonitoringScriptResult{}, err
	}

	return action.NewMonitoringScriptResult(script), nil
}

// GetByID retrieves a monitoring script by ID
func (s *MonitoringScriptService) GetByID(ctx context.Context, id string) (action.MonitoringScriptResult, error) {
	script, err := s.scriptRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.MonitoringScriptResult{}, domainerrors.ErrNotFound
		}
		return action.MonitoringScriptResult{}, err
	}
	return action.NewMonitoringScriptResult(script), nil
}

// GetByName retrieves a monitoring script by name
func (s *MonitoringScriptService) GetByName(ctx context.Context, name string) (action.MonitoringScriptResult, error) {
	script, err := s.scriptRepo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.MonitoringScriptResult{}, domainerrors.ErrNotFound
		}
		return action.MonitoringScriptResult{}, err
	}
	return action.NewMonitoringScriptResult(script), nil
}

// GetByScriptType retrieves all scripts for a specific script type
func (s *MonitoringScriptService) GetByScriptType(ctx context.Context, scriptType string) ([]action.MonitoringScriptResult, error) {
	scripts, err := s.scriptRepo.GetByScriptType(ctx, scriptType)
	if err != nil {
		return nil, err
	}
	return action.NewMonitoringScriptResultList(scripts), nil
}

// Update updates an existing monitoring script
func (s *MonitoringScriptService) Update(ctx context.Context, id string, act action.UpdateMonitoringScript) (action.MonitoringScriptResult, error) {
	script, err := s.scriptRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.MonitoringScriptResult{}, domainerrors.ErrNotFound
		}
		return action.MonitoringScriptResult{}, err
	}

	// Update fields if provided
	if act.Name != nil {
		script.Name = *act.Name
	}
	if act.ScriptType != nil {
		script.ScriptType = *act.ScriptType
	}
	if act.Description != nil {
		script.Description = *act.Description
	}
	if act.ScriptContent != nil {
		script.ScriptContent = *act.ScriptContent
		// Recalculate hash when script content changes
		script.UpdateHash()
	}
	if act.Language != nil {
		script.Language = *act.Language
	}
	if act.DefaultConfig != nil {
		script.DefaultConfig = act.DefaultConfig
	}
	if act.Version != nil {
		script.Version = *act.Version
	}

	// Validate updated script
	if err := script.Validate(); err != nil {
		return action.MonitoringScriptResult{}, err
	}

	if err := s.scriptRepo.Update(ctx, script); err != nil {
		return action.MonitoringScriptResult{}, err
	}

	return action.NewMonitoringScriptResult(script), nil
}

// Delete performs soft delete on a monitoring script
func (s *MonitoringScriptService) Delete(ctx context.Context, id string) error {
	// Check if script exists
	_, err := s.scriptRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return domainerrors.ErrNotFound
		}
		return err
	}

	// Check if script is in use by any instances
	instances, err := s.instanceRepo.GetByTemplateID(ctx, id)
	if err != nil {
		return err
	}
	if len(instances) > 0 {
		return domainerrors.ErrInUse
	}

	return s.scriptRepo.Delete(ctx, id)
}

// Purge permanently removes a monitoring script (hard delete, admin operation)
func (s *MonitoringScriptService) Purge(ctx context.Context, id string) error {
	return s.scriptRepo.Purge(ctx, id)
}

// Restore restores a soft-deleted monitoring script
func (s *MonitoringScriptService) Restore(ctx context.Context, id string) error {
	return s.scriptRepo.Restore(ctx, id)
}

// List retrieves a paginated list of monitoring scripts
func (s *MonitoringScriptService) List(ctx context.Context, pagination action.Pagination) (action.ListResult[action.MonitoringScriptResult], error) {
	scripts, total, err := s.scriptRepo.List(ctx, pagination.Page, pagination.Limit)
	if err != nil {
		return action.ListResult[action.MonitoringScriptResult]{}, err
	}

	results := action.NewMonitoringScriptResultList(scripts)
	return action.NewListResult(results, pagination, total), nil
}

// ListActive retrieves all active (non-deleted) scripts
func (s *MonitoringScriptService) ListActive(ctx context.Context) ([]action.MonitoringScriptResult, error) {
	scripts, err := s.scriptRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	return action.NewMonitoringScriptResultList(scripts), nil
}

// VerifyHash checks if the stored hash matches the script content
func (s *MonitoringScriptService) VerifyHash(ctx context.Context, id string) (bool, error) {
	script, err := s.scriptRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return false, domainerrors.ErrNotFound
		}
		return false, err
	}
	return script.VerifyHash(), nil
}
