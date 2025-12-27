package dto

import (
	"github.com/fregataa/aami/config-server/internal/domain"
)

// CreateCheckSettingRequest represents a request to create a new check setting
type CreateCheckSettingRequest struct {
	GroupID       string                 `json:"group_id" binding:"required,uuid"`
	CheckType     string                 `json:"check_type" binding:"required,min=1,max=100"`
	Config        map[string]interface{} `json:"config,omitempty"`
	MergeStrategy string                 `json:"merge_strategy" binding:"omitempty,oneof=override merge"`
	Priority      int                    `json:"priority" binding:"omitempty,min=0,max=1000"`
}

// UpdateCheckSettingRequest represents a request to update an existing check setting
type UpdateCheckSettingRequest struct {
	CheckType     *string                `json:"check_type,omitempty" binding:"omitempty,min=1,max=100"`
	Config        map[string]interface{} `json:"config,omitempty"`
	MergeStrategy *string                `json:"merge_strategy,omitempty" binding:"omitempty,oneof=override merge"`
	Priority      *int                   `json:"priority,omitempty" binding:"omitempty,min=0,max=1000"`
}

// CheckSettingResponse represents a check setting in API responses
type CheckSettingResponse struct {
	ID            string                 `json:"id"`
	GroupID       string                 `json:"group_id"`
	Group         *GroupResponse         `json:"group,omitempty"`
	CheckType     string                 `json:"check_type"`
	Config        map[string]interface{} `json:"config"`
	MergeStrategy string                 `json:"merge_strategy"`
	Priority      int                    `json:"priority"`
	TimestampResponse
}

// ToCheckSettingResponse converts a domain.CheckSetting to CheckSettingResponse
func ToCheckSettingResponse(setting *domain.CheckSetting) CheckSettingResponse {
	resp := CheckSettingResponse{
		ID:            setting.ID,
		GroupID:       setting.GroupID,
		CheckType:     setting.CheckType,
		Config:        setting.Config,
		MergeStrategy: setting.MergeStrategy,
		Priority:      setting.Priority,
		TimestampResponse: TimestampResponse{
			CreatedAt: setting.CreatedAt,
			UpdatedAt: setting.UpdatedAt,
		},
	}

	// Include group if loaded
	if setting.Group.ID != "" {
		group := ToGroupResponse(&setting.Group)
		resp.Group = &group
	}

	return resp
}

// ToCheckSettingResponseList converts a slice of domain.CheckSetting to slice of CheckSettingResponse
func ToCheckSettingResponseList(settings []domain.CheckSetting) []CheckSettingResponse {
	responses := make([]CheckSettingResponse, len(settings))
	for i, setting := range settings {
		responses[i] = ToCheckSettingResponse(&setting)
	}
	return responses
}
