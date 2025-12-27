package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// CheckSetting represents configuration settings at the group level
// These settings are inherited by targets in the group hierarchy
type CheckSetting struct {
	ID            string                 `json:"id"`
	GroupID       string                 `json:"group_id"`
	Group         Group                  `json:"group,omitempty"`
	CheckType     string                 `json:"check_type"` // e.g., "mount", "disk", "network"
	Config        map[string]interface{} `json:"config"`
	MergeStrategy string                 `json:"merge_strategy"` // 'override' or 'merge'
	Priority      int                    `json:"priority"`
	DeletedAt     *time.Time             `json:"deleted_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// MergeWith merges this check setting with a parent setting based on merge strategy
func (cs *CheckSetting) MergeWith(parent *CheckSetting) *CheckSetting {
	if parent == nil {
		return cs
	}

	merged := &CheckSetting{
		ID:            cs.ID,
		GroupID:       cs.GroupID,
		CheckType:     cs.CheckType,
		Config:        make(map[string]interface{}),
		MergeStrategy: cs.MergeStrategy,
		Priority:      cs.Priority,
	}

	if cs.MergeStrategy == "override" {
		// Override: use child config entirely
		merged.Config = cs.Config
	} else {
		// Merge: combine parent and child configs (child takes precedence)
		for k, v := range parent.Config {
			merged.Config[k] = v
		}
		for k, v := range cs.Config {
			merged.Config[k] = v
		}
	}

	return merged
}

// ToJSON converts the config to JSON string
func (cs *CheckSetting) ToJSON() (string, error) {
	data, err := json.Marshal(cs.Config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}
	return string(data), nil
}

// GetConfigValue retrieves a specific configuration value
func (cs *CheckSetting) GetConfigValue(key string) (interface{}, bool) {
	val, exists := cs.Config[key]
	return val, exists
}

// SetConfigValue sets a specific configuration value
func (cs *CheckSetting) SetConfigValue(key string, value interface{}) {
	if cs.Config == nil {
		cs.Config = make(map[string]interface{})
	}
	cs.Config[key] = value
	cs.UpdatedAt = time.Now()
}
