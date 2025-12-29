package dto

import (
	"github.com/fregataa/aami/config-server/internal/action"
)

// CreateMonitoringScriptRequest represents a request to create a new monitoring script
type CreateMonitoringScriptRequest struct {
	Name          string                 `json:"name" binding:"required,min=1,max=255"`
	ScriptType    string                 `json:"script_type" binding:"required,min=1,max=255"`
	ScriptContent string                 `json:"script_content" binding:"required,min=1"`
	Language      string                 `json:"language" binding:"required,oneof=bash python shell"`
	DefaultConfig map[string]interface{} `json:"default_config" binding:"omitempty"`
	Description   string                 `json:"description" binding:"omitempty,max=1000"`
	Version       string                 `json:"version" binding:"required,min=1,max=50"`
}

// Validate validates the CreateMonitoringScriptRequest
func (req *CreateMonitoringScriptRequest) Validate() error {
	if req.DefaultConfig == nil {
		req.DefaultConfig = make(map[string]interface{})
	}
	return nil
}

// ToAction converts CreateMonitoringScriptRequest to action.CreateMonitoringScript
func (r *CreateMonitoringScriptRequest) ToAction() action.CreateMonitoringScript {
	return action.CreateMonitoringScript{
		Name:          r.Name,
		ScriptType:    r.ScriptType,
		ScriptContent: r.ScriptContent,
		Language:      r.Language,
		DefaultConfig: r.DefaultConfig,
		Description:   r.Description,
		Version:       r.Version,
	}
}

// UpdateMonitoringScriptRequest represents a request to update an existing monitoring script
type UpdateMonitoringScriptRequest struct {
	ScriptContent *string                 `json:"script_content,omitempty" binding:"omitempty,min=1"`
	Language      *string                 `json:"language,omitempty" binding:"omitempty,oneof=bash python shell"`
	DefaultConfig map[string]interface{}  `json:"default_config,omitempty" binding:"omitempty"`
	Description   *string                 `json:"description,omitempty" binding:"omitempty,max=1000"`
	Version       *string                 `json:"version,omitempty" binding:"omitempty,min=1,max=50"`
}

// ToAction converts UpdateMonitoringScriptRequest to action.UpdateMonitoringScript
func (r *UpdateMonitoringScriptRequest) ToAction() action.UpdateMonitoringScript {
	return action.UpdateMonitoringScript{
		ScriptContent: r.ScriptContent,
		Language:      r.Language,
		DefaultConfig: r.DefaultConfig,
		Description:   r.Description,
		Version:       r.Version,
	}
}

// MonitoringScriptResponse represents a monitoring script in API responses
type MonitoringScriptResponse struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	ScriptType    string                 `json:"script_type"`
	ScriptContent string                 `json:"script_content"`
	Language      string                 `json:"language"`
	DefaultConfig map[string]interface{} `json:"default_config"`
	Description   string                 `json:"description"`
	Version       string                 `json:"version"`
	Hash          string                 `json:"hash"`
	TimestampResponse
}

// MonitoringScriptSummaryResponse represents a minimal monitoring script in API responses
// Used for list views where script content is not needed
type MonitoringScriptSummaryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ScriptType  string `json:"script_type"`
	Language    string `json:"language"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Hash        string `json:"hash"`
	TimestampResponse
}

// ToMonitoringScriptResponse converts action.MonitoringScriptResult to MonitoringScriptResponse
func ToMonitoringScriptResponse(result action.MonitoringScriptResult) MonitoringScriptResponse {
	return MonitoringScriptResponse{
		ID:            result.ID,
		Name:          result.Name,
		ScriptType:    result.ScriptType,
		ScriptContent: result.ScriptContent,
		Language:      result.Language,
		DefaultConfig: result.DefaultConfig,
		Description:   result.Description,
		Version:       result.Version,
		Hash:          result.Hash,
		TimestampResponse: TimestampResponse{
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
		},
	}
}

// ToMonitoringScriptSummaryResponse converts action.MonitoringScriptResult to MonitoringScriptSummaryResponse
func ToMonitoringScriptSummaryResponse(result action.MonitoringScriptResult) MonitoringScriptSummaryResponse {
	return MonitoringScriptSummaryResponse{
		ID:          result.ID,
		Name:        result.Name,
		ScriptType:  result.ScriptType,
		Language:    result.Language,
		Description: result.Description,
		Version:     result.Version,
		Hash:        result.Hash,
		TimestampResponse: TimestampResponse{
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
		},
	}
}

// ToMonitoringScriptResponseList converts a slice of action.MonitoringScriptResult to slice of MonitoringScriptResponse
func ToMonitoringScriptResponseList(results []action.MonitoringScriptResult) []MonitoringScriptResponse {
	responses := make([]MonitoringScriptResponse, len(results))
	for i, result := range results {
		responses[i] = ToMonitoringScriptResponse(result)
	}
	return responses
}

// ToMonitoringScriptSummaryResponseList converts a slice of action.MonitoringScriptResult to slice of MonitoringScriptSummaryResponse
func ToMonitoringScriptSummaryResponseList(results []action.MonitoringScriptResult) []MonitoringScriptSummaryResponse {
	responses := make([]MonitoringScriptSummaryResponse, len(results))
	for i, result := range results {
		responses[i] = ToMonitoringScriptSummaryResponse(result)
	}
	return responses
}
