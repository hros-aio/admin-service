package interfaces

import (
	"context"
	"time"
)

// PasswordResetCache defines the contract for caching single-use password reset tokens.
type PasswordResetCache interface {
	// StoreToken associates a reset token with an admin's ID for a specific TTL.
	StoreToken(ctx context.Context, token string, adminID string, ttl time.Duration) error

	// GetAdminID retrieves the cached admin ID associated with the reset token.
	// It returns ErrTokenExpired if the token is not found or has expired.
	GetAdminID(ctx context.Context, token string) (string, error)

	// DeleteToken invalidates/removes the cached reset token.
	DeleteToken(ctx context.Context, token string) error
}
