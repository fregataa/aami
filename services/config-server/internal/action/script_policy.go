package action

import (
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
)

// ============================================================================
// Actions (Input)
// ============================================================================

// CreateScriptPolicyFromTemplate represents the action to create a script policy from a template
type CreateScriptPolicyFromTemplate struct {
	TemplateID string
	Scope      domain.PolicyScope
	GroupID    *string
	Config     map[string]interface{}
	Priority   int
	IsActive   bool
}

// CreateScriptPolicyDirect represents the action to create a script policy directly
type CreateScriptPolicyDirect struct {
	Name          string
	ScriptType    string
	ScriptContent string
	Language      string
	DefaultConfig map[string]interface{}
	Description   string
	Version       string
	Hash          string
	Scope         domain.PolicyScope
	GroupID       *string
	Config        map[string]interface{}
	Priority      int
	IsActive      bool
}

// UpdateScriptPolicy represents the action to update a script policy
// nil fields mean "do not update"
type UpdateScriptPolicy struct {
	Config   map[string]interface{}
	Priority *int
	IsActive *bool
}

// ============================================================================
// Action Results (Output)
// ============================================================================

// ScriptPolicyResult represents the result of script policy operations
type ScriptPolicyResult struct {
	ID string

	// Template fields (copied at creation time)
	Name          string
	ScriptType    string
	ScriptContent string
	Language      string
	DefaultConfig map[string]interface{}
	Description   string
	Version       string
	Hash          string

	// Instance-specific fields
	Scope    domain.PolicyScope
	GroupID  *string
	Group    *GroupResult
	Config   map[string]interface{}
	Priority int
	IsActive bool

	// Metadata
	CreatedFromTemplateID   *string
	CreatedFromTemplateName *string
	TemplateVersion         *string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// FromDomain converts domain.ScriptPolicy to ScriptPolicyResult
func (r *ScriptPolicyResult) FromDomain(p *domain.ScriptPolicy) {
	r.ID = p.ID
	r.Name = p.Name
	r.ScriptType = p.ScriptType
	r.ScriptContent = p.ScriptContent
	r.Language = p.Language
	r.DefaultConfig = p.DefaultConfig
	r.Description = p.Description
	r.Version = p.Version
	r.Hash = p.Hash
	r.Scope = p.Scope
	r.GroupID = p.GroupID
	r.Config = p.Config
	r.Priority = p.Priority
	r.IsActive = p.IsActive
	r.CreatedFromTemplateID = p.CreatedFromTemplateID
	r.CreatedFromTemplateName = p.CreatedFromTemplateName
	r.TemplateVersion = p.TemplateVersion
	r.CreatedAt = p.CreatedAt
	r.UpdatedAt = p.UpdatedAt

	// Convert nested group if loaded
	if p.Group != nil {
		g := GroupResult{}
		g.FromDomain(p.Group)
		r.Group = &g
	}
}

// NewScriptPolicyResult creates ScriptPolicyResult from domain.ScriptPolicy
func NewScriptPolicyResult(p *domain.ScriptPolicy) ScriptPolicyResult {
	var result ScriptPolicyResult
	result.FromDomain(p)
	return result
}

// NewScriptPolicyResultList creates []ScriptPolicyResult from []domain.ScriptPolicy
func NewScriptPolicyResultList(policies []domain.ScriptPolicy) []ScriptPolicyResult {
	results := make([]ScriptPolicyResult, len(policies))
	for i, p := range policies {
		results[i] = NewScriptPolicyResult(&p)
	}
	return results
}
