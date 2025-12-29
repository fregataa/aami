package action

import (
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
)

// ============================================================================
// Actions (Input)
// ============================================================================

// CreateBootstrapToken represents the action to create a bootstrap token
type CreateBootstrapToken struct {
	Name      string
	MaxUses   int
	ExpiresAt time.Time
	Labels    map[string]string
}

// UpdateBootstrapToken represents the action to update a bootstrap token
// nil fields mean "do not update"
type UpdateBootstrapToken struct {
	Name      *string
	MaxUses   *int
	ExpiresAt *time.Time
	Labels    map[string]string
}

// ValidateToken represents the action to validate and use a bootstrap token
type ValidateToken struct {
	Token string
}

// BootstrapRegister represents the action to register a new target using bootstrap token
type BootstrapRegister struct {
	Token     string
	Hostname  string
	IPAddress string
	GroupID   string
	Labels    map[string]string
	Metadata  map[string]string
}

// ============================================================================
// Action Results (Output)
// ============================================================================

// BootstrapTokenResult represents the result of bootstrap token operations
type BootstrapTokenResult struct {
	ID        string
	Token     string
	Name      string
	MaxUses   int
	Uses      int
	ExpiresAt time.Time
	Labels    map[string]string
	IsValid   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// FromDomain converts domain.BootstrapToken to BootstrapTokenResult
func (r *BootstrapTokenResult) FromDomain(t *domain.BootstrapToken) {
	r.ID = t.ID
	r.Token = t.Token
	r.Name = t.Name
	r.MaxUses = t.MaxUses
	r.Uses = t.Uses
	r.ExpiresAt = t.ExpiresAt
	r.Labels = t.Labels
	r.IsValid = t.IsValid()
	r.CreatedAt = t.CreatedAt
	r.UpdatedAt = t.UpdatedAt
}

// NewBootstrapTokenResult creates BootstrapTokenResult from domain.BootstrapToken
func NewBootstrapTokenResult(t *domain.BootstrapToken) BootstrapTokenResult {
	var result BootstrapTokenResult
	result.FromDomain(t)
	return result
}

// NewBootstrapTokenResultList creates []BootstrapTokenResult from []domain.BootstrapToken
func NewBootstrapTokenResultList(tokens []domain.BootstrapToken) []BootstrapTokenResult {
	results := make([]BootstrapTokenResult, len(tokens))
	for i, t := range tokens {
		results[i] = NewBootstrapTokenResult(&t)
	}
	return results
}

// BootstrapRegisterResult represents the result of bootstrap registration
type BootstrapRegisterResult struct {
	Target        TargetResult
	TokenUsage    int
	RemainingUses int
}
