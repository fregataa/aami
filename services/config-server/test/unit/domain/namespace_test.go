package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/fregataa/aami/config-server/internal/domain"
)

func TestNamespace_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		namespace domain.Namespace
		want      bool
	}{
		{
			name:      "infrastructure namespace is valid",
			namespace: domain.NamespaceInfrastructure,
			want:      true,
		},
		{
			name:      "logical namespace is valid",
			namespace: domain.NamespaceLogical,
			want:      true,
		},
		{
			name:      "environment namespace is valid",
			namespace: domain.NamespaceEnvironment,
			want:      true,
		},
		{
			name:      "invalid namespace returns false",
			namespace: domain.Namespace("invalid"),
			want:      false,
		},
		{
			name:      "empty namespace returns false",
			namespace: domain.Namespace(""),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.namespace.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNamespace_Priority(t *testing.T) {
	tests := []struct {
		name      string
		namespace domain.Namespace
		want      int
	}{
		{
			name:      "environment has priority 10",
			namespace: domain.NamespaceEnvironment,
			want:      10,
		},
		{
			name:      "logical has priority 50",
			namespace: domain.NamespaceLogical,
			want:      50,
		},
		{
			name:      "infrastructure has priority 100",
			namespace: domain.NamespaceInfrastructure,
			want:      100,
		},
		{
			name:      "invalid namespace defaults to 100",
			namespace: domain.Namespace("invalid"),
			want:      100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.namespace.Priority()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNamespace_Constants(t *testing.T) {
	// Verify the constant values are as expected
	assert.Equal(t, domain.Namespace("infrastructure"), domain.NamespaceInfrastructure)
	assert.Equal(t, domain.Namespace("logical"), domain.NamespaceLogical)
	assert.Equal(t, domain.Namespace("environment"), domain.NamespaceEnvironment)
}
