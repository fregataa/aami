package dto

import (
	"github.com/fregataa/aami/config-server/internal/action"
	"github.com/fregataa/aami/config-server/internal/domain"
)

// CreateAlertTemplateRequest represents a request to create a new alert template
type CreateAlertTemplateRequest struct {
	ID            string                `json:"id" binding:"required,min=1,max=100"`
	Name          string                `json:"name" binding:"required,min=1,max=255"`
	Description   string                `json:"description" binding:"omitempty,max=500"`
	Severity      domain.AlertSeverity  `json:"severity" binding:"required"`
	QueryTemplate string                `json:"query_template" binding:"required"`
	DefaultConfig domain.AlertRuleConfig `json:"default_config,omitempty"`
}

// ToAction converts CreateAlertTemplateRequest to action.CreateAlertTemplate
func (r *CreateAlertTemplateRequest) ToAction() action.CreateAlertTemplate {
	return action.CreateAlertTemplate{
		ID:            r.ID,
		Name:          r.Name,
		Description:   r.Description,
		Severity:      r.Severity,
		QueryTemplate: r.QueryTemplate,
		DefaultConfig: r.DefaultConfig,
	}
}

// UpdateAlertTemplateRequest represents a request to update an existing alert template
type UpdateAlertTemplateRequest struct {
	Name          *string                 `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description   *string                 `json:"description,omitempty" binding:"omitempty,max=500"`
	Severity      *domain.AlertSeverity   `json:"severity,omitempty"`
	QueryTemplate *string                 `json:"query_template,omitempty"`
	DefaultConfig *domain.AlertRuleConfig `json:"default_config,omitempty"`
}

// ToAction converts UpdateAlertTemplateRequest to action.UpdateAlertTemplate
func (r *UpdateAlertTemplateRequest) ToAction() action.UpdateAlertTemplate {
	return action.UpdateAlertTemplate{
		Name:          r.Name,
		Description:   r.Description,
		Severity:      r.Severity,
		QueryTemplate: r.QueryTemplate,
		DefaultConfig: r.DefaultConfig,
	}
}

// AlertTemplateResponse represents an alert template in API responses
type AlertTemplateResponse struct {
	ID            string                `json:"id"`
	Name          string                `json:"name"`
	Description   string                `json:"description"`
	Severity      domain.AlertSeverity  `json:"severity"`
	QueryTemplate string                `json:"query_template"`
	DefaultConfig domain.AlertRuleConfig `json:"default_config"`
	TimestampResponse
}

// ToAlertTemplateResponse converts action.AlertTemplateResult to AlertTemplateResponse
func ToAlertTemplateResponse(result action.AlertTemplateResult) AlertTemplateResponse {
	return AlertTemplateResponse{
		ID:            result.ID,
		Name:          result.Name,
		Description:   result.Description,
		Severity:      result.Severity,
		QueryTemplate: result.QueryTemplate,
		DefaultConfig: result.DefaultConfig,
		TimestampResponse: TimestampResponse{
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
		},
	}
}

// ToAlertTemplateResponseList converts a slice of action.AlertTemplateResult to slice of AlertTemplateResponse
func ToAlertTemplateResponseList(results []action.AlertTemplateResult) []AlertTemplateResponse {
	responses := make([]AlertTemplateResponse, len(results))
	for i, result := range results {
		responses[i] = ToAlertTemplateResponse(result)
	}
	return responses
}

// CreateAlertRuleFromTemplateRequest represents a request to create an alert rule from a template
type CreateAlertRuleFromTemplateRequest struct {
	GroupID       string                `json:"group_id" binding:"required,uuid"`
	TemplateID    string                `json:"template_id" binding:"required,uuid"`
	Enabled       bool                  `json:"enabled"`
	Config        domain.AlertRuleConfig `json:"config,omitempty"`
	MergeStrategy string                `json:"merge_strategy" binding:"omitempty,oneof=override merge"`
	Priority      int                   `json:"priority" binding:"omitempty,min=0,max=1000"`
}

// ToAction converts CreateAlertRuleFromTemplateRequest to action.CreateAlertRuleFromTemplate
func (r *CreateAlertRuleFromTemplateRequest) ToAction() action.CreateAlertRuleFromTemplate {
	return action.CreateAlertRuleFromTemplate{
		GroupID:       r.GroupID,
		TemplateID:    r.TemplateID,
		Enabled:       r.Enabled,
		Config:        r.Config,
		MergeStrategy: r.MergeStrategy,
		Priority:      r.Priority,
	}
}

// CreateAlertRuleDirectRequest represents a request to create an alert rule directly (without template)
type CreateAlertRuleDirectRequest struct {
	GroupID       string                `json:"group_id" binding:"required,uuid"`
	Name          string                `json:"name" binding:"required"`
	Description   string                `json:"description" binding:"required"`
	Severity      domain.AlertSeverity  `json:"severity" binding:"required"`
	QueryTemplate string                `json:"query_template" binding:"required"`
	DefaultConfig domain.AlertRuleConfig `json:"default_config,omitempty"`
	Enabled       bool                  `json:"enabled"`
	Config        domain.AlertRuleConfig `json:"config,omitempty"`
	MergeStrategy string                `json:"merge_strategy" binding:"omitempty,oneof=override merge"`
	Priority      int                   `json:"priority" binding:"omitempty,min=0,max=1000"`
}

// ToAction converts CreateAlertRuleDirectRequest to action.CreateAlertRuleDirect
func (r *CreateAlertRuleDirectRequest) ToAction() action.CreateAlertRuleDirect {
	return action.CreateAlertRuleDirect{
		GroupID:       r.GroupID,
		Name:          r.Name,
		Description:   r.Description,
		Severity:      r.Severity,
		QueryTemplate: r.QueryTemplate,
		DefaultConfig: r.DefaultConfig,
		Enabled:       r.Enabled,
		Config:        r.Config,
		MergeStrategy: r.MergeStrategy,
		Priority:      r.Priority,
	}
}

// CreateAlertRuleRequest represents a request to create a new alert rule
// Supports two modes: from template (template_id) or direct creation (all fields)
// Deprecated: Use CreateAlertRuleFromTemplateRequest or CreateAlertRuleDirectRequest instead
type CreateAlertRuleRequest struct {
	GroupID string `json:"group_id" binding:"required,uuid"`

	// Option 1: Create from template
	TemplateID *string `json:"template_id,omitempty"`

	// Option 2: Direct creation (required if template_id not provided)
	Name          *string                  `json:"name,omitempty"`
	Description   *string                  `json:"description,omitempty"`
	Severity      *domain.AlertSeverity    `json:"severity,omitempty"`
	QueryTemplate *string                  `json:"query_template,omitempty"`
	DefaultConfig *domain.AlertRuleConfig  `json:"default_config,omitempty"`

	// Common fields
	Enabled       bool                   `json:"enabled"`
	Config        domain.AlertRuleConfig `json:"config,omitempty"`
	MergeStrategy string                 `json:"merge_strategy" binding:"omitempty,oneof=override merge"`
	Priority      int                    `json:"priority" binding:"omitempty,min=0,max=1000"`
}

// UpdateAlertRuleRequest represents a request to update an existing alert rule
type UpdateAlertRuleRequest struct {
	Enabled       *bool                   `json:"enabled,omitempty"`
	Config        *domain.AlertRuleConfig `json:"config,omitempty"`
	MergeStrategy *string                 `json:"merge_strategy,omitempty" binding:"omitempty,oneof=override merge"`
	Priority      *int                    `json:"priority,omitempty" binding:"omitempty,min=0,max=1000"`
}

// ToAction converts UpdateAlertRuleRequest to action.UpdateAlertRule
func (r *UpdateAlertRuleRequest) ToAction() action.UpdateAlertRule {
	return action.UpdateAlertRule{
		Enabled:       r.Enabled,
		Config:        r.Config,
		MergeStrategy: r.MergeStrategy,
		Priority:      r.Priority,
	}
}

// AlertRuleResponse represents an alert rule in API responses
type AlertRuleResponse struct {
	ID      string         `json:"id"`
	GroupID string         `json:"group_id"`
	Group   *GroupResponse `json:"group,omitempty"`

	// Template fields (copied from template at creation)
	Name          string                `json:"name"`
	Description   string                `json:"description"`
	Severity      domain.AlertSeverity  `json:"severity"`
	QueryTemplate string                `json:"query_template"`
	DefaultConfig domain.AlertRuleConfig `json:"default_config"`

	// Rule-specific fields
	Enabled       bool                  `json:"enabled"`
	Config        domain.AlertRuleConfig `json:"config"`
	MergeStrategy string                `json:"merge_strategy"`
	Priority      int                   `json:"priority"`

	// Metadata
	CreatedFromTemplateID   *string `json:"created_from_template_id,omitempty"`
	CreatedFromTemplateName *string `json:"created_from_template_name,omitempty"`

	TimestampResponse
}

// ToAlertRuleResponse converts action.AlertRuleResult to AlertRuleResponse
func ToAlertRuleResponse(result action.AlertRuleResult) AlertRuleResponse {
	resp := AlertRuleResponse{
		ID:      result.ID,
		GroupID: result.GroupID,

		// Template fields
		Name:          result.Name,
		Description:   result.Description,
		Severity:      result.Severity,
		QueryTemplate: result.QueryTemplate,
		DefaultConfig: result.DefaultConfig,

		// Rule-specific fields
		Enabled:       result.Enabled,
		Config:        result.Config,
		MergeStrategy: result.MergeStrategy,
		Priority:      result.Priority,

		// Metadata
		CreatedFromTemplateID:   result.CreatedFromTemplateID,
		CreatedFromTemplateName: result.CreatedFromTemplateName,

		TimestampResponse: TimestampResponse{
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
		},
	}

	// Include group if loaded
	if result.Group != nil {
		group := ToGroupResponse(*result.Group)
		resp.Group = &group
	}

	return resp
}

// ToAlertRuleResponseList converts a slice of action.AlertRuleResult to slice of AlertRuleResponse
func ToAlertRuleResponseList(results []action.AlertRuleResult) []AlertRuleResponse {
	responses := make([]AlertRuleResponse, len(results))
	for i, result := range results {
		responses[i] = ToAlertRuleResponse(result)
	}
	return responses
}

// ActiveAlertResponse represents an active alert from Alertmanager
type ActiveAlertResponse struct {
	Fingerprint  string            `json:"fingerprint"`
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"starts_at"`
	GeneratorURL string            `json:"generator_url"`
}

// ActiveAlertsResponse represents the response for active alerts endpoint
type ActiveAlertsResponse struct {
	Alerts []ActiveAlertResponse `json:"alerts"`
	Total  int                   `json:"total"`
}
