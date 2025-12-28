package dto

import (
	"github.com/fregataa/aami/config-server/internal/domain"
)

// CreateCheckTemplateRequest represents a request to create a new check template
type CreateCheckTemplateRequest struct {
	Name          string                 `json:"name" binding:"required,min=1,max=255"`
	CheckType     string                 `json:"check_type" binding:"required,min=1,max=255"`
	ScriptContent string                 `json:"script_content" binding:"required,min=1"`
	Language      string                 `json:"language" binding:"required,oneof=bash python shell"`
	DefaultConfig map[string]interface{} `json:"default_config" binding:"omitempty"`
	Description   string                 `json:"description" binding:"omitempty,max=1000"`
	Version       string                 `json:"version" binding:"required,min=1,max=50"`
}

// Validate validates the CreateCheckTemplateRequest
func (req *CreateCheckTemplateRequest) Validate() error {
	if req.DefaultConfig == nil {
		req.DefaultConfig = make(map[string]interface{})
	}
	return nil
}

// UpdateCheckTemplateRequest represents a request to update an existing check template
type UpdateCheckTemplateRequest struct {
	ScriptContent *string                 `json:"script_content,omitempty" binding:"omitempty,min=1"`
	Language      *string                 `json:"language,omitempty" binding:"omitempty,oneof=bash python shell"`
	DefaultConfig map[string]interface{}  `json:"default_config,omitempty" binding:"omitempty"`
	Description   *string                 `json:"description,omitempty" binding:"omitempty,max=1000"`
	Version       *string                 `json:"version,omitempty" binding:"omitempty,min=1,max=50"`
}

// CheckTemplateResponse represents a check template in API responses
type CheckTemplateResponse struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	CheckType     string                 `json:"check_type"`
	ScriptContent string                 `json:"script_content"`
	Language      string                 `json:"language"`
	DefaultConfig map[string]interface{} `json:"default_config"`
	Description   string                 `json:"description"`
	Version       string                 `json:"version"`
	Hash          string                 `json:"hash"`
	TimestampResponse
}

// CheckTemplateSummaryResponse represents a minimal check template in API responses
// Used for list views where script content is not needed
type CheckTemplateSummaryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CheckType   string `json:"check_type"`
	Language    string `json:"language"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Hash        string `json:"hash"`
	TimestampResponse
}

// ToCheckTemplateResponse converts a domain.CheckTemplate to CheckTemplateResponse
func ToCheckTemplateResponse(template *domain.CheckTemplate) CheckTemplateResponse {
	return CheckTemplateResponse{
		ID:            template.ID,
		Name:          template.Name,
		CheckType:     template.CheckType,
		ScriptContent: template.ScriptContent,
		Language:      template.Language,
		DefaultConfig: template.DefaultConfig,
		Description:   template.Description,
		Version:       template.Version,
		Hash:          template.Hash,
		TimestampResponse: TimestampResponse{
			CreatedAt: template.CreatedAt,
			UpdatedAt: template.UpdatedAt,
		},
	}
}

// ToCheckTemplateSummaryResponse converts a domain.CheckTemplate to CheckTemplateSummaryResponse
func ToCheckTemplateSummaryResponse(template *domain.CheckTemplate) CheckTemplateSummaryResponse {
	return CheckTemplateSummaryResponse{
		ID:          template.ID,
		Name:        template.Name,
		CheckType:   template.CheckType,
		Language:    template.Language,
		Description: template.Description,
		Version:     template.Version,
		Hash:        template.Hash,
		TimestampResponse: TimestampResponse{
			CreatedAt: template.CreatedAt,
			UpdatedAt: template.UpdatedAt,
		},
	}
}

// ToCheckTemplateResponseList converts a slice of domain.CheckTemplate to slice of CheckTemplateResponse
func ToCheckTemplateResponseList(templates []domain.CheckTemplate) []CheckTemplateResponse {
	responses := make([]CheckTemplateResponse, len(templates))
	for i, template := range templates {
		responses[i] = ToCheckTemplateResponse(&template)
	}
	return responses
}

// ToCheckTemplateSummaryResponseList converts a slice of domain.CheckTemplate to slice of CheckTemplateSummaryResponse
func ToCheckTemplateSummaryResponseList(templates []domain.CheckTemplate) []CheckTemplateSummaryResponse {
	responses := make([]CheckTemplateSummaryResponse, len(templates))
	for i, template := range templates {
		responses[i] = ToCheckTemplateSummaryResponse(&template)
	}
	return responses
}
