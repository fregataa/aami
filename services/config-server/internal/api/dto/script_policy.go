package dto

import (
	"github.com/fregataa/aami/config-server/internal/domain"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
)

// CreateScriptPolicyFromTemplateRequest represents a request to create a script policy from a template
type CreateScriptPolicyFromTemplateRequest struct {
	TemplateID  string                 `json:"template_id" binding:"required,uuid"`
	Scope       domain.PolicyScope     `json:"scope" binding:"required,oneof=global namespace group"`
	NamespaceID *string                `json:"namespace_id,omitempty" binding:"omitempty,uuid"`
	GroupID     *string                `json:"group_id,omitempty" binding:"omitempty,uuid"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Priority    int                    `json:"priority" binding:"omitempty,min=0,max=1000"`
	IsActive    bool                   `json:"is_active"`
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
	Scope         domain.PolicyScope     `json:"scope" binding:"required,oneof=global namespace group"`
	NamespaceID   *string                `json:"namespace_id,omitempty" binding:"omitempty,uuid"`
	GroupID       *string                `json:"group_id,omitempty" binding:"omitempty,uuid"`
	Config        map[string]interface{} `json:"config,omitempty"`
	Priority      int                    `json:"priority" binding:"omitempty,min=0,max=1000"`
	IsActive      bool                   `json:"is_active"`
}

// CreateScriptPolicyRequest represents a request to create a new check instance
// Supports two modes: from template (template_id) or direct creation (all fields)
// Deprecated: Use CreateScriptPolicyFromTemplateRequest or CreateScriptPolicyDirectRequest instead
type CreateScriptPolicyRequest struct {
	// Option 1: Create from template
	TemplateID *string `json:"template_id,omitempty" binding:"omitempty,uuid"`

	// Option 2: Direct creation (required if template_id not provided)
	Name          *string                 `json:"name,omitempty"`
	ScriptType     *string                 `json:"script_type,omitempty"`
	ScriptContent *string                 `json:"script_content,omitempty"`
	Language      *string                 `json:"language,omitempty"`
	DefaultConfig *map[string]interface{} `json:"default_config,omitempty"`
	Description   *string                 `json:"description,omitempty"`
	Version       *string                 `json:"version,omitempty"`

	// Common fields
	Scope       domain.PolicyScope   `json:"scope" binding:"required,oneof=global namespace group"`
	NamespaceID *string                `json:"namespace_id,omitempty" binding:"omitempty,uuid"`
	GroupID     *string                `json:"group_id,omitempty" binding:"omitempty,uuid"`
	Config      map[string]interface{} `json:"config" binding:"omitempty"`
	Priority    int                    `json:"priority" binding:"omitempty,min=0,max=1000"`
	IsActive    bool                   `json:"is_active"`
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
		if req.NamespaceID != nil || req.GroupID != nil {
			return domainerrors.NewValidationError("scope", "global scope must not have namespace_id or group_id")
		}
	case domain.ScopeNamespace:
		if req.NamespaceID == nil {
			return domainerrors.NewValidationError("namespace_id", "namespace_id is required for namespace scope")
		}
		if req.GroupID != nil {
			return domainerrors.NewValidationError("group_id", "namespace scope must not have group_id")
		}
	case domain.ScopeGroup:
		if req.GroupID == nil {
			return domainerrors.NewValidationError("group_id", "group_id is required for group scope")
		}
		if req.NamespaceID == nil {
			return domainerrors.NewValidationError("namespace_id", "namespace_id is required for group scope")
		}
	default:
		return domainerrors.NewValidationError("scope", "invalid scope value")
	}

	return nil
}

// UpdateScriptPolicyRequest represents a request to update an existing check instance
type UpdateScriptPolicyRequest struct {
	Config   map[string]interface{} `json:"config,omitempty" binding:"omitempty"`
	Priority *int                   `json:"priority,omitempty" binding:"omitempty,min=0,max=1000"`
	IsActive *bool                  `json:"is_active,omitempty"`
}

// ScriptPolicyResponse represents a check instance in API responses
type ScriptPolicyResponse struct {
	ID string `json:"id"`

	// Template fields (copied from template at creation)
	Name          string                 `json:"name"`
	ScriptType     string                 `json:"script_type"`
	ScriptContent string                 `json:"script_content"`
	Language      string                 `json:"language"`
	DefaultConfig map[string]interface{} `json:"default_config"`
	Description   string                 `json:"description"`
	Version       string                 `json:"version"`
	Hash          string                 `json:"hash"`

	// Instance-specific fields
	Scope       string                 `json:"scope"`
	NamespaceID *string                `json:"namespace_id,omitempty"`
	GroupID     *string                `json:"group_id,omitempty"`
	Config      map[string]interface{} `json:"config"`
	Priority    int                    `json:"priority"`
	IsActive    bool                   `json:"is_active"`

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

// ToScriptPolicyResponse converts a domain.ScriptPolicy to ScriptPolicyResponse
func ToScriptPolicyResponse(instance *domain.ScriptPolicy) ScriptPolicyResponse {
	return ScriptPolicyResponse{
		ID: instance.ID,

		// Template fields
		Name:          instance.Name,
		ScriptType:     instance.ScriptType,
		ScriptContent: instance.ScriptContent,
		Language:      instance.Language,
		DefaultConfig: instance.DefaultConfig,
		Description:   instance.Description,
		Version:       instance.Version,
		Hash:          instance.Hash,

		// Instance-specific fields
		Scope:       string(instance.Scope),
		NamespaceID: instance.NamespaceID,
		GroupID:     instance.GroupID,
		Config:      instance.Config,
		Priority:    instance.Priority,
		IsActive:    instance.IsActive,

		// Metadata
		CreatedFromTemplateID:   instance.CreatedFromTemplateID,
		CreatedFromTemplateName: instance.CreatedFromTemplateName,
		TemplateVersion:         instance.TemplateVersion,

		TimestampResponse: TimestampResponse{
			CreatedAt: instance.CreatedAt,
			UpdatedAt: instance.UpdatedAt,
		},
	}
}

// ToScriptPolicyResponseList converts a slice of domain.ScriptPolicy to slice of ScriptPolicyResponse
func ToScriptPolicyResponseList(instances []domain.ScriptPolicy) []ScriptPolicyResponse {
	responses := make([]ScriptPolicyResponse, len(instances))
	for i, instance := range instances {
		responses[i] = ToScriptPolicyResponse(&instance)
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
