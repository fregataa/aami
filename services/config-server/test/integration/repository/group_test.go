package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/fregataa/aami/config-server/test/testutil"
)

func TestGroupRepository_Create(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := repoManager.Group
	ctx := context.Background()

	group := testutil.NewTestGroup("production")

	err := repo.Create(ctx, group)
	require.NoError(t, err)
	assert.NotEmpty(t, group.ID)
}

func TestGroupRepository_GetByID(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := repoManager.Group
	ctx := context.Background()

	// Create a group
	group := testutil.NewTestGroup("staging")
	err := repo.Create(ctx, group)
	require.NoError(t, err)

	// Retrieve it
	retrieved, err := repo.GetByID(ctx, group.ID)
	require.NoError(t, err)
	assert.Equal(t, group.ID, retrieved.ID)
	assert.Equal(t, group.Name, retrieved.Name)
}

func TestGroupRepository_GetByID_NotFound(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := repoManager.Group
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "nonexistent-id")
	assert.Error(t, err)
}

func TestGroupRepository_Update(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := repoManager.Group
	ctx := context.Background()

	// Create a group
	group := testutil.NewTestGroup("dev")
	err := repo.Create(ctx, group)
	require.NoError(t, err)

	// Update it
	group.Description = "Updated description"
	group.Priority = 200
	err = repo.Update(ctx, group)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(ctx, group.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated description", retrieved.Description)
	assert.Equal(t, 200, retrieved.Priority)
}

func TestGroupRepository_Delete(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := repoManager.Group
	ctx := context.Background()

	// Create a group
	group := testutil.NewTestGroup("temp")
	err := repo.Create(ctx, group)
	require.NoError(t, err)

	// Delete it
	err = repo.Delete(ctx, group.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(ctx, group.ID)
	assert.Error(t, err)
}

func TestGroupRepository_List(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := repoManager.Group
	ctx := context.Background()

	// Create multiple groups
	group1 := testutil.NewTestGroup("prod")
	group2 := testutil.NewTestGroup("staging")
	group3 := testutil.NewTestGroup("dev")

	require.NoError(t, repo.Create(ctx, group1))
	require.NoError(t, repo.Create(ctx, group2))
	require.NoError(t, repo.Create(ctx, group3))

	// List all groups (page 1, limit 10)
	groups, total, err := repo.List(ctx, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(groups), 3)
	assert.GreaterOrEqual(t, total, 3)
}

func TestGroupRepository_Purge(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := repoManager.Group
	ctx := context.Background()

	// Create group
	group := testutil.NewTestGroup("temp-group")
	require.NoError(t, repo.Create(ctx, group))

	// Purge (hard delete)
	err := repo.Purge(ctx, group.ID)
	require.NoError(t, err)

	// Verify it's deleted
	_, err = repo.GetByID(ctx, group.ID)
	assert.Error(t, err)
}
