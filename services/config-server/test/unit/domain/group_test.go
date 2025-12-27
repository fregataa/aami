package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/fregataa/aami/config-server/internal/domain"
	"github.com/fregataa/aami/config-server/test/testutil"
)

func TestGroup_IsRoot(t *testing.T) {
	tests := []struct {
		name     string
		parentID *string
		want     bool
	}{
		{
			name:     "root group has no parent",
			parentID: nil,
			want:     true,
		},
		{
			name:     "child group has parent",
			parentID: testutil.StringPtr("parent-id"),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &domain.Group{
				ParentID: tt.parentID,
			}
			got := group.IsRoot()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGroup_GetPriority(t *testing.T) {
	tests := []struct {
		name      string
		priority  int
		namespace domain.Namespace
		want      int
	}{
		{
			name:      "returns custom priority when set",
			priority:  150,
			namespace: domain.NamespaceEnvironment,
			want:      150,
		},
		{
			name:      "returns namespace priority when not set",
			priority:  0,
			namespace: domain.NamespaceEnvironment,
			want:      10, // Environment namespace priority
		},
		{
			name:      "returns namespace priority for logical",
			priority:  0,
			namespace: domain.NamespaceLogical,
			want:      50, // Logical namespace priority
		},
		{
			name:      "returns namespace priority for infrastructure",
			priority:  0,
			namespace: domain.NamespaceInfrastructure,
			want:      100, // Infrastructure namespace priority
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &domain.Group{
				Priority:  tt.priority,
				Namespace: tt.namespace,
			}
			got := group.GetPriority()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGroup_Creation(t *testing.T) {
	// Test creating a group with testutil
	group := testutil.NewTestGroup("production", domain.NamespaceEnvironment)

	assert.NotEmpty(t, group.ID)
	assert.Equal(t, "production", group.Name)
	assert.Equal(t, domain.NamespaceEnvironment, group.Namespace)
	assert.Contains(t, group.Description, "production")
	assert.Equal(t, 10, group.Priority) // Environment priority
	assert.NotNil(t, group.Metadata)
	assert.NotZero(t, group.CreatedAt)
	assert.NotZero(t, group.UpdatedAt)
	assert.Nil(t, group.ParentID) // Root group
}

func TestGroup_CreationWithParent(t *testing.T) {
	// Test creating a child group
	parentID := "parent-group-id"
	group := testutil.NewTestGroupWithParent("us-west", domain.NamespaceEnvironment, parentID)

	assert.NotEmpty(t, group.ID)
	assert.Equal(t, "us-west", group.Name)
	assert.NotNil(t, group.ParentID)
	assert.Equal(t, parentID, *group.ParentID)
	assert.False(t, group.IsRoot())
}

func TestGroup_ValidNamespaces(t *testing.T) {
	namespaces := []domain.Namespace{
		domain.NamespaceInfrastructure,
		domain.NamespaceLogical,
		domain.NamespaceEnvironment,
	}

	for _, ns := range namespaces {
		t.Run(string(ns), func(t *testing.T) {
			group := testutil.NewTestGroup("test", ns)
			assert.Equal(t, ns, group.Namespace)
			assert.True(t, group.Namespace.IsValid())
		})
	}
}
