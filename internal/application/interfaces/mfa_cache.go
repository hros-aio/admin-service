package interfaces

import (
	"context"
	"time"
)

// MFACache defines the temporary caching contract for partially authenticated user contexts during MFA verification.
type MFACache interface {
	// StoreToken caches the admin ID associated with the MFA token for the specified TTL.
	StoreToken(ctx context.Context, mfaToken string, adminID string, ttl time.Duration) error

	// GetAdminID retrieves the cached admin ID associated with the MFA token.
	// It returns ErrMFATokenExpired (or a translation error) if the token is not found or has expired.
	GetAdminID(ctx context.Context, mfaToken string) (string, error)

	// DeleteToken invalidates/removes the cached admin ID for the given MFA token.
	DeleteToken(ctx context.Context, mfaToken string) error
}
