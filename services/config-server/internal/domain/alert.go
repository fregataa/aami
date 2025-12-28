package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
	"time"
)

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	// AlertSeverityCritical for critical alerts requiring immediate attention
	AlertSeverityCritical AlertSeverity = "critical"

	// AlertSeverityWarning for warning alerts
	AlertSeverityWarning AlertSeverity = "warning"

	// AlertSeverityInfo for informational alerts
	AlertSeverityInfo AlertSeverity = "info"
)

// IsValid checks if the severity is one of the allowed values
func (s AlertSeverity) IsValid() bool {
	switch s {
	case AlertSeverityCritical, AlertSeverityWarning, AlertSeverityInfo:
		return true
	default:
		return false
	}
}

// AlertTemplate represents a reusable alert definition
type AlertTemplate struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Severity      AlertSeverity          `json:"severity"`
	QueryTemplate string                 `json:"query_template"`
	DefaultConfig map[string]interface{} `json:"default_config"`
	DeletedAt     *time.Time             `json:"deleted_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// RenderQuery renders the query template with the given configuration
func (at *AlertTemplate) RenderQuery(config map[string]interface{}) (string, error) {
	// Merge default config with provided config
	mergedConfig := make(map[string]interface{})
	for k, v := range at.DefaultConfig {
		mergedConfig[k] = v
	}
	for k, v := range config {
		mergedConfig[k] = v
	}

	// Parse and execute template
	tmpl, err := template.New("query").Parse(at.QueryTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse query template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, mergedConfig); err != nil {
		return "", fmt.Errorf("failed to render query template: %w", err)
	}

	return buf.String(), nil
}

// AlertRule represents a group-specific alert configuration
// Rule is independent from template - it contains a snapshot of template at creation time
type AlertRule struct {
	ID      string `json:"id"`
	GroupID string `json:"group_id"`
	Group   Group  `json:"group,omitempty"`

	// Template fields (copied at creation time - deep copy from AlertTemplate)
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Severity      AlertSeverity          `json:"severity"`
	QueryTemplate string                 `json:"query_template"`
	DefaultConfig map[string]interface{} `json:"default_config"`

	// Rule-specific fields
	Enabled       bool                   `json:"enabled"`
	Config        map[string]interface{} `json:"config"`
	MergeStrategy string                 `json:"merge_strategy"` // 'override' or 'merge'
	Priority      int                    `json:"priority"`

	// Metadata (optional, for tracking origin)
	CreatedFromTemplateID   *string `json:"created_from_template_id,omitempty"`
	CreatedFromTemplateName *string `json:"created_from_template_name,omitempty"`

	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// RenderQuery renders the alert query using the rule's query template and merged config
func (ar *AlertRule) RenderQuery() (string, error) {
	// Merge default config with rule config
	mergedConfig := make(map[string]interface{})
	for k, v := range ar.DefaultConfig {
		mergedConfig[k] = v
	}
	for k, v := range ar.Config {
		mergedConfig[k] = v
	}

	// Parse and execute template
	tmpl, err := template.New("query").Parse(ar.QueryTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse query template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, mergedConfig); err != nil {
		return "", fmt.Errorf("failed to render query template: %w", err)
	}

	return buf.String(), nil
}

// MergeWith merges this rule with a parent rule based on merge strategy
func (ar *AlertRule) MergeWith(parent *AlertRule) *AlertRule {
	if parent == nil {
		return ar
	}

	merged := &AlertRule{
		ID:      ar.ID,
		GroupID: ar.GroupID,

		// Template fields (copy from current rule)
		Name:          ar.Name,
		Description:   ar.Description,
		Severity:      ar.Severity,
		QueryTemplate: ar.QueryTemplate,
		DefaultConfig: ar.DefaultConfig,

		// Rule fields
		Enabled:       ar.Enabled,
		Config:        make(map[string]interface{}),
		MergeStrategy: ar.MergeStrategy,
		Priority:      ar.Priority,

		// Metadata
		CreatedFromTemplateID:   ar.CreatedFromTemplateID,
		CreatedFromTemplateName: ar.CreatedFromTemplateName,

		CreatedAt: ar.CreatedAt,
		UpdatedAt: ar.UpdatedAt,
	}

	if ar.MergeStrategy == "override" {
		// Override: use child config entirely
		merged.Config = ar.Config
	} else {
		// Merge: combine parent and child configs (child takes precedence)
		for k, v := range parent.Config {
			merged.Config[k] = v
		}
		for k, v := range ar.Config {
			merged.Config[k] = v
		}
	}

	return merged
}

// ToJSON converts the config to JSON string
func (ar *AlertRule) ToJSON() (string, error) {
	data, err := json.Marshal(ar.Config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}
	return string(data), nil
}

// NewAlertRuleFromTemplate creates a new AlertRule from an AlertTemplate
// This performs a deep copy of the template's fields into the rule
func NewAlertRuleFromTemplate(
	template *AlertTemplate,
	groupID string,
	overrideConfig map[string]interface{},
) *AlertRule {
	// Deep copy template's default config
	defaultConfig := make(map[string]interface{})
	for k, v := range template.DefaultConfig {
		defaultConfig[k] = deepCopyAlertValue(v)
	}

	// Deep copy override config
	config := make(map[string]interface{})
	for k, v := range overrideConfig {
		config[k] = deepCopyAlertValue(v)
	}

	now := time.Now()

	return &AlertRule{
		GroupID: groupID,

		// Template fields (deep copied)
		Name:          template.Name,
		Description:   template.Description,
		Severity:      template.Severity,
		QueryTemplate: template.QueryTemplate,
		DefaultConfig: defaultConfig,

		// Rule-specific fields
		Enabled:       true,
		Config:        config,
		MergeStrategy: "merge",
		Priority:      0,

		// Metadata (track origin)
		CreatedFromTemplateID:   &template.ID,
		CreatedFromTemplateName: &template.Name,

		CreatedAt: now,
		UpdatedAt: now,
	}
}

// deepCopyAlertValue performs a deep copy of interface{} values
// Handles maps and slices recursively
func deepCopyAlertValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case map[string]interface{}:
		newMap := make(map[string]interface{}, len(val))
		for k, v := range val {
			newMap[k] = deepCopyAlertValue(v)
		}
		return newMap
	case []interface{}:
		newSlice := make([]interface{}, len(val))
		for i, item := range val {
			newSlice[i] = deepCopyAlertValue(item)
		}
		return newSlice
	case map[interface{}]interface{}:
		newMap := make(map[interface{}]interface{}, len(val))
		for k, v := range val {
			newMap[k] = deepCopyAlertValue(v)
		}
		return newMap
	default:
		// For basic types (string, int, float, bool), direct assignment is safe
		return v
	}
}
