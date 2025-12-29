package dto

import (
	"github.com/fregataa/aami/config-server/internal/domain"
)

// CreateAlertTemplateRequest represents a request to create a new alert template
type CreateAlertTemplateRequest struct {
	ID            string                 `json:"id" binding:"required,min=1,max=100"`
	Name          string                 `json:"name" binding:"required,min=1,max=255"`
	Description   string                 `json:"description" binding:"omitempty,max=500"`
	Severity      domain.AlertSeverity   `json:"severity" binding:"required"`
	QueryTemplate string                 `json:"query_template" binding:"required"`
	DefaultConfig map[string]interface{} `json:"default_config,omitempty"`
}

// UpdateAlertTemplateRequest represents a request to update an existing alert template
type UpdateAlertTemplateRequest struct {
	Name          *string                `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description   *string                `json:"description,omitempty" binding:"omitempty,max=500"`
	Severity      *domain.AlertSeverity  `json:"severity,omitempty"`
	QueryTemplate *string                `json:"query_template,omitempty"`
	DefaultConfig map[string]interface{} `json:"default_config,omitempty"`
}

// AlertTemplateResponse represents an alert template in API responses
type AlertTemplateResponse struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Severity      domain.AlertSeverity   `json:"severity"`
	QueryTemplate string                 `json:"query_template"`
	DefaultConfig map[string]interface{} `json:"default_config"`
	TimestampResponse
}

// ToAlertTemplateResponse converts a domain.AlertTemplate to AlertTemplateResponse
func ToAlertTemplateResponse(template *domain.AlertTemplate) AlertTemplateResponse {
	return AlertTemplateResponse{
		ID:            template.ID,
		Name:          template.Name,
		Description:   template.Description,
		Severity:      template.Severity,
		QueryTemplate: template.QueryTemplate,
		DefaultConfig: template.DefaultConfig,
		TimestampResponse: TimestampResponse{
			CreatedAt: template.CreatedAt,
			UpdatedAt: template.UpdatedAt,
		},
	}
}

// ToAlertTemplateResponseList converts a slice of domain.AlertTemplate to slice of AlertTemplateResponse
func ToAlertTemplateResponseList(templates []domain.AlertTemplate) []AlertTemplateResponse {
	responses := make([]AlertTemplateResponse, len(templates))
	for i, template := range templates {
		responses[i] = ToAlertTemplateResponse(&template)
	}
	return responses
}

// CreateAlertRuleFromTemplateRequest represents a request to create an alert rule from a template
type CreateAlertRuleFromTemplateRequest struct {
	GroupID       string                 `json:"group_id" binding:"required,uuid"`
	TemplateID    string                 `json:"template_id" binding:"required,uuid"`
	Enabled       bool                   `json:"enabled"`
	Config        map[string]interface{} `json:"config,omitempty"`
	MergeStrategy string                 `json:"merge_strategy" binding:"omitempty,oneof=override merge"`
	Priority      int                    `json:"priority" binding:"omitempty,min=0,max=1000"`
}

// CreateAlertRuleDirectRequest represents a request to create an alert rule directly (without template)
type CreateAlertRuleDirectRequest struct {
	GroupID       string                 `json:"group_id" binding:"required,uuid"`
	Name          string                 `json:"name" binding:"required"`
	Description   string                 `json:"description" binding:"required"`
	Severity      domain.AlertSeverity   `json:"severity" binding:"required"`
	QueryTemplate string                 `json:"query_template" binding:"required"`
	DefaultConfig map[string]interface{} `json:"default_config,omitempty"`
	Enabled       bool                   `json:"enabled"`
	Config        map[string]interface{} `json:"config,omitempty"`
	MergeStrategy string                 `json:"merge_strategy" binding:"omitempty,oneof=override merge"`
	Priority      int                    `json:"priority" binding:"omitempty,min=0,max=1000"`
}

// CreateAlertRuleRequest represents a request to create a new alert rule
// Supports two modes: from template (template_id) or direct creation (all fields)
// Deprecated: Use CreateAlertRuleFromTemplateRequest or CreateAlertRuleDirectRequest instead
type CreateAlertRuleRequest struct {
	GroupID string `json:"group_id" binding:"required,uuid"`

	// Option 1: Create from template
	TemplateID *string `json:"template_id,omitempty"`

	// Option 2: Direct creation (required if template_id not provided)
	Name          *string                 `json:"name,omitempty"`
	Description   *string                 `json:"description,omitempty"`
	Severity      *domain.AlertSeverity   `json:"severity,omitempty"`
	QueryTemplate *string                 `json:"query_template,omitempty"`
	DefaultConfig *map[string]interface{} `json:"default_config,omitempty"`

	// Common fields
	Enabled       bool                   `json:"enabled"`
	Config        map[string]interface{} `json:"config,omitempty"`
	MergeStrategy string                 `json:"merge_strategy" binding:"omitempty,oneof=override merge"`
	Priority      int                    `json:"priority" binding:"omitempty,min=0,max=1000"`
}

// UpdateAlertRuleRequest represents a request to update an existing alert rule
type UpdateAlertRuleRequest struct {
	Enabled       *bool                  `json:"enabled,omitempty"`
	Config        map[string]interface{} `json:"config,omitempty"`
	MergeStrategy *string                `json:"merge_strategy,omitempty" binding:"omitempty,oneof=override merge"`
	Priority      *int                   `json:"priority,omitempty" binding:"omitempty,min=0,max=1000"`
}

// AlertRuleResponse represents an alert rule in API responses
type AlertRuleResponse struct {
	ID      string         `json:"id"`
	GroupID string         `json:"group_id"`
	Group   *GroupResponse `json:"group,omitempty"`

	// Template fields (copied from template at creation)
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Severity      domain.AlertSeverity   `json:"severity"`
	QueryTemplate string                 `json:"query_template"`
	DefaultConfig map[string]interface{} `json:"default_config"`

	// Rule-specific fields
	Enabled       bool                   `json:"enabled"`
	Config        map[string]interface{} `json:"config"`
	MergeStrategy string                 `json:"merge_strategy"`
	Priority      int                    `json:"priority"`

	// Metadata
	CreatedFromTemplateID   *string `json:"created_from_template_id,omitempty"`
	CreatedFromTemplateName *string `json:"created_from_template_name,omitempty"`

	TimestampResponse
}

// ToAlertRuleResponse converts a domain.AlertRule to AlertRuleResponse
func ToAlertRuleResponse(rule *domain.AlertRule) AlertRuleResponse {
	resp := AlertRuleResponse{
		ID:      rule.ID,
		GroupID: rule.GroupID,

		// Template fields
		Name:          rule.Name,
		Description:   rule.Description,
		Severity:      rule.Severity,
		QueryTemplate: rule.QueryTemplate,
		DefaultConfig: rule.DefaultConfig,

		// Rule-specific fields
		Enabled:       rule.Enabled,
		Config:        rule.Config,
		MergeStrategy: rule.MergeStrategy,
		Priority:      rule.Priority,

		// Metadata
		CreatedFromTemplateID:   rule.CreatedFromTemplateID,
		CreatedFromTemplateName: rule.CreatedFromTemplateName,

		TimestampResponse: TimestampResponse{
			CreatedAt: rule.CreatedAt,
			UpdatedAt: rule.UpdatedAt,
		},
	}

	// Include group if loaded
	if rule.Group.ID != "" {
		group := ToGroupResponse(&rule.Group)
		resp.Group = &group
	}

	return resp
}

// ToAlertRuleResponseList converts a slice of domain.AlertRule to slice of AlertRuleResponse
func ToAlertRuleResponseList(rules []domain.AlertRule) []AlertRuleResponse {
	responses := make([]AlertRuleResponse, len(rules))
	for i, rule := range rules {
		responses[i] = ToAlertRuleResponse(&rule)
	}
	return responses
}
