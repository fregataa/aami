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

func TestGroupRepository_GetChildren(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := repoManager.Group
	ctx := context.Background()

	// Create parent group
	parent := testutil.NewTestGroup("us")
	require.NoError(t, repo.Create(ctx, parent))

	// Create child groups
	child1 := testutil.NewTestGroupWithParent("us-west", parent.ID)
	child2 := testutil.NewTestGroupWithParent("us-east", parent.ID)

	require.NoError(t, repo.Create(ctx, child1))
	require.NoError(t, repo.Create(ctx, child2))

	// Get children
	children, err := repo.GetChildren(ctx, parent.ID)
	require.NoError(t, err)
	assert.Len(t, children, 2)

	// Verify parent IDs
	for _, child := range children {
		assert.NotNil(t, child.ParentID)
		assert.Equal(t, parent.ID, *child.ParentID)
	}
}

func TestGroupRepository_GetAncestors(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := repoManager.Group
	ctx := context.Background()

	// Create hierarchy: root -> child -> grandchild
	root := testutil.NewTestGroup("root")
	require.NoError(t, repo.Create(ctx, root))

	child := testutil.NewTestGroupWithParent("child", root.ID)
	require.NoError(t, repo.Create(ctx, child))

	grandchild := testutil.NewTestGroupWithParent("grandchild", child.ID)
	require.NoError(t, repo.Create(ctx, grandchild))

	// Get ancestors of grandchild
	ancestors, err := repo.GetAncestors(ctx, grandchild.ID)
	require.NoError(t, err)
	assert.Len(t, ancestors, 2) // Should include child and root

	// Verify order (should be from direct parent upwards)
	assert.Equal(t, child.ID, ancestors[0].ID)
	assert.Equal(t, root.ID, ancestors[1].ID)
}

func TestGroupRepository_HierarchicalOperations(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := repoManager.Group
	ctx := context.Background()

	// Create a complex hierarchy
	// root
	// ├── branch1
	// │   ├── leaf1
	// │   └── leaf2
	// └── branch2
	//     └── leaf3

	root := testutil.NewTestGroup("root")
	require.NoError(t, repo.Create(ctx, root))

	branch1 := testutil.NewTestGroupWithParent("branch1", root.ID)
	branch2 := testutil.NewTestGroupWithParent("branch2", root.ID)
	require.NoError(t, repo.Create(ctx, branch1))
	require.NoError(t, repo.Create(ctx, branch2))

	leaf1 := testutil.NewTestGroupWithParent("leaf1", branch1.ID)
	leaf2 := testutil.NewTestGroupWithParent("leaf2", branch1.ID)
	leaf3 := testutil.NewTestGroupWithParent("leaf3", branch2.ID)
	require.NoError(t, repo.Create(ctx, leaf1))
	require.NoError(t, repo.Create(ctx, leaf2))
	require.NoError(t, repo.Create(ctx, leaf3))

	// Test getting children at different levels
	rootChildren, err := repo.GetChildren(ctx, root.ID)
	require.NoError(t, err)
	assert.Len(t, rootChildren, 2)

	branch1Children, err := repo.GetChildren(ctx, branch1.ID)
	require.NoError(t, err)
	assert.Len(t, branch1Children, 2)

	// Test getting ancestors from leaf
	leaf1Ancestors, err := repo.GetAncestors(ctx, leaf1.ID)
	require.NoError(t, err)
	assert.Len(t, leaf1Ancestors, 2) // branch1 and root
}

func TestGroupRepository_CascadeDelete(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := repoManager.Group
	ctx := context.Background()

	// Create parent and child
	parent := testutil.NewTestGroup("parent")
	require.NoError(t, repo.Create(ctx, parent))

	child := testutil.NewTestGroupWithParent("child", parent.ID)
	require.NoError(t, repo.Create(ctx, child))

	// Purge parent (hard delete - should cascade to child due to ON DELETE CASCADE)
	err := repo.Purge(ctx, parent.ID)
	require.NoError(t, err)

	// Verify both are deleted (hard deleted, not found)
	_, err = repo.GetByID(ctx, parent.ID)
	assert.Error(t, err)

	_, err = repo.GetByID(ctx, child.ID)
	assert.Error(t, err)
}
