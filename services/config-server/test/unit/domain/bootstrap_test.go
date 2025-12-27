package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/fregataa/aami/config-server/internal/domain"
	"github.com/fregataa/aami/config-server/test/testutil"
)

func TestBootstrapToken_IsValid(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		token     *domain.BootstrapToken
		want      bool
	}{
		{
			name: "valid token",
			token: &domain.BootstrapToken{
				MaxUses:   10,
				Uses:      5,
				ExpiresAt: now.Add(24 * time.Hour),
			},
			want: true,
		},
		{
			name: "expired token",
			token: &domain.BootstrapToken{
				MaxUses:   10,
				Uses:      5,
				ExpiresAt: now.Add(-1 * time.Hour),
			},
			want: false,
		},
		{
			name: "exhausted token",
			token: &domain.BootstrapToken{
				MaxUses:   10,
				Uses:      10,
				ExpiresAt: now.Add(24 * time.Hour),
			},
			want: false,
		},
		{
			name: "over-used token",
			token: &domain.BootstrapToken{
				MaxUses:   10,
				Uses:      15,
				ExpiresAt: now.Add(24 * time.Hour),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.token.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBootstrapToken_IsExpired(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "not expired",
			expiresAt: now.Add(24 * time.Hour),
			want:      false,
		},
		{
			name:      "expired 1 hour ago",
			expiresAt: now.Add(-1 * time.Hour),
			want:      true,
		},
		{
			name:      "expires in 1 second",
			expiresAt: now.Add(1 * time.Second),
			want:      false,
		},
		{
			name:      "expired 1 second ago",
			expiresAt: now.Add(-1 * time.Second),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &domain.BootstrapToken{
				ExpiresAt: tt.expiresAt,
			}
			got := token.IsExpired()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBootstrapToken_CanUse(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		maxUses   int
		uses      int
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "can use valid token",
			maxUses:   10,
			uses:      5,
			expiresAt: now.Add(24 * time.Hour),
			want:      true,
		},
		{
			name:      "cannot use expired token",
			maxUses:   10,
			uses:      5,
			expiresAt: now.Add(-1 * time.Hour),
			want:      false,
		},
		{
			name:      "cannot use exhausted token",
			maxUses:   10,
			uses:      10,
			expiresAt: now.Add(24 * time.Hour),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &domain.BootstrapToken{
				MaxUses:   tt.maxUses,
				Uses:      tt.uses,
				ExpiresAt: tt.expiresAt,
			}
			got := token.CanUse()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBootstrapToken_IncrementUses(t *testing.T) {
	now := time.Now()
	token := &domain.BootstrapToken{
		MaxUses:   10,
		Uses:      5,
		ExpiresAt: now.Add(24 * time.Hour),
	}

	err := token.IncrementUses()
	assert.NoError(t, err)
	assert.Equal(t, 6, token.Uses)

	err = token.IncrementUses()
	assert.NoError(t, err)
	assert.Equal(t, 7, token.Uses)
}

func TestBootstrapToken_Creation(t *testing.T) {
	groupID := "group-123"
	token := testutil.NewTestBootstrapToken("test-token", groupID)

	assert.NotEmpty(t, token.ID)
	assert.NotEmpty(t, token.Token)
	assert.Equal(t, "test-token", token.Name)
	assert.Equal(t, groupID, token.DefaultGroupID)
	assert.Equal(t, 10, token.MaxUses)
	assert.Equal(t, 0, token.Uses)
	assert.True(t, token.ExpiresAt.After(time.Now()))
	assert.NotNil(t, token.Labels)
	assert.NotZero(t, token.CreatedAt)
	assert.NotZero(t, token.UpdatedAt)
}

func TestGenerateToken(t *testing.T) {
	// Test token generation
	token1, err := domain.GenerateToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token1)
	// Token is base64 URL encoded from 32 bytes, resulting in 44 characters
	assert.Len(t, token1, 44)

	// Test that tokens are unique
	token2, err := domain.GenerateToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token2)
	assert.NotEqual(t, token1, token2)
}

func TestBootstrapToken_UsageScenario(t *testing.T) {
	// Scenario: Create token, use it multiple times until exhausted
	now := time.Now()
	token := &domain.BootstrapToken{
		MaxUses:   3,
		Uses:      0,
		ExpiresAt: now.Add(24 * time.Hour),
	}

	// Initially valid and can be used
	assert.True(t, token.IsValid())
	assert.True(t, token.CanUse())
	assert.False(t, token.IsExpired())

	// Use 1
	err := token.IncrementUses()
	assert.NoError(t, err)
	assert.Equal(t, 1, token.Uses)
	assert.True(t, token.IsValid())
	assert.True(t, token.CanUse())

	// Use 2
	err = token.IncrementUses()
	assert.NoError(t, err)
	assert.Equal(t, 2, token.Uses)
	assert.True(t, token.IsValid())
	assert.True(t, token.CanUse())

	// Use 3 (exhausted)
	err = token.IncrementUses()
	assert.NoError(t, err)
	assert.Equal(t, 3, token.Uses)
	assert.False(t, token.IsValid())
	assert.False(t, token.CanUse())

	// Try to use again (should fail)
	err = token.IncrementUses()
	assert.Error(t, err)
	assert.Equal(t, 3, token.Uses) // Uses should not increment
}

func TestBootstrapToken_ExpirationScenario(t *testing.T) {
	// Scenario: Token expires before being exhausted
	now := time.Now()
	token := &domain.BootstrapToken{
		MaxUses:   10,
		Uses:      3,
		ExpiresAt: now.Add(-1 * time.Hour), // Expired 1 hour ago
	}

	assert.True(t, token.IsExpired())
	assert.False(t, token.IsValid())
	assert.False(t, token.CanUse())
}

func TestBootstrapToken_RemainingUses(t *testing.T) {
	// Test remaining uses calculation
	token := &domain.BootstrapToken{
		MaxUses: 10,
		Uses:    3,
	}

	remaining := token.RemainingUses()
	assert.Equal(t, 7, remaining)

	// After more uses
	token.Uses = 9
	remaining = token.RemainingUses()
	assert.Equal(t, 1, remaining)

	// When exhausted
	token.Uses = 10
	remaining = token.RemainingUses()
	assert.Equal(t, 0, remaining)

	// When over-used (should return 0)
	token.Uses = 15
	remaining = token.RemainingUses()
	assert.Equal(t, 0, remaining)
}
