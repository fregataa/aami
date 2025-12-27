package domain

import "time"

// Namespace represents a dimension for organizing groups
// Namespaces enable multi-dimensional classification of infrastructure
type Namespace struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	PolicyPriority int        `json:"policy_priority"` // Lower = higher priority
	MergeStrategy  string     `json:"merge_strategy"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// MergeStrategy constants
const (
	MergeStrategyOverride = "override"
	MergeStrategyMerge    = "merge"
	MergeStrategyAppend   = "append"
)

// Default namespace names (for initial seeding)
const (
	NamespaceNameInfrastructure = "infrastructure"
	NamespaceNameLogical        = "logical"
	NamespaceNameEnvironment    = "environment"
)

// IsValidMergeStrategy checks if the merge strategy is valid
func IsValidMergeStrategy(strategy string) bool {
	switch strategy {
	case MergeStrategyOverride, MergeStrategyMerge, MergeStrategyAppend:
		return true
	default:
		return false
	}
}

// GetDefaultPriority returns the default priority for well-known namespace names
// This is used during seeding and validation
func GetDefaultPriority(name string) int {
	switch name {
	case NamespaceNameEnvironment:
		return 10 // Highest priority
	case NamespaceNameLogical:
		return 50
	case NamespaceNameInfrastructure:
		return 100
	default:
		return 100
	}
}
