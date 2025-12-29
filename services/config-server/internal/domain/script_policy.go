package domain

import (
	"time"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
)

// ScriptPolicy represents an application of a check at a specific scope
// Instance is independent from template - it contains a snapshot of template at creation time
type ScriptPolicy struct {
	ID string `json:"id"`

	// Template fields (copied at creation time - deep copy from ScriptTemplate)
	Name          string                 `json:"name"`
	ScriptType     string                 `json:"script_type"`
	ScriptContent string                 `json:"script_content"`
	Language      string                 `json:"language"`
	DefaultConfig map[string]interface{} `json:"default_config"`
	Description   string                 `json:"description"`
	Version       string                 `json:"version"`
	Hash          string                 `json:"hash"` // SHA256 hash of script_content

	// Instance-specific fields
	Scope   PolicyScope `json:"scope"` // "global", "group"
	GroupID *string     `json:"group_id,omitempty"`
	Group   *Group      `json:"group,omitempty"`
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

// PolicyScope represents the scope level of a script policy
type PolicyScope string

const (
	ScopeGlobal PolicyScope = "global"
	ScopeGroup  PolicyScope = "group"
)

// GetScope returns the scope level of this script policy
func (sp *ScriptPolicy) GetScope() PolicyScope {
	return sp.Scope
}

// IsGlobal returns true if this is a global-scope instance
func (sp *ScriptPolicy) IsGlobal() bool {
	return sp.Scope == ScopeGlobal
}

// IsGroupLevel returns true if this is a group-level instance
func (sp *ScriptPolicy) IsGroupLevel() bool {
	return sp.Scope == ScopeGroup
}

// MergeConfig merges instance's default config with override config
// Instance config takes precedence over default config
func (sp *ScriptPolicy) MergeConfig() map[string]interface{} {
	merged := make(map[string]interface{})

	// Start with instance's default config (copied from template at creation)
	for k, v := range sp.DefaultConfig {
		merged[k] = deepCopyValue(v)
	}

	// Override with instance config
	for k, v := range sp.Config {
		merged[k] = deepCopyValue(v)
	}

	return merged
}

// GetScopeIdentifier returns a string representation of the scope
// Used for logging and debugging
func (sp *ScriptPolicy) GetScopeIdentifier() string {
	switch sp.Scope {
	case ScopeGlobal:
		return "global"
	case ScopeGroup:
		if sp.GroupID != nil {
			return "group:" + *sp.GroupID
		}
		return "group:unknown"
	default:
		return "unknown"
	}
}

// Validate performs basic validation on the script policy
func (sp *ScriptPolicy) Validate() error {
	// Validate required fields
	if sp.Name == "" {
		return domainerrors.NewValidationError("name", "name is required")
	}
	if sp.ScriptType == "" {
		return domainerrors.NewValidationError("script_type", "script_type is required")
	}
	if sp.ScriptContent == "" {
		return domainerrors.NewValidationError("script_content", "script_content is required")
	}
	if sp.Language == "" {
		return domainerrors.NewValidationError("language", "language is required")
	}
	if sp.Version == "" {
		return domainerrors.NewValidationError("version", "version is required")
	}
	if sp.Hash == "" {
		return domainerrors.NewValidationError("hash", "hash is required")
	}

	if sp.Scope == "" {
		return domainerrors.NewValidationError("scope", "scope is required")
	}

	// Validate scope consistency
	switch sp.Scope {
	case ScopeGlobal:
		if sp.GroupID != nil {
			return domainerrors.NewValidationError("scope", "global scope must not have group_id")
		}
	case ScopeGroup:
		if sp.GroupID == nil {
			return domainerrors.NewValidationError("group_id", "group_id is required for group scope")
		}
	default:
		return domainerrors.NewValidationError("scope", "invalid scope value: must be 'global' or 'group'")
	}

	return nil
}

// EffectiveCheck represents the resolved check for a node
type EffectiveCheck struct {
	Name          string                 `json:"name"`
	ScriptType     string                 `json:"script_type"`
	ScriptContent string                 `json:"script_content"`
	Language      string                 `json:"language"`
	Config        map[string]interface{} `json:"config"` // Merged config
	Version       string                 `json:"version"`
	Hash          string                 `json:"hash"`
	InstanceID    string                 `json:"instance_id"`
}

// EffectivePoliciesResult contains effective script policies grouped by scope
// Used by repository to return structured data for target's effective checks
type EffectivePoliciesResult struct {
	GlobalInstances []ScriptPolicy `json:"global_instances"`
	GroupInstances  []ScriptPolicy `json:"group_instances"`
}

// NewScriptPolicyFromTemplate creates a new ScriptPolicy from a ScriptTemplate
// This performs a deep copy of the script's fields into the instance
func NewScriptPolicyFromTemplate(
	template *ScriptTemplate,
	scope PolicyScope,
	groupID *string,
	overrideConfig map[string]interface{},
) *ScriptPolicy {
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

	return &ScriptPolicy{
		// Template fields (deep copied)
		Name:          template.Name,
		ScriptType:    template.ScriptType,
		ScriptContent: template.ScriptContent,
		Language:      template.Language,
		DefaultConfig: defaultConfig,
		Description:   template.Description,
		Version:       template.Version,
		Hash:          template.Hash,

		// Instance-specific fields
		Scope:    scope,
		GroupID:  groupID,
		Config:   config,
		Priority: 0,
		IsActive: true,

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
