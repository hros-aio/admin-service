package interfaces

import (
	"context"
	"time"

	"github.com/hros/admin-service/internal/domain"
)

// MFACache defines the temporary caching contract for partially authenticated user contexts during MFA verification.
type MFACache interface {
	// Store caches the partially authenticated user context associated with the MFA token for the specified TTL.
	Store(ctx context.Context, token string, user *domain.AdminUser, ttl time.Duration) error

	// Get retrieves the cached user context associated with the MFA token.
	// It returns ErrMFATokenExpired if the token is not found or has expired.
	Get(ctx context.Context, token string) (*domain.AdminUser, error)

	// Delete invalidates/removes the cached user context for the given MFA token.
	Delete(ctx context.Context, token string) error
}
