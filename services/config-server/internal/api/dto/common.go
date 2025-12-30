package dto

import (
	"time"

	"github.com/fregataa/aami/config-server/internal/action"
)

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// PaginationRequest represents pagination parameters in API requests
type PaginationRequest struct {
	Page  int `form:"page" binding:"omitempty,min=1" json:"page"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100" json:"limit"`
}

// DefaultPagination returns default pagination values
func DefaultPagination() PaginationRequest {
	return PaginationRequest{
		Page:  1,
		Limit: 20,
	}
}

// Normalize ensures pagination values are within valid ranges
func (p *PaginationRequest) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 20
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
}

// ToAction converts PaginationRequest to action.Pagination
func (p *PaginationRequest) ToAction() action.Pagination {
	return action.NewPagination(p.Page, p.Limit)
}

// PaginationResponse represents pagination metadata in API responses
type PaginationResponse struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// NewPaginationResponse creates a pagination response from request and total count
func NewPaginationResponse(req PaginationRequest, total int) PaginationResponse {
	totalPages := total / req.Limit
	if total%req.Limit > 0 {
		totalPages++
	}
	return PaginationResponse{
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// ListResponse represents a generic paginated list response
type ListResponse struct {
	Data       interface{}        `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// TimestampResponse contains common timestamp fields
type TimestampResponse struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DeleteRequest represents a request to soft delete a resource
type DeleteRequest struct {
	ID string `json:"id" binding:"required,uuid"`
}

// PurgeRequest represents a request to permanently delete a resource
type PurgeRequest struct {
	ID string `json:"id" binding:"required,uuid"`
}

// RestoreRequest represents a request to restore a soft-deleted resource
type RestoreRequest struct {
	ID string `json:"id" binding:"required,uuid"`
}

// =============================================================================
// URI Parameter DTOs (for ShouldBindUri)
// =============================================================================

// IDUri represents a URI parameter for resource ID
type IDUri struct {
	ID string `uri:"id" binding:"required"`
}

// GroupIDUri represents a URI parameter for group ID
type GroupIDUri struct {
	GroupID string `uri:"group_id" binding:"required"`
}

// GroupIdUri represents a URI parameter for group ID (camelCase variant)
type GroupIdUri struct {
	GroupID string `uri:"groupId" binding:"required"`
}

// TargetIDUri represents a URI parameter for target ID
type TargetIDUri struct {
	TargetID string `uri:"target_id" binding:"required"`
}

// TargetIdUri represents a URI parameter for target ID (camelCase variant)
type TargetIdUri struct {
	TargetID string `uri:"targetId" binding:"required"`
}

// TemplateIDUri represents a URI parameter for template ID
type TemplateIDUri struct {
	TemplateID string `uri:"template_id" binding:"required"`
}

// TemplateIdUri represents a URI parameter for template ID (camelCase variant)
type TemplateIdUri struct {
	TemplateID string `uri:"templateId" binding:"required"`
}

// HostnameUri represents a URI parameter for hostname
type HostnameUri struct {
	Hostname string `uri:"hostname" binding:"required"`
}

// NameUri represents a URI parameter for name
type NameUri struct {
	Name string `uri:"name" binding:"required"`
}

// TokenUri represents a URI parameter for token
type TokenUri struct {
	Token string `uri:"token" binding:"required"`
}

// SeverityUri represents a URI parameter for severity
type SeverityUri struct {
	Severity string `uri:"severity" binding:"required"`
}

// TypeUri represents a URI parameter for type
type TypeUri struct {
	Type string `uri:"type" binding:"required"`
}

// ScriptTypeUri represents a URI parameter for script type
type ScriptTypeUri struct {
	ScriptType string `uri:"scriptType" binding:"required"`
}
