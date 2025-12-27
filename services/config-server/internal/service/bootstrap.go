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

// BootstrapTokenService handles business logic for bootstrap tokens
type BootstrapTokenService struct {
	tokenRepo repository.BootstrapTokenRepository
	groupRepo repository.GroupRepository
}

// NewBootstrapTokenService creates a new BootstrapTokenService
func NewBootstrapTokenService(
	tokenRepo repository.BootstrapTokenRepository,
	groupRepo repository.GroupRepository,
) *BootstrapTokenService {
	return &BootstrapTokenService{
		tokenRepo: tokenRepo,
		groupRepo: groupRepo,
	}
}

// Create creates a new bootstrap token
func (s *BootstrapTokenService) Create(ctx context.Context, req dto.CreateBootstrapTokenRequest) (*domain.BootstrapToken, error) {
	// Validate default group exists
	_, err := s.groupRepo.GetByID(ctx, req.DefaultGroupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewValidationError("default_group_id", "group not found")
		}
		return nil, err
	}

	// Generate token
	tokenStr, err := domain.GenerateToken()
	if err != nil {
		return nil, err
	}

	// Set default expiry if not provided
	expiresAt := req.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(30 * 24 * time.Hour) // 30 days default
	}

	token := &domain.BootstrapToken{
		ID:             uuid.New().String(),
		Token:          tokenStr,
		Name:           req.Name,
		DefaultGroupID: req.DefaultGroupID,
		MaxUses:        req.MaxUses,
		Uses:           0,
		ExpiresAt:      expiresAt,
		Labels:         req.Labels,
	}

	if token.MaxUses == 0 {
		token.MaxUses = 10 // Default max uses
	}
	if token.Labels == nil {
		token.Labels = make(map[string]string)
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return nil, err
	}

	return token, nil
}

// GetByID retrieves a bootstrap token by ID
func (s *BootstrapTokenService) GetByID(ctx context.Context, id string) (*domain.BootstrapToken, error) {
	token, err := s.tokenRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return token, nil
}

// GetByToken retrieves a bootstrap token by token string
func (s *BootstrapTokenService) GetByToken(ctx context.Context, tokenStr string) (*domain.BootstrapToken, error) {
	token, err := s.tokenRepo.GetByToken(ctx, tokenStr)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return token, nil
}

// ValidateAndUse validates a token and increments its usage counter
func (s *BootstrapTokenService) ValidateAndUse(ctx context.Context, req dto.ValidateTokenRequest) (*domain.BootstrapToken, error) {
	token, err := s.tokenRepo.GetByToken(ctx, req.Token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Check if token can be used
	if !token.CanUse() {
		return nil, NewValidationError("token", "token is expired or exhausted")
	}

	// Increment uses
	if err := token.IncrementUses(); err != nil {
		return nil, err
	}

	if err := s.tokenRepo.Update(ctx, token); err != nil {
		return nil, err
	}

	return token, nil
}

// Update updates an existing bootstrap token
func (s *BootstrapTokenService) Update(ctx context.Context, id string, req dto.UpdateBootstrapTokenRequest) (*domain.BootstrapToken, error) {
	token, err := s.tokenRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		token.Name = *req.Name
	}
	if req.MaxUses != nil {
		token.MaxUses = *req.MaxUses
	}
	if req.ExpiresAt != nil {
		token.ExpiresAt = *req.ExpiresAt
	}
	if req.Labels != nil {
		token.Labels = req.Labels
	}

	if err := s.tokenRepo.Update(ctx, token); err != nil {
		return nil, err
	}

	return token, nil
}

// Delete performs soft delete on a bootstrap token
func (s *BootstrapTokenService) Delete(ctx context.Context, id string) error {
	_, err := s.tokenRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
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
func (s *BootstrapTokenService) List(ctx context.Context, pagination dto.PaginationRequest) ([]domain.BootstrapToken, int, error) {
	pagination.Normalize()
	return s.tokenRepo.List(ctx, pagination.Page, pagination.Limit)
}

// GetByGroupID retrieves all bootstrap tokens for a group
func (s *BootstrapTokenService) GetByGroupID(ctx context.Context, groupID string) ([]domain.BootstrapToken, error) {
	return s.tokenRepo.GetByGroupID(ctx, groupID)
}
