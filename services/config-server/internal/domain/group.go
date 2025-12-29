package domain

import "time"

// Group represents a hierarchical organizational unit within a namespace
type Group struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	NamespaceID  string            `json:"namespace_id"`
	Namespace    *Namespace        `json:"namespace,omitempty"`
	ParentID     *string           `json:"parent_id,omitempty"`
	Parent       *Group            `json:"-"`
	Children     []Group           `json:"-"`
	Description  string            `json:"description"`
	Priority     int               `json:"priority"`
	IsDefaultOwn bool              `json:"is_default_own"`
	Metadata     map[string]string `json:"metadata"`
	DeletedAt    *time.Time        `json:"deleted_at,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// IsRoot returns true if this group has no parent
func (g *Group) IsRoot() bool {
	return g.ParentID == nil
}

// GetPriority returns the calculated priority for this group
// Priority is based on namespace priority and can be overridden at group level
func (g *Group) GetPriority() int {
	if g.Priority != 0 {
		return g.Priority
	}
	// If namespace is loaded, use its priority
	if g.Namespace != nil {
		return g.Namespace.PolicyPriority
	}
	// Default priority
	return 100
}
