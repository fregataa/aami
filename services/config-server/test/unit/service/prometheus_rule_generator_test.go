package service_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Mock AlertRuleRepository
type mockAlertRuleRepo struct {
	getRulesByGroupFunc func(ctx context.Context, groupID string) ([]domain.AlertRule, error)
	listFunc            func(ctx context.Context, page, pageSize int) ([]domain.AlertRule, int, error)
}

func (m *mockAlertRuleRepo) Create(ctx context.Context, rule *domain.AlertRule) error {
	return nil
}

func (m *mockAlertRuleRepo) GetByID(ctx context.Context, id string) (*domain.AlertRule, error) {
	return nil, nil
}

func (m *mockAlertRuleRepo) GetByGroupID(ctx context.Context, groupID string) ([]domain.AlertRule, error) {
	if m.getRulesByGroupFunc != nil {
		return m.getRulesByGroupFunc(ctx, groupID)
	}
	return nil, nil
}

func (m *mockAlertRuleRepo) GetByTemplateID(ctx context.Context, templateID string) ([]domain.AlertRule, error) {
	return nil, nil
}

func (m *mockAlertRuleRepo) Update(ctx context.Context, rule *domain.AlertRule) error {
	return nil
}

func (m *mockAlertRuleRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockAlertRuleRepo) Purge(ctx context.Context, id string) error {
	return nil
}

func (m *mockAlertRuleRepo) Restore(ctx context.Context, id string) error {
	return nil
}

func (m *mockAlertRuleRepo) List(ctx context.Context, page, pageSize int) ([]domain.AlertRule, int, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, page, pageSize)
	}
	return nil, 0, nil
}

// Mock GroupRepository
type mockGroupRepo struct {
	getByIDFunc func(ctx context.Context, id string) (*domain.Group, error)
}

func (m *mockGroupRepo) Create(ctx context.Context, group *domain.Group) error {
	return nil
}

func (m *mockGroupRepo) GetByID(ctx context.Context, id string) (*domain.Group, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockGroupRepo) GetByNamespaceID(ctx context.Context, namespaceID string) ([]domain.Group, error) {
	return nil, nil
}

func (m *mockGroupRepo) Update(ctx context.Context, group *domain.Group) error {
	return nil
}

func (m *mockGroupRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockGroupRepo) Purge(ctx context.Context, id string) error {
	return nil
}

func (m *mockGroupRepo) Restore(ctx context.Context, id string) error {
	return nil
}

func (m *mockGroupRepo) List(ctx context.Context, page, pageSize int) ([]domain.Group, int, error) {
	return nil, 0, nil
}

func (m *mockGroupRepo) GetChildren(ctx context.Context, parentID string) ([]domain.Group, error) {
	return nil, nil
}

func (m *mockGroupRepo) GetAncestors(ctx context.Context, groupID string) ([]domain.Group, error) {
	return nil, nil
}

func (m *mockGroupRepo) CountByNamespaceID(ctx context.Context, namespaceID string) (int64, error) {
	return 0, nil
}

