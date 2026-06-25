package interfaces

import (
	"context"
	"time"
)

// PasswordResetCache defines the contract for caching single-use password reset tokens.
type PasswordResetCache interface {
	// StoreToken associates a reset token with an admin's email for a specific TTL.
	StoreToken(ctx context.Context, token string, email string, ttl time.Duration) error

	// GetEmail retrieves the cached email associated with the reset token.
	// It returns ErrTokenExpired if the token is not found or has expired.
	GetEmail(ctx context.Context, token string) (string, error)

	// DeleteToken invalidates/removes the cached reset token.
	DeleteToken(ctx context.Context, token string) error
}
