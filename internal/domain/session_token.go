package domain

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

// SessionToken represents an active or historical refresh token.
type SessionToken struct {
	ID           string
	AdminID      string
	RefreshToken string
	ExpiresAt    time.Time
	IsPersistent bool
	IPAddress    string
	UserAgent    string
	CreatedAt    time.Time
	RevokedAt    *time.Time
	RevokeReason string
}

// Rotate updates the refresh token value with a cryptographically secure random string
// and sets a new expiration time, returning the new token value.
func (t *SessionToken) Rotate(newExpiry time.Time) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Reader.Read(bytes); err != nil {
		return "", err
	}
	newToken := hex.EncodeToString(bytes)
	t.RefreshToken = newToken
	t.ExpiresAt = newExpiry
	return newToken, nil
}

// IsExpired checks if the session token has expired.
func (t *SessionToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsRevoked checks if the session token has been revoked.
func (t *SessionToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

// SessionTokenRepository defines the interface for persisting and retrieving session tokens.
type SessionTokenRepository interface {
	Save(ctx context.Context, token *SessionToken) error
	FindByToken(ctx context.Context, token string) (*SessionToken, error)
	DeleteByToken(ctx context.Context, token string) error
	DeleteByAdminID(ctx context.Context, adminID string) error
	Revoke(ctx context.Context, token string, reason string) error
}
