package client

import "time"

// Common types

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

// Namespace types

// Namespace represents a namespace
type Namespace struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	PolicyPriority int                    `json:"policy_priority"`
	MergeStrategy  string                 `json:"merge_strategy"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// CreateNamespaceRequest represents a create namespace request
type CreateNamespaceRequest struct {
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	PolicyPriority int                    `json:"policy_priority"`
	MergeStrategy  string                 `json:"merge_strategy,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateNamespaceRequest represents an update namespace request
type UpdateNamespaceRequest struct {
	Name           *string                 `json:"name,omitempty"`
	Description    *string                 `json:"description,omitempty"`
	PolicyPriority *int                    `json:"policy_priority,omitempty"`
	MergeStrategy  *string                 `json:"merge_strategy,omitempty"`
	Metadata       *map[string]interface{} `json:"metadata,omitempty"`
}

// Group types

// Group represents a group
type Group struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	NamespaceID  string                 `json:"namespace_id"`
	ParentID     *string                `json:"parent_id,omitempty"`
	Description  string                 `json:"description"`
	Priority     int                    `json:"priority"`
	IsDefaultOwn bool                   `json:"is_default_own"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// CreateGroupRequest represents a create group request
type CreateGroupRequest struct {
	Name        string                  `json:"name"`
	NamespaceID string                  `json:"namespace_id"`
	ParentID    *string                 `json:"parent_id,omitempty"`
	Description string                  `json:"description,omitempty"`
	Priority    int                     `json:"priority"`
	Metadata    *map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateGroupRequest represents an update group request
type UpdateGroupRequest struct {
	Name        *string                 `json:"name,omitempty"`
	ParentID    *string                 `json:"parent_id,omitempty"`
	Description *string                 `json:"description,omitempty"`
	Priority    *int                    `json:"priority,omitempty"`
	Metadata    *map[string]interface{} `json:"metadata,omitempty"`
}

// Target types

// Target represents a target
type Target struct {
	ID         string                 `json:"id"`
	Hostname   string                 `json:"hostname"`
	IPAddress  string                 `json:"ip_address"`
	Status     string                 `json:"status"`
	Labels     map[string]string      `json:"labels,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	LastSeen   *time.Time             `json:"last_seen,omitempty"`
	Groups     []Group                `json:"groups,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// CreateTargetRequest represents a create target request
type CreateTargetRequest struct {
	Hostname  string                  `json:"hostname"`
	IPAddress string                  `json:"ip_address"`
	GroupIDs  []string                `json:"group_ids,omitempty"`
	Labels    *map[string]string      `json:"labels,omitempty"`
	Metadata  *map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTargetRequest represents an update target request
type UpdateTargetRequest struct {
	Hostname  *string                 `json:"hostname,omitempty"`
	IPAddress *string                 `json:"ip_address,omitempty"`
	GroupIDs  *[]string               `json:"group_ids,omitempty"`
	Labels    *map[string]string      `json:"labels,omitempty"`
	Metadata  *map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTargetStatusRequest represents an update target status request
type UpdateTargetStatusRequest struct {
	Status string `json:"status"`
}

// Bootstrap Token types

// BootstrapToken represents a bootstrap token
type BootstrapToken struct {
	ID        string            `json:"id"`
	Token     string            `json:"token"`
	Name      string            `json:"name"`
	MaxUses   int               `json:"max_uses"`
	Uses      int               `json:"uses"`
	ExpiresAt time.Time         `json:"expires_at"`
	Labels    map[string]string `json:"labels,omitempty"`
	IsValid   bool              `json:"is_valid"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// CreateBootstrapTokenRequest represents a create bootstrap token request
type CreateBootstrapTokenRequest struct {
	Name      string            `json:"name"`
	MaxUses   int               `json:"max_uses"`
	ExpiresAt time.Time         `json:"expires_at"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// UpdateBootstrapTokenRequest represents an update bootstrap token request
type UpdateBootstrapTokenRequest struct {
	Name      *string           `json:"name,omitempty"`
	MaxUses   *int              `json:"max_uses,omitempty"`
	ExpiresAt *time.Time        `json:"expires_at,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// ValidateTokenRequest represents a validate token request
type ValidateTokenRequest struct {
	Token string `json:"token"`
}

// BootstrapRegisterRequest represents a bootstrap register request
type BootstrapRegisterRequest struct {
	Token     string                 `json:"token"`
	Hostname  string                 `json:"hostname"`
	IPAddress string                 `json:"ip_address"`
	GroupID   string                 `json:"group_id,omitempty"`
	Labels    map[string]string      `json:"labels,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// BootstrapRegisterResponse represents a bootstrap register response
type BootstrapRegisterResponse struct {
	Target        Target `json:"target"`
	TokenUsage    int    `json:"token_usage"`
	RemainingUses int    `json:"remaining_uses"`
}
