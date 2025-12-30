package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/fregataa/aami/config-server/internal/domain"
	"github.com/fregataa/aami/config-server/test/testutil"
)

func TestTargetStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status domain.TargetStatus
		want   bool
	}{
		{
			name:   "active status is valid",
			status: domain.TargetStatusActive,
			want:   true,
		},
		{
			name:   "inactive status is valid",
			status: domain.TargetStatusInactive,
			want:   true,
		},
		{
			name:   "down status is valid",
			status: domain.TargetStatusDown,
			want:   true,
		},
		{
			name:   "invalid status returns false",
			status: domain.TargetStatus("invalid"),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTarget_IsHealthy(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		status   domain.TargetStatus
		lastSeen *time.Time
		want     bool
	}{
		{
			name:     "active target with recent last seen is healthy",
			status:   domain.TargetStatusActive,
			lastSeen: &now,
			want:     true,
		},
		{
			name:     "active target with old last seen is unhealthy",
			status:   domain.TargetStatusActive,
			lastSeen: testutil.TimePtr(now.Add(-10 * time.Minute)),
			want:     false,
		},
		{
			name:     "inactive target is unhealthy",
			status:   domain.TargetStatusInactive,
			lastSeen: &now,
			want:     false,
		},
		{
			name:     "down target is unhealthy",
			status:   domain.TargetStatusDown,
			lastSeen: &now,
			want:     false,
		},
		{
			name:     "target with nil last seen is unhealthy",
			status:   domain.TargetStatusActive,
			lastSeen: nil,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := &domain.Target{
				Status:   tt.status,
				LastSeen: tt.lastSeen,
			}
			got := target.IsHealthy()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTarget_Creation(t *testing.T) {
	group := testutil.NewTestGroup("production")
	target := testutil.NewTestTarget("server-01", "192.168.1.1", []domain.Group{*group})

	assert.NotEmpty(t, target.ID)
	assert.Equal(t, "server-01", target.Hostname)
	assert.Equal(t, "192.168.1.1", target.IPAddress)
	assert.Len(t, target.Groups, 1)
	assert.Equal(t, group.ID, target.Groups[0].ID)
	assert.Equal(t, domain.TargetStatusActive, target.Status)
	assert.NotNil(t, target.Labels)
	assert.NotNil(t, target.Metadata)
	assert.NotZero(t, target.CreatedAt)
	assert.NotZero(t, target.UpdatedAt)
}

func TestTarget_StatusConstants(t *testing.T) {
	// Verify the constant values are as expected
	assert.Equal(t, domain.TargetStatus("active"), domain.TargetStatusActive)
	assert.Equal(t, domain.TargetStatus("inactive"), domain.TargetStatusInactive)
	assert.Equal(t, domain.TargetStatus("down"), domain.TargetStatusDown)
}

func TestTarget_ValidStatuses(t *testing.T) {
	statuses := []domain.TargetStatus{
		domain.TargetStatusActive,
		domain.TargetStatusInactive,
		domain.TargetStatusDown,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			assert.True(t, status.IsValid())
		})
	}
}
