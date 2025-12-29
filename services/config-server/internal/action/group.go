package action

import (
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
)

// ============================================================================
// Actions (Input)
// ============================================================================

// CreateGroup represents the action to create a group
type CreateGroup struct {
	Name        string
	NamespaceID string
	ParentID    *string
	Description string
	Priority    int
	Metadata    map[string]string
}

// UpdateGroup represents the action to update a group
// nil fields mean "do not update"
type UpdateGroup struct {
	Name        *string
	ParentID    *string
	Description *string
	Priority    *int
	Metadata    map[string]string
}

// ============================================================================
// Action Results (Output)
// ============================================================================

// GroupResult represents the result of group operations
type GroupResult struct {
	ID           string
	Name         string
	NamespaceID  string
	Namespace    *NamespaceResult
	ParentID     *string
	Description  string
	Priority     int
	IsDefaultOwn bool
	Metadata     map[string]string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// FromDomain converts domain.Group to GroupResult
func (r *GroupResult) FromDomain(g *domain.Group) {
	r.ID = g.ID
	r.Name = g.Name
	r.NamespaceID = g.NamespaceID
	r.ParentID = g.ParentID
	r.Description = g.Description
	r.Priority = g.Priority
	r.IsDefaultOwn = g.IsDefaultOwn
	r.Metadata = g.Metadata
	r.CreatedAt = g.CreatedAt
	r.UpdatedAt = g.UpdatedAt

	// Convert nested namespace if loaded
	if g.Namespace != nil {
		ns := NamespaceResult{}
		ns.FromDomain(g.Namespace)
		r.Namespace = &ns
	}
}

// NewGroupResult creates GroupResult from domain.Group
func NewGroupResult(g *domain.Group) GroupResult {
	var result GroupResult
	result.FromDomain(g)
	return result
}

// NewGroupResultList creates []GroupResult from []domain.Group
func NewGroupResultList(groups []domain.Group) []GroupResult {
	results := make([]GroupResult, len(groups))
	for i, g := range groups {
		results[i] = NewGroupResult(&g)
	}
	return results
}
