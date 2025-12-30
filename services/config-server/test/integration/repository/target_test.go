package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/fregataa/aami/config-server/internal/domain"
	"github.com/fregataa/aami/config-server/test/testutil"
)

func TestTargetRepository_Create(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	ctx := context.Background()

	// Create a group first
	group := testutil.NewTestGroup("production")
	err := groupRepo.Create(ctx, group)
	require.NoError(t, err)

	// Create a target
	target := testutil.NewTestTarget("server1", "192.168.1.10", []domain.Group{*group})

	err = targetRepo.Create(ctx, target)
	require.NoError(t, err)
	assert.NotEmpty(t, target.ID)
	assert.Equal(t, domain.TargetStatusActive, target.Status)
}

func TestTargetRepository_GetByID(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	ctx := context.Background()

	// Create group and target
	group := testutil.NewTestGroup("staging")
	require.NoError(t, groupRepo.Create(ctx, group))

	target := testutil.NewTestTarget("server2", "192.168.1.20", []domain.Group{*group})
	require.NoError(t, targetRepo.Create(ctx, target))

	// Retrieve it
	retrieved, err := targetRepo.GetByID(ctx, target.ID)
	require.NoError(t, err)
	assert.Equal(t, target.ID, retrieved.ID)
	assert.Equal(t, target.Hostname, retrieved.Hostname)
	assert.Equal(t, target.IPAddress, retrieved.IPAddress)
	assert.Len(t, retrieved.Groups, 1)
	assert.Equal(t, group.ID, retrieved.Groups[0].ID)
}

func TestTargetRepository_GetByID_NotFound(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	targetRepo := repoManager.Target
	ctx := context.Background()

	_, err := targetRepo.GetByID(ctx, "nonexistent-id")
	assert.Error(t, err)
}

func TestTargetRepository_Update(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	ctx := context.Background()

	// Create group and target
	group := testutil.NewTestGroup("dev")
	require.NoError(t, groupRepo.Create(ctx, group))

	target := testutil.NewTestTarget("server3", "192.168.1.30", []domain.Group{*group})
	require.NoError(t, targetRepo.Create(ctx, target))

	// Update it
	target.IPAddress = "192.168.1.31"
	target.Status = domain.TargetStatusInactive
	target.Labels["environment"] = "development"
	err := targetRepo.Update(ctx, target)
	require.NoError(t, err)

	// Verify update
	retrieved, err := targetRepo.GetByID(ctx, target.ID)
	require.NoError(t, err)
	assert.Equal(t, "192.168.1.31", retrieved.IPAddress)
	assert.Equal(t, domain.TargetStatusInactive, retrieved.Status)
	assert.Equal(t, "development", retrieved.Labels["environment"])
}

