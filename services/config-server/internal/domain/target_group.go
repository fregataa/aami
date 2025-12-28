package domain

import (
	"errors"
	"time"
)

// TargetGroup represents the many-to-many relationship between targets and groups
type TargetGroup struct {
	TargetID     string    `json:"target_id"`
	GroupID      string    `json:"group_id"`
	IsDefaultOwn bool      `json:"is_default_own"`
	CreatedAt    time.Time `json:"created_at"`
}

// Validate validates the TargetGroup fields
func (tg *TargetGroup) Validate() error {
	if tg.TargetID == "" {
		return errors.New("target_id is required")
	}
	if tg.GroupID == "" {
		return errors.New("group_id is required")
	}
	return nil
}