// Helper function to create test alert rules
func createTestAlertRule(groupID string, name string, enabled bool) domain.AlertRule {
	return domain.AlertRule{
		ID:            uuid.New().String(),
		GroupID:       groupID,
		Name:          name,
		Description:   "Test alert rule: " + name,
		Severity:      domain.AlertSeverityCritical,
		QueryTemplate: "cpu_usage > {{ .threshold }}",
		DefaultConfig: map[string]interface{}{
			"threshold": 80,
		},
		Enabled: enabled,
		Config: map[string]interface{}{
			"threshold":    90,
			"for_duration": "5m",
			"labels": map[string]interface{}{
				"team": "platform",
			},
			"annotations": map[string]interface{}{
				"runbook_url": "https://example.com/runbook",
			},
		},
		MergeStrategy: "override",
		Priority:      100,
		DeletedAt:     nil,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func TestPrometheusRuleGenerator_GenerateRulesForGroup(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "prometheus-rules-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	groupID := uuid.New().String()
	groupName := "test-group"

	tests := []struct {
		name          string
		mockRules     []domain.AlertRule
		mockGroup     *domain.Group
		expectFile    bool
		expectError   bool
		validateYAML  bool
	}{
		{
			name: "generate rules for group with enabled rules",
			mockRules: []domain.AlertRule{
				createTestAlertRule(groupID, "HighCPU", true),
				createTestAlertRule(groupID, "HighMemory", true),
			},
			mockGroup: &domain.Group{
				ID:   groupID,
				Name: groupName,
			},
			expectFile:   true,
			expectError:  false,
			validateYAML: true,
		},
		{
			name: "skip disabled rules",
			mockRules: []domain.AlertRule{
				createTestAlertRule(groupID, "EnabledRule", true),
				createTestAlertRule(groupID, "DisabledRule", false),
			},
			mockGroup: &domain.Group{
				ID:   groupID,
				Name: groupName,
			},
			expectFile:   true,
			expectError:  false,
			validateYAML: true,
		},
		{
			name:      "delete file when no enabled rules",
			mockRules: []domain.AlertRule{},
			mockGroup: &domain.Group{
				ID:   groupID,
				Name: groupName,
			},
			expectFile:  false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Setup mocks
			alertRuleRepo := &mockAlertRuleRepo{
				getRulesByGroupFunc: func(ctx context.Context, gid string) ([]domain.AlertRule, error) {
					assert.Equal(t, groupID, gid)
					return tt.mockRules, nil
				},
			}

			groupRepo := &mockGroupRepo{
				getByIDFunc: func(ctx context.Context, id string) (*domain.Group, error) {
					assert.Equal(t, groupID, id)
					return tt.mockGroup, nil
				},
			}

			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
			generator := service.NewPrometheusRuleGenerator(alertRuleRepo, groupRepo, tmpDir, logger)

			// Execute
			err := generator.GenerateRulesForGroup(ctx, groupID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check file existence
			filename := filepath.Join(tmpDir, "group-"+groupID+".yml")
			_, statErr := os.Stat(filename)

			if tt.expectFile {
				assert.NoError(t, statErr, "Expected rule file to exist")

				// Validate YAML if requested
				if tt.validateYAML {
					data, readErr := os.ReadFile(filename)
					require.NoError(t, readErr)

					var ruleFile domain.PrometheusRuleFile
					yamlErr := yaml.Unmarshal(data, &ruleFile)
					require.NoError(t, yamlErr)

					// Verify structure
					assert.Len(t, ruleFile.Groups, 1)
					assert.Equal(t, "group_"+groupName+"_"+groupID, ruleFile.Groups[0].Name)

					// Count enabled rules
					enabledCount := 0
					for _, rule := range tt.mockRules {
						if rule.Enabled && rule.DeletedAt == nil {
							enabledCount++
						}
					}
					assert.Len(t, ruleFile.Groups[0].Rules, enabledCount)

					// Verify each rule
					for _, promRule := range ruleFile.Groups[0].Rules {
						assert.NotEmpty(t, promRule.Alert)
						assert.NotEmpty(t, promRule.Expr)
						assert.NotEmpty(t, promRule.Labels)
						assert.NotEmpty(t, promRule.Annotations)
						assert.Equal(t, groupID, promRule.Labels["group_id"])
						assert.Contains(t, promRule.Alert, "_Group_"+groupID)
					}
				}
			} else {
				assert.True(t, os.IsNotExist(statErr), "Expected rule file to not exist")
			}
		})
	}
}

func TestPrometheusRuleGenerator_ConvertToPrometheusRule(t *testing.T) {
	tests := []struct {
		name        string
		alertRule   domain.AlertRule
		expectError bool
		validate    func(t *testing.T, rule *domain.PrometheusRule)
	}{
		{
			name: "convert with full config",
			alertRule: domain.AlertRule{
				ID:            uuid.New().String(),
				GroupID:       "group-123",
				Name:          "HighCPU",
				Description:   "CPU usage is high",
				Severity:      domain.AlertSeverityCritical,
				QueryTemplate: "cpu_usage > {{ .threshold }}",
				Enabled:       true,
				Config: map[string]interface{}{
					"threshold":    90,
					"for_duration": "5m",
					"labels": map[string]interface{}{
						"team":        "platform",
						"environment": "production",
					},
					"annotations": map[string]interface{}{
						"runbook_url": "https://example.com/runbook",
						"dashboard":   "https://grafana.example.com/d/cpu",
					},
				},
			},
			expectError: false,
			validate: func(t *testing.T, rule *domain.PrometheusRule) {
				assert.Equal(t, "HighCPU_Group_group-123", rule.Alert)
				assert.Equal(t, "cpu_usage > 90", rule.Expr)
				assert.Equal(t, "5m", rule.For)
				assert.Equal(t, "critical", rule.Labels["severity"])
				assert.Equal(t, "group-123", rule.Labels["group_id"])
				assert.Equal(t, "platform", rule.Labels["team"])
				assert.Equal(t, "production", rule.Labels["environment"])
				assert.Equal(t, "HighCPU", rule.Annotations["summary"])
				assert.Equal(t, "CPU usage is high", rule.Annotations["description"])
				assert.Equal(t, "https://example.com/runbook", rule.Annotations["runbook_url"])
				assert.Equal(t, "https://grafana.example.com/d/cpu", rule.Annotations["dashboard"])
			},
		},
		{
			name: "convert with minimal config",
			alertRule: domain.AlertRule{
				ID:            uuid.New().String(),
				GroupID:       "group-456",
				Name:          "BasicAlert",
				Description:   "Basic alert",
				Severity:      domain.AlertSeverityWarning,
				QueryTemplate: "metric_value > 100",
				Enabled:       true,
				Config:        map[string]interface{}{},
			},
			expectError: false,
			validate: func(t *testing.T, rule *domain.PrometheusRule) {
				assert.Equal(t, "BasicAlert_Group_group-456", rule.Alert)
				assert.Equal(t, "metric_value > 100", rule.Expr)
				assert.Empty(t, rule.For)
				assert.Equal(t, "warning", rule.Labels["severity"])
				assert.Equal(t, "group-456", rule.Labels["group_id"])
				assert.Equal(t, "BasicAlert", rule.Annotations["summary"])
				assert.Equal(t, "Basic alert", rule.Annotations["description"])
			},
		},
		{
			name: "convert with template variables",
			alertRule: domain.AlertRule{
				ID:            uuid.New().String(),
				GroupID:       "group-789",
				Name:          "DiskFull",
				Description:   "Disk is full",
				Severity:      domain.AlertSeverityCritical,
				QueryTemplate: "disk_usage{mount=\"{{ .mount }}\"} > {{ .threshold }}",
				Enabled:       true,
				Config: map[string]interface{}{
					"mount":     "/data",
					"threshold": 95,
				},
			},
			expectError: false,
			validate: func(t *testing.T, rule *domain.PrometheusRule) {
				assert.Equal(t, "DiskFull_Group_group-789", rule.Alert)
				assert.Contains(t, rule.Expr, "mount=\"/data\"")
				assert.Contains(t, rule.Expr, "> 95")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create generator (we need to use reflection or expose the method)
			// For now, we'll test through the full flow
			ctx := context.Background()
			tmpDir, err := os.MkdirTemp("", "prometheus-convert-test-")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			alertRuleRepo := &mockAlertRuleRepo{
				getRulesByGroupFunc: func(ctx context.Context, gid string) ([]domain.AlertRule, error) {
					return []domain.AlertRule{tt.alertRule}, nil
				},
			}

			groupRepo := &mockGroupRepo{
				getByIDFunc: func(ctx context.Context, id string) (*domain.Group, error) {
					return &domain.Group{
						ID:   tt.alertRule.GroupID,
						Name: "test-group",
					}, nil
				},
			}

			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
			generator := service.NewPrometheusRuleGenerator(alertRuleRepo, groupRepo, tmpDir, logger)

			// Generate rules
			err = generator.GenerateRulesForGroup(ctx, tt.alertRule.GroupID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Read and validate the generated file
				filename := filepath.Join(tmpDir, "group-"+tt.alertRule.GroupID+".yml")
				data, readErr := os.ReadFile(filename)
				require.NoError(t, readErr)

				var ruleFile domain.PrometheusRuleFile
				yamlErr := yaml.Unmarshal(data, &ruleFile)
				require.NoError(t, yamlErr)

				require.Len(t, ruleFile.Groups, 1)
				require.Len(t, ruleFile.Groups[0].Rules, 1)

				// Validate using custom validation function
				if tt.validate != nil {
					tt.validate(t, &ruleFile.Groups[0].Rules[0])
				}
			}
		})
	}
}

func TestPrometheusRuleGenerator_GenerateAllRules(t *testing.T) {
	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "prometheus-all-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	group1ID := uuid.New().String()
	group2ID := uuid.New().String()

	allRules := []domain.AlertRule{
		createTestAlertRule(group1ID, "Rule1", true),
		createTestAlertRule(group1ID, "Rule2", true),
		createTestAlertRule(group2ID, "Rule3", true),
	}

	alertRuleRepo := &mockAlertRuleRepo{
		listFunc: func(ctx context.Context, page, pageSize int) ([]domain.AlertRule, int, error) {
			return allRules, len(allRules), nil
		},
		getRulesByGroupFunc: func(ctx context.Context, groupID string) ([]domain.AlertRule, error) {
			var rules []domain.AlertRule
			for _, rule := range allRules {
				if rule.GroupID == groupID && rule.Enabled && rule.DeletedAt == nil {
					rules = append(rules, rule)
				}
			}
			return rules, nil
		},
	}

	groupRepo := &mockGroupRepo{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Group, error) {
			return &domain.Group{
				ID:   id,
				Name: "test-group-" + id[:8],
			}, nil
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	generator := service.NewPrometheusRuleGenerator(alertRuleRepo, groupRepo, tmpDir, logger)

	// Execute
	err = generator.GenerateAllRules(ctx)
	assert.NoError(t, err)

	// Verify files were created for each group
	file1 := filepath.Join(tmpDir, "group-"+group1ID+".yml")
	file2 := filepath.Join(tmpDir, "group-"+group2ID+".yml")

	_, err1 := os.Stat(file1)
	_, err2 := os.Stat(file2)

	assert.NoError(t, err1, "Expected file for group1")
	assert.NoError(t, err2, "Expected file for group2")
}

func TestPrometheusRuleGenerator_DeleteRulesForGroup(t *testing.T) {
	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "prometheus-delete-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	groupID := uuid.New().String()

	// Create a test file
	filename := filepath.Join(tmpDir, "group-"+groupID+".yml")
	err = os.WriteFile(filename, []byte("test"), 0644)
	require.NoError(t, err)

	alertRuleRepo := &mockAlertRuleRepo{}
	groupRepo := &mockGroupRepo{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	generator := service.NewPrometheusRuleGenerator(alertRuleRepo, groupRepo, tmpDir, logger)

	// Execute delete
	err = generator.DeleteRulesForGroup(ctx, groupID)
	assert.NoError(t, err)

	// Verify file is deleted
	_, statErr := os.Stat(filename)
	assert.True(t, os.IsNotExist(statErr), "Expected file to be deleted")

	// Delete again (should not error)
	err = generator.DeleteRulesForGroup(ctx, groupID)
	assert.NoError(t, err)
}

func TestPrometheusRuleGenerator_YAMLMarshaling(t *testing.T) {
	// Create a PrometheusRuleFile and verify YAML output
	rule := domain.PrometheusRule{
		Alert: "TestAlert",
		Expr:  "cpu_usage > 80",
		For:   "5m",
		Labels: map[string]string{
			"severity": "warning",
			"team":     "platform",
		},
		Annotations: map[string]string{
			"summary":     "Test alert",
			"description": "This is a test alert",
		},
	}

	group := domain.NewPrometheusRuleGroup("test-group", []domain.PrometheusRule{rule})
	ruleFile := domain.NewPrometheusRuleFile([]domain.PrometheusRuleGroup{*group})

	// Marshal to YAML
	data, err := yaml.Marshal(ruleFile)
	require.NoError(t, err)

	// Unmarshal and verify
	var unmarshaled domain.PrometheusRuleFile
	err = yaml.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Len(t, unmarshaled.Groups, 1)
	assert.Equal(t, "test-group", unmarshaled.Groups[0].Name)
	assert.Len(t, unmarshaled.Groups[0].Rules, 1)
	assert.Equal(t, "TestAlert", unmarshaled.Groups[0].Rules[0].Alert)
	assert.Equal(t, "cpu_usage > 80", unmarshaled.Groups[0].Rules[0].Expr)
	assert.Equal(t, "5m", unmarshaled.Groups[0].Rules[0].For)
	assert.Equal(t, "warning", unmarshaled.Groups[0].Rules[0].Labels["severity"])
	assert.Equal(t, "Test alert", unmarshaled.Groups[0].Rules[0].Annotations["summary"])
}

func TestPrometheusRuleGenerator_FilterDeletedRules(t *testing.T) {
	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "prometheus-deleted-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	groupID := uuid.New().String()
	now := time.Now()

	rules := []domain.AlertRule{
		createTestAlertRule(groupID, "EnabledRule", true),
		{
			ID:            uuid.New().String(),
			GroupID:       groupID,
			Name:          "DeletedRule",
			QueryTemplate: "test > 1",
			Enabled:       true,
			DeletedAt:     &now, // Soft deleted
		},
	}

	alertRuleRepo := &mockAlertRuleRepo{
		getRulesByGroupFunc: func(ctx context.Context, gid string) ([]domain.AlertRule, error) {
			return rules, nil
		},
	}

	groupRepo := &mockGroupRepo{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Group, error) {
			return &domain.Group{ID: groupID, Name: "test-group"}, nil
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	generator := service.NewPrometheusRuleGenerator(alertRuleRepo, groupRepo, tmpDir, logger)

	err = generator.GenerateRulesForGroup(ctx, groupID)
	assert.NoError(t, err)

	// Read and verify only 1 rule (deleted rule should be filtered)
	filename := filepath.Join(tmpDir, "group-"+groupID+".yml")
	data, readErr := os.ReadFile(filename)
	require.NoError(t, readErr)

	var ruleFile domain.PrometheusRuleFile
	yamlErr := yaml.Unmarshal(data, &ruleFile)
	require.NoError(t, yamlErr)

	assert.Len(t, ruleFile.Groups[0].Rules, 1, "Should only have 1 rule (deleted rule filtered)")
	assert.Equal(t, "EnabledRule_Group_"+groupID, ruleFile.Groups[0].Rules[0].Alert)
}
