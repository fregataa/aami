package service

import (
	"context"
	"errors"
	"time"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/domain"
	"github.com/fregataa/aami/config-server/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TargetService handles business logic for targets
type TargetService struct {
	targetRepo repository.TargetRepository
	groupRepo  repository.GroupRepository
}

// NewTargetService creates a new TargetService
func NewTargetService(
	targetRepo repository.TargetRepository,
	groupRepo repository.GroupRepository,
) *TargetService {
	return &TargetService{
		targetRepo: targetRepo,
		groupRepo:  groupRepo,
	}
}

// Create creates a new target
func (s *TargetService) Create(ctx context.Context, req dto.CreateTargetRequest) (*domain.Target, error) {
	// Validate primary group exists
	_, err := s.groupRepo.GetByID(ctx, req.PrimaryGroupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewValidationError("primary_group_id", "group not found")
		}
		return nil, err
	}

	// Check hostname uniqueness
	existing, err := s.targetRepo.GetByHostname(ctx, req.Hostname)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyExists
	}

	target := &domain.Target{
		ID:             uuid.New().String(),
		Hostname:       req.Hostname,
		IPAddress:      req.IPAddress,
		PrimaryGroupID: req.PrimaryGroupID,
		Status:         domain.TargetStatusActive,
		Labels:         req.Labels,
		Metadata:       req.Metadata,
	}

	if req.Labels == nil {
		target.Labels = make(map[string]string)
	}
	if req.Metadata == nil {
		target.Metadata = make(map[string]interface{})
	}

	if err := s.targetRepo.Create(ctx, target); err != nil {
		return nil, err
	}

	return target, nil
}

// GetByID retrieves a target by ID
func (s *TargetService) GetByID(ctx context.Context, id string) (*domain.Target, error) {
	target, err := s.targetRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return target, nil
}

// GetByHostname retrieves a target by hostname
func (s *TargetService) GetByHostname(ctx context.Context, hostname string) (*domain.Target, error) {
	target, err := s.targetRepo.GetByHostname(ctx, hostname)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return target, nil
}

// Update updates an existing target
func (s *TargetService) Update(ctx context.Context, id string, req dto.UpdateTargetRequest) (*domain.Target, error) {
	target, err := s.targetRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if req.Hostname != nil {
		target.Hostname = *req.Hostname
	}
	if req.IPAddress != nil {
		target.IPAddress = *req.IPAddress
	}
	if req.PrimaryGroupID != nil {
		// Validate group exists
		_, err := s.groupRepo.GetByID(ctx, *req.PrimaryGroupID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, NewValidationError("primary_group_id", "group not found")
			}
			return nil, err
		}
		target.PrimaryGroupID = *req.PrimaryGroupID
	}
	if req.Status != nil {
		target.Status = *req.Status
	}
	if req.Labels != nil {
		target.Labels = req.Labels
	}
	if req.Metadata != nil {
		target.Metadata = req.Metadata
	}

	if err := s.targetRepo.Update(ctx, target); err != nil {
		return nil, err
	}

	return target, nil
}

// Delete performs soft delete on a target
func (s *TargetService) Delete(ctx context.Context, id string) error {
	_, err := s.targetRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	return s.targetRepo.Delete(ctx, id)
}

// Purge permanently removes a target (hard delete, admin operation)
func (s *TargetService) Purge(ctx context.Context, id string) error {
	return s.targetRepo.Purge(ctx, id)
}

// Restore restores a soft-deleted target
func (s *TargetService) Restore(ctx context.Context, id string) error {
	return s.targetRepo.Restore(ctx, id)
}

// List retrieves a paginated list of targets
func (s *TargetService) List(ctx context.Context, pagination dto.PaginationRequest) ([]domain.Target, int, error) {
	pagination.Normalize()
	return s.targetRepo.List(ctx, pagination.Page, pagination.Limit)
}

// GetByGroupID retrieves all targets in a group
func (s *TargetService) GetByGroupID(ctx context.Context, groupID string) ([]domain.Target, error) {
	return s.targetRepo.GetByGroupID(ctx, groupID)
}

// UpdateStatus updates the status of a target
func (s *TargetService) UpdateStatus(ctx context.Context, id string, req dto.UpdateTargetStatusRequest) error {
	if !req.Status.IsValid() {
		return NewValidationError("status", "invalid status")
	}

	return s.targetRepo.UpdateStatus(ctx, id, req.Status)
}

// Heartbeat updates the last_seen timestamp of a target
func (s *TargetService) Heartbeat(ctx context.Context, id string) error {
	return s.targetRepo.UpdateLastSeen(ctx, id, time.Now())
}
