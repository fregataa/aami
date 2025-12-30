package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/fregataa/aami/config-server/internal/domain"
	"github.com/fregataa/aami/config-server/test/testutil"
)

func TestGroup_GetPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority int
		want     int
	}{
		{
			name:     "returns custom priority when set",
			priority: 150,
			want:     150,
		},
		{
			name:     "returns default priority (100) when not set",
			priority: 0,
			want:     100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &domain.Group{
				Priority: tt.priority,
			}
			got := group.GetPriority()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGroup_Creation(t *testing.T) {
	// Test creating a group with testutil
	group := testutil.NewTestGroup("production")

	assert.NotEmpty(t, group.ID)
	assert.Equal(t, "production", group.Name)
	assert.Contains(t, group.Description, "production")
	assert.Equal(t, 100, group.Priority) // Default priority
	assert.NotNil(t, group.Metadata)
	assert.NotZero(t, group.CreatedAt)
	assert.NotZero(t, group.UpdatedAt)
}
