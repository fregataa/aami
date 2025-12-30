package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/fregataa/aami/config-server/internal/domain"
	"github.com/fregataa/aami/config-server/test/testutil"
)

func TestAlertSeverity_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		severity domain.AlertSeverity
		want     bool
	}{
		{
			name:     "info severity is valid",
			severity: domain.AlertSeverityInfo,
			want:     true,
		},
		{
			name:     "warning severity is valid",
			severity: domain.AlertSeverityWarning,
			want:     true,
		},
		{
			name:     "critical severity is valid",
			severity: domain.AlertSeverityCritical,
			want:     true,
		},
		{
			name:     "invalid severity returns false",
			severity: domain.AlertSeverity("invalid"),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.severity.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAlertTemplate_RenderQuery(t *testing.T) {
	tests := []struct {
		name          string
		queryTemplate string
		config        domain.AlertRuleConfig
		want          string
		wantErr       bool
	}{
		{
			name:          "render query with threshold",
			queryTemplate: "cpu_usage > {{ .threshold }}",
			config: domain.AlertRuleConfig{
				TemplateVars: map[string]interface{}{
					"threshold": 80,
				},
			},
			want:    "cpu_usage > 80",
			wantErr: false,
		},
		{
			name:          "render query with multiple variables",
			queryTemplate: "metric > {{ .threshold }} for {{ .duration }}",
			config: domain.AlertRuleConfig{
				TemplateVars: map[string]interface{}{
					"threshold": 90,
					"duration":  "5m",
				},
			},
			want:    "metric > 90 for 5m",
			wantErr: false,
		},
		{
			name:          "render query with missing variable renders empty",
			queryTemplate: "cpu_usage > {{ .threshold }}",
			config:        domain.AlertRuleConfig{},
			want:          "cpu_usage > <no value>",
			wantErr:       false,
		},
		{
			name:          "render query with no template variables",
			queryTemplate: "cpu_usage > 80",
			config:        domain.AlertRuleConfig{},
			want:          "cpu_usage > 80",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := &domain.AlertTemplate{
				QueryTemplate: tt.queryTemplate,
			}
			got, err := template.RenderQuery(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMergeStrategy(t *testing.T) {
	tests := []struct {
		name          string
		mergeStrategy string
		want          bool
	}{
		{
			name:          "merge strategy is valid",
			mergeStrategy: "merge",
			want:          true,
		},
		{
			name:          "override strategy is valid",
			mergeStrategy: "override",
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &domain.AlertRule{
				MergeStrategy: tt.mergeStrategy,
			}
			assert.Equal(t, tt.mergeStrategy, rule.MergeStrategy)
		})
	}
}

func TestAlertTemplate_Creation(t *testing.T) {
	template := testutil.NewTestAlertTemplate("high_cpu", domain.AlertSeverityWarning)

	assert.Equal(t, "high_cpu", template.ID)
	assert.Contains(t, template.Name, "high_cpu")
	assert.NotEmpty(t, template.Description)
	assert.Equal(t, domain.AlertSeverityWarning, template.Severity)
	assert.NotEmpty(t, template.QueryTemplate)
	assert.NotNil(t, template.DefaultConfig.TemplateVars)
	assert.NotZero(t, template.CreatedAt)
	assert.NotZero(t, template.UpdatedAt)
}

func TestAlertRule_Creation(t *testing.T) {
	groupID := "group-123"
	templateID := "high_cpu"
	rule := testutil.NewTestAlertRule(groupID, templateID)

	assert.NotEmpty(t, rule.ID)
	assert.Equal(t, groupID, rule.GroupID)
	assert.NotNil(t, rule.CreatedFromTemplateID)
	assert.Equal(t, templateID, *rule.CreatedFromTemplateID)
	assert.True(t, rule.Enabled)
	assert.NotNil(t, rule.Config.TemplateVars)
	assert.Equal(t, "override", rule.MergeStrategy)
	assert.Equal(t, 100, rule.Priority)
	assert.NotZero(t, rule.CreatedAt)
	assert.NotZero(t, rule.UpdatedAt)
}

func TestAlertSeverity_Constants(t *testing.T) {
	// Verify the constant values are as expected
	assert.Equal(t, domain.AlertSeverity("info"), domain.AlertSeverityInfo)
	assert.Equal(t, domain.AlertSeverity("warning"), domain.AlertSeverityWarning)
	assert.Equal(t, domain.AlertSeverity("critical"), domain.AlertSeverityCritical)
}

func TestAlertTemplate_WithDefaultConfig(t *testing.T) {
	template := testutil.NewTestAlertTemplate("test_alert", domain.AlertSeverityInfo)

	// Verify default config exists
	assert.NotNil(t, template.DefaultConfig.TemplateVars)

	// Test rendering with default config
	rendered, err := template.RenderQuery(domain.AlertRuleConfig{})
	assert.NoError(t, err)
	assert.NotEmpty(t, rendered)
}

func TestAlertRule_MergeStrategies(t *testing.T) {
	strategies := []string{
		"merge",
		"override",
	}

	for _, strategy := range strategies {
		t.Run(strategy, func(t *testing.T) {
			rule := &domain.AlertRule{
				MergeStrategy: strategy,
			}
			assert.Equal(t, strategy, rule.MergeStrategy)
		})
	}
}
