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
type AlertRule struct {
	ID            string                 `json:"id"`
	GroupID       string                 `json:"group_id"`
	Group         Group                  `json:"group,omitempty"`
	TemplateID    string                 `json:"template_id"`
	Template      AlertTemplate          `json:"template,omitempty"`
	Enabled       bool                   `json:"enabled"`
	Config        map[string]interface{} `json:"config"`
	MergeStrategy string                 `json:"merge_strategy"` // 'override' or 'merge'
	Priority      int                    `json:"priority"`
	DeletedAt     *time.Time             `json:"deleted_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// RenderQuery renders the alert query using the template and this rule's config
func (ar *AlertRule) RenderQuery() (string, error) {
	return ar.Template.RenderQuery(ar.Config)
}

// MergeWith merges this rule with a parent rule based on merge strategy
func (ar *AlertRule) MergeWith(parent *AlertRule) *AlertRule {
	if parent == nil {
		return ar
	}

	merged := &AlertRule{
		ID:            ar.ID,
		GroupID:       ar.GroupID,
		TemplateID:    ar.TemplateID,
		Enabled:       ar.Enabled,
		Config:        make(map[string]interface{}),
		MergeStrategy: ar.MergeStrategy,
		Priority:      ar.Priority,
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
