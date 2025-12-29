package action

import (
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
)

// ============================================================================
// Actions (Input)
// ============================================================================

// CreateScriptTemplate represents the action to create a script template
type CreateScriptTemplate struct {
	Name          string
	ScriptType    string
	ScriptContent string
	Language      string
	DefaultConfig map[string]interface{}
	Description   string
	Version       string
}

// UpdateScriptTemplate represents the action to update a script template
// nil fields mean "do not update"
type UpdateScriptTemplate struct {
	Name          *string
	ScriptType    *string
	ScriptContent *string
	Language      *string
	DefaultConfig map[string]interface{}
	Description   *string
	Version       *string
}

// ============================================================================
// Action Results (Output)
// ============================================================================

// ScriptTemplateResult represents the result of script template operations
type ScriptTemplateResult struct {
	ID            string
	Name          string
	ScriptType    string
	ScriptContent string
	Language      string
	DefaultConfig map[string]interface{}
	Description   string
	Version       string
	Hash          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// FromDomain converts domain.ScriptTemplate to ScriptTemplateResult
func (r *ScriptTemplateResult) FromDomain(s *domain.ScriptTemplate) {
	r.ID = s.ID
	r.Name = s.Name
	r.ScriptType = s.ScriptType
	r.ScriptContent = s.ScriptContent
	r.Language = s.Language
	r.DefaultConfig = s.DefaultConfig
	r.Description = s.Description
	r.Version = s.Version
	r.Hash = s.Hash
	r.CreatedAt = s.CreatedAt
	r.UpdatedAt = s.UpdatedAt
}

// NewScriptTemplateResult creates ScriptTemplateResult from domain.ScriptTemplate
func NewScriptTemplateResult(s *domain.ScriptTemplate) ScriptTemplateResult {
	var result ScriptTemplateResult
	result.FromDomain(s)
	return result
}

// NewScriptTemplateResultList creates []ScriptTemplateResult from []domain.ScriptTemplate
func NewScriptTemplateResultList(templates []domain.ScriptTemplate) []ScriptTemplateResult {
	results := make([]ScriptTemplateResult, len(templates))
	for i, t := range templates {
		results[i] = NewScriptTemplateResult(&t)
	}
	return results
}
