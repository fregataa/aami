package dto

import (
	"github.com/fregataa/aami/config-server/internal/domain"
)

// CreateGroupRequest represents a request to create a new group
type CreateGroupRequest struct {
	Name        string                 `json:"name" binding:"required,min=1,max=100"`
	NamespaceID string                 `json:"namespace_id" binding:"required,uuid"`
	ParentID    *string                `json:"parent_id,omitempty" binding:"omitempty,uuid"`
	Description string                 `json:"description" binding:"omitempty,max=500"`
	Priority    int                    `json:"priority" binding:"omitempty,min=0,max=1000"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateGroupRequest represents a request to update an existing group
type UpdateGroupRequest struct {
	Name        *string                `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	ParentID    *string                `json:"parent_id,omitempty" binding:"omitempty,uuid"`
	Description *string                `json:"description,omitempty" binding:"omitempty,max=500"`
	Priority    *int                   `json:"priority,omitempty" binding:"omitempty,min=0,max=1000"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NamespaceInfo represents namespace information in responses
type NamespaceInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	PolicyPriority int    `json:"policy_priority"`
}

// GroupResponse represents a group in API responses
type GroupResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	NamespaceID string                 `json:"namespace_id"`
	Namespace   *NamespaceInfo         `json:"namespace,omitempty"`
	ParentID    *string                `json:"parent_id,omitempty"`
	Description string                 `json:"description"`
	Priority    int                    `json:"priority"`
	Metadata    map[string]interface{} `json:"metadata"`
	TimestampResponse
}

// ToGroupResponse converts a domain.Group to GroupResponse
func ToGroupResponse(group *domain.Group) GroupResponse {
	resp := GroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		NamespaceID: group.NamespaceID,
		ParentID:    group.ParentID,
		Description: group.Description,
		Priority:    group.Priority,
		Metadata:    group.Metadata,
		TimestampResponse: TimestampResponse{
			CreatedAt: group.CreatedAt,
			UpdatedAt: group.UpdatedAt,
		},
	}

	// Include namespace info if loaded
	if group.Namespace != nil {
		resp.Namespace = &NamespaceInfo{
			ID:             group.Namespace.ID,
			Name:           group.Namespace.Name,
			PolicyPriority: group.Namespace.PolicyPriority,
		}
	}

	return resp
}

// ToGroupResponseList converts a slice of domain.Group to slice of GroupResponse
func ToGroupResponseList(groups []domain.Group) []GroupResponse {
	responses := make([]GroupResponse, len(groups))
	for i, group := range groups {
		responses[i] = ToGroupResponse(&group)
	}
	return responses
}
