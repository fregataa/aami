package dto

import (
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
)

// CreateBootstrapTokenRequest represents a request to create a new bootstrap token
type CreateBootstrapTokenRequest struct {
	Name           string            `json:"name" binding:"required,min=1,max=255"`
	DefaultGroupID string            `json:"default_group_id" binding:"required,uuid"`
	MaxUses        int               `json:"max_uses" binding:"required,min=1"`
	ExpiresAt      time.Time         `json:"expires_at" binding:"required"`
	Labels         map[string]string `json:"labels,omitempty"`
}

// UpdateBootstrapTokenRequest represents a request to update an existing bootstrap token
type UpdateBootstrapTokenRequest struct {
	Name           *string           `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	DefaultGroupID *string           `json:"default_group_id,omitempty" binding:"omitempty,uuid"`
	MaxUses        *int              `json:"max_uses,omitempty" binding:"omitempty,min=1"`
	ExpiresAt      *time.Time        `json:"expires_at,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
}

// ValidateTokenRequest represents a request to validate and use a bootstrap token
type ValidateTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// BootstrapTokenResponse represents a bootstrap token in API responses
type BootstrapTokenResponse struct {
	ID             string         `json:"id"`
	Token          string         `json:"token"`
	Name           string         `json:"name"`
	DefaultGroupID string         `json:"default_group_id"`
	DefaultGroup   *GroupResponse `json:"default_group,omitempty"`
	MaxUses        int            `json:"max_uses"`
	Uses           int            `json:"uses"`
	ExpiresAt      time.Time      `json:"expires_at"`
	Labels         map[string]string `json:"labels"`
	IsValid        bool           `json:"is_valid"`
	TimestampResponse
}

// ToBootstrapTokenResponse converts a domain.BootstrapToken to BootstrapTokenResponse
func ToBootstrapTokenResponse(token *domain.BootstrapToken) BootstrapTokenResponse {
	resp := BootstrapTokenResponse{
		ID:             token.ID,
		Token:          token.Token,
		Name:           token.Name,
		DefaultGroupID: token.DefaultGroupID,
		MaxUses:        token.MaxUses,
		Uses:           token.Uses,
		ExpiresAt:      token.ExpiresAt,
		Labels:         token.Labels,
		IsValid:        token.IsValid(),
		TimestampResponse: TimestampResponse{
			CreatedAt: token.CreatedAt,
			UpdatedAt: token.UpdatedAt,
		},
	}

	// Include default group if loaded
	if token.DefaultGroup.ID != "" {
		group := ToGroupResponse(&token.DefaultGroup)
		resp.DefaultGroup = &group
	}

	return resp
}

// ToBootstrapTokenResponseList converts a slice of domain.BootstrapToken to slice of BootstrapTokenResponse
func ToBootstrapTokenResponseList(tokens []domain.BootstrapToken) []BootstrapTokenResponse {
	responses := make([]BootstrapTokenResponse, len(tokens))
	for i, token := range tokens {
		responses[i] = ToBootstrapTokenResponse(&token)
	}
	return responses
}
