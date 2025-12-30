package client

import "time"

// RegisterRequest represents a registration request to the config-server
type RegisterRequest struct {
	Token     string            `json:"token"`
	Hostname  string            `json:"hostname"`
	IPAddress string            `json:"ip_address"`
	GroupID   string            `json:"group_id,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// RegisterResponse represents the response from registration
type RegisterResponse struct {
	Target        TargetInfo `json:"target"`
	TokenUsage    int        `json:"token_usage"`
	RemainingUses int        `json:"remaining_uses"`
}

// TargetInfo represents target information returned from the server
type TargetInfo struct {
	ID        string            `json:"id"`
	Hostname  string            `json:"hostname"`
	IPAddress string            `json:"ip_address"`
	Status    string            `json:"status"`
	Labels    map[string]string `json:"labels"`
	Metadata  map[string]string `json:"metadata"`
	Groups    []GroupInfo       `json:"groups,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// GroupInfo represents group information
type GroupInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Priority    int    `json:"priority"`
}

// EffectiveCheck represents a check configuration to execute
type EffectiveCheck struct {
	Name          string                 `json:"name"`
	ScriptType    string                 `json:"script_type"`
	ScriptContent string                 `json:"script_content"`
	Language      string                 `json:"language"`
	Config        map[string]interface{} `json:"config"`
	Version       string                 `json:"version"`
	Hash          string                 `json:"hash"`
	InstanceID    string                 `json:"instance_id"`
}

// CheckResult represents the result of a check execution
type CheckResult struct {
	InstanceID string    `json:"instance_id"`
	TargetID   string    `json:"target_id"`
	ExitCode   int       `json:"exit_code"`
	Stdout     string    `json:"stdout,omitempty"`
	Stderr     string    `json:"stderr,omitempty"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
	ExecutedAt time.Time `json:"executed_at"`
	DurationMs int64     `json:"duration_ms"`
}

// SubmitCheckResultsRequest represents a batch of check results
type SubmitCheckResultsRequest struct {
	Results []CheckResult `json:"results"`
}

// ErrorResponse represents an error response from the server
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}
