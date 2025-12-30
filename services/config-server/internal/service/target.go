package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fregataa/aami/config-server/internal/action"
	"github.com/fregataa/aami/config-server/internal/domain"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
	"github.com/fregataa/aami/config-server/internal/repository"
	"github.com/google/uuid"
)

// TargetService handles business logic for targets
type TargetService struct {
	targetRepo      repository.TargetRepository
	targetGroupRepo repository.TargetGroupRepository
	groupRepo       repository.GroupRepository
}

// NewTargetService creates a new TargetService
func NewTargetService(
	targetRepo repository.TargetRepository,
	targetGroupRepo repository.TargetGroupRepository,
	groupRepo repository.GroupRepository,
) *TargetService {
	return &TargetService{
		targetRepo:      targetRepo,
		targetGroupRepo: targetGroupRepo,
		groupRepo:       groupRepo,
	}
}

// Create creates a new target
func (s *TargetService) Create(ctx context.Context, act action.CreateTarget) (action.TargetResult, error) {
	// Check hostname uniqueness
	existing, err := s.targetRepo.GetByHostname(ctx, act.Hostname)
	if err != nil && !errors.Is(err, domainerrors.ErrNotFound) {
		return action.TargetResult{}, err
	}
	if existing != nil {
		return action.TargetResult{}, domainerrors.ErrAlreadyExists
	}

	var groupIDs []string
	var shouldCreateDefaultOwn bool

	if len(act.GroupIDs) == 0 {
		// Case A: No groups provided - create Default Own Group
		shouldCreateDefaultOwn = true
	} else {
		// Case B: Groups provided - validate all exist
		for _, gid := range act.GroupIDs {
			_, err := s.groupRepo.GetByID(ctx, gid)
			if err != nil {
				if errors.Is(err, domainerrors.ErrNotFound) {
					return action.TargetResult{}, domainerrors.ErrForeignKeyViolation
				}
				return action.TargetResult{}, err
			}
		}
		groupIDs = act.GroupIDs
	}

	// Create target domain object from action
	status := act.Status
	if status == "" {
		status = domain.TargetStatusActive
	}

	target := &domain.Target{
		ID:        uuid.New().String(),
		Hostname:  act.Hostname,
		IPAddress: act.IPAddress,
		Status:    status,
		Labels:    act.Labels,
		Metadata:  act.Metadata,
	}

	if target.Labels == nil {
		target.Labels = make(map[string]string)
	}
	if target.Metadata == nil {
		target.Metadata = make(map[string]string)
	}

	if err := s.targetRepo.Create(ctx, target); err != nil {
		return action.TargetResult{}, err
	}

	// Create Default Own Group if needed
	if shouldCreateDefaultOwn {
		// Create default own group
		defaultGroup := &domain.Group{
			ID:           uuid.New().String(),
			Name:         fmt.Sprintf("target-%s", target.Hostname),
			Description:  fmt.Sprintf("Default group for target %s", target.Hostname),
			Priority:     100,
			IsDefaultOwn: true,
			Metadata:     make(map[string]string),
		}

		if err := s.groupRepo.Create(ctx, defaultGroup); err != nil {
			return action.TargetResult{}, fmt.Errorf("failed to create default group: %w", err)
		}

		groupIDs = []string{defaultGroup.ID}
	}

	// Create target-group mappings
	mappings := make([]domain.TargetGroup, len(groupIDs))
	for i, gid := range groupIDs {
		mappings[i] = domain.TargetGroup{
			TargetID:     target.ID,
			GroupID:      gid,
			IsDefaultOwn: shouldCreateDefaultOwn && i == 0,
		}
	}

	if err := s.targetGroupRepo.CreateBatch(ctx, mappings); err != nil {
		return action.TargetResult{}, fmt.Errorf("failed to create group mappings: %w", err)
	}

	// Load target with groups and convert to result
	target, err = s.targetRepo.GetByID(ctx, target.ID)
	if err != nil {
		return action.TargetResult{}, err
	}

	return action.NewTargetResult(target), nil
}

// GetByID retrieves a target by ID
func (s *TargetService) GetByID(ctx context.Context, id string) (action.TargetResult, error) {
	target, err := s.targetRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.TargetResult{}, domainerrors.ErrNotFound
		}
		return action.TargetResult{}, err
	}
	return action.NewTargetResult(target), nil
}

// GetByHostname retrieves a target by hostname
func (s *TargetService) GetByHostname(ctx context.Context, hostname string) (action.TargetResult, error) {
	target, err := s.targetRepo.GetByHostname(ctx, hostname)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.TargetResult{}, domainerrors.ErrNotFound
		}
		return action.TargetResult{}, err
	}
	return action.NewTargetResult(target), nil
}

