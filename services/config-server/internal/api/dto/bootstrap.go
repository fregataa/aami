package dto

import (
	"time"

	"github.com/fregataa/aami/config-server/internal/action"
)

// CreateBootstrapTokenRequest represents a request to create a new bootstrap token
type CreateBootstrapTokenRequest struct {
	Name      string            `json:"name" binding:"required,min=1,max=255"`
	MaxUses   int               `json:"max_uses" binding:"required,min=1"`
	ExpiresAt time.Time         `json:"expires_at" binding:"required"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// ToAction converts CreateBootstrapTokenRequest to action.CreateBootstrapToken
func (r *CreateBootstrapTokenRequest) ToAction() action.CreateBootstrapToken {
	return action.CreateBootstrapToken{
		Name:      r.Name,
		MaxUses:   r.MaxUses,
		ExpiresAt: r.ExpiresAt,
		Labels:    r.Labels,
	}
}

// UpdateBootstrapTokenRequest represents a request to update an existing bootstrap token
type UpdateBootstrapTokenRequest struct {
	Name      *string           `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	MaxUses   *int              `json:"max_uses,omitempty" binding:"omitempty,min=1"`
	ExpiresAt *time.Time        `json:"expires_at,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// ToAction converts UpdateBootstrapTokenRequest to action.UpdateBootstrapToken
func (r *UpdateBootstrapTokenRequest) ToAction() action.UpdateBootstrapToken {
	return action.UpdateBootstrapToken{
		Name:      r.Name,
		MaxUses:   r.MaxUses,
		ExpiresAt: r.ExpiresAt,
		Labels:    r.Labels,
	}
}

// ValidateTokenRequest represents a request to validate and use a bootstrap token
type ValidateTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// ToAction converts ValidateTokenRequest to action.ValidateToken
func (r *ValidateTokenRequest) ToAction() action.ValidateToken {
	return action.ValidateToken{
		Token: r.Token,
	}
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

// ToAction converts BootstrapRegisterRequest to action.BootstrapRegister
func (r *BootstrapRegisterRequest) ToAction() action.BootstrapRegister {
	return action.BootstrapRegister{
		Token:     r.Token,
		Hostname:  r.Hostname,
		IPAddress: r.IPAddress,
		GroupID:   r.GroupID,
		Labels:    r.Labels,
		Metadata:  r.Metadata,
	}
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

// ToBootstrapTokenResponse converts action.BootstrapTokenResult to BootstrapTokenResponse
func ToBootstrapTokenResponse(result action.BootstrapTokenResult) BootstrapTokenResponse {
	return BootstrapTokenResponse{
		ID:        result.ID,
		Token:     result.Token,
		Name:      result.Name,
		MaxUses:   result.MaxUses,
		Uses:      result.Uses,
		ExpiresAt: result.ExpiresAt,
		Labels:    result.Labels,
		IsValid:   result.IsValid,
		TimestampResponse: TimestampResponse{
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
		},
	}
}

// ToBootstrapTokenResponseList converts a slice of action.BootstrapTokenResult to slice of BootstrapTokenResponse
func ToBootstrapTokenResponseList(results []action.BootstrapTokenResult) []BootstrapTokenResponse {
	responses := make([]BootstrapTokenResponse, len(results))
	for i, result := range results {
		responses[i] = ToBootstrapTokenResponse(result)
	}
	return responses
}

// BootstrapRegisterResponse represents the response for node registration
type BootstrapRegisterResponse struct {
	Target        TargetResponse `json:"target"`
	TokenUsage    int            `json:"token_usage"`
	RemainingUses int            `json:"remaining_uses"`
}
