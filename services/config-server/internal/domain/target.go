package domain

import "time"

// TargetStatus represents the operational status of a target
type TargetStatus string

const (
	// TargetStatusActive indicates the target is active and monitored
	TargetStatusActive TargetStatus = "active"

	// TargetStatusInactive indicates the target is inactive but registered
	TargetStatusInactive TargetStatus = "inactive"

	// TargetStatusDown indicates the target is unreachable
	TargetStatusDown TargetStatus = "down"
)

// IsValid checks if the status is one of the allowed values
func (s TargetStatus) IsValid() bool {
	switch s {
	case TargetStatusActive, TargetStatusInactive, TargetStatusDown:
		return true
	default:
		return false
	}
}

// Target represents a monitored server/node
type Target struct {
	ID        string            `json:"id"`
	Hostname  string            `json:"hostname"`
	IPAddress string            `json:"ip_address"`
	Groups    []Group           `json:"groups,omitempty"`
	Status    TargetStatus      `json:"status"`
	Exporters []Exporter        `json:"exporters,omitempty"`
	Labels    map[string]string `json:"labels"`
	Metadata  map[string]string `json:"metadata"`
	LastSeen  *time.Time        `json:"last_seen,omitempty"`
	DeletedAt *time.Time        `json:"deleted_at,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// GetAllGroups returns all groups this target belongs to
func (t *Target) GetAllGroups() []Group {
	return t.Groups
}

// GetGroupIDs returns the IDs of all groups this target belongs to
func (t *Target) GetGroupIDs() []string {
	ids := make([]string, len(t.Groups))
	for i, g := range t.Groups {
		ids[i] = g.ID
	}
	return ids
}

// HasGroup checks if the target belongs to a specific group
func (t *Target) HasGroup(groupID string) bool {
	for _, group := range t.Groups {
		if group.ID == groupID {
			return true
		}
	}
	return false
}

// HasDefaultOwnGroup checks if the target has a default own group
func (t *Target) HasDefaultOwnGroup() bool {
	for _, g := range t.Groups {
		if g.IsDefaultOwn {
			return true
		}
	}
	return false
}

// IsHealthy returns true if the target is active and recently seen
func (t *Target) IsHealthy() bool {
	if t.Status != TargetStatusActive {
		return false
	}
	if t.LastSeen == nil {
		return false
	}
	// Consider healthy if seen in last 5 minutes
	return time.Since(*t.LastSeen) < 5*time.Minute
}

// UpdateStatus updates the target status
func (t *Target) UpdateStatus(status TargetStatus) {
	t.Status = status
	t.UpdatedAt = time.Now()
}

// UpdateLastSeen updates the last seen timestamp
func (t *Target) UpdateLastSeen() {
	now := time.Now()
	t.LastSeen = &now
	t.UpdatedAt = now
}
