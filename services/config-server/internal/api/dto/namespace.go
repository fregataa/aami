package dto

import (
	"github.com/fregataa/aami/config-server/internal/action"
)

// CreateNamespaceRequest represents a request to create a new namespace
type CreateNamespaceRequest struct {
	Name           string `json:"name" binding:"required,min=1,max=50"`
	Description    string `json:"description" binding:"omitempty,max=500"`
	PolicyPriority int    `json:"policy_priority" binding:"required,min=1,max=1000"`
	MergeStrategy  string `json:"merge_strategy" binding:"required,oneof=override merge append"`
}

// ToAction converts CreateNamespaceRequest to action.CreateNamespace
func (r *CreateNamespaceRequest) ToAction() action.CreateNamespace {
	return action.CreateNamespace{
		Name:           r.Name,
		Description:    r.Description,
		PolicyPriority: r.PolicyPriority,
		MergeStrategy:  r.MergeStrategy,
	}
}

// UpdateNamespaceRequest represents a request to update an existing namespace
type UpdateNamespaceRequest struct {
	Description    *string `json:"description,omitempty" binding:"omitempty,max=500"`
	PolicyPriority *int    `json:"policy_priority,omitempty" binding:"omitempty,min=1,max=1000"`
	MergeStrategy  *string `json:"merge_strategy,omitempty" binding:"omitempty,oneof=override merge append"`
}

// ToAction converts UpdateNamespaceRequest to action.UpdateNamespace
func (r *UpdateNamespaceRequest) ToAction() action.UpdateNamespace {
	return action.UpdateNamespace{
		Description:    r.Description,
		PolicyPriority: r.PolicyPriority,
		MergeStrategy:  r.MergeStrategy,
	}
}

// NamespaceResponse represents a namespace in API responses
type NamespaceResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	PolicyPriority int    `json:"policy_priority"`
	MergeStrategy  string `json:"merge_strategy"`
	TimestampResponse
}

// NamespaceStatsResponse represents namespace statistics
type NamespaceStatsResponse struct {
	NamespaceID    string `json:"namespace_id"`
	NamespaceName  string `json:"namespace_name"`
	GroupCount     int64  `json:"group_count"`
	TargetCount    int64  `json:"target_count"`
	PolicyPriority int    `json:"policy_priority"`
}

// ToNamespaceResponse converts action.NamespaceResult to NamespaceResponse
func ToNamespaceResponse(result action.NamespaceResult) NamespaceResponse {
	return NamespaceResponse{
		ID:             result.ID,
		Name:           result.Name,
		Description:    result.Description,
		PolicyPriority: result.PolicyPriority,
		MergeStrategy:  result.MergeStrategy,
		TimestampResponse: TimestampResponse{
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
		},
	}
}

// ToNamespaceResponseList converts a slice of action.NamespaceResult to slice of NamespaceResponse
func ToNamespaceResponseList(results []action.NamespaceResult) []NamespaceResponse {
	responses := make([]NamespaceResponse, len(results))
	for i, result := range results {
		responses[i] = ToNamespaceResponse(result)
	}
	return responses
}
