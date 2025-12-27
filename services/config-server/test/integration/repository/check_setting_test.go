package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/fregataa/aami/config-server/internal/domain"
	"github.com/fregataa/aami/config-server/test/testutil"
)

func TestCheckSettingRepository_Create(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	checkRepo := repoManager.CheckSetting
	ctx := context.Background()

	// Create a group
	group := testutil.NewTestGroup("production", domain.NamespaceEnvironment)
	err := groupRepo.Create(ctx, group)
	require.NoError(t, err)

	// Create a check setting
	checkSetting := testutil.NewTestCheckSetting(group.ID, "mount")

	err = checkRepo.Create(ctx, checkSetting)
	require.NoError(t, err)
	assert.NotEmpty(t, checkSetting.ID)
}

func TestCheckSettingRepository_GetByID(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	checkRepo := repoManager.CheckSetting
	ctx := context.Background()

	// Create group and check setting
	group := testutil.NewTestGroup("staging", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	checkSetting := testutil.NewTestCheckSetting(group.ID, "disk")
	require.NoError(t, checkRepo.Create(ctx, checkSetting))

	// Retrieve it
	retrieved, err := checkRepo.GetByID(ctx, checkSetting.ID)
	require.NoError(t, err)
	assert.Equal(t, checkSetting.ID, retrieved.ID)
	assert.Equal(t, checkSetting.GroupID, retrieved.GroupID)
	assert.Equal(t, checkSetting.CheckType, retrieved.CheckType)
}

func TestCheckSettingRepository_GetByID_NotFound(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	checkRepo := repoManager.CheckSetting
	ctx := context.Background()

	_, err := checkRepo.GetByID(ctx, "nonexistent-id")
	assert.Error(t, err)
}

func TestCheckSettingRepository_Update(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	checkRepo := repoManager.CheckSetting
	ctx := context.Background()

	// Create group and check setting
	group := testutil.NewTestGroup("dev", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	checkSetting := testutil.NewTestCheckSetting(group.ID, "network")
	require.NoError(t, checkRepo.Create(ctx, checkSetting))

	// Update it
	checkSetting.Config["timeout"] = "30s"
	checkSetting.MergeStrategy = "override"
	checkSetting.Priority = 200
	err := checkRepo.Update(ctx, checkSetting)
	require.NoError(t, err)

	// Verify update
	retrieved, err := checkRepo.GetByID(ctx, checkSetting.ID)
	require.NoError(t, err)
	assert.Equal(t, "30s", retrieved.Config["timeout"])
	assert.Equal(t, "override", retrieved.MergeStrategy)
	assert.Equal(t, 200, retrieved.Priority)
}

func TestCheckSettingRepository_Delete(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	checkRepo := repoManager.CheckSetting
	ctx := context.Background()

	// Create group and check setting
	group := testutil.NewTestGroup("temp", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	checkSetting := testutil.NewTestCheckSetting(group.ID, "temp-check")
	require.NoError(t, checkRepo.Create(ctx, checkSetting))

	// Delete it
	err := checkRepo.Delete(ctx, checkSetting.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = checkRepo.GetByID(ctx, checkSetting.ID)
	assert.Error(t, err)
}

func TestCheckSettingRepository_GetByGroupID(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	checkRepo := repoManager.CheckSetting
	ctx := context.Background()

	// Create groups
	group1 := testutil.NewTestGroup("group1", domain.NamespaceEnvironment)
	group2 := testutil.NewTestGroup("group2", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group1))
	require.NoError(t, groupRepo.Create(ctx, group2))

	// Create check settings for different groups
	check1 := testutil.NewTestCheckSetting(group1.ID, "mount")
	check2 := testutil.NewTestCheckSetting(group1.ID, "disk")
	check3 := testutil.NewTestCheckSetting(group2.ID, "network")

	require.NoError(t, checkRepo.Create(ctx, check1))
	require.NoError(t, checkRepo.Create(ctx, check2))
	require.NoError(t, checkRepo.Create(ctx, check3))

	// Get check settings by group
	group1Checks, err := checkRepo.GetByGroupID(ctx, group1.ID)
	require.NoError(t, err)
	assert.Len(t, group1Checks, 2)

	// Verify all checks belong to group1
	for _, c := range group1Checks {
		assert.Equal(t, group1.ID, c.GroupID)
	}
}

func TestCheckSettingRepository_GetByCheckType(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	checkRepo := repoManager.CheckSetting
	ctx := context.Background()

	// Create groups
	group1 := testutil.NewTestGroup("group1", domain.NamespaceEnvironment)
	group2 := testutil.NewTestGroup("group2", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group1))
	require.NoError(t, groupRepo.Create(ctx, group2))

	// Create check settings of different types
	mountCheck1 := testutil.NewTestCheckSetting(group1.ID, "mount")
	mountCheck2 := testutil.NewTestCheckSetting(group2.ID, "mount")
	diskCheck := testutil.NewTestCheckSetting(group1.ID, "disk")

	require.NoError(t, checkRepo.Create(ctx, mountCheck1))
	require.NoError(t, checkRepo.Create(ctx, mountCheck2))
	require.NoError(t, checkRepo.Create(ctx, diskCheck))

	// Get mount checks
	mountChecks, err := checkRepo.GetByCheckType(ctx, "mount")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(mountChecks), 2)

	// Verify all are mount checks
	for _, c := range mountChecks {
		assert.Equal(t, "mount", c.CheckType)
	}
}

func TestCheckSettingRepository_MergeStrategy(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	checkRepo := repoManager.CheckSetting
	ctx := context.Background()

	// Create parent and child groups
	parent := testutil.NewTestGroup("parent", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, parent))

	child := testutil.NewTestGroupWithParent("child", domain.NamespaceEnvironment, parent.ID)
	require.NoError(t, groupRepo.Create(ctx, child))

	// Create parent check setting
	parentCheck := testutil.NewTestCheckSetting(parent.ID, "mount")
	parentCheck.Config["path"] = "/data"
	parentCheck.Config["threshold"] = 80
	parentCheck.MergeStrategy = "merge"
	require.NoError(t, checkRepo.Create(ctx, parentCheck))

	// Create child check setting with merge strategy
	childCheck := testutil.NewTestCheckSetting(child.ID, "mount")
	childCheck.Config["threshold"] = 90
	childCheck.MergeStrategy = "merge"
	require.NoError(t, checkRepo.Create(ctx, childCheck))

	// Test MergeWith method
	merged := childCheck.MergeWith(parentCheck)
	assert.Equal(t, "/data", merged.Config["path"])    // From parent
	assert.Equal(t, 90, merged.Config["threshold"])    // From child (overrides)
}

