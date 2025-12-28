package dto

import (
	"github.com/fregataa/aami/config-server/internal/domain"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
)

// CreateCheckInstanceRequest represents a request to create a new check instance
// Supports two modes: from template (template_id) or direct creation (all fields)
type CreateCheckInstanceRequest struct {
	// Option 1: Create from template
	TemplateID *string `json:"template_id,omitempty" binding:"omitempty,uuid"`

	// Option 2: Direct creation (required if template_id not provided)
	Name          *string                 `json:"name,omitempty"`
	CheckType     *string                 `json:"check_type,omitempty"`
	ScriptContent *string                 `json:"script_content,omitempty"`
	Language      *string                 `json:"language,omitempty"`
	DefaultConfig *map[string]interface{} `json:"default_config,omitempty"`
	Description   *string                 `json:"description,omitempty"`
	Version       *string                 `json:"version,omitempty"`

	// Common fields
	Scope       domain.InstanceScope   `json:"scope" binding:"required,oneof=global namespace group"`
	NamespaceID *string                `json:"namespace_id,omitempty" binding:"omitempty,uuid"`
	GroupID     *string                `json:"group_id,omitempty" binding:"omitempty,uuid"`
	Config      map[string]interface{} `json:"config" binding:"omitempty"`
	Priority    int                    `json:"priority" binding:"omitempty,min=0,max=1000"`
	IsActive    bool                   `json:"is_active"`
}

// Validate validates the CreateCheckInstanceRequest
func (req *CreateCheckInstanceRequest) Validate() error {
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

// UpdateCheckInstanceRequest represents a request to update an existing check instance
type UpdateCheckInstanceRequest struct {
	Config   map[string]interface{} `json:"config,omitempty" binding:"omitempty"`
	Priority *int                   `json:"priority,omitempty" binding:"omitempty,min=0,max=1000"`
	IsActive *bool                  `json:"is_active,omitempty"`
}

// CheckInstanceResponse represents a check instance in API responses
type CheckInstanceResponse struct {
	ID string `json:"id"`

	// Template fields (copied from template at creation)
	Name          string                 `json:"name"`
	CheckType     string                 `json:"check_type"`
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
	CheckType     string                 `json:"check_type"`
	ScriptContent string                 `json:"script_content"`
	Language      string                 `json:"language"`
	Config        map[string]interface{} `json:"config"`
	Version       string                 `json:"version"`
	Hash          string                 `json:"hash"`
	InstanceID    string                 `json:"instance_id"`
}

// ToCheckInstanceResponse converts a domain.CheckInstance to CheckInstanceResponse
func ToCheckInstanceResponse(instance *domain.CheckInstance) CheckInstanceResponse {
	return CheckInstanceResponse{
		ID: instance.ID,

		// Template fields
		Name:          instance.Name,
		CheckType:     instance.CheckType,
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

// ToCheckInstanceResponseList converts a slice of domain.CheckInstance to slice of CheckInstanceResponse
func ToCheckInstanceResponseList(instances []domain.CheckInstance) []CheckInstanceResponse {
	responses := make([]CheckInstanceResponse, len(instances))
	for i, instance := range instances {
		responses[i] = ToCheckInstanceResponse(&instance)
	}
	return responses
}

// ToEffectiveCheckResponse converts a domain.EffectiveCheck to EffectiveCheckResponse
func ToEffectiveCheckResponse(check *domain.EffectiveCheck) EffectiveCheckResponse {
	return EffectiveCheckResponse{
		Name:          check.Name,
		CheckType:     check.CheckType,
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
