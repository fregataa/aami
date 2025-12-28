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

func TestBootstrapTokenRepository_Create(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	tokenRepo := repoManager.BootstrapToken
	ctx := context.Background()

	// Create a group
	group := testutil.NewTestGroup("production", domain.NamespaceEnvironment)
	err := groupRepo.Create(ctx, group)
	require.NoError(t, err)

	// Create a bootstrap token
	token := testutil.NewTestBootstrapToken("test-token")

	err = tokenRepo.Create(ctx, token)
	require.NoError(t, err)
	assert.NotEmpty(t, token.ID)
	assert.NotEmpty(t, token.Token)
}

func TestBootstrapTokenRepository_GetByID(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	tokenRepo := repoManager.BootstrapToken
	ctx := context.Background()

	// Create group and token
	group := testutil.NewTestGroup("staging", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	token := testutil.NewTestBootstrapToken("staging-token")
	require.NoError(t, tokenRepo.Create(ctx, token))

	// Retrieve it
	retrieved, err := tokenRepo.GetByID(ctx, token.ID)
	require.NoError(t, err)
	assert.Equal(t, token.ID, retrieved.ID)
	assert.Equal(t, token.Name, retrieved.Name)
	assert.Equal(t, token.Token, retrieved.Token)
	assert.Equal(t, token.DefaultGroupID, retrieved.DefaultGroupID)
}

func TestBootstrapTokenRepository_GetByID_NotFound(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	tokenRepo := repoManager.BootstrapToken
	ctx := context.Background()

	_, err := tokenRepo.GetByID(ctx, "nonexistent-id")
	assert.Error(t, err)
}

func TestBootstrapTokenRepository_GetByToken(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	tokenRepo := repoManager.BootstrapToken
	ctx := context.Background()

	// Create group and token
	group := testutil.NewTestGroup("prod", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	token := testutil.NewTestBootstrapToken("prod-token")
	require.NoError(t, tokenRepo.Create(ctx, token))

	// Retrieve by token string
	retrieved, err := tokenRepo.GetByToken(ctx, token.Token)
	require.NoError(t, err)
	assert.Equal(t, token.ID, retrieved.ID)
	assert.Equal(t, token.Token, retrieved.Token)
}

func TestBootstrapTokenRepository_GetByToken_NotFound(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	tokenRepo := repoManager.BootstrapToken
	ctx := context.Background()

	_, err := tokenRepo.GetByToken(ctx, "nonexistent-token")
	assert.Error(t, err)
}

func TestBootstrapTokenRepository_Update(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	tokenRepo := repoManager.BootstrapToken
	ctx := context.Background()

	// Create group and token
	group := testutil.NewTestGroup("dev", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	token := testutil.NewTestBootstrapToken("dev-token")
	require.NoError(t, tokenRepo.Create(ctx, token))

	// Update it
	token.Uses = 5
	token.MaxUses = 20
	token.Labels["environment"] = "development"
	err := tokenRepo.Update(ctx, token)
	require.NoError(t, err)

	// Verify update
	retrieved, err := tokenRepo.GetByID(ctx, token.ID)
	require.NoError(t, err)
	assert.Equal(t, 5, retrieved.Uses)
	assert.Equal(t, 20, retrieved.MaxUses)
	assert.Equal(t, "development", retrieved.Labels["environment"])
}

func TestBootstrapTokenRepository_Delete(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	tokenRepo := repoManager.BootstrapToken
	ctx := context.Background()

	// Create group and token
	group := testutil.NewTestGroup("temp", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	token := testutil.NewTestBootstrapToken("temp-token")
	require.NoError(t, tokenRepo.Create(ctx, token))

	// Delete it
	err := tokenRepo.Delete(ctx, token.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = tokenRepo.GetByID(ctx, token.ID)
	assert.Error(t, err)
}

func TestBootstrapTokenRepository_List(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	tokenRepo := repoManager.BootstrapToken
	ctx := context.Background()

	// Create a group
	group := testutil.NewTestGroup("prod", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	// Create multiple tokens
	token1 := testutil.NewTestBootstrapToken("token1")
	token2 := testutil.NewTestBootstrapToken("token2")
	token3 := testutil.NewTestBootstrapToken("token3")

	require.NoError(t, tokenRepo.Create(ctx, token1))
	require.NoError(t, tokenRepo.Create(ctx, token2))
	require.NoError(t, tokenRepo.Create(ctx, token3))

	// List all tokens
	tokens, total, err := tokenRepo.List(ctx, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(tokens), 3)
	assert.GreaterOrEqual(t, total, 3)
}

func TestBootstrapTokenRepository_IsValid(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	tokenRepo := repoManager.BootstrapToken
	ctx := context.Background()

	// Create group
	group := testutil.NewTestGroup("prod", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	// Create valid token
	validToken := testutil.NewTestBootstrapToken("valid-token")
	require.NoError(t, tokenRepo.Create(ctx, validToken))

	// Create expired token
	expiredToken := testutil.NewTestBootstrapToken("expired-token")
	expiredToken.ExpiresAt = time.Now().Add(-1 * time.Hour)
	require.NoError(t, tokenRepo.Create(ctx, expiredToken))

	// Create exhausted token
	exhaustedToken := testutil.NewTestBootstrapToken("exhausted-token")
	exhaustedToken.Uses = exhaustedToken.MaxUses
	require.NoError(t, tokenRepo.Create(ctx, exhaustedToken))

	// Test validation
	assert.True(t, validToken.IsValid())
	assert.False(t, expiredToken.IsValid())
	assert.False(t, exhaustedToken.IsValid())
}

func TestBootstrapTokenRepository_CanUse(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	tokenRepo := repoManager.BootstrapToken
	ctx := context.Background()

	// Create group and token
	group := testutil.NewTestGroup("prod", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	token := testutil.NewTestBootstrapToken("test-token")
	require.NoError(t, tokenRepo.Create(ctx, token))

	// Should be usable initially
	assert.True(t, token.CanUse())

	// Use it multiple times
	for i := 0; i < token.MaxUses; i++ {
		assert.True(t, token.CanUse())
		err := token.IncrementUses()
		require.NoError(t, err)
		err = tokenRepo.Update(ctx, token)
		require.NoError(t, err)
	}

	// Should not be usable after exhaustion
	retrieved, err := tokenRepo.GetByID(ctx, token.ID)
	require.NoError(t, err)
	assert.False(t, retrieved.CanUse())
}

func TestBootstrapTokenRepository_IncrementUses(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	tokenRepo := repoManager.BootstrapToken
	ctx := context.Background()

	// Create group and token
	group := testutil.NewTestGroup("prod", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	token := testutil.NewTestBootstrapToken("test-token")
	require.NoError(t, tokenRepo.Create(ctx, token))
	assert.Equal(t, 0, token.Uses)

	// Increment uses
	err := token.IncrementUses()
	require.NoError(t, err)
	assert.Equal(t, 1, token.Uses)

	err = tokenRepo.Update(ctx, token)
	require.NoError(t, err)

	// Verify persistence
	retrieved, err := tokenRepo.GetByID(ctx, token.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, retrieved.Uses)

	// Increment again
	err = retrieved.IncrementUses()
	require.NoError(t, err)
	assert.Equal(t, 2, retrieved.Uses)

	err = tokenRepo.Update(ctx, retrieved)
	require.NoError(t, err)

	// Verify persistence again
	retrieved2, err := tokenRepo.GetByID(ctx, token.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, retrieved2.Uses)
}

func TestBootstrapTokenRepository_IncrementUses_Exhausted(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	tokenRepo := repoManager.BootstrapToken
	ctx := context.Background()

	// Create group and token
	group := testutil.NewTestGroup("prod", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	token := testutil.NewTestBootstrapToken("test-token")
	token.Uses = token.MaxUses // Already exhausted
	require.NoError(t, tokenRepo.Create(ctx, token))

	// Try to increment uses (should fail)
	err := token.IncrementUses()
	assert.Error(t, err)
	assert.Equal(t, token.MaxUses, token.Uses) // Should not increment
}

func TestBootstrapTokenRepository_IsExpired(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	tokenRepo := repoManager.BootstrapToken
	ctx := context.Background()

	// Create group
	group := testutil.NewTestGroup("prod", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	// Create token that expires in the future
	futureToken := testutil.NewTestBootstrapToken("future-token")
	futureToken.ExpiresAt = time.Now().Add(24 * time.Hour)
	require.NoError(t, tokenRepo.Create(ctx, futureToken))
	assert.False(t, futureToken.IsExpired())

	// Create token that expired in the past
	pastToken := testutil.NewTestBootstrapToken("past-token")
	pastToken.ExpiresAt = time.Now().Add(-24 * time.Hour)
	require.NoError(t, tokenRepo.Create(ctx, pastToken))
	assert.True(t, pastToken.IsExpired())
}

func TestBootstrapTokenRepository_RemainingUses(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	tokenRepo := repoManager.BootstrapToken
	ctx := context.Background()

	// Create group and token
	group := testutil.NewTestGroup("prod", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	token := testutil.NewTestBootstrapToken("test-token")
	require.NoError(t, tokenRepo.Create(ctx, token))

	// Initially: 0 uses, 10 max
	assert.Equal(t, 10, token.RemainingUses())

	// Use it 3 times
	for i := 0; i < 3; i++ {
		err := token.IncrementUses()
		require.NoError(t, err)
	}
	assert.Equal(t, 7, token.RemainingUses())

	// Use it until exhausted
	for i := 0; i < 7; i++ {
		err := token.IncrementUses()
		require.NoError(t, err)
	}
	assert.Equal(t, 0, token.RemainingUses())
}

func TestBootstrapTokenRepository_UsageScenario(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	tokenRepo := repoManager.BootstrapToken
	ctx := context.Background()

	// Create group and token with limited uses
	group := testutil.NewTestGroup("prod", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group))

	token := testutil.NewTestBootstrapToken("limited-token")
	token.MaxUses = 3
	require.NoError(t, tokenRepo.Create(ctx, token))

	// Scenario: Use token multiple times until exhausted
	for i := 0; i < 3; i++ {
		// Check if can use
		retrieved, err := tokenRepo.GetByToken(ctx, token.Token)
		require.NoError(t, err)
		assert.True(t, retrieved.CanUse())

		// Use it
		err = retrieved.IncrementUses()
		require.NoError(t, err)
		err = tokenRepo.Update(ctx, retrieved)
		require.NoError(t, err)
	}

	// After exhaustion, should not be usable
	retrieved, err := tokenRepo.GetByToken(ctx, token.Token)
	require.NoError(t, err)
	assert.False(t, retrieved.CanUse())
	assert.Equal(t, 0, retrieved.RemainingUses())

	// Try to use again (should fail)
	err = retrieved.IncrementUses()
	assert.Error(t, err)
}

func TestBootstrapTokenRepository_GetByGroupID(t *testing.T) {
	repoManager, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	groupRepo := repoManager.Group
	tokenRepo := repoManager.BootstrapToken
	ctx := context.Background()

	// Create groups
	group1 := testutil.NewTestGroup("group1", domain.NamespaceEnvironment)
	group2 := testutil.NewTestGroup("group2", domain.NamespaceEnvironment)
	require.NoError(t, groupRepo.Create(ctx, group1))
	require.NoError(t, groupRepo.Create(ctx, group2))

	// Create tokens for different groups
	token1 := testutil.NewTestBootstrapToken("token1")
	token2 := testutil.NewTestBootstrapToken("token2")
	token3 := testutil.NewTestBootstrapToken("token3")

	require.NoError(t, tokenRepo.Create(ctx, token1))
	require.NoError(t, tokenRepo.Create(ctx, token2))
	require.NoError(t, tokenRepo.Create(ctx, token3))

	// Get tokens by group
	group1Tokens, err := tokenRepo.GetByGroupID(ctx, group1.ID)
	require.NoError(t, err)
	assert.Len(t, group1Tokens, 2)

	// Verify all tokens belong to group1
	for _, tok := range group1Tokens {
		assert.Equal(t, group1.ID, tok.DefaultGroupID)
	}
}
