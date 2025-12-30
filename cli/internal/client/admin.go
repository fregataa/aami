package client

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

// Seed calls the admin seed API to load default templates
func (c *Client) Seed(req SeedRequest) (*SeedResponse, error) {
	var response SeedResponse
	if err := c.Post("/api/v1/admin/seed", req, &response); err != nil {
		return nil, err
	}
	return &response, nil
}
