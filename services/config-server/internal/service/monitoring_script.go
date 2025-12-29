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
func (s *MonitoringScriptService) Create(ctx context.Context, req dto.CreateMonitoringScriptRequest) (*domain.MonitoringScript, error) {
	// Validate required fields
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if script name already exists
	existing, err := s.scriptRepo.GetByName(ctx, req.Name)
	if err != nil && !errors.Is(err, domainerrors.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domainerrors.ErrAlreadyExists
	}

	script := &domain.MonitoringScript{
		ID:            uuid.New().String(),
		Name:          req.Name,
		ScriptType:    req.ScriptType,
		ScriptContent: req.ScriptContent,
		Language:      req.Language,
		DefaultConfig: req.DefaultConfig,
		Description:   req.Description,
		Version:       req.Version,
	}

	// Compute hash from script content
	script.UpdateHash()

	// Validate domain object
	if err := script.Validate(); err != nil {
		return nil, err
	}

	if err := s.scriptRepo.Create(ctx, script); err != nil {
		return nil, err
	}

	return script, nil
}

// GetByID retrieves a monitoring script by ID
func (s *MonitoringScriptService) GetByID(ctx context.Context, id string) (*domain.MonitoringScript, error) {
	script, err := s.scriptRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return script, nil
}

// GetByName retrieves a monitoring script by name
func (s *MonitoringScriptService) GetByName(ctx context.Context, name string) (*domain.MonitoringScript, error) {
	script, err := s.scriptRepo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return script, nil
}

// GetByScriptType retrieves all scripts for a specific script type
func (s *MonitoringScriptService) GetByScriptType(ctx context.Context, scriptType string) ([]domain.MonitoringScript, error) {
	return s.scriptRepo.GetByScriptType(ctx, scriptType)
}

// Update updates an existing monitoring script
func (s *MonitoringScriptService) Update(ctx context.Context, id string, req dto.UpdateMonitoringScriptRequest) (*domain.MonitoringScript, error) {
	script, err := s.scriptRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}

	// Update fields if provided
	if req.Description != nil {
		script.Description = *req.Description
	}
	if req.ScriptContent != nil {
		script.ScriptContent = *req.ScriptContent
		// Recalculate hash when script content changes
		script.UpdateHash()
	}
	if req.Language != nil {
		script.Language = *req.Language
	}
	if req.DefaultConfig != nil {
		script.DefaultConfig = req.DefaultConfig
	}
	if req.Version != nil {
		script.Version = *req.Version
	}

	// Validate updated script
	if err := script.Validate(); err != nil {
		return nil, err
	}

	if err := s.scriptRepo.Update(ctx, script); err != nil {
		return nil, err
	}

	return script, nil
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
func (s *MonitoringScriptService) List(ctx context.Context, pagination dto.PaginationRequest) ([]domain.MonitoringScript, int, error) {
	pagination.Normalize()
	return s.scriptRepo.List(ctx, pagination.Page, pagination.Limit)
}

// ListActive retrieves all active (non-deleted) scripts
func (s *MonitoringScriptService) ListActive(ctx context.Context) ([]domain.MonitoringScript, error) {
	return s.scriptRepo.ListActive(ctx)
}

// VerifyHash checks if the stored hash matches the script content
func (s *MonitoringScriptService) VerifyHash(ctx context.Context, id string) (bool, error) {
	script, err := s.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	return script.VerifyHash(), nil
}
