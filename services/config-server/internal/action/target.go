package action

import (
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
)

// ============================================================================
// Actions (Input)
// ============================================================================

// CreateTarget represents the action to create a target
type CreateTarget struct {
	Hostname  string
	IPAddress string
	GroupIDs  []string
	Status    domain.TargetStatus
	Labels    map[string]string
	Metadata  map[string]string
}

// UpdateTarget represents the action to update a target
// nil fields mean "do not update"
type UpdateTarget struct {
	Hostname  *string
	IPAddress *string
	Status    *domain.TargetStatus
	Labels    map[string]string
	Metadata  map[string]string
}

// UpdateTargetStatus represents the action to update target status only
type UpdateTargetStatus struct {
	Status domain.TargetStatus
}

// AddTargetToGroup represents the action to add a target to a group
type AddTargetToGroup struct {
	TargetID string
	GroupID  string
}

// RemoveTargetFromGroup represents the action to remove a target from a group
type RemoveTargetFromGroup struct {
	TargetID string
	GroupID  string
}

// ReplaceTargetGroups represents the action to replace all group mappings for a target
type ReplaceTargetGroups struct {
	GroupIDs []string
}

// ============================================================================
// Action Results (Output)
// ============================================================================

// TargetResult represents the result of target operations
type TargetResult struct {
	ID        string
	Hostname  string
	IPAddress string
	Groups    []GroupResult
	Status    domain.TargetStatus
	Exporters []ExporterResult
	Labels    map[string]string
	Metadata  map[string]string
	LastSeen  *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// FromDomain converts domain.Target to TargetResult
func (r *TargetResult) FromDomain(t *domain.Target) {
	r.ID = t.ID
	r.Hostname = t.Hostname
	r.IPAddress = t.IPAddress
	r.Status = t.Status
	r.Labels = t.Labels
	r.Metadata = t.Metadata
	r.LastSeen = t.LastSeen
	r.CreatedAt = t.CreatedAt
	r.UpdatedAt = t.UpdatedAt

	// Convert nested groups
	if len(t.Groups) > 0 {
		r.Groups = make([]GroupResult, len(t.Groups))
		for i, g := range t.Groups {
			r.Groups[i].FromDomain(&g)
		}
	}

	// Convert nested exporters
	if len(t.Exporters) > 0 {
		r.Exporters = make([]ExporterResult, len(t.Exporters))
		for i, e := range t.Exporters {
			r.Exporters[i].FromDomain(&e)
		}
	}
}

// NewTargetResult creates TargetResult from domain.Target
func NewTargetResult(t *domain.Target) TargetResult {
	var result TargetResult
	result.FromDomain(t)
	return result
}

// NewTargetResultList creates []TargetResult from []domain.Target
func NewTargetResultList(targets []domain.Target) []TargetResult {
	results := make([]TargetResult, len(targets))
	for i, t := range targets {
		results[i] = NewTargetResult(&t)
	}
	return results
}
