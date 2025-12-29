package action

import (
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
)

// ============================================================================
// Actions (Input)
// ============================================================================

// CreateMonitoringScript represents the action to create a monitoring script
type CreateMonitoringScript struct {
	Name          string
	ScriptType    string
	ScriptContent string
	Language      string
	DefaultConfig map[string]interface{}
	Description   string
	Version       string
}

// UpdateMonitoringScript represents the action to update a monitoring script
// nil fields mean "do not update"
type UpdateMonitoringScript struct {
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

// MonitoringScriptResult represents the result of monitoring script operations
type MonitoringScriptResult struct {
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

// FromDomain converts domain.MonitoringScript to MonitoringScriptResult
func (r *MonitoringScriptResult) FromDomain(s *domain.MonitoringScript) {
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

// NewMonitoringScriptResult creates MonitoringScriptResult from domain.MonitoringScript
func NewMonitoringScriptResult(s *domain.MonitoringScript) MonitoringScriptResult {
	var result MonitoringScriptResult
	result.FromDomain(s)
	return result
}

// NewMonitoringScriptResultList creates []MonitoringScriptResult from []domain.MonitoringScript
func NewMonitoringScriptResultList(scripts []domain.MonitoringScript) []MonitoringScriptResult {
	results := make([]MonitoringScriptResult, len(scripts))
	for i, s := range scripts {
		results[i] = NewMonitoringScriptResult(&s)
	}
	return results
}
