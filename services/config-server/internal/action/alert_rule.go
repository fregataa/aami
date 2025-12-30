package action

import (
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
)

// ============================================================================
// Alert Rule Actions (Input)
// ============================================================================

// CreateAlertRuleFromTemplate represents the action to create an alert rule from a template
type CreateAlertRuleFromTemplate struct {
	GroupID       string
	TemplateID    string
	Enabled       bool
	Config        domain.AlertRuleConfig
	MergeStrategy string
	Priority      int
}

// CreateAlertRuleDirect represents the action to create an alert rule directly
type CreateAlertRuleDirect struct {
	GroupID       string
	Name          string
	Description   string
	Severity      domain.AlertSeverity
	QueryTemplate string
	DefaultConfig domain.AlertRuleConfig
	Enabled       bool
	Config        domain.AlertRuleConfig
	MergeStrategy string
	Priority      int
}

// UpdateAlertRule represents the action to update an alert rule
// nil fields mean "do not update"
type UpdateAlertRule struct {
	Name          *string
	Description   *string
	Severity      *domain.AlertSeverity
	QueryTemplate *string
	DefaultConfig *domain.AlertRuleConfig
	Enabled       *bool
	Config        *domain.AlertRuleConfig
	MergeStrategy *string
	Priority      *int
}

// ============================================================================
// Action Results (Output)
// ============================================================================

// AlertRuleResult represents the result of alert rule operations
type AlertRuleResult struct {
	ID      string
	GroupID string
	Group   *GroupResult

	// Template fields (copied at creation)
	Name          string
	Description   string
	Severity      domain.AlertSeverity
	QueryTemplate string
	DefaultConfig domain.AlertRuleConfig

	// Rule-specific fields
	Enabled       bool
	Config        domain.AlertRuleConfig
	MergeStrategy string
	Priority      int

	// Metadata
	CreatedFromTemplateID   *string
	CreatedFromTemplateName *string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// FromDomain converts domain.AlertRule to AlertRuleResult
func (r *AlertRuleResult) FromDomain(rule *domain.AlertRule) {
	r.ID = rule.ID
	r.GroupID = rule.GroupID
	r.Name = rule.Name
	r.Description = rule.Description
	r.Severity = rule.Severity
	r.QueryTemplate = rule.QueryTemplate
	r.DefaultConfig = rule.DefaultConfig
	r.Enabled = rule.Enabled
	r.Config = rule.Config
	r.MergeStrategy = rule.MergeStrategy
	r.Priority = rule.Priority
	r.CreatedFromTemplateID = rule.CreatedFromTemplateID
	r.CreatedFromTemplateName = rule.CreatedFromTemplateName
	r.CreatedAt = rule.CreatedAt
	r.UpdatedAt = rule.UpdatedAt

	// Convert nested group if loaded
	if rule.Group.ID != "" {
		g := GroupResult{}
		g.FromDomain(&rule.Group)
		r.Group = &g
	}
}

// NewAlertRuleResult creates AlertRuleResult from domain.AlertRule
func NewAlertRuleResult(rule *domain.AlertRule) AlertRuleResult {
	var result AlertRuleResult
	result.FromDomain(rule)
	return result
}

// NewAlertRuleResultList creates []AlertRuleResult from []domain.AlertRule
func NewAlertRuleResultList(rules []domain.AlertRule) []AlertRuleResult {
	results := make([]AlertRuleResult, len(rules))
	for i, rule := range rules {
		results[i] = NewAlertRuleResult(&rule)
	}
	return results
}
