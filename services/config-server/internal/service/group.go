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

// GroupService handles business logic for groups
type GroupService struct {
	groupRepo repository.GroupRepository
}

// NewGroupService creates a new GroupService
func NewGroupService(groupRepo repository.GroupRepository) *GroupService {
	return &GroupService{
		groupRepo: groupRepo,
	}
}

// Create creates a new group
func (s *GroupService) Create(ctx context.Context, act action.CreateGroup) (action.GroupResult, error) {
	// Validate parent exists if specified
	if act.ParentID != nil {
		_, err := s.groupRepo.GetByID(ctx, *act.ParentID)
		if err != nil {
			if errors.Is(err, domainerrors.ErrNotFound) {
				return action.GroupResult{}, domainerrors.ErrForeignKeyViolation
			}
			return action.GroupResult{}, err
		}

		// Check for circular references
		if err := s.checkCircularReference(ctx, *act.ParentID, ""); err != nil {
			return action.GroupResult{}, err
		}
	}

	// Set default priority if not specified
	priority := act.Priority
	if priority == 0 {
		priority = 100 // Default priority
	}

	// Initialize metadata if nil
	metadata := act.Metadata
	if metadata == nil {
		metadata = make(map[string]string)
	}

	group := &domain.Group{
		ID:          uuid.New().String(),
		Name:        act.Name,
		ParentID:    act.ParentID,
		Description: act.Description,
		Priority:    priority,
		Metadata:    metadata,
	}

	if err := s.groupRepo.Create(ctx, group); err != nil {
		return action.GroupResult{}, err
	}

	return action.NewGroupResult(group), nil
}

// GetByID retrieves a group by ID
func (s *GroupService) GetByID(ctx context.Context, id string) (action.GroupResult, error) {
	group, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.GroupResult{}, domainerrors.ErrNotFound
		}
		return action.GroupResult{}, err
	}
	return action.NewGroupResult(group), nil
}

// Update updates an existing group
func (s *GroupService) Update(ctx context.Context, id string, act action.UpdateGroup) (action.GroupResult, error) {
	// Get existing group
	group, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.GroupResult{}, domainerrors.ErrNotFound
		}
		return action.GroupResult{}, err
	}

	// Update fields if provided
	if act.Name != nil {
		group.Name = *act.Name
	}

	if act.ParentID != nil {
		// Validate parent exists
		_, err := s.groupRepo.GetByID(ctx, *act.ParentID)
		if err != nil {
			if errors.Is(err, domainerrors.ErrNotFound) {
				return action.GroupResult{}, domainerrors.ErrForeignKeyViolation
			}
			return action.GroupResult{}, err
		}

		// Check for circular references
		if err := s.checkCircularReference(ctx, *act.ParentID, id); err != nil {
			return action.GroupResult{}, err
		}

		group.ParentID = act.ParentID
	}

	if act.Description != nil {
		group.Description = *act.Description
	}

	if act.Priority != nil {
		group.Priority = *act.Priority
	}

	if act.Metadata != nil {
		group.Metadata = act.Metadata
	}

	if err := s.groupRepo.Update(ctx, group); err != nil {
		return action.GroupResult{}, err
	}

	return action.NewGroupResult(group), nil
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
func (s *GroupService) List(ctx context.Context, pagination action.Pagination) (action.ListResult[action.GroupResult], error) {
	groups, total, err := s.groupRepo.List(ctx, pagination.Page, pagination.Limit)
	if err != nil {
		return action.ListResult[action.GroupResult]{}, err
	}

	results := action.NewGroupResultList(groups)
	return action.NewListResult(results, pagination, total), nil
}

// GetChildren retrieves child groups of a parent group
func (s *GroupService) GetChildren(ctx context.Context, parentID string) ([]action.GroupResult, error) {
	// Verify parent exists
	if _, err := s.GetByID(ctx, parentID); err != nil {
		return nil, err
	}
	groups, err := s.groupRepo.GetChildren(ctx, parentID)
	if err != nil {
		return nil, err
	}
	return action.NewGroupResultList(groups), nil
}

// GetAncestors retrieves all ancestors of a group
func (s *GroupService) GetAncestors(ctx context.Context, id string) ([]action.GroupResult, error) {
	// Verify group exists
	if _, err := s.GetByID(ctx, id); err != nil {
		return nil, err
	}
	groups, err := s.groupRepo.GetAncestors(ctx, id)
	if err != nil {
		return nil, err
	}
	return action.NewGroupResultList(groups), nil
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
