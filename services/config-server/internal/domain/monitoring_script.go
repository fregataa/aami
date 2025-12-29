package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
)

// MonitoringScript represents a reusable monitoring script definition
// Consistent with AlertTemplate pattern
type MonitoringScript struct {
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
func (ms *MonitoringScript) ComputeHash() string {
	hash := sha256.Sum256([]byte(ms.ScriptContent))
	return hex.EncodeToString(hash[:])
}

// UpdateHash recalculates and updates the hash field
func (ms *MonitoringScript) UpdateHash() {
	ms.Hash = ms.ComputeHash()
}

// VerifyHash checks if the stored hash matches the script content
func (ms *MonitoringScript) VerifyHash() bool {
	return ms.Hash == ms.ComputeHash()
}

// Validate performs basic validation on the monitoring script
func (ms *MonitoringScript) Validate() error {
	if ms.Name == "" {
		return domainerrors.NewValidationError("name", "name is required")
	}
	if ms.ScriptType == "" {
		return domainerrors.NewValidationError("script_type", "script_type is required")
	}
	if ms.ScriptContent == "" {
		return domainerrors.NewValidationError("script_content", "script_content is required")
	}
	if ms.Language == "" {
		return domainerrors.NewValidationError("language", "language is required")
	}
	if ms.Version == "" {
		return domainerrors.NewValidationError("version", "version is required")
	}

	return nil
}
