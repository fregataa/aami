package domain

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

// BootstrapToken represents a token used for auto-registration of new nodes
type BootstrapToken struct {
	ID        string            `json:"id"`
	Token     string            `json:"token"`
	Name      string            `json:"name"`
	MaxUses   int               `json:"max_uses"`
	Uses      int               `json:"uses"`
	ExpiresAt time.Time         `json:"expires_at"`
	Labels    map[string]string `json:"labels"`
	DeletedAt *time.Time        `json:"deleted_at,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// IsValid checks if the token is valid (not expired and has remaining uses)
func (bt *BootstrapToken) IsValid() bool {
	return bt.CanUse()
}

// CanUse checks if the token can still be used
func (bt *BootstrapToken) CanUse() bool {
	if bt.IsExpired() {
		return false
	}
	if bt.Uses >= bt.MaxUses {
		return false
	}
	return true
}

// IsExpired checks if the token has expired
func (bt *BootstrapToken) IsExpired() bool {
	return time.Now().After(bt.ExpiresAt)
}

// IncrementUses increments the usage counter
func (bt *BootstrapToken) IncrementUses() error {
	if !bt.CanUse() {
		return fmt.Errorf("token cannot be used: expired or max uses reached")
	}
	bt.Uses++
	bt.UpdatedAt = time.Now()
	return nil
}

// RemainingUses returns the number of remaining uses
func (bt *BootstrapToken) RemainingUses() int {
	remaining := bt.MaxUses - bt.Uses
	if remaining < 0 {
		return 0
	}
	return remaining
}

// TimeUntilExpiry returns the duration until the token expires
func (bt *BootstrapToken) TimeUntilExpiry() time.Duration {
	return time.Until(bt.ExpiresAt)
}

// GenerateToken generates a new random token string
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// NewBootstrapToken creates a new bootstrap token with generated token string
func NewBootstrapToken(name string, maxUses int, expiresAt time.Time) (*BootstrapToken, error) {
	token, err := GenerateToken()
	if err != nil {
		return nil, err
	}

	return &BootstrapToken{
		Token:     token,
		Name:      name,
		MaxUses:   maxUses,
		Uses:      0,
		ExpiresAt: expiresAt,
		Labels:    make(map[string]string),
	}, nil
}
