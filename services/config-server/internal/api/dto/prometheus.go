package dto

// RegenerateRulesResponse represents the result of rule regeneration
type RegenerateRulesResponse struct {
	Message       string `json:"message"`
	GroupsUpdated int    `json:"groups_updated,omitempty"`
	Success       bool   `json:"success"`
}

// RuleFileInfo represents information about a generated rule file
type RuleFileInfo struct {
	GroupID  string `json:"group_id"`
	FilePath string `json:"file_path"`
}

// ListRuleFilesResponse represents the response for listing rule files
type ListRuleFilesResponse struct {
	Files []RuleFileInfo `json:"files"`
	Total int            `json:"total"`
}

// ReloadPrometheusResponse represents the result of Prometheus reload
type ReloadPrometheusResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
	Healthy bool   `json:"healthy"`
}

// PrometheusStatusResponse represents Prometheus status
type PrometheusStatusResponse struct {
	Reachable bool                   `json:"reachable"`
	Healthy   bool                   `json:"healthy"`
	Status    map[string]interface{} `json:"status,omitempty"`
}

// EffectiveAlertRule represents a single effective alert rule with source info
type EffectiveAlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Severity    string                 `json:"severity"`
	Query       string                 `json:"query"`
	ForDuration string                 `json:"for_duration,omitempty"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Annotations map[string]string      `json:"annotations,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Source      string                 `json:"source"`
	SourceID    string                 `json:"source_id"`
	SourceName  string                 `json:"source_name"`
}

// EffectiveAlertRulesResponse is the response for effective rules endpoint
type EffectiveAlertRulesResponse struct {
	TargetID string               `json:"target_id"`
	Hostname string               `json:"hostname"`
	Rules    []EffectiveAlertRule `json:"rules"`
	Total    int                  `json:"total"`
}
