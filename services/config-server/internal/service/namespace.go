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

// NamespaceService handles business logic for namespaces
type NamespaceService struct {
	namespaceRepo repository.NamespaceRepository
	groupRepo     repository.GroupRepository
	targetRepo    repository.TargetRepository
}

// NewNamespaceService creates a new NamespaceService
func NewNamespaceService(
	namespaceRepo repository.NamespaceRepository,
	groupRepo repository.GroupRepository,
	targetRepo repository.TargetRepository,
) *NamespaceService {
	return &NamespaceService{
		namespaceRepo: namespaceRepo,
		groupRepo:     groupRepo,
		targetRepo:    targetRepo,
	}
}

// Create creates a new namespace
func (s *NamespaceService) Create(ctx context.Context, req dto.CreateNamespaceRequest) (*domain.Namespace, error) {
	// Validate merge strategy
	if !domain.IsValidMergeStrategy(req.MergeStrategy) {
		return nil, domainerrors.NewValidationError("merge_strategy", "invalid merge strategy")
	}

	// Check if namespace name already exists
	existing, err := s.namespaceRepo.GetByName(ctx, req.Name)
	if err != nil && !errors.Is(err, domainerrors.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domainerrors.ErrAlreadyExists
	}

	namespace := &domain.Namespace{
		ID:             uuid.New().String(),
		Name:           req.Name,
		Description:    req.Description,
		PolicyPriority: req.PolicyPriority,
		MergeStrategy:  req.MergeStrategy,
	}

	if err := s.namespaceRepo.Create(ctx, namespace); err != nil {
		return nil, err
	}

	return namespace, nil
}

// GetByID retrieves a namespace by ID
func (s *NamespaceService) GetByID(ctx context.Context, id string) (*domain.Namespace, error) {
	namespace, err := s.namespaceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return namespace, nil
}

// GetByName retrieves a namespace by name
func (s *NamespaceService) GetByName(ctx context.Context, name string) (*domain.Namespace, error) {
	namespace, err := s.namespaceRepo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return namespace, nil
}

// Update updates an existing namespace
func (s *NamespaceService) Update(ctx context.Context, id string, req dto.UpdateNamespaceRequest) (*domain.Namespace, error) {
	namespace, err := s.namespaceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}

	// Update fields if provided
	if req.Description != nil {
		namespace.Description = *req.Description
	}
	if req.PolicyPriority != nil {
		namespace.PolicyPriority = *req.PolicyPriority
	}
	if req.MergeStrategy != nil {
		if !domain.IsValidMergeStrategy(*req.MergeStrategy) {
			return nil, domainerrors.NewValidationError("merge_strategy", "invalid merge strategy")
		}
		namespace.MergeStrategy = *req.MergeStrategy
	}

	if err := s.namespaceRepo.Update(ctx, namespace); err != nil {
		return nil, err
	}

	return namespace, nil
}

// Delete performs soft delete on a namespace
func (s *NamespaceService) Delete(ctx context.Context, id string) error {
	// Check if namespace exists
	_, err := s.namespaceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return domainerrors.ErrNotFound
		}
		return err
	}

	// Check if namespace is in use by groups
	groupCount, err := s.groupRepo.CountByNamespaceID(ctx, id)
	if err != nil {
		return err
	}
	if groupCount > 0 {
		return domainerrors.ErrInUse
	}

	return s.namespaceRepo.Delete(ctx, id)
}

// Purge permanently removes a namespace (hard delete, admin operation)
func (s *NamespaceService) Purge(ctx context.Context, id string) error {
	return s.namespaceRepo.Purge(ctx, id)
}

// Restore restores a soft-deleted namespace
func (s *NamespaceService) Restore(ctx context.Context, id string) error {
	return s.namespaceRepo.Restore(ctx, id)
}

// List retrieves a paginated list of namespaces
func (s *NamespaceService) List(ctx context.Context, pagination dto.PaginationRequest) ([]domain.Namespace, int, error) {
	pagination.Normalize()
	return s.namespaceRepo.List(ctx, pagination.Page, pagination.Limit)
}

// GetAll retrieves all namespaces without pagination
func (s *NamespaceService) GetAll(ctx context.Context) ([]domain.Namespace, error) {
	return s.namespaceRepo.GetAll(ctx)
}

// GetStats retrieves statistics for a namespace
func (s *NamespaceService) GetStats(ctx context.Context, id string) (*dto.NamespaceStatsResponse, error) {
	namespace, err := s.namespaceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}

	groupCount, err := s.groupRepo.CountByNamespaceID(ctx, id)
	if err != nil {
		return nil, err
	}

	targetCount, err := s.targetRepo.CountByNamespaceID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.NamespaceStatsResponse{
		NamespaceID:    namespace.ID,
		NamespaceName:  namespace.Name,
		GroupCount:     groupCount,
		TargetCount:    targetCount,
		PolicyPriority: namespace.PolicyPriority,
	}, nil
}

// GetAllStats retrieves statistics for all namespaces
func (s *NamespaceService) GetAllStats(ctx context.Context) ([]dto.NamespaceStatsResponse, error) {
	namespaces, err := s.namespaceRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	stats := make([]dto.NamespaceStatsResponse, len(namespaces))
	for i, ns := range namespaces {
		groupCount, err := s.groupRepo.CountByNamespaceID(ctx, ns.ID)
		if err != nil {
			return nil, err
		}

		targetCount, err := s.targetRepo.CountByNamespaceID(ctx, ns.ID)
		if err != nil {
			return nil, err
		}

		stats[i] = dto.NamespaceStatsResponse{
			NamespaceID:    ns.ID,
			NamespaceName:  ns.Name,
			GroupCount:     groupCount,
			TargetCount:    targetCount,
			PolicyPriority: ns.PolicyPriority,
		}
	}

	return stats, nil
}