// Update updates an existing target
func (s *TargetService) Update(ctx context.Context, id string, act action.UpdateTarget) (action.TargetResult, error) {
	target, err := s.targetRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.TargetResult{}, domainerrors.ErrNotFound
		}
		return action.TargetResult{}, err
	}

	if act.Hostname != nil {
		target.Hostname = *act.Hostname
	}
	if act.IPAddress != nil {
		target.IPAddress = *act.IPAddress
	}
	if act.Status != nil {
		target.Status = *act.Status
	}
	if act.Labels != nil {
		target.Labels = act.Labels
	}
	if act.Metadata != nil {
		target.Metadata = act.Metadata
	}

	if err := s.targetRepo.Update(ctx, target); err != nil {
		return action.TargetResult{}, err
	}

	target, err = s.targetRepo.GetByID(ctx, id)
	if err != nil {
		return action.TargetResult{}, err
	}

	return action.NewTargetResult(target), nil
}

// Delete performs soft delete on a target
func (s *TargetService) Delete(ctx context.Context, id string) error {
	_, err := s.targetRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return domainerrors.ErrNotFound
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
func (s *TargetService) List(ctx context.Context, pagination action.Pagination) (action.ListResult[action.TargetResult], error) {
	targets, total, err := s.targetRepo.List(ctx, pagination.Page, pagination.Limit)
	if err != nil {
		return action.ListResult[action.TargetResult]{}, err
	}

	results := action.NewTargetResultList(targets)
	return action.NewListResult(results, pagination, total), nil
}

// GetByGroupID retrieves all targets in a group
func (s *TargetService) GetByGroupID(ctx context.Context, groupID string) ([]action.TargetResult, error) {
	targets, err := s.targetRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return action.NewTargetResultList(targets), nil
}

// UpdateStatus updates the status of a target
func (s *TargetService) UpdateStatus(ctx context.Context, id string, act action.UpdateTargetStatus) error {
	if !act.Status.IsValid() {
		return domainerrors.NewValidationError("status", "invalid status")
	}

	return s.targetRepo.UpdateStatus(ctx, id, act.Status)
}

// Heartbeat updates the last_seen timestamp of a target
func (s *TargetService) Heartbeat(ctx context.Context, id string) error {
	return s.targetRepo.UpdateLastSeen(ctx, id, time.Now())
}

// AddGroupMapping adds a target to a group
func (s *TargetService) AddGroupMapping(ctx context.Context, targetID, groupID string) error {
	// Validate target exists
	_, err := s.targetRepo.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return domainerrors.ErrNotFound
		}
		return err
	}

	// Validate group exists
	_, err = s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return domainerrors.ErrForeignKeyViolation
		}
		return err
	}

	// Check if mapping already exists
	exists, err := s.targetGroupRepo.Exists(ctx, targetID, groupID)
	if err != nil {
		return err
	}
	if exists {
		return domainerrors.ErrAlreadyExists
	}

	// Create mapping
	mapping := &domain.TargetGroup{
		TargetID:     targetID,
		GroupID:      groupID,
		IsDefaultOwn: false,
	}

	return s.targetGroupRepo.Create(ctx, mapping)
}

// RemoveGroupMapping removes a target from a group
func (s *TargetService) RemoveGroupMapping(ctx context.Context, targetID, groupID string) error {
	// Check if target exists
	_, err := s.targetRepo.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return domainerrors.ErrNotFound
		}
		return err
	}

	// Count existing mappings
	count, err := s.targetGroupRepo.CountByTarget(ctx, targetID)
	if err != nil {
		return err
	}

	// Prevent removal of last group
	if count <= 1 {
		return domainerrors.ErrCannotRemoveLastGroup
	}

	// Delete mapping
	return s.targetGroupRepo.Delete(ctx, targetID, groupID)
}

// GetTargetGroups retrieves all groups for a target
func (s *TargetService) GetTargetGroups(ctx context.Context, targetID string) ([]action.GroupResult, error) {
	// Validate target exists
	target, err := s.targetRepo.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}

	return action.NewGroupResultList(target.Groups), nil
}

// ReplaceGroupMappings replaces all group mappings for a target
func (s *TargetService) ReplaceGroupMappings(ctx context.Context, targetID string, groupIDs []string) error {
	if len(groupIDs) == 0 {
		return errors.New("at least one group is required")
	}

	// Validate target exists
	_, err := s.targetRepo.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return domainerrors.ErrNotFound
		}
		return err
	}

	// Validate all groups exist
	for _, gid := range groupIDs {
		_, err := s.groupRepo.GetByID(ctx, gid)
		if err != nil {
			if errors.Is(err, domainerrors.ErrNotFound) {
				return domainerrors.ErrForeignKeyViolation
			}
			return err
		}
	}

	// Delete all existing mappings
	if err := s.targetGroupRepo.DeleteByTarget(ctx, targetID); err != nil {
		return err
	}

	// Create new mappings
	mappings := make([]domain.TargetGroup, len(groupIDs))
	for i, gid := range groupIDs {
		mappings[i] = domain.TargetGroup{
			TargetID:     targetID,
			GroupID:      gid,
			IsDefaultOwn: false,
		}
	}

	return s.targetGroupRepo.CreateBatch(ctx, mappings)
}