func TestTargetRepository_Delete(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	ctx := context.Background()

	// Create group and target
	group := testutil.NewTestGroup("temp")
	require.NoError(t, groupRepo.Create(ctx, group))

	target := testutil.NewTestTarget("server-temp", "192.168.1.99", []domain.Group{*group})
	require.NoError(t, targetRepo.Create(ctx, target))

	// Delete it
	err := targetRepo.Delete(ctx, target.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = targetRepo.GetByID(ctx, target.ID)
	assert.Error(t, err)
}

func TestTargetRepository_List(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	ctx := context.Background()

	// Create a group
	group := testutil.NewTestGroup("prod")
	require.NoError(t, groupRepo.Create(ctx, group))

	// Create multiple targets
	target1 := testutil.NewTestTarget("server1", "192.168.1.1", []domain.Group{*group})
	target2 := testutil.NewTestTarget("server2", "192.168.1.2", []domain.Group{*group})
	target3 := testutil.NewTestTarget("server3", "192.168.1.3", []domain.Group{*group})

	require.NoError(t, targetRepo.Create(ctx, target1))
	require.NoError(t, targetRepo.Create(ctx, target2))
	require.NoError(t, targetRepo.Create(ctx, target3))

	// List all targets
	targets, total, err := targetRepo.List(ctx, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(targets), 3)
	assert.GreaterOrEqual(t, total, 3)
}

func TestTargetRepository_GetByGroupID(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	ctx := context.Background()

	// Create groups
	group1 := testutil.NewTestGroup("group1")
	group2 := testutil.NewTestGroup("group2")
	require.NoError(t, groupRepo.Create(ctx, group1))
	require.NoError(t, groupRepo.Create(ctx, group2))

	// Create targets in different groups
	target1 := testutil.NewTestTarget("server1", "192.168.1.1", []domain.Group{*group1})
	target2 := testutil.NewTestTarget("server2", "192.168.1.2", []domain.Group{*group1})
	target3 := testutil.NewTestTarget("server3", "192.168.1.3", []domain.Group{*group2})

	require.NoError(t, targetRepo.Create(ctx, target1))
	require.NoError(t, targetRepo.Create(ctx, target2))
	require.NoError(t, targetRepo.Create(ctx, target3))

	// Get targets by group
	group1Targets, err := targetRepo.GetByGroupID(ctx, group1.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(group1Targets), 2)

	// Verify all targets belong to group1
	for _, tgt := range group1Targets {
		hasGroup := false
		for _, g := range tgt.Groups {
			if g.ID == group1.ID {
				hasGroup = true
				break
			}
		}
		assert.True(t, hasGroup, "Target should belong to group1")
	}
}

func TestTargetRepository_UpdateStatus(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	ctx := context.Background()

	// Create group and target
	group := testutil.NewTestGroup("prod")
	require.NoError(t, groupRepo.Create(ctx, group))

	target := testutil.NewTestTarget("server1", "192.168.1.1", []domain.Group{*group})
	require.NoError(t, targetRepo.Create(ctx, target))
	assert.Equal(t, domain.TargetStatusActive, target.Status)

	// Update status using domain method
	target.UpdateStatus(domain.TargetStatusDown)
	err := targetRepo.Update(ctx, target)
	require.NoError(t, err)

	// Verify status update
	retrieved, err := targetRepo.GetByID(ctx, target.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.TargetStatusDown, retrieved.Status)
}

func TestTargetRepository_UpdateLastSeen(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	ctx := context.Background()

	// Create group and target
	group := testutil.NewTestGroup("prod")
	require.NoError(t, groupRepo.Create(ctx, group))

	target := testutil.NewTestTarget("server1", "192.168.1.1", []domain.Group{*group})
	require.NoError(t, targetRepo.Create(ctx, target))
	assert.Nil(t, target.LastSeen)

	// Update last seen using domain method
	target.UpdateLastSeen()
	err := targetRepo.Update(ctx, target)
	require.NoError(t, err)

	// Verify last seen update
	retrieved, err := targetRepo.GetByID(ctx, target.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved.LastSeen)
	assert.True(t, time.Since(*retrieved.LastSeen) < 1*time.Second)
}

func TestTargetRepository_GetByHostname(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	ctx := context.Background()

	// Create group and target
	group := testutil.NewTestGroup("prod")
	require.NoError(t, groupRepo.Create(ctx, group))

	target := testutil.NewTestTarget("unique-hostname", "192.168.1.1", []domain.Group{*group})
	require.NoError(t, targetRepo.Create(ctx, target))

	// Get by hostname
	retrieved, err := targetRepo.GetByHostname(ctx, "unique-hostname")
	require.NoError(t, err)
	assert.Equal(t, target.ID, retrieved.ID)
	assert.Equal(t, "unique-hostname", retrieved.Hostname)
}

func TestTargetRepository_GetByHostname_NotFound(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	targetRepo := repoManager.Target
	ctx := context.Background()

	_, err := targetRepo.GetByHostname(ctx, "nonexistent-hostname")
	assert.Error(t, err)
}

func TestTargetRepository_HealthCheck(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	targetRepo := repoManager.Target
	ctx := context.Background()

	// Create group
	group := testutil.NewTestGroup("prod")
	require.NoError(t, groupRepo.Create(ctx, group))

	// Create healthy target (recently seen)
	healthyTarget := testutil.NewTestTarget("healthy", "192.168.1.1", []domain.Group{*group})
	healthyTarget.UpdateLastSeen()
	require.NoError(t, targetRepo.Create(ctx, healthyTarget))

	// Create unhealthy target (not seen)
	unhealthyTarget := testutil.NewTestTarget("unhealthy", "192.168.1.2", []domain.Group{*group})
	require.NoError(t, targetRepo.Create(ctx, unhealthyTarget))

	// Create down target
	downTarget := testutil.NewTestTarget("down", "192.168.1.3", []domain.Group{*group})
	downTarget.UpdateStatus(domain.TargetStatusDown)
	require.NoError(t, targetRepo.Create(ctx, downTarget))

	// Verify health status
	retrieved1, err := targetRepo.GetByID(ctx, healthyTarget.ID)
	require.NoError(t, err)
	assert.True(t, retrieved1.IsHealthy())

	retrieved2, err := targetRepo.GetByID(ctx, unhealthyTarget.ID)
	require.NoError(t, err)
	assert.False(t, retrieved2.IsHealthy())

	retrieved3, err := targetRepo.GetByID(ctx, downTarget.ID)
	require.NoError(t, err)
	assert.False(t, retrieved3.IsHealthy())
}
