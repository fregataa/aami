package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/fregataa/aami/config-server/internal/domain"
	"github.com/fregataa/aami/config-server/test/testutil"
)

// AlertTemplate Tests

func TestAlertTemplateRepository_Create(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	templateRepo := repoManager.AlertTemplate
	ctx := context.Background()

	template := testutil.NewTestAlertTemplate("high-cpu", domain.AlertSeverityWarning)

	err := templateRepo.Create(ctx, template)
	require.NoError(t, err)
	assert.NotEmpty(t, template.ID)
}

func TestAlertTemplateRepository_GetByID(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	templateRepo := repoManager.AlertTemplate
	ctx := context.Background()

	template := testutil.NewTestAlertTemplate("memory-alert", domain.AlertSeverityCritical)
	require.NoError(t, templateRepo.Create(ctx, template))

	// Retrieve it
	retrieved, err := templateRepo.GetByID(ctx, template.ID)
	require.NoError(t, err)
	assert.Equal(t, template.ID, retrieved.ID)
	assert.Equal(t, template.Name, retrieved.Name)
	assert.Equal(t, template.Severity, retrieved.Severity)
}

func TestAlertTemplateRepository_GetByID_NotFound(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	templateRepo := repoManager.AlertTemplate
	ctx := context.Background()

	_, err := templateRepo.GetByID(ctx, "nonexistent-id")
	assert.Error(t, err)
}

