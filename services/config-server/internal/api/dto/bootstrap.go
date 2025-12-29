package dto

import (
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
)

// CreateBootstrapTokenRequest represents a request to create a new bootstrap token
type CreateBootstrapTokenRequest struct {
	Name      string            `json:"name" binding:"required,min=1,max=255"`
	MaxUses   int               `json:"max_uses" binding:"required,min=1"`
	ExpiresAt time.Time         `json:"expires_at" binding:"required"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// UpdateBootstrapTokenRequest represents a request to update an existing bootstrap token
type UpdateBootstrapTokenRequest struct {
	Name      *string           `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	MaxUses   *int              `json:"max_uses,omitempty" binding:"omitempty,min=1"`
	ExpiresAt *time.Time        `json:"expires_at,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// ValidateTokenRequest represents a request to validate and use a bootstrap token
type ValidateTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// BootstrapRegisterRequest represents a request to register a new node using bootstrap token
type BootstrapRegisterRequest struct {
	Token     string            `json:"token" binding:"required"`
	Hostname  string            `json:"hostname" binding:"required,min=1,max=255"`
	IPAddress string            `json:"ip_address" binding:"required,ip"`
	GroupID   string            `json:"group_id,omitempty" binding:"omitempty,uuid"`
	Labels    map[string]string `json:"labels,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// BootstrapTokenResponse represents a bootstrap token in API responses
type BootstrapTokenResponse struct {
	ID        string            `json:"id"`
	Token     string            `json:"token"`
	Name      string            `json:"name"`
	MaxUses   int               `json:"max_uses"`
	Uses      int               `json:"uses"`
	ExpiresAt time.Time         `json:"expires_at"`
	Labels    map[string]string `json:"labels"`
	IsValid   bool              `json:"is_valid"`
	TimestampResponse
}

// ToBootstrapTokenResponse converts a domain.BootstrapToken to BootstrapTokenResponse
func ToBootstrapTokenResponse(token *domain.BootstrapToken) BootstrapTokenResponse {
	return BootstrapTokenResponse{
		ID:        token.ID,
		Token:     token.Token,
		Name:      token.Name,
		MaxUses:   token.MaxUses,
		Uses:      token.Uses,
		ExpiresAt: token.ExpiresAt,
		Labels:    token.Labels,
		IsValid:   token.IsValid(),
		TimestampResponse: TimestampResponse{
			CreatedAt: token.CreatedAt,
			UpdatedAt: token.UpdatedAt,
		},
	}
}

// ToBootstrapTokenResponseList converts a slice of domain.BootstrapToken to slice of BootstrapTokenResponse
func ToBootstrapTokenResponseList(tokens []domain.BootstrapToken) []BootstrapTokenResponse {
	responses := make([]BootstrapTokenResponse, len(tokens))
	for i, token := range tokens {
		responses[i] = ToBootstrapTokenResponse(&token)
	}
	return responses
}

// BootstrapRegisterResponse represents the response for node registration
type BootstrapRegisterResponse struct {
	Target        TargetResponse `json:"target"`
	TokenUsage    int            `json:"token_usage"`
	RemainingUses int            `json:"remaining_uses"`
}
