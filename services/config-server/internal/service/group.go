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

// GroupService handles business logic for groups
type GroupService struct {
	groupRepo     repository.GroupRepository
	namespaceRepo repository.NamespaceRepository
}

// NewGroupService creates a new GroupService
func NewGroupService(groupRepo repository.GroupRepository, namespaceRepo repository.NamespaceRepository) *GroupService {
	return &GroupService{
		groupRepo:     groupRepo,
		namespaceRepo: namespaceRepo,
	}
}

// Create creates a new group
func (s *GroupService) Create(ctx context.Context, req dto.CreateGroupRequest) (*domain.Group, error) {
	// Validate namespace exists
	namespace, err := s.namespaceRepo.GetByID(ctx, req.NamespaceID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.NewValidationError("namespace_id", "namespace not found")
		}
		return nil, err
	}

	// Validate parent exists if specified
	if req.ParentID != nil {
		parent, err := s.groupRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			if errors.Is(err, domainerrors.ErrNotFound) {
				return nil, domainerrors.ErrForeignKeyViolation
			}
			return nil, err
		}

		// Ensure parent is in the same namespace
		if parent.NamespaceID != req.NamespaceID {
			return nil, domainerrors.NewValidationError("parent_id", "parent must be in the same namespace")
		}

		// Check for circular references
		if err := s.checkCircularReference(ctx, *req.ParentID, ""); err != nil {
			return nil, err
		}
	}

	// Set default priority based on namespace if not specified
	priority := req.Priority
	if priority == 0 {
		priority = namespace.PolicyPriority
	}

	// Initialize metadata if nil
	metadata := req.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	group := &domain.Group{
		ID:          uuid.New().String(),
		Name:        req.Name,
		NamespaceID: req.NamespaceID,
		ParentID:    req.ParentID,
		Description: req.Description,
		Priority:    priority,
		Metadata:    metadata,
	}

	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	// Load namespace for response
	group.Namespace = namespace

	return group, nil
}

// GetByID retrieves a group by ID
func (s *GroupService) GetByID(ctx context.Context, id string) (*domain.Group, error) {
	group, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return group, nil
}

// Update updates an existing group
func (s *GroupService) Update(ctx context.Context, id string, req dto.UpdateGroupRequest) (*domain.Group, error) {
	// Get existing group
	group, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		group.Name = *req.Name
	}

	if req.ParentID != nil {
		// Validate parent exists
		parent, err := s.groupRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			if errors.Is(err, domainerrors.ErrNotFound) {
				return nil, domainerrors.ErrForeignKeyViolation
			}
			return nil, err
		}

		// Ensure parent is in the same namespace
		if parent.Namespace != group.Namespace {
			return nil, domainerrors.NewValidationError("parent_id", "parent must be in the same namespace")
		}

		// Check for circular references
		if err := s.checkCircularReference(ctx, *req.ParentID, id); err != nil {
			return nil, err
		}

		group.ParentID = req.ParentID
	}

	if req.Description != nil {
		group.Description = *req.Description
	}

	if req.Priority != nil {
		group.Priority = *req.Priority
	}

	if req.Metadata != nil {
		group.Metadata = req.Metadata
	}

	if err := s.groupRepo.Update(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

// Delete deletes a group by ID
// Delete performs soft delete on a group
func (s *GroupService) Delete(ctx context.Context, id string) error {
	// Check if group exists
	_, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return domainerrors.ErrNotFound
		}
		return err
	}

	// Check if group has children
	children, err := s.groupRepo.GetChildren(ctx, id)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return domainerrors.ErrInUse
	}

	return s.groupRepo.Delete(ctx, id)
}

// Purge permanently removes a group (hard delete, admin operation)
func (s *GroupService) Purge(ctx context.Context, id string) error {
	return s.groupRepo.Purge(ctx, id)
}

// Restore restores a soft-deleted group
func (s *GroupService) Restore(ctx context.Context, id string) error {
	return s.groupRepo.Restore(ctx, id)
}

// List retrieves a paginated list of groups
func (s *GroupService) List(ctx context.Context, pagination dto.PaginationRequest) ([]domain.Group, int, error) {
	pagination.Normalize()
	return s.groupRepo.List(ctx, pagination.Page, pagination.Limit)
}

// GetByNamespaceID retrieves groups by namespace ID
func (s *GroupService) GetByNamespaceID(ctx context.Context, namespaceID string) ([]domain.Group, error) {
	// Validate namespace exists
	_, err := s.namespaceRepo.GetByID(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.NewValidationError("namespace_id", "namespace not found")
		}
		return nil, err
	}
	return s.groupRepo.GetByNamespaceID(ctx, namespaceID)
}

// GetChildren retrieves child groups of a parent group
func (s *GroupService) GetChildren(ctx context.Context, parentID string) ([]domain.Group, error) {
	// Verify parent exists
	if _, err := s.GetByID(ctx, parentID); err != nil {
		return nil, err
	}
	return s.groupRepo.GetChildren(ctx, parentID)
}

// GetAncestors retrieves all ancestors of a group
func (s *GroupService) GetAncestors(ctx context.Context, id string) ([]domain.Group, error) {
	// Verify group exists
	if _, err := s.GetByID(ctx, id); err != nil {
		return nil, err
	}
	return s.groupRepo.GetAncestors(ctx, id)
}

// checkCircularReference checks if setting parentID would create a circular reference
func (s *GroupService) checkCircularReference(ctx context.Context, parentID, currentID string) error {
	// If we're setting the parent to be the current group itself
	if parentID == currentID {
		return domainerrors.ErrCircularReference
	}

	// Get all ancestors of the proposed parent
	ancestors, err := s.groupRepo.GetAncestors(ctx, parentID)
	if err != nil {
		return err
	}

	// Check if any ancestor is the current group (would create a cycle)
	for _, ancestor := range ancestors {
		if ancestor.ID == currentID {
			return domainerrors.ErrCircularReference
		}
	}

	return nil
}
