package dto

import (
	"github.com/fregataa/aami/config-server/internal/action"
)

// CreateScriptTemplateRequest represents a request to create a new script template
type CreateScriptTemplateRequest struct {
	Name          string                 `json:"name" binding:"required,min=1,max=255"`
	ScriptType    string                 `json:"script_type" binding:"required,min=1,max=255"`
	ScriptContent string                 `json:"script_content" binding:"required,min=1"`
	Language      string                 `json:"language" binding:"required,oneof=bash python shell"`
	DefaultConfig map[string]interface{} `json:"default_config" binding:"omitempty"`
	Description   string                 `json:"description" binding:"omitempty,max=1000"`
	Version       string                 `json:"version" binding:"required,min=1,max=50"`
}

// Validate validates the CreateScriptTemplateRequest
func (req *CreateScriptTemplateRequest) Validate() error {
	if req.DefaultConfig == nil {
		req.DefaultConfig = make(map[string]interface{})
	}
	return nil
}

// ToAction converts CreateScriptTemplateRequest to action.CreateScriptTemplate
func (r *CreateScriptTemplateRequest) ToAction() action.CreateScriptTemplate {
	return action.CreateScriptTemplate{
		Name:          r.Name,
		ScriptType:    r.ScriptType,
		ScriptContent: r.ScriptContent,
		Language:      r.Language,
		DefaultConfig: r.DefaultConfig,
		Description:   r.Description,
		Version:       r.Version,
	}
}

// UpdateScriptTemplateRequest represents a request to update an existing script template
type UpdateScriptTemplateRequest struct {
	ScriptContent *string                `json:"script_content,omitempty" binding:"omitempty,min=1"`
	Language      *string                `json:"language,omitempty" binding:"omitempty,oneof=bash python shell"`
	DefaultConfig map[string]interface{} `json:"default_config,omitempty" binding:"omitempty"`
	Description   *string                `json:"description,omitempty" binding:"omitempty,max=1000"`
	Version       *string                `json:"version,omitempty" binding:"omitempty,min=1,max=50"`
}

// ToAction converts UpdateScriptTemplateRequest to action.UpdateScriptTemplate
func (r *UpdateScriptTemplateRequest) ToAction() action.UpdateScriptTemplate {
	return action.UpdateScriptTemplate{
		ScriptContent: r.ScriptContent,
		Language:      r.Language,
		DefaultConfig: r.DefaultConfig,
		Description:   r.Description,
		Version:       r.Version,
	}
}

// ScriptTemplateResponse represents a script template in API responses
type ScriptTemplateResponse struct {
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

// ScriptTemplateSummaryResponse represents a minimal script template in API responses
// Used for list views where script content is not needed
type ScriptTemplateSummaryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ScriptType  string `json:"script_type"`
	Language    string `json:"language"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Hash        string `json:"hash"`
	TimestampResponse
}

// ToScriptTemplateResponse converts action.ScriptTemplateResult to ScriptTemplateResponse
func ToScriptTemplateResponse(result action.ScriptTemplateResult) ScriptTemplateResponse {
	return ScriptTemplateResponse{
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

// ToScriptTemplateSummaryResponse converts action.ScriptTemplateResult to ScriptTemplateSummaryResponse
func ToScriptTemplateSummaryResponse(result action.ScriptTemplateResult) ScriptTemplateSummaryResponse {
	return ScriptTemplateSummaryResponse{
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

// ToScriptTemplateResponseList converts a slice of action.ScriptTemplateResult to slice of ScriptTemplateResponse
func ToScriptTemplateResponseList(results []action.ScriptTemplateResult) []ScriptTemplateResponse {
	responses := make([]ScriptTemplateResponse, len(results))
	for i, result := range results {
		responses[i] = ToScriptTemplateResponse(result)
	}
	return responses
}

// ToScriptTemplateSummaryResponseList converts a slice of action.ScriptTemplateResult to slice of ScriptTemplateSummaryResponse
func ToScriptTemplateSummaryResponseList(results []action.ScriptTemplateResult) []ScriptTemplateSummaryResponse {
	responses := make([]ScriptTemplateSummaryResponse, len(results))
	for i, result := range results {
		responses[i] = ToScriptTemplateSummaryResponse(result)
	}
	return responses
}
