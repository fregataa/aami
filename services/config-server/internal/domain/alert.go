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

// AlertRuleConfig represents structured configuration for alert rules
type AlertRuleConfig struct {
	// ForDuration specifies how long the condition must be true before firing
	ForDuration string `json:"for_duration,omitempty"`

	// Labels are custom Prometheus labels attached to the alert
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations are custom Prometheus annotations attached to the alert
	Annotations map[string]string `json:"annotations,omitempty"`

	// TemplateVars contains variables used for query template rendering
	// These are merged with the query template during execution
	TemplateVars map[string]interface{} `json:"template_vars,omitempty"`
}

// ToMap converts AlertRuleConfig to a map for template rendering
func (c *AlertRuleConfig) ToMap() map[string]interface{} {
	result := make(map[string]interface{})

	if c.ForDuration != "" {
		result["for_duration"] = c.ForDuration
	}

	if len(c.Labels) > 0 {
		// Convert to map[string]interface{} for template compatibility
		labels := make(map[string]interface{})
		for k, v := range c.Labels {
			labels[k] = v
		}
		result["labels"] = labels
	}

	if len(c.Annotations) > 0 {
		// Convert to map[string]interface{} for template compatibility
		annotations := make(map[string]interface{})
		for k, v := range c.Annotations {
			annotations[k] = v
		}
		result["annotations"] = annotations
	}

	// Add template vars
	for k, v := range c.TemplateVars {
		result[k] = v
	}

	return result
}

// Merge merges another config into this one, with other taking precedence
func (c *AlertRuleConfig) Merge(other AlertRuleConfig) AlertRuleConfig {
	merged := AlertRuleConfig{
		ForDuration:  c.ForDuration,
		Labels:       make(map[string]string),
		Annotations:  make(map[string]string),
		TemplateVars: make(map[string]interface{}),
	}

	// Merge labels
	for k, v := range c.Labels {
		merged.Labels[k] = v
	}
	for k, v := range other.Labels {
		merged.Labels[k] = v
	}

	// Merge annotations
	for k, v := range c.Annotations {
		merged.Annotations[k] = v
	}
	for k, v := range other.Annotations {
		merged.Annotations[k] = v
	}

	// Merge template vars
	for k, v := range c.TemplateVars {
		merged.TemplateVars[k] = v
	}
	for k, v := range other.TemplateVars {
		merged.TemplateVars[k] = v
	}

	// Override ForDuration if provided
	if other.ForDuration != "" {
		merged.ForDuration = other.ForDuration
	}

	return merged
}

// AlertTemplate represents a reusable alert definition
type AlertTemplate struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Severity      AlertSeverity   `json:"severity"`
	QueryTemplate string          `json:"query_template"`
	DefaultConfig AlertRuleConfig `json:"default_config"`
	DeletedAt     *time.Time      `json:"deleted_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// RenderQuery renders the query template with the given configuration
func (at *AlertTemplate) RenderQuery(config AlertRuleConfig) (string, error) {
	// Merge default config with provided config
	mergedConfig := at.DefaultConfig.Merge(config)

	// Parse and execute template
	tmpl, err := template.New("query").Parse(at.QueryTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse query template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, mergedConfig.ToMap()); err != nil {
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
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Severity      AlertSeverity   `json:"severity"`
	QueryTemplate string          `json:"query_template"`
	DefaultConfig AlertRuleConfig `json:"default_config"`

	// Rule-specific fields
	Enabled       bool            `json:"enabled"`
	Config        AlertRuleConfig `json:"config"`
	MergeStrategy string          `json:"merge_strategy"` // 'override' or 'merge'
	Priority      int             `json:"priority"`

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
	mergedConfig := ar.DefaultConfig.Merge(ar.Config)

	// Parse and execute template
	tmpl, err := template.New("query").Parse(ar.QueryTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse query template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, mergedConfig.ToMap()); err != nil {
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
		merged.Config = parent.Config.Merge(ar.Config)
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
	overrideConfig AlertRuleConfig,
) *AlertRule {
	now := time.Now()

	return &AlertRule{
		GroupID: groupID,

		// Template fields (copied from template)
		Name:          template.Name,
		Description:   template.Description,
		Severity:      template.Severity,
		QueryTemplate: template.QueryTemplate,
		DefaultConfig: template.DefaultConfig,

		// Rule-specific fields
		Enabled:       true,
		Config:        overrideConfig,
		MergeStrategy: "merge",
		Priority:      0,

		// Metadata (track origin)
		CreatedFromTemplateID:   &template.ID,
		CreatedFromTemplateName: &template.Name,

		CreatedAt: now,
		UpdatedAt: now,
	}
}

// EffectiveAlertRule represents an alert rule with its source information
type EffectiveAlertRule struct {
	AlertRule
	RenderedQuery string // Rendered PromQL with config values
	SourceGroup   *Group // Group where this rule is defined
}

// EffectiveAlertRulesResult contains effective rules for a target
type EffectiveAlertRulesResult struct {
	Target *Target
	Rules  []EffectiveAlertRule
}

