package service

import (
	"context"
	"errors"
	"time"

	"github.com/fregataa/aami/config-server/internal/action"
	"github.com/fregataa/aami/config-server/internal/domain"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
	"github.com/fregataa/aami/config-server/internal/repository"
	"github.com/google/uuid"
)

// BootstrapTokenService handles business logic for bootstrap tokens
type BootstrapTokenService struct {
	tokenRepo     repository.BootstrapTokenRepository
	groupRepo     repository.GroupRepository
	targetService *TargetService
}

// NewBootstrapTokenService creates a new BootstrapTokenService
func NewBootstrapTokenService(
	tokenRepo repository.BootstrapTokenRepository,
	groupRepo repository.GroupRepository,
	targetService *TargetService,
) *BootstrapTokenService {
	return &BootstrapTokenService{
		tokenRepo:     tokenRepo,
		groupRepo:     groupRepo,
		targetService: targetService,
	}
}

// Create creates a new bootstrap token
func (s *BootstrapTokenService) Create(ctx context.Context, act action.CreateBootstrapToken) (action.BootstrapTokenResult, error) {
	// Generate token
	tokenStr, err := domain.GenerateToken()
	if err != nil {
		return action.BootstrapTokenResult{}, err
	}

	// Set default expiry if not provided
	expiresAt := act.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(30 * 24 * time.Hour) // 30 days default
	}

	token := &domain.BootstrapToken{
		ID:        uuid.New().String(),
		Token:     tokenStr,
		Name:      act.Name,
		MaxUses:   act.MaxUses,
		Uses:      0,
		ExpiresAt: expiresAt,
		Labels:    act.Labels,
	}

	if token.MaxUses == 0 {
		token.MaxUses = 10 // Default max uses
	}
	if token.Labels == nil {
		token.Labels = make(map[string]string)
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return action.BootstrapTokenResult{}, err
	}

	return action.NewBootstrapTokenResult(token), nil
}

// GetByID retrieves a bootstrap token by ID
func (s *BootstrapTokenService) GetByID(ctx context.Context, id string) (action.BootstrapTokenResult, error) {
	token, err := s.tokenRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.BootstrapTokenResult{}, domainerrors.ErrNotFound
		}
		return action.BootstrapTokenResult{}, err
	}
	return action.NewBootstrapTokenResult(token), nil
}

// GetByToken retrieves a bootstrap token by token string
func (s *BootstrapTokenService) GetByToken(ctx context.Context, tokenStr string) (action.BootstrapTokenResult, error) {
	token, err := s.tokenRepo.GetByToken(ctx, tokenStr)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.BootstrapTokenResult{}, domainerrors.ErrNotFound
		}
		return action.BootstrapTokenResult{}, err
	}
	return action.NewBootstrapTokenResult(token), nil
}

// ValidateAndUse validates a token and increments its usage counter
func (s *BootstrapTokenService) ValidateAndUse(ctx context.Context, act action.ValidateToken) (action.BootstrapTokenResult, error) {
	token, err := s.tokenRepo.GetByToken(ctx, act.Token)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.BootstrapTokenResult{}, domainerrors.ErrNotFound
		}
		return action.BootstrapTokenResult{}, err
	}

	// Check if token can be used
	if !token.CanUse() {
		return action.BootstrapTokenResult{}, domainerrors.NewValidationError("token", "token is expired or exhausted")
	}

	// Increment uses
	if err := token.IncrementUses(); err != nil {
		return action.BootstrapTokenResult{}, err
	}

	if err := s.tokenRepo.Update(ctx, token); err != nil {
		return action.BootstrapTokenResult{}, err
	}

	return action.NewBootstrapTokenResult(token), nil
}

// Update updates an existing bootstrap token
func (s *BootstrapTokenService) Update(ctx context.Context, id string, act action.UpdateBootstrapToken) (action.BootstrapTokenResult, error) {
	token, err := s.tokenRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.BootstrapTokenResult{}, domainerrors.ErrNotFound
		}
		return action.BootstrapTokenResult{}, err
	}

	if act.Name != nil {
		token.Name = *act.Name
	}
	if act.MaxUses != nil {
		token.MaxUses = *act.MaxUses
	}
	if act.ExpiresAt != nil {
		token.ExpiresAt = *act.ExpiresAt
	}
	if act.Labels != nil {
		token.Labels = act.Labels
	}

	if err := s.tokenRepo.Update(ctx, token); err != nil {
		return action.BootstrapTokenResult{}, err
	}

	return action.NewBootstrapTokenResult(token), nil
}

// Delete performs soft delete on a bootstrap token
func (s *BootstrapTokenService) Delete(ctx context.Context, id string) error {
	_, err := s.tokenRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return domainerrors.ErrNotFound
		}
		return err
	}

	return s.tokenRepo.Delete(ctx, id)
}

// Purge permanently removes a bootstrap token (hard delete, admin operation)
func (s *BootstrapTokenService) Purge(ctx context.Context, id string) error {
	return s.tokenRepo.Purge(ctx, id)
}

// Restore restores a soft-deleted bootstrap token
func (s *BootstrapTokenService) Restore(ctx context.Context, id string) error {
	return s.tokenRepo.Restore(ctx, id)
}

// List retrieves a paginated list of bootstrap tokens
func (s *BootstrapTokenService) List(ctx context.Context, pagination action.Pagination) (action.ListResult[action.BootstrapTokenResult], error) {
	tokens, total, err := s.tokenRepo.List(ctx, pagination.Page, pagination.Limit)
	if err != nil {
		return action.ListResult[action.BootstrapTokenResult]{}, err
	}

	results := action.NewBootstrapTokenResultList(tokens)
	return action.NewListResult(results, pagination, total), nil
}

// RegisterNode validates token and creates target with specified or own group
func (s *BootstrapTokenService) RegisterNode(
	ctx context.Context,
	act action.BootstrapRegister,
) (action.TargetResult, action.BootstrapTokenResult, error) {
	// 1. Validate and use token
	tokenResult, err := s.ValidateAndUse(ctx, action.ValidateToken{
		Token: act.Token,
	})
	if err != nil {
		return action.TargetResult{}, action.BootstrapTokenResult{}, err
	}

	// 2. Prepare group IDs
	var groupIDs []string
	if act.GroupID != "" {
		// Use specified group
		groupIDs = []string{act.GroupID}
	}
	// If GroupID is empty, target service will create own group automatically

	// 3. Create target
	targetResult, err := s.targetService.Create(ctx, action.CreateTarget{
		Hostname:  act.Hostname,
		IPAddress: act.IPAddress,
		GroupIDs:  groupIDs,
		Labels:    act.Labels,
		Metadata:  act.Metadata,
	})
	if err != nil {
		// Rollback: Decrement token usage on failure
		// Need to get token from repo to rollback
		token, getErr := s.tokenRepo.GetByToken(ctx, act.Token)
		if getErr == nil {
			token.Uses--
			_ = s.tokenRepo.Update(ctx, token)
		}
		return action.TargetResult{}, action.BootstrapTokenResult{}, err
	}

	return targetResult, tokenResult, nil
}
