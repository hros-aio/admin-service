package interfaces

import (
	"context"
	"time"
)

// PasswordResetCache defines the contract for caching single-use password reset tokens.
type PasswordResetCache interface {
	// StoreToken associates a reset token with an admin's ID for a specific TTL.
	StoreToken(ctx context.Context, token string, adminID string, ttl time.Duration) error

	// ConsumeToken atomically retrieves the cached admin ID associated with the reset token and marks it as used.
	// It returns ErrTokenExpired if the token is not found or has expired.
	// It returns ErrTokenUsed if the token has already been consumed.
	ConsumeToken(ctx context.Context, token string) (string, error)

	// DeleteToken invalidates/removes the cached reset token without marking it as used (used for rollback/cleanup).
	DeleteToken(ctx context.Context, token string) error
}
