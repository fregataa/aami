package dto

import (
	"github.com/fregataa/aami/config-server/internal/domain"
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

// UpdateMonitoringScriptRequest represents a request to update an existing monitoring script
type UpdateMonitoringScriptRequest struct {
	ScriptContent *string                 `json:"script_content,omitempty" binding:"omitempty,min=1"`
	Language      *string                 `json:"language,omitempty" binding:"omitempty,oneof=bash python shell"`
	DefaultConfig map[string]interface{}  `json:"default_config,omitempty" binding:"omitempty"`
	Description   *string                 `json:"description,omitempty" binding:"omitempty,max=1000"`
	Version       *string                 `json:"version,omitempty" binding:"omitempty,min=1,max=50"`
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

// ToMonitoringScriptResponse converts a domain.MonitoringScript to MonitoringScriptResponse
func ToMonitoringScriptResponse(script *domain.MonitoringScript) MonitoringScriptResponse {
	return MonitoringScriptResponse{
		ID:            script.ID,
		Name:          script.Name,
		ScriptType:    script.ScriptType,
		ScriptContent: script.ScriptContent,
		Language:      script.Language,
		DefaultConfig: script.DefaultConfig,
		Description:   script.Description,
		Version:       script.Version,
		Hash:          script.Hash,
		TimestampResponse: TimestampResponse{
			CreatedAt: script.CreatedAt,
			UpdatedAt: script.UpdatedAt,
		},
	}
}

// ToMonitoringScriptSummaryResponse converts a domain.MonitoringScript to MonitoringScriptSummaryResponse
func ToMonitoringScriptSummaryResponse(script *domain.MonitoringScript) MonitoringScriptSummaryResponse {
	return MonitoringScriptSummaryResponse{
		ID:          script.ID,
		Name:        script.Name,
		ScriptType:  script.ScriptType,
		Language:    script.Language,
		Description: script.Description,
		Version:     script.Version,
		Hash:        script.Hash,
		TimestampResponse: TimestampResponse{
			CreatedAt: script.CreatedAt,
			UpdatedAt: script.UpdatedAt,
		},
	}
}

// ToMonitoringScriptResponseList converts a slice of domain.MonitoringScript to slice of MonitoringScriptResponse
func ToMonitoringScriptResponseList(scripts []domain.MonitoringScript) []MonitoringScriptResponse {
	responses := make([]MonitoringScriptResponse, len(scripts))
	for i, script := range scripts {
		responses[i] = ToMonitoringScriptResponse(&script)
	}
	return responses
}

// ToMonitoringScriptSummaryResponseList converts a slice of domain.MonitoringScript to slice of MonitoringScriptSummaryResponse
func ToMonitoringScriptSummaryResponseList(scripts []domain.MonitoringScript) []MonitoringScriptSummaryResponse {
	responses := make([]MonitoringScriptSummaryResponse, len(scripts))
	for i, script := range scripts {
		responses[i] = ToMonitoringScriptSummaryResponse(&script)
	}
	return responses
}
