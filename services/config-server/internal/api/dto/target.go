package dto

import (
	"time"

	"github.com/fregataa/aami/config-server/internal/action"
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

// ToAction converts CreateTargetRequest to action.CreateTarget
func (r *CreateTargetRequest) ToAction() action.CreateTarget {
	return action.CreateTarget{
		Hostname:  r.Hostname,
		IPAddress: r.IPAddress,
		GroupIDs:  r.GroupIDs,
		Status:    r.Status,
		Labels:    r.Labels,
		Metadata:  r.Metadata,
	}
}

// UpdateTargetRequest represents a request to update an existing target
type UpdateTargetRequest struct {
	Hostname  *string              `json:"hostname,omitempty" binding:"omitempty,min=1,max=255"`
	IPAddress *string              `json:"ip_address,omitempty" binding:"omitempty,ip"`
	Status    *domain.TargetStatus `json:"status,omitempty"`
	Labels    map[string]string    `json:"labels,omitempty"`
	Metadata  map[string]string    `json:"metadata,omitempty"`
}

// ToAction converts UpdateTargetRequest to action.UpdateTarget
func (r *UpdateTargetRequest) ToAction() action.UpdateTarget {
	return action.UpdateTarget{
		Hostname:  r.Hostname,
		IPAddress: r.IPAddress,
		Status:    r.Status,
		Labels:    r.Labels,
		Metadata:  r.Metadata,
	}
}

// UpdateTargetStatusRequest represents a request to update target status
type UpdateTargetStatusRequest struct {
	Status domain.TargetStatus `json:"status" binding:"required"`
}

// ToAction converts UpdateTargetStatusRequest to action.UpdateTargetStatus
func (r *UpdateTargetStatusRequest) ToAction() action.UpdateTargetStatus {
	return action.UpdateTargetStatus{
		Status: r.Status,
	}
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

// ToTargetResponse converts action.TargetResult to TargetResponse
func ToTargetResponse(result action.TargetResult) TargetResponse {
	resp := TargetResponse{
		ID:        result.ID,
		Hostname:  result.Hostname,
		IPAddress: result.IPAddress,
		Status:    result.Status,
		Labels:    result.Labels,
		Metadata:  result.Metadata,
		LastSeen:  result.LastSeen,
		TimestampResponse: TimestampResponse{
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
		},
	}

	// Include groups if loaded
	if len(result.Groups) > 0 {
		resp.Groups = ToGroupResponseList(result.Groups)
	}

	// Include exporters if loaded
	if len(result.Exporters) > 0 {
		resp.Exporters = ToExporterResponseList(result.Exporters)
	}

	return resp
}

// ToTargetResponseList converts a slice of action.TargetResult to slice of TargetResponse
func ToTargetResponseList(results []action.TargetResult) []TargetResponse {
	responses := make([]TargetResponse, len(results))
	for i, result := range results {
		responses[i] = ToTargetResponse(result)
	}
	return responses
}
