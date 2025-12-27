package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/fregataa/aami/config-server/internal/domain"
	"github.com/fregataa/aami/config-server/test/testutil"
)

func TestCheckSetting_Creation(t *testing.T) {
	groupID := "group-123"
	checkType := "disk_usage"
	setting := testutil.NewTestCheckSetting(groupID, checkType)

	assert.NotEmpty(t, setting.ID)
	assert.Equal(t, groupID, setting.GroupID)
	assert.Equal(t, checkType, setting.CheckType)
	assert.NotNil(t, setting.Config)
	assert.Equal(t, "merge", setting.MergeStrategy)
	assert.Equal(t, 100, setting.Priority)
	assert.NotZero(t, setting.CreatedAt)
	assert.NotZero(t, setting.UpdatedAt)
}

func TestCheckSetting_ConfigMap(t *testing.T) {
	// Test that config is a proper map
	setting := testutil.NewTestCheckSetting("group-123", "memory_check")

	assert.NotNil(t, setting.Config)
	assert.IsType(t, map[string]interface{}{}, setting.Config)

	// Verify default config has enabled key
	enabled, ok := setting.Config["enabled"]
	assert.True(t, ok)
	assert.Equal(t, true, enabled)
}

func TestCheckSetting_MergeStrategy(t *testing.T) {
	tests := []struct {
		name         string
		mergeStrategy string
	}{
		{
			name:         "merge strategy",
			mergeStrategy: "merge",
		},
		{
			name:         "override strategy",
			mergeStrategy: "override",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setting := &domain.CheckSetting{
				MergeStrategy: tt.mergeStrategy,
			}
			assert.Equal(t, tt.mergeStrategy, setting.MergeStrategy)
		})
	}
}

func TestCheckSetting_Priority(t *testing.T) {
	tests := []struct {
		name     string
		priority int
	}{
		{
			name:     "default priority",
			priority: 100,
		},
		{
			name:     "high priority",
			priority: 200,
		},
		{
			name:     "low priority",
			priority: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setting := &domain.CheckSetting{
				GroupID:       "group-123",
				CheckType:     "test",
				Config:        make(map[string]interface{}),
				MergeStrategy: "merge",
				Priority:      tt.priority,
			}
			assert.Equal(t, tt.priority, setting.Priority)
		})
	}
}

func TestCheckSetting_CheckTypes(t *testing.T) {
	// Test various check types
	checkTypes := []string{
		"cpu_usage",
		"memory_usage",
		"disk_usage",
		"network_latency",
		"gpu_temperature",
		"custom_check",
	}

	groupID := "group-123"
	for _, checkType := range checkTypes {
		t.Run(checkType, func(t *testing.T) {
			setting := testutil.NewTestCheckSetting(groupID, checkType)
			assert.Equal(t, checkType, setting.CheckType)
			assert.NotNil(t, setting.Config)
		})
	}
}

func TestCheckSetting_DefaultMergeStrategy(t *testing.T) {
	// Test that default check setting uses merge strategy
	setting := testutil.NewTestCheckSetting("group-123", "test_check")

	assert.Equal(t, "merge", setting.MergeStrategy)
}

func TestCheckSetting_ConfigFlexibility(t *testing.T) {
	// Test that config can hold various types
	setting := &domain.CheckSetting{
		GroupID:   "group-123",
		CheckType: "flexible_check",
		Config: map[string]interface{}{
			"string_value":  "test",
			"int_value":     42,
			"float_value":   3.14,
			"bool_value":    true,
			"nested_object": map[string]interface{}{
				"key": "value",
			},
			"array_value": []string{"a", "b", "c"},
		},
		MergeStrategy: "merge",
		Priority:      100,
	}

	assert.Equal(t, "test", setting.Config["string_value"])
	assert.Equal(t, 42, setting.Config["int_value"])
	assert.Equal(t, 3.14, setting.Config["float_value"])
	assert.Equal(t, true, setting.Config["bool_value"])
	assert.NotNil(t, setting.Config["nested_object"])
	assert.NotNil(t, setting.Config["array_value"])
}
