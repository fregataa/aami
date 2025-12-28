package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
)

// CheckTemplate represents a reusable check script definition
// Consistent with AlertTemplate pattern
type CheckTemplate struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`           // Unique template name
	CheckType     string                 `json:"check_type"`     // e.g., "disk", "mount"
	ScriptContent string                 `json:"script_content"` // Script code
	Language      string                 `json:"language"`       // "bash", "python"
	DefaultConfig map[string]interface{} `json:"default_config"` // Default parameters
	Description   string                 `json:"description"`
	Version       string                 `json:"version"`
	Hash          string                 `json:"hash"` // SHA256 hash of script_content
	DeletedAt     *time.Time             `json:"deleted_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// ComputeHash calculates the SHA256 hash of the script content
func (ct *CheckTemplate) ComputeHash() string {
	hash := sha256.Sum256([]byte(ct.ScriptContent))
	return hex.EncodeToString(hash[:])
}

// UpdateHash recalculates and updates the hash field
func (ct *CheckTemplate) UpdateHash() {
	ct.Hash = ct.ComputeHash()
}

// VerifyHash checks if the stored hash matches the script content
func (ct *CheckTemplate) VerifyHash() bool {
	return ct.Hash == ct.ComputeHash()
}

// Validate performs basic validation on the check template
func (ct *CheckTemplate) Validate() error {
	if ct.Name == "" {
		return domainerrors.NewValidationError("name", "name is required")
	}
	if ct.CheckType == "" {
		return domainerrors.NewValidationError("check_type", "check_type is required")
	}
	if ct.ScriptContent == "" {
		return domainerrors.NewValidationError("script_content", "script_content is required")
	}
	if ct.Language == "" {
		return domainerrors.NewValidationError("language", "language is required")
	}
	if ct.Version == "" {
		return domainerrors.NewValidationError("version", "version is required")
	}

	return nil
}