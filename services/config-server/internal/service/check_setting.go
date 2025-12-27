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

// CheckSettingService handles business logic for check settings
type CheckSettingService struct {
	checkSettingRepo repository.CheckSettingRepository
	groupRepo        repository.GroupRepository
}

// NewCheckSettingService creates a new CheckSettingService
func NewCheckSettingService(
	checkSettingRepo repository.CheckSettingRepository,
	groupRepo repository.GroupRepository,
) *CheckSettingService {
	return &CheckSettingService{
		checkSettingRepo: checkSettingRepo,
		groupRepo:        groupRepo,
	}
}

// Create creates a new check setting
func (s *CheckSettingService) Create(ctx context.Context, req dto.CreateCheckSettingRequest) (*domain.CheckSetting, error) {
	// Validate group exists
	_, err := s.groupRepo.GetByID(ctx, req.GroupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewValidationError("group_id", "group not found")
		}
		return nil, err
	}

	setting := &domain.CheckSetting{
		ID:            uuid.New().String(),
		GroupID:       req.GroupID,
		CheckType:     req.CheckType,
		Config:        req.Config,
		MergeStrategy: req.MergeStrategy,
		Priority:      req.Priority,
	}

	if setting.Config == nil {
		setting.Config = make(map[string]interface{})
	}
	if setting.MergeStrategy == "" {
		setting.MergeStrategy = "merge"
	}
	if setting.Priority == 0 {
		setting.Priority = 100
	}

	if err := s.checkSettingRepo.Create(ctx, setting); err != nil {
		return nil, err
	}

	return setting, nil
}

// GetByID retrieves a check setting by ID
func (s *CheckSettingService) GetByID(ctx context.Context, id string) (*domain.CheckSetting, error) {
	setting, err := s.checkSettingRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return setting, nil
}

// Update updates an existing check setting
func (s *CheckSettingService) Update(ctx context.Context, id string, req dto.UpdateCheckSettingRequest) (*domain.CheckSetting, error) {
	setting, err := s.checkSettingRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if req.Config != nil {
		setting.Config = req.Config
	}
	if req.MergeStrategy != nil {
		setting.MergeStrategy = *req.MergeStrategy
	}
	if req.Priority != nil {
		setting.Priority = *req.Priority
	}

	if err := s.checkSettingRepo.Update(ctx, setting); err != nil {
		return nil, err
	}

	return setting, nil
}

// Delete performs soft delete on a check setting
func (s *CheckSettingService) Delete(ctx context.Context, id string) error {
	_, err := s.checkSettingRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	return s.checkSettingRepo.Delete(ctx, id)
}

// Purge permanently removes a check setting (hard delete, admin operation)
func (s *CheckSettingService) Purge(ctx context.Context, id string) error {
	return s.checkSettingRepo.Purge(ctx, id)
}

// Restore restores a soft-deleted check setting
func (s *CheckSettingService) Restore(ctx context.Context, id string) error {
	return s.checkSettingRepo.Restore(ctx, id)
}

// List retrieves a paginated list of check settings
func (s *CheckSettingService) List(ctx context.Context, pagination dto.PaginationRequest) ([]domain.CheckSetting, int, error) {
	pagination.Normalize()
	return s.checkSettingRepo.List(ctx, pagination.Page, pagination.Limit)
}

// GetByGroupID retrieves all check settings for a group
func (s *CheckSettingService) GetByGroupID(ctx context.Context, groupID string) ([]domain.CheckSetting, error) {
	return s.checkSettingRepo.GetByGroupID(ctx, groupID)
}

// GetByCheckType retrieves all check settings of a specific type
func (s *CheckSettingService) GetByCheckType(ctx context.Context, checkType string) ([]domain.CheckSetting, error) {
	return s.checkSettingRepo.GetByCheckType(ctx, checkType)
}
