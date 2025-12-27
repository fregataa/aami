package dto

import (
	"github.com/fregataa/aami/config-server/internal/domain"
)

// CreateNamespaceRequest represents a request to create a new namespace
type CreateNamespaceRequest struct {
	Name           string `json:"name" binding:"required,min=1,max=50"`
	Description    string `json:"description" binding:"omitempty,max=500"`
	PolicyPriority int    `json:"policy_priority" binding:"required,min=1,max=1000"`
	MergeStrategy  string `json:"merge_strategy" binding:"required,oneof=override merge append"`
}

// UpdateNamespaceRequest represents a request to update an existing namespace
type UpdateNamespaceRequest struct {
	Description    *string `json:"description,omitempty" binding:"omitempty,max=500"`
	PolicyPriority *int    `json:"policy_priority,omitempty" binding:"omitempty,min=1,max=1000"`
	MergeStrategy  *string `json:"merge_strategy,omitempty" binding:"omitempty,oneof=override merge append"`
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

// ToNamespaceResponse converts a domain.Namespace to NamespaceResponse
func ToNamespaceResponse(namespace *domain.Namespace) NamespaceResponse {
	return NamespaceResponse{
		ID:             namespace.ID,
		Name:           namespace.Name,
		Description:    namespace.Description,
		PolicyPriority: namespace.PolicyPriority,
		MergeStrategy:  namespace.MergeStrategy,
		TimestampResponse: TimestampResponse{
			CreatedAt: namespace.CreatedAt,
			UpdatedAt: namespace.UpdatedAt,
		},
	}
}

// ToNamespaceResponseList converts a slice of domain.Namespace to slice of NamespaceResponse
func ToNamespaceResponseList(namespaces []domain.Namespace) []NamespaceResponse {
	responses := make([]NamespaceResponse, len(namespaces))
	for i, namespace := range namespaces {
		responses[i] = ToNamespaceResponse(&namespace)
	}
	return responses
}
