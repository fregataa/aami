package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
)

// ScriptTemplate represents a reusable monitoring script definition (template)
// Consistent with AlertTemplate pattern
type ScriptTemplate struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`           // Unique script name
	ScriptType    string                 `json:"script_type"`    // e.g., "disk", "mount"
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
func (st *ScriptTemplate) ComputeHash() string {
	hash := sha256.Sum256([]byte(st.ScriptContent))
	return hex.EncodeToString(hash[:])
}

// UpdateHash recalculates and updates the hash field
func (st *ScriptTemplate) UpdateHash() {
	st.Hash = st.ComputeHash()
}

// VerifyHash checks if the stored hash matches the script content
func (st *ScriptTemplate) VerifyHash() bool {
	return st.Hash == st.ComputeHash()
}

// Validate performs basic validation on the script template
func (st *ScriptTemplate) Validate() error {
	if st.Name == "" {
		return domainerrors.NewValidationError("name", "name is required")
	}
	if st.ScriptType == "" {
		return domainerrors.NewValidationError("script_type", "script_type is required")
	}
	if st.ScriptContent == "" {
		return domainerrors.NewValidationError("script_content", "script_content is required")
	}
	if st.Language == "" {
		return domainerrors.NewValidationError("language", "language is required")
	}
	if st.Version == "" {
		return domainerrors.NewValidationError("version", "version is required")
	}

	return nil
}
