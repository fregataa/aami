package dto

import (
	"github.com/fregataa/aami/config-server/internal/action"
	"github.com/fregataa/aami/config-server/internal/domain"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
)

// CreateScriptPolicyFromTemplateRequest represents a request to create a script policy from a template
type CreateScriptPolicyFromTemplateRequest struct {
	TemplateID string                 `json:"template_id" binding:"required,uuid"`
	Scope      domain.PolicyScope     `json:"scope" binding:"required,oneof=global group"`
	GroupID    *string                `json:"group_id,omitempty" binding:"omitempty,uuid"`
	Config     map[string]interface{} `json:"config,omitempty"`
	Priority   int                    `json:"priority" binding:"omitempty,min=0,max=1000"`
	IsActive   bool                   `json:"is_active"`
}

// ToAction converts CreateScriptPolicyFromTemplateRequest to action.CreateScriptPolicyFromTemplate
func (r *CreateScriptPolicyFromTemplateRequest) ToAction() action.CreateScriptPolicyFromTemplate {
	return action.CreateScriptPolicyFromTemplate{
		TemplateID: r.TemplateID,
		Scope:      r.Scope,
		GroupID:    r.GroupID,
		Config:     r.Config,
		Priority:   r.Priority,
		IsActive:   r.IsActive,
	}
}

// CreateScriptPolicyDirectRequest represents a request to create a script policy directly (without template)
type CreateScriptPolicyDirectRequest struct {
	Name          string                 `json:"name" binding:"required"`
	ScriptType    string                 `json:"script_type" binding:"required"`
	ScriptContent string                 `json:"script_content" binding:"required"`
	Language      string                 `json:"language" binding:"required"`
	DefaultConfig map[string]interface{} `json:"default_config,omitempty"`
	Description   string                 `json:"description,omitempty"`
	Version       string                 `json:"version" binding:"required"`
	Scope         domain.PolicyScope     `json:"scope" binding:"required,oneof=global group"`
	GroupID       *string                `json:"group_id,omitempty" binding:"omitempty,uuid"`
	Config        map[string]interface{} `json:"config,omitempty"`
	Priority      int                    `json:"priority" binding:"omitempty,min=0,max=1000"`
	IsActive      bool                   `json:"is_active"`
}

// ToAction converts CreateScriptPolicyDirectRequest to action.CreateScriptPolicyDirect
func (r *CreateScriptPolicyDirectRequest) ToAction() action.CreateScriptPolicyDirect {
	return action.CreateScriptPolicyDirect{
		Name:          r.Name,
		ScriptType:    r.ScriptType,
		ScriptContent: r.ScriptContent,
		Language:      r.Language,
		DefaultConfig: r.DefaultConfig,
		Description:   r.Description,
		Version:       r.Version,
		Hash:          "", // Hash will be computed by service layer
		Scope:         r.Scope,
		GroupID:       r.GroupID,
		Config:        r.Config,
		Priority:      r.Priority,
		IsActive:      r.IsActive,
	}
}

// CreateScriptPolicyRequest represents a request to create a new check instance
// Supports two modes: from template (template_id) or direct creation (all fields)
// Deprecated: Use CreateScriptPolicyFromTemplateRequest or CreateScriptPolicyDirectRequest instead
type CreateScriptPolicyRequest struct {
	// Option 1: Create from template
	TemplateID *string `json:"template_id,omitempty" binding:"omitempty,uuid"`

	// Option 2: Direct creation (required if template_id not provided)
	Name          *string                 `json:"name,omitempty"`
	ScriptType    *string                 `json:"script_type,omitempty"`
	ScriptContent *string                 `json:"script_content,omitempty"`
	Language      *string                 `json:"language,omitempty"`
	DefaultConfig *map[string]interface{} `json:"default_config,omitempty"`
	Description   *string                 `json:"description,omitempty"`
	Version       *string                 `json:"version,omitempty"`

	// Common fields
	Scope    domain.PolicyScope     `json:"scope" binding:"required,oneof=global group"`
	GroupID  *string                `json:"group_id,omitempty" binding:"omitempty,uuid"`
	Config   map[string]interface{} `json:"config" binding:"omitempty"`
	Priority int                    `json:"priority" binding:"omitempty,min=0,max=1000"`
	IsActive bool                   `json:"is_active"`
}

// Validate validates the CreateScriptPolicyRequest
func (req *CreateScriptPolicyRequest) Validate() error {
	if req.Config == nil {
		req.Config = make(map[string]interface{})
	}

	// Default priority
	if req.Priority == 0 {
		req.Priority = 100
	}

	// Validate scope consistency
	switch req.Scope {
	case domain.ScopeGlobal:
		if req.GroupID != nil {
			return domainerrors.NewValidationError("scope", "global scope must not have group_id")
		}
	case domain.ScopeGroup:
		if req.GroupID == nil {
			return domainerrors.NewValidationError("group_id", "group_id is required for group scope")
		}
	default:
		return domainerrors.NewValidationError("scope", "invalid scope value: must be 'global' or 'group'")
	}

	return nil
}

