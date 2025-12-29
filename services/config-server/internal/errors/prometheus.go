package errors

import "errors"

// Prometheus rule generation errors
var (
	ErrRuleGenerationFailed   = errors.New("prometheus rule generation failed")
	ErrRuleMarshalingFailed   = errors.New("failed to marshal rules to YAML")
	ErrRuleWriteFailed        = errors.New("failed to write rule file")
	ErrRuleDeleteFailed       = errors.New("failed to delete rule file")
	ErrQueryRenderingFailed   = errors.New("failed to render query template")
	ErrNoValidRules           = errors.New("no valid rules could be converted")
	ErrGroupRetrievalFailed   = errors.New("failed to retrieve group")
	ErrAlertRulesListFailed   = errors.New("failed to list alert rules")
	ErrBulkGenerationFailed   = errors.New("failed to generate rules for multiple groups")
)
