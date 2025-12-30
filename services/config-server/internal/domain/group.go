package domain

import "time"

// Group represents a hierarchical organizational unit
type Group struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
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

// GetPriority returns the priority for this group
func (g *Group) GetPriority() int {
	if g.Priority != 0 {
		return g.Priority
	}
	// Default priority
	return 100
}