// UpdateScriptPolicyRequest represents a request to update an existing check instance
type UpdateScriptPolicyRequest struct {
	Config   map[string]interface{} `json:"config,omitempty" binding:"omitempty"`
	Priority *int                   `json:"priority,omitempty" binding:"omitempty,min=0,max=1000"`
	IsActive *bool                  `json:"is_active,omitempty"`
}

// ToAction converts UpdateScriptPolicyRequest to action.UpdateScriptPolicy
func (r *UpdateScriptPolicyRequest) ToAction() action.UpdateScriptPolicy {
	return action.UpdateScriptPolicy{
		Config:   r.Config,
		Priority: r.Priority,
		IsActive: r.IsActive,
	}
}

// ScriptPolicyResponse represents a check instance in API responses
type ScriptPolicyResponse struct {
	ID string `json:"id"`

	// Template fields (copied from template at creation)
	Name          string                 `json:"name"`
	ScriptType    string                 `json:"script_type"`
	ScriptContent string                 `json:"script_content"`
	Language      string                 `json:"language"`
	DefaultConfig map[string]interface{} `json:"default_config"`
	Description   string                 `json:"description"`
	Version       string                 `json:"version"`
	Hash          string                 `json:"hash"`

	// Instance-specific fields
	Scope    string                 `json:"scope"`
	GroupID  *string                `json:"group_id,omitempty"`
	Config   map[string]interface{} `json:"config"`
	Priority int                    `json:"priority"`
	IsActive bool                   `json:"is_active"`

	// Metadata
	CreatedFromTemplateID   *string `json:"created_from_template_id,omitempty"`
	CreatedFromTemplateName *string `json:"created_from_template_name,omitempty"`
	TemplateVersion         *string `json:"template_version,omitempty"`

	TimestampResponse
}

// EffectiveCheckResponse represents an effective check for a node
// Contains merged configuration and script content
type EffectiveCheckResponse struct {
	Name          string                 `json:"name"`
	ScriptType     string                 `json:"script_type"`
	ScriptContent string                 `json:"script_content"`
	Language      string                 `json:"language"`
	Config        map[string]interface{} `json:"config"`
	Version       string                 `json:"version"`
	Hash          string                 `json:"hash"`
	InstanceID    string                 `json:"instance_id"`
}

// ToScriptPolicyResponse converts action.ScriptPolicyResult to ScriptPolicyResponse
func ToScriptPolicyResponse(result action.ScriptPolicyResult) ScriptPolicyResponse {
	return ScriptPolicyResponse{
		ID: result.ID,

		// Template fields
		Name:          result.Name,
		ScriptType:    result.ScriptType,
		ScriptContent: result.ScriptContent,
		Language:      result.Language,
		DefaultConfig: result.DefaultConfig,
		Description:   result.Description,
		Version:       result.Version,
		Hash:          result.Hash,

		// Instance-specific fields
		Scope:    string(result.Scope),
		GroupID:  result.GroupID,
		Config:   result.Config,
		Priority: result.Priority,
		IsActive: result.IsActive,

		// Metadata
		CreatedFromTemplateID:   result.CreatedFromTemplateID,
		CreatedFromTemplateName: result.CreatedFromTemplateName,
		TemplateVersion:         result.TemplateVersion,

		TimestampResponse: TimestampResponse{
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
		},
	}
}

// ToScriptPolicyResponseList converts a slice of action.ScriptPolicyResult to slice of ScriptPolicyResponse
func ToScriptPolicyResponseList(results []action.ScriptPolicyResult) []ScriptPolicyResponse {
	responses := make([]ScriptPolicyResponse, len(results))
	for i, result := range results {
		responses[i] = ToScriptPolicyResponse(result)
	}
	return responses
}

// ToEffectiveCheckResponse converts a domain.EffectiveCheck to EffectiveCheckResponse
func ToEffectiveCheckResponse(check *domain.EffectiveCheck) EffectiveCheckResponse {
	return EffectiveCheckResponse{
		Name:          check.Name,
		ScriptType:     check.ScriptType,
		ScriptContent: check.ScriptContent,
		Language:      check.Language,
		Config:        check.Config,
		Version:       check.Version,
		Hash:          check.Hash,
		InstanceID:    check.InstanceID,
	}
}

// ToEffectiveCheckResponseList converts a slice of domain.EffectiveCheck to slice of EffectiveCheckResponse
func ToEffectiveCheckResponseList(checks []domain.EffectiveCheck) []EffectiveCheckResponse {
	responses := make([]EffectiveCheckResponse, len(checks))
	for i, check := range checks {
		responses[i] = ToEffectiveCheckResponse(&check)
	}
	return responses
}
