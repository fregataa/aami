package dto

import (
	"github.com/fregataa/aami/config-server/internal/action"
)

// CreateGroupRequest represents a request to create a new group
type CreateGroupRequest struct {
	Name        string            `json:"name" binding:"required,min=1,max=100"`
	NamespaceID string            `json:"namespace_id" binding:"required,uuid"`
	ParentID    *string           `json:"parent_id,omitempty" binding:"omitempty,uuid"`
	Description string            `json:"description" binding:"omitempty,max=500"`
	Priority    int               `json:"priority" binding:"omitempty,min=0,max=1000"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ToAction converts CreateGroupRequest to action.CreateGroup
func (r *CreateGroupRequest) ToAction() action.CreateGroup {
	return action.CreateGroup{
		Name:        r.Name,
		NamespaceID: r.NamespaceID,
		ParentID:    r.ParentID,
		Description: r.Description,
		Priority:    r.Priority,
		Metadata:    r.Metadata,
	}
}

// UpdateGroupRequest represents a request to update an existing group
type UpdateGroupRequest struct {
	Name        *string           `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	ParentID    *string           `json:"parent_id,omitempty" binding:"omitempty,uuid"`
	Description *string           `json:"description,omitempty" binding:"omitempty,max=500"`
	Priority    *int              `json:"priority,omitempty" binding:"omitempty,min=0,max=1000"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ToAction converts UpdateGroupRequest to action.UpdateGroup
func (r *UpdateGroupRequest) ToAction() action.UpdateGroup {
	return action.UpdateGroup{
		Name:        r.Name,
		ParentID:    r.ParentID,
		Description: r.Description,
		Priority:    r.Priority,
		Metadata:    r.Metadata,
	}
}

// NamespaceInfo represents namespace information in responses
type NamespaceInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	PolicyPriority int    `json:"policy_priority"`
}

// GroupResponse represents a group in API responses
type GroupResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	NamespaceID string            `json:"namespace_id"`
	Namespace   *NamespaceInfo    `json:"namespace,omitempty"`
	ParentID    *string           `json:"parent_id,omitempty"`
	Description string            `json:"description"`
	Priority    int               `json:"priority"`
	Metadata    map[string]string `json:"metadata"`
	TimestampResponse
}

// ToGroupResponse converts action.GroupResult to GroupResponse
func ToGroupResponse(result action.GroupResult) GroupResponse {
	resp := GroupResponse{
		ID:          result.ID,
		Name:        result.Name,
		NamespaceID: result.NamespaceID,
		ParentID:    result.ParentID,
		Description: result.Description,
		Priority:    result.Priority,
		Metadata:    result.Metadata,
		TimestampResponse: TimestampResponse{
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
		},
	}

	// Include namespace info if loaded
	if result.Namespace != nil {
		resp.Namespace = &NamespaceInfo{
			ID:             result.Namespace.ID,
			Name:           result.Namespace.Name,
			PolicyPriority: result.Namespace.PolicyPriority,
		}
	}

	return resp
}

// ToGroupResponseList converts a slice of action.GroupResult to slice of GroupResponse
func ToGroupResponseList(results []action.GroupResult) []GroupResponse {
	responses := make([]GroupResponse, len(results))
	for i, result := range results {
		responses[i] = ToGroupResponse(result)
	}
	return responses
}
