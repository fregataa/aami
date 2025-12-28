package domain

import (
	"time"
)

// CheckInstance represents an application of a check at a specific scope
// Instance is independent from template - it contains a snapshot of template at creation time
type CheckInstance struct {
	ID string `json:"id"`

	// Template fields (copied at creation time - deep copy from CheckTemplate)
	Name          string                 `json:"name"`
	CheckType     string                 `json:"check_type"`
	ScriptContent string                 `json:"script_content"`
	Language      string                 `json:"language"`
	DefaultConfig map[string]interface{} `json:"default_config"`
	Description   string                 `json:"description"`
	Version       string                 `json:"version"`
	Hash          string                 `json:"hash"` // SHA256 hash of script_content

	// Instance-specific fields
	Scope       InstanceScope          `json:"scope"` // "global", "namespace", "group"
	NamespaceID *string                `json:"namespace_id,omitempty"`
	Namespace   *Namespace             `json:"namespace,omitempty"`
	GroupID     *string                `json:"group_id,omitempty"`
	Group       *Group                 `json:"group,omitempty"`
	Config      map[string]interface{} `json:"config"`   // Override parameters
	Priority    int                    `json:"priority"` // Higher number = higher priority
	IsActive    bool                   `json:"is_active"`

	// Metadata (optional, for tracking origin)
	CreatedFromTemplateID   *string `json:"created_from_template_id,omitempty"`
	CreatedFromTemplateName *string `json:"created_from_template_name,omitempty"`
	TemplateVersion         *string `json:"template_version,omitempty"`

	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// InstanceScope represents the scope level of a check instance
type InstanceScope string

const (
	ScopeGlobal    InstanceScope = "global"
	ScopeNamespace InstanceScope = "namespace"
	ScopeGroup     InstanceScope = "group"
)

// GetScope returns the scope level of this check instance
func (ci *CheckInstance) GetScope() InstanceScope {
	return ci.Scope
}

// IsGlobal returns true if this is a global-scope instance
func (ci *CheckInstance) IsGlobal() bool {
	return ci.Scope == ScopeGlobal
}

// IsNamespaceLevel returns true if this is a namespace-level instance
func (ci *CheckInstance) IsNamespaceLevel() bool {
	return ci.Scope == ScopeNamespace
}

// IsGroupLevel returns true if this is a group-level instance
func (ci *CheckInstance) IsGroupLevel() bool {
	return ci.Scope == ScopeGroup
}

// MergeConfig merges instance's default config with override config
// Instance config takes precedence over default config
func (ci *CheckInstance) MergeConfig() map[string]interface{} {
	merged := make(map[string]interface{})

	// Start with instance's default config (copied from template at creation)
	for k, v := range ci.DefaultConfig {
		merged[k] = deepCopyValue(v)
	}

	// Override with instance config
	for k, v := range ci.Config {
		merged[k] = deepCopyValue(v)
	}

	return merged
}

// GetScopeIdentifier returns a string representation of the scope
// Used for logging and debugging
func (ci *CheckInstance) GetScopeIdentifier() string {
	switch ci.Scope {
	case ScopeGlobal:
		return "global"
	case ScopeNamespace:
		if ci.NamespaceID != nil {
			return "namespace:" + *ci.NamespaceID
		}
		return "namespace:unknown"
	case ScopeGroup:
		if ci.GroupID != nil {
			return "group:" + *ci.GroupID
		}
		return "group:unknown"
	default:
		return "unknown"
	}
}

// Validate performs basic validation on the check instance
func (ci *CheckInstance) Validate() error {
	// Validate required fields
	if ci.Name == "" {
		return NewValidationError("name", "name is required")
	}
	if ci.CheckType == "" {
		return NewValidationError("check_type", "check_type is required")
	}
	if ci.ScriptContent == "" {
		return NewValidationError("script_content", "script_content is required")
	}
	if ci.Language == "" {
		return NewValidationError("language", "language is required")
	}
	if ci.Version == "" {
		return NewValidationError("version", "version is required")
	}
	if ci.Hash == "" {
		return NewValidationError("hash", "hash is required")
	}

	if ci.Scope == "" {
		return NewValidationError("scope", "scope is required")
	}

	// Validate scope consistency
	switch ci.Scope {
	case ScopeGlobal:
		if ci.NamespaceID != nil || ci.GroupID != nil {
			return NewValidationError("scope", "global scope must not have namespace_id or group_id")
		}
	case ScopeNamespace:
		if ci.NamespaceID == nil {
			return NewValidationError("namespace_id", "namespace_id is required for namespace scope")
		}
		if ci.GroupID != nil {
			return NewValidationError("group_id", "namespace scope must not have group_id")
		}
	case ScopeGroup:
		if ci.GroupID == nil {
			return NewValidationError("group_id", "group_id is required for group scope")
		}
		if ci.NamespaceID == nil {
			return NewValidationError("namespace_id", "namespace_id is required for group scope")
		}
	default:
		return NewValidationError("scope", "invalid scope value")
	}

	return nil
}

// EffectiveCheck represents the resolved check for a node
type EffectiveCheck struct {
	Name          string                 `json:"name"`
	CheckType     string                 `json:"check_type"`
	ScriptContent string                 `json:"script_content"`
	Language      string                 `json:"language"`
	Config        map[string]interface{} `json:"config"` // Merged config
	Version       string                 `json:"version"`
	Hash          string                 `json:"hash"`
	InstanceID    string                 `json:"instance_id"`
}

// EffectiveCheckInstancesResult contains effective check instances grouped by scope
// Used by repository to return structured data for target's effective checks
type EffectiveCheckInstancesResult struct {
	NamespaceInstances []CheckInstance `json:"namespace_instances"`
	GroupInstances     []CheckInstance `json:"group_instances"`
}

// NewCheckInstanceFromTemplate creates a new CheckInstance from a CheckTemplate
// This performs a deep copy of the template's fields into the instance
func NewCheckInstanceFromTemplate(
	template *CheckTemplate,
	scope InstanceScope,
	namespaceID *string,
	groupID *string,
	overrideConfig map[string]interface{},
) *CheckInstance {
	// Deep copy template's default config
	defaultConfig := make(map[string]interface{})
	for k, v := range template.DefaultConfig {
		defaultConfig[k] = deepCopyValue(v)
	}

	// Deep copy override config
	config := make(map[string]interface{})
	for k, v := range overrideConfig {
		config[k] = deepCopyValue(v)
	}

	now := time.Now()

	return &CheckInstance{
		// Template fields (deep copied)
		Name:          template.Name,
		CheckType:     template.CheckType,
		ScriptContent: template.ScriptContent,
		Language:      template.Language,
		DefaultConfig: defaultConfig,
		Description:   template.Description,
		Version:       template.Version,
		Hash:          template.Hash,

		// Instance-specific fields
		Scope:       scope,
		NamespaceID: namespaceID,
		GroupID:     groupID,
		Config:      config,
		Priority:    0,
		IsActive:    true,

		// Metadata (track origin)
		CreatedFromTemplateID:   &template.ID,
		CreatedFromTemplateName: &template.Name,
		TemplateVersion:         &template.Version,

		CreatedAt: now,
		UpdatedAt: now,
	}
}

// deepCopyValue performs a deep copy of interface{} values
// Handles maps and slices recursively
func deepCopyValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case map[string]interface{}:
		newMap := make(map[string]interface{}, len(val))
		for k, v := range val {
			newMap[k] = deepCopyValue(v)
		}
		return newMap
	case []interface{}:
		newSlice := make([]interface{}, len(val))
		for i, item := range val {
			newSlice[i] = deepCopyValue(item)
		}
		return newSlice
	case map[interface{}]interface{}:
		newMap := make(map[interface{}]interface{}, len(val))
		for k, v := range val {
			newMap[k] = deepCopyValue(v)
		}
		return newMap
	default:
		// For basic types (string, int, float, bool), direct assignment is safe
		return v
	}
}
