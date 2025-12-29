package dto

import (
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
)

// CreateTargetRequest represents a request to create a new target
type CreateTargetRequest struct {
	Hostname  string              `json:"hostname" binding:"required,min=1,max=255"`
	IPAddress string              `json:"ip_address" binding:"required,ip"`
	GroupIDs  []string            `json:"group_ids,omitempty" binding:"omitempty,dive,uuid"`
	Status    domain.TargetStatus `json:"status" binding:"omitempty"`
	Labels    map[string]string   `json:"labels,omitempty"`
	Metadata  map[string]string   `json:"metadata,omitempty"`
}

// UpdateTargetRequest represents a request to update an existing target
type UpdateTargetRequest struct {
	Hostname  *string              `json:"hostname,omitempty" binding:"omitempty,min=1,max=255"`
	IPAddress *string              `json:"ip_address,omitempty" binding:"omitempty,ip"`
	Status    *domain.TargetStatus `json:"status,omitempty"`
	Labels    map[string]string    `json:"labels,omitempty"`
	Metadata  map[string]string    `json:"metadata,omitempty"`
}

// UpdateTargetStatusRequest represents a request to update target status
type UpdateTargetStatusRequest struct {
	Status domain.TargetStatus `json:"status" binding:"required"`
}

// UpdateTargetGroupsRequest represents a request to replace all group mappings for a target
type UpdateTargetGroupsRequest struct {
	GroupIDs []string `json:"group_ids" binding:"required,min=1,dive,uuid"`
}

// AddGroupMappingRequest represents a request to add a target to a group
type AddGroupMappingRequest struct {
	TargetID string `json:"target_id" binding:"required,uuid"`
	GroupID  string `json:"group_id" binding:"required,uuid"`
}

// RemoveGroupMappingRequest represents a request to remove a target from a group
type RemoveGroupMappingRequest struct {
	TargetID string `json:"target_id" binding:"required,uuid"`
	GroupID  string `json:"group_id" binding:"required,uuid"`
}

// TargetResponse represents a target in API responses
type TargetResponse struct {
	ID        string              `json:"id"`
	Hostname  string              `json:"hostname"`
	IPAddress string              `json:"ip_address"`
	Groups    []GroupResponse     `json:"groups,omitempty"`
	Status    domain.TargetStatus `json:"status"`
	Exporters []ExporterResponse  `json:"exporters,omitempty"`
	Labels    map[string]string   `json:"labels"`
	Metadata  map[string]string   `json:"metadata"`
	LastSeen  *time.Time          `json:"last_seen,omitempty"`
	TimestampResponse
}

// ToTargetResponse converts a domain.Target to TargetResponse
func ToTargetResponse(target *domain.Target) TargetResponse {
	resp := TargetResponse{
		ID:        target.ID,
		Hostname:  target.Hostname,
		IPAddress: target.IPAddress,
		Status:    target.Status,
		Labels:    target.Labels,
		Metadata:  target.Metadata,
		LastSeen:  target.LastSeen,
		TimestampResponse: TimestampResponse{
			CreatedAt: target.CreatedAt,
			UpdatedAt: target.UpdatedAt,
		},
	}

	// Include groups if loaded
	if len(target.Groups) > 0 {
		resp.Groups = ToGroupResponseList(target.Groups)
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