func TestCheckSettingRepository_OverrideStrategy(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	checkRepo := repoManager.CheckSetting
	ctx := context.Background()

	// Create parent and child groups
	parent := testutil.NewTestGroup("parent", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, parent))

	child := testutil.NewTestGroupWithParent("child", domain.NamespaceEnvironment, parent.ID)
	require.NoError(t, groupRepo.Create(ctx, child))

	// Create parent check setting
	parentCheck := testutil.NewTestCheckSetting(parent.ID, "disk")
	parentCheck.Config["path"] = "/data"
	parentCheck.Config["threshold"] = 80
	require.NoError(t, checkRepo.Create(ctx, parentCheck))

	// Create child check setting with override strategy
	childCheck := testutil.NewTestCheckSetting(child.ID, "disk")
	childCheck.Config["threshold"] = 90
	childCheck.MergeStrategy = "override"
	require.NoError(t, checkRepo.Create(ctx, childCheck))

	// Test MergeWith method with override
	merged := childCheck.MergeWith(parentCheck)
	assert.Nil(t, merged.Config["path"])            // Not from parent
	assert.Equal(t, 90, merged.Config["threshold"]) // Only from child
}

func TestCheckSettingRepository_ConfigHelpers(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	checkRepo := repoManager.CheckSetting
	ctx := context.Background()

	// Create group and check setting
	group := testutil.NewTestGroup("prod", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	checkSetting := testutil.NewTestCheckSetting(group.ID, "mount")
	require.NoError(t, checkRepo.Create(ctx, checkSetting))

	// Test GetConfigValue
	value, exists := checkSetting.GetConfigValue("path")
	assert.True(t, exists)
	assert.Equal(t, "/mnt/data", value)

	// Test SetConfigValue
	checkSetting.SetConfigValue("new_key", "new_value")
	value, exists = checkSetting.GetConfigValue("new_key")
	assert.True(t, exists)
	assert.Equal(t, "new_value", value)

	// Update in database
	err := checkRepo.Update(ctx, checkSetting)
	require.NoError(t, err)

	// Verify persistence
	retrieved, err := checkRepo.GetByID(ctx, checkSetting.ID)
	require.NoError(t, err)
	value, exists = retrieved.GetConfigValue("new_key")
	assert.True(t, exists)
	assert.Equal(t, "new_value", value)
}

func TestCheckSettingRepository_ToJSON(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	checkRepo := repoManager.CheckSetting
	ctx := context.Background()

	// Create group and check setting
	group := testutil.NewTestGroup("prod", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	checkSetting := testutil.NewTestCheckSetting(group.ID, "mount")
	checkSetting.Config["path"] = "/mnt/data"
	checkSetting.Config["threshold"] = 80
	require.NoError(t, checkRepo.Create(ctx, checkSetting))

	// Test ToJSON method
	jsonStr, err := checkSetting.ToJSON()
	require.NoError(t, err)
	assert.Contains(t, jsonStr, "/mnt/data")
	assert.Contains(t, jsonStr, "80")
}

func TestCheckSettingRepository_List(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	checkRepo := repoManager.CheckSetting
	ctx := context.Background()

	// Create a group
	group := testutil.NewTestGroup("prod", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	// Create multiple check settings
	check1 := testutil.NewTestCheckSetting(group.ID, "mount")
	check2 := testutil.NewTestCheckSetting(group.ID, "disk")
	check3 := testutil.NewTestCheckSetting(group.ID, "network")

	require.NoError(t, checkRepo.Create(ctx, check1))
	require.NoError(t, checkRepo.Create(ctx, check2))
	require.NoError(t, checkRepo.Create(ctx, check3))

	// List all check settings
	checks, total, err := checkRepo.List(ctx, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(checks), 3)
	assert.GreaterOrEqual(t, total, 3)
}
