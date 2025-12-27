package dto

import (
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
)

// CreateTargetRequest represents a request to create a new target
type CreateTargetRequest struct {
	Hostname         string            `json:"hostname" binding:"required,min=1,max=255"`
	IPAddress        string            `json:"ip_address" binding:"required,ip"`
	PrimaryGroupID   string            `json:"primary_group_id" binding:"required,uuid"`
	SecondaryGroupIDs []string         `json:"secondary_group_ids,omitempty" binding:"omitempty,dive,uuid"`
	Status           domain.TargetStatus `json:"status" binding:"omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTargetRequest represents a request to update an existing target
type UpdateTargetRequest struct {
	Hostname          *string            `json:"hostname,omitempty" binding:"omitempty,min=1,max=255"`
	IPAddress         *string            `json:"ip_address,omitempty" binding:"omitempty,ip"`
	PrimaryGroupID    *string            `json:"primary_group_id,omitempty" binding:"omitempty,uuid"`
	SecondaryGroupIDs []string           `json:"secondary_group_ids,omitempty" binding:"omitempty,dive,uuid"`
	Status            *domain.TargetStatus `json:"status,omitempty"`
	Labels            map[string]string  `json:"labels,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTargetStatusRequest represents a request to update target status
type UpdateTargetStatusRequest struct {
	Status domain.TargetStatus `json:"status" binding:"required"`
}

// TargetResponse represents a target in API responses
type TargetResponse struct {
	ID               string                 `json:"id"`
	Hostname         string                 `json:"hostname"`
	IPAddress        string                 `json:"ip_address"`
	PrimaryGroupID   string                 `json:"primary_group_id"`
	PrimaryGroup     *GroupResponse         `json:"primary_group,omitempty"`
	SecondaryGroups  []GroupResponse        `json:"secondary_groups,omitempty"`
	Status           domain.TargetStatus    `json:"status"`
	Exporters        []ExporterResponse     `json:"exporters,omitempty"`
	Labels           map[string]string      `json:"labels"`
	Metadata         map[string]interface{} `json:"metadata"`
	LastSeen         *time.Time             `json:"last_seen,omitempty"`
	TimestampResponse
}

// ToTargetResponse converts a domain.Target to TargetResponse
func ToTargetResponse(target *domain.Target) TargetResponse {
	resp := TargetResponse{
		ID:             target.ID,
		Hostname:       target.Hostname,
		IPAddress:      target.IPAddress,
		PrimaryGroupID: target.PrimaryGroupID,
		Status:         target.Status,
		Labels:         target.Labels,
		Metadata:       target.Metadata,
		LastSeen:       target.LastSeen,
		TimestampResponse: TimestampResponse{
			CreatedAt: target.CreatedAt,
			UpdatedAt: target.UpdatedAt,
		},
	}

	// Include primary group if loaded
	if target.PrimaryGroup.ID != "" {
		group := ToGroupResponse(&target.PrimaryGroup)
		resp.PrimaryGroup = &group
	}

	// Include secondary groups if loaded
	if len(target.SecondaryGroups) > 0 {
		resp.SecondaryGroups = ToGroupResponseList(target.SecondaryGroups)
	}

	// Include exporters if loaded
	if len(target.Exporters) > 0 {
		resp.Exporters = ToExporterResponseList(target.Exporters)
	}

	return resp
}

// ToTargetResponseList converts a slice of domain.Target to slice of TargetResponse
func ToTargetResponseList(targets []domain.Target) []TargetResponse {
	responses := make([]TargetResponse, len(targets))
	for i, target := range targets {
		responses[i] = ToTargetResponse(&target)
	}
	return responses
}
