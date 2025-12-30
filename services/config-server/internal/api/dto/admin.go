package dto

// SeedRequest represents the request body for seed operation
type SeedRequest struct {
	Force  bool `json:"force"`
	DryRun bool `json:"dry_run"`
}

// SeedStats represents statistics for a single entity type
type SeedStats struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
	Skipped int `json:"skipped"`
}

// SeedResponse represents the response for seed operation
type SeedResponse struct {
	AlertTemplates  SeedStats `json:"alert_templates"`
	ScriptTemplates SeedStats `json:"script_templates"`
	Errors          []string  `json:"errors,omitempty"`
}
