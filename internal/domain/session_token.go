package domain

import (
	"context"
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
	DeleteByAdminID(ctx context.Context, adminID string) error
	Revoke(ctx context.Context, token string, reason string) error
}
