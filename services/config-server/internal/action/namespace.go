package action

import (
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
)

// ============================================================================
// Actions (Input)
// ============================================================================

// CreateNamespace represents the action to create a namespace
type CreateNamespace struct {
	Name           string
	Description    string
	PolicyPriority int
	MergeStrategy  string
}

// UpdateNamespace represents the action to update a namespace
// nil fields mean "do not update"
type UpdateNamespace struct {
	Description    *string
	PolicyPriority *int
	MergeStrategy  *string
}

// ============================================================================
// Action Results (Output)
// ============================================================================

// NamespaceResult represents the result of namespace operations
type NamespaceResult struct {
	ID             string
	Name           string
	Description    string
	PolicyPriority int
	MergeStrategy  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// FromDomain converts domain.Namespace to NamespaceResult
func (r *NamespaceResult) FromDomain(ns *domain.Namespace) {
	r.ID = ns.ID
	r.Name = ns.Name
	r.Description = ns.Description
	r.PolicyPriority = ns.PolicyPriority
	r.MergeStrategy = ns.MergeStrategy
	r.CreatedAt = ns.CreatedAt
	r.UpdatedAt = ns.UpdatedAt
}

// NewNamespaceResult creates NamespaceResult from domain.Namespace
func NewNamespaceResult(ns *domain.Namespace) NamespaceResult {
	var result NamespaceResult
	result.FromDomain(ns)
	return result
}

// NewNamespaceResultList creates []NamespaceResult from []domain.Namespace
func NewNamespaceResultList(namespaces []domain.Namespace) []NamespaceResult {
	results := make([]NamespaceResult, len(namespaces))
	for i, ns := range namespaces {
		results[i] = NewNamespaceResult(&ns)
	}
	return results
}

// NamespaceStats represents aggregated namespace statistics
// This is NOT a domain entity - it's a computed view
type NamespaceStats struct {
	NamespaceID    string
	NamespaceName  string
	GroupCount     int64
	TargetCount    int64
	PolicyPriority int
}
