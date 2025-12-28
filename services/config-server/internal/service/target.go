package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fregataa/aami/config-server/internal/api/dto"
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
	namespaceRepo   repository.NamespaceRepository
}

// NewTargetService creates a new TargetService
func NewTargetService(
	targetRepo repository.TargetRepository,
	targetGroupRepo repository.TargetGroupRepository,
	groupRepo repository.GroupRepository,
	namespaceRepo repository.NamespaceRepository,
) *TargetService {
	return &TargetService{
		targetRepo:      targetRepo,
		targetGroupRepo: targetGroupRepo,
		groupRepo:       groupRepo,
		namespaceRepo:   namespaceRepo,
	}
}

// Create creates a new target
func (s *TargetService) Create(ctx context.Context, req dto.CreateTargetRequest) (*domain.Target, error) {
	// Check hostname uniqueness
	existing, err := s.targetRepo.GetByHostname(ctx, req.Hostname)
	if err != nil && !errors.Is(err, domainerrors.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domainerrors.ErrAlreadyExists
	}

	var groupIDs []string
	var shouldCreateDefaultOwn bool

	if len(req.GroupIDs) == 0 {
		// Case A: No groups provided - create Default Own Group
		shouldCreateDefaultOwn = true
	} else {
		// Case B: Groups provided - validate all exist
		for _, gid := range req.GroupIDs {
			_, err := s.groupRepo.GetByID(ctx, gid)
			if err != nil {
				if errors.Is(err, domainerrors.ErrNotFound) {
					return nil, domainerrors.ErrForeignKeyViolation
				}
				return nil, err
			}
		}
		groupIDs = req.GroupIDs
	}

	// Create target
	target := &domain.Target{
		ID:        uuid.New().String(),
		Hostname:  req.Hostname,
		IPAddress: req.IPAddress,
		Status:    domain.TargetStatusActive,
		Labels:    req.Labels,
		Metadata:  req.Metadata,
	}

	if target.Labels == nil {
		target.Labels = make(map[string]string)
	}
	if target.Metadata == nil {
		target.Metadata = make(map[string]interface{})
	}

	if err := s.targetRepo.Create(ctx, target); err != nil {
		return nil, err
	}

	// Create Default Own Group if needed
	if shouldCreateDefaultOwn {
		// Find or create default namespace
		namespace, err := s.getOrCreateDefaultNamespace(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get default namespace: %w", err)
		}

		// Create default own group
		defaultGroup := &domain.Group{
			ID:           uuid.New().String(),
			Name:         fmt.Sprintf("target-%s", target.Hostname),
			NamespaceID:  namespace.ID,
			Description:  fmt.Sprintf("Default group for target %s", target.Hostname),
			Priority:     100,
			IsDefaultOwn: true,
			Metadata:     make(map[string]interface{}),
		}

		if err := s.groupRepo.Create(ctx, defaultGroup); err != nil {
			return nil, fmt.Errorf("failed to create default group: %w", err)
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
		return nil, fmt.Errorf("failed to create group mappings: %w", err)
	}

	// Load target with groups
	return s.targetRepo.GetByID(ctx, target.ID)
}

// getOrCreateDefaultNamespace finds or creates the default namespace
func (s *TargetService) getOrCreateDefaultNamespace(ctx context.Context) (*domain.Namespace, error) {
	const defaultNamespaceName = "default"

	ns, err := s.namespaceRepo.GetByName(ctx, defaultNamespaceName)
	if err == nil {
		return ns, nil
	}

	if !errors.Is(err, domainerrors.ErrNotFound) {
		return nil, err
	}

	// Create default namespace
	newNS := &domain.Namespace{
		ID:             uuid.New().String(),
		Name:           defaultNamespaceName,
		Description:    "Default namespace for auto-created groups",
		PolicyPriority: 100,
		MergeStrategy:  domain.MergeStrategyMerge,
	}

	if err := s.namespaceRepo.Create(ctx, newNS); err != nil {
		return nil, err
	}

	return newNS, nil
}

// GetByID retrieves a target by ID
func (s *TargetService) GetByID(ctx context.Context, id string) (*domain.Target, error) {
	target, err := s.targetRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return target, nil
}

// GetByHostname retrieves a target by hostname
func (s *TargetService) GetByHostname(ctx context.Context, hostname string) (*domain.Target, error) {
	target, err := s.targetRepo.GetByHostname(ctx, hostname)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return target, nil
}

// Update updates an existing target
func (s *TargetService) Update(ctx context.Context, id string, req dto.UpdateTargetRequest) (*domain.Target, error) {
	target, err := s.targetRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}

	if req.Hostname != nil {
		target.Hostname = *req.Hostname
	}
	if req.IPAddress != nil {
		target.IPAddress = *req.IPAddress
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

	return s.targetRepo.GetByID(ctx, id)
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
		return domainerrors.NewValidationError("status", "invalid status")
	}

	return s.targetRepo.UpdateStatus(ctx, id, req.Status)
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
func (s *TargetService) GetTargetGroups(ctx context.Context, targetID string) ([]domain.Group, error) {
	// Validate target exists
	target, err := s.targetRepo.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}

	return target.Groups, nil
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