func TestAlertTemplateRepository_Update(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	templateRepo := repoManager.AlertTemplate
	ctx := context.Background()

	template := testutil.NewTestAlertTemplate("disk-alert", domain.AlertSeverityWarning)
	require.NoError(t, templateRepo.Create(ctx, template))

	// Update it
	template.Description = "Updated description"
	template.Severity = domain.AlertSeverityCritical
	template.DefaultConfig["new_key"] = "new_value"
	err := templateRepo.Update(ctx, template)
	require.NoError(t, err)

	// Verify update
	retrieved, err := templateRepo.GetByID(ctx, template.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated description", retrieved.Description)
	assert.Equal(t, domain.AlertSeverityCritical, retrieved.Severity)
	assert.Equal(t, "new_value", retrieved.DefaultConfig["new_key"])
}

func TestAlertTemplateRepository_Delete(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	templateRepo := repoManager.AlertTemplate
	ctx := context.Background()

	template := testutil.NewTestAlertTemplate("temp-alert", domain.AlertSeverityInfo)
	require.NoError(t, templateRepo.Create(ctx, template))

	// Delete it
	err := templateRepo.Delete(ctx, template.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = templateRepo.GetByID(ctx, template.ID)
	assert.Error(t, err)
}

func TestAlertTemplateRepository_List(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	templateRepo := repoManager.AlertTemplate
	ctx := context.Background()

	// Create multiple templates
	template1 := testutil.NewTestAlertTemplate("alert1", domain.AlertSeverityWarning)
	template2 := testutil.NewTestAlertTemplate("alert2", domain.AlertSeverityCritical)
	template3 := testutil.NewTestAlertTemplate("alert3", domain.AlertSeverityInfo)

	require.NoError(t, templateRepo.Create(ctx, template1))
	require.NoError(t, templateRepo.Create(ctx, template2))
	require.NoError(t, templateRepo.Create(ctx, template3))

	// List all templates
	templates, total, err := templateRepo.List(ctx, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(templates), 3)
	assert.GreaterOrEqual(t, total, 3)
}

func TestAlertTemplateRepository_GetBySeverity(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	templateRepo := repoManager.AlertTemplate
	ctx := context.Background()

	// Create templates with different severities
	critical1 := testutil.NewTestAlertTemplate("critical1", domain.AlertSeverityCritical)
	critical2 := testutil.NewTestAlertTemplate("critical2", domain.AlertSeverityCritical)
	warning := testutil.NewTestAlertTemplate("warning1", domain.AlertSeverityWarning)

	require.NoError(t, templateRepo.Create(ctx, critical1))
	require.NoError(t, templateRepo.Create(ctx, critical2))
	require.NoError(t, templateRepo.Create(ctx, warning))

	// Get critical templates
	criticals, err := templateRepo.GetBySeverity(ctx, domain.AlertSeverityCritical)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(criticals), 2)

	// Verify all are critical
	for _, tmpl := range criticals {
		assert.Equal(t, domain.AlertSeverityCritical, tmpl.Severity)
	}
}

func TestAlertTemplateRepository_RenderQuery(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	templateRepo := repoManager.AlertTemplate
	ctx := context.Background()

	// Create template with query template
	template := testutil.NewTestAlertTemplate("cpu-alert", domain.AlertSeverityWarning)
	template.QueryTemplate = "cpu_usage > {{.threshold}}"
	template.DefaultConfig["threshold"] = 80
	require.NoError(t, templateRepo.Create(ctx, template))

	// Test RenderQuery with default config
	query, err := template.RenderQuery(nil)
	require.NoError(t, err)
	assert.Equal(t, "cpu_usage > 80", query)

	// Test RenderQuery with custom config
	query, err = template.RenderQuery(map[string]interface{}{"threshold": 90})
	require.NoError(t, err)
	assert.Equal(t, "cpu_usage > 90", query)
}

// AlertRule Tests

func TestAlertRuleRepository_Create(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	templateRepo := repoManager.AlertTemplate
	ruleRepo := repoManager.AlertRule
	ctx := context.Background()

	// Create group and template
	group := testutil.NewTestGroup("production", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	template := testutil.NewTestAlertTemplate("high-cpu", domain.AlertSeverityWarning)
	require.NoError(t, templateRepo.Create(ctx, template))

	// Create alert rule
	rule := testutil.NewTestAlertRule(group.ID, template.ID)

	err := ruleRepo.Create(ctx, rule)
	require.NoError(t, err)
	assert.NotEmpty(t, rule.ID)
}

func TestAlertRuleRepository_GetByID(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	templateRepo := repoManager.AlertTemplate
	ruleRepo := repoManager.AlertRule
	ctx := context.Background()

	// Create group, template, and rule
	group := testutil.NewTestGroup("staging", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	template := testutil.NewTestAlertTemplate("memory-alert", domain.AlertSeverityCritical)
	require.NoError(t, templateRepo.Create(ctx, template))

	rule := testutil.NewTestAlertRule(group.ID, template.ID)
	require.NoError(t, ruleRepo.Create(ctx, rule))

	// Retrieve it
	retrieved, err := ruleRepo.GetByID(ctx, rule.ID)
	require.NoError(t, err)
	assert.Equal(t, rule.ID, retrieved.ID)
	assert.Equal(t, rule.GroupID, retrieved.GroupID)
	assert.Equal(t, rule.TemplateID, retrieved.TemplateID)
}

func TestAlertRuleRepository_GetByID_NotFound(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	ruleRepo := repoManager.AlertRule
	ctx := context.Background()

	_, err := ruleRepo.GetByID(ctx, "nonexistent-id")
	assert.Error(t, err)
}

func TestAlertRuleRepository_Update(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	templateRepo := repoManager.AlertTemplate
	ruleRepo := repoManager.AlertRule
	ctx := context.Background()

	// Create group, template, and rule
	group := testutil.NewTestGroup("dev", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	template := testutil.NewTestAlertTemplate("disk-alert", domain.AlertSeverityWarning)
	require.NoError(t, templateRepo.Create(ctx, template))

	rule := testutil.NewTestAlertRule(group.ID, template.ID)
	require.NoError(t, ruleRepo.Create(ctx, rule))

	// Update it
	rule.Enabled = false
	rule.Config["threshold"] = 95
	rule.MergeStrategy = "override"
	rule.Priority = 200
	err := ruleRepo.Update(ctx, rule)
	require.NoError(t, err)

	// Verify update
	retrieved, err := ruleRepo.GetByID(ctx, rule.ID)
	require.NoError(t, err)
	assert.False(t, retrieved.Enabled)
	assert.Equal(t, 95, retrieved.Config["threshold"])
	assert.Equal(t, "override", retrieved.MergeStrategy)
	assert.Equal(t, 200, retrieved.Priority)
}

func TestAlertRuleRepository_Delete(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	templateRepo := repoManager.AlertTemplate
	ruleRepo := repoManager.AlertRule
	ctx := context.Background()

	// Create group, template, and rule
	group := testutil.NewTestGroup("temp", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	template := testutil.NewTestAlertTemplate("temp-alert", domain.AlertSeverityInfo)
	require.NoError(t, templateRepo.Create(ctx, template))

	rule := testutil.NewTestAlertRule(group.ID, template.ID)
	require.NoError(t, ruleRepo.Create(ctx, rule))

	// Delete it
	err := ruleRepo.Delete(ctx, rule.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = ruleRepo.GetByID(ctx, rule.ID)
	assert.Error(t, err)
}

func TestAlertRuleRepository_GetByGroupID(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	templateRepo := repoManager.AlertTemplate
	ruleRepo := repoManager.AlertRule
	ctx := context.Background()

	// Create groups and template
	group1 := testutil.NewTestGroup("group1", domain.NamespaceEnvironment)
	group2 := testutil.NewTestGroup("group2", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group1))
	require.NoError(t, groupRepo.Create(ctx, group2))

	template := testutil.NewTestAlertTemplate("cpu-alert", domain.AlertSeverityWarning)
	require.NoError(t, templateRepo.Create(ctx, template))

	// Create rules for different groups
	rule1 := testutil.NewTestAlertRule(group1.ID, template.ID)
	rule2 := testutil.NewTestAlertRule(group1.ID, template.ID)
	rule3 := testutil.NewTestAlertRule(group2.ID, template.ID)

	require.NoError(t, ruleRepo.Create(ctx, rule1))
	require.NoError(t, ruleRepo.Create(ctx, rule2))
	require.NoError(t, ruleRepo.Create(ctx, rule3))

	// Get rules by group
	group1Rules, err := ruleRepo.GetByGroupID(ctx, group1.ID)
	require.NoError(t, err)
	assert.Len(t, group1Rules, 2)

	// Verify all rules belong to group1
	for _, r := range group1Rules {
		assert.Equal(t, group1.ID, r.GroupID)
	}
}

func TestAlertRuleRepository_GetByTemplateID(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	templateRepo := repoManager.AlertTemplate
	ruleRepo := repoManager.AlertRule
	ctx := context.Background()

	// Create group and templates
	group := testutil.NewTestGroup("prod", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	template1 := testutil.NewTestAlertTemplate("template1", domain.AlertSeverityWarning)
	template2 := testutil.NewTestAlertTemplate("template2", domain.AlertSeverityCritical)
	require.NoError(t, templateRepo.Create(ctx, template1))
	require.NoError(t, templateRepo.Create(ctx, template2))

	// Create rules for different templates
	rule1 := testutil.NewTestAlertRule(group.ID, template1.ID)
	rule2 := testutil.NewTestAlertRule(group.ID, template1.ID)
	rule3 := testutil.NewTestAlertRule(group.ID, template2.ID)

	require.NoError(t, ruleRepo.Create(ctx, rule1))
	require.NoError(t, ruleRepo.Create(ctx, rule2))
	require.NoError(t, ruleRepo.Create(ctx, rule3))

	// Get rules by template
	template1Rules, err := ruleRepo.GetByTemplateID(ctx, template1.ID)
	require.NoError(t, err)
	assert.Len(t, template1Rules, 2)

	// Verify all rules use template1
	for _, r := range template1Rules {
		assert.Equal(t, template1.ID, r.TemplateID)
	}
}

func TestAlertRuleRepository_MergeStrategy(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	templateRepo := repoManager.AlertTemplate
	ruleRepo := repoManager.AlertRule
	ctx := context.Background()

	// Create parent and child groups
	parent := testutil.NewTestGroup("parent", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, parent))

	child := testutil.NewTestGroupWithParent("child", domain.NamespaceEnvironment, parent.ID)
	require.NoError(t, groupRepo.Create(ctx, child))

	template := testutil.NewTestAlertTemplate("cpu-alert", domain.AlertSeverityWarning)
	require.NoError(t, templateRepo.Create(ctx, template))

	// Create parent rule
	parentRule := testutil.NewTestAlertRule(parent.ID, template.ID)
	parentRule.Config["threshold"] = 80
	parentRule.Config["duration"] = "5m"
	parentRule.MergeStrategy = "merge"
	require.NoError(t, ruleRepo.Create(ctx, parentRule))

	// Create child rule with merge strategy
	childRule := testutil.NewTestAlertRule(child.ID, template.ID)
	childRule.Config["threshold"] = 90
	childRule.MergeStrategy = "merge"
	require.NoError(t, ruleRepo.Create(ctx, childRule))

	// Test MergeWith method
	merged := childRule.MergeWith(parentRule)
	assert.Equal(t, "5m", merged.Config["duration"])  // From parent
	assert.Equal(t, 90, merged.Config["threshold"])   // From child (overrides)
}

func TestAlertRuleRepository_List(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	templateRepo := repoManager.AlertTemplate
	ruleRepo := repoManager.AlertRule
	ctx := context.Background()

	// Create group and template
	group := testutil.NewTestGroup("prod", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	template := testutil.NewTestAlertTemplate("cpu-alert", domain.AlertSeverityWarning)
	require.NoError(t, templateRepo.Create(ctx, template))

	// Create multiple rules
	rule1 := testutil.NewTestAlertRule(group.ID, template.ID)
	rule2 := testutil.NewTestAlertRule(group.ID, template.ID)

	require.NoError(t, ruleRepo.Create(ctx, rule1))
	require.NoError(t, ruleRepo.Create(ctx, rule2))

	// List all rules
	rules, total, err := ruleRepo.List(ctx, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(rules), 2)
	assert.GreaterOrEqual(t, total, 2)
}
