package action

import (
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
)

// ============================================================================
// Alert Template Actions (Input)
// ============================================================================

// CreateAlertTemplate represents the action to create an alert template
type CreateAlertTemplate struct {
	Name          string
	Description   string
	Severity      domain.AlertSeverity
	QueryTemplate string
	DefaultConfig domain.AlertRuleConfig
}

// UpdateAlertTemplate represents the action to update an alert template
// nil fields mean "do not update"
type UpdateAlertTemplate struct {
	Name          *string
	Description   *string
	Severity      *domain.AlertSeverity
	QueryTemplate *string
	DefaultConfig *domain.AlertRuleConfig
}

// ============================================================================
// Action Results (Output)
// ============================================================================

// AlertTemplateResult represents the result of alert template operations
type AlertTemplateResult struct {
	ID            string
	Name          string
	Description   string
	Severity      domain.AlertSeverity
	QueryTemplate string
	DefaultConfig domain.AlertRuleConfig
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// FromDomain converts domain.AlertTemplate to AlertTemplateResult
func (r *AlertTemplateResult) FromDomain(t *domain.AlertTemplate) {
	r.ID = t.ID
	r.Name = t.Name
	r.Description = t.Description
	r.Severity = t.Severity
	r.QueryTemplate = t.QueryTemplate
	r.DefaultConfig = t.DefaultConfig
	r.CreatedAt = t.CreatedAt
	r.UpdatedAt = t.UpdatedAt
}

// NewAlertTemplateResult creates AlertTemplateResult from domain.AlertTemplate
func NewAlertTemplateResult(t *domain.AlertTemplate) AlertTemplateResult {
	var result AlertTemplateResult
	result.FromDomain(t)
	return result
}

// NewAlertTemplateResultList creates []AlertTemplateResult from []domain.AlertTemplate
func NewAlertTemplateResultList(templates []domain.AlertTemplate) []AlertTemplateResult {
	results := make([]AlertTemplateResult, len(templates))
	for i, t := range templates {
		results[i] = NewAlertTemplateResult(&t)
	}
	return results
}
