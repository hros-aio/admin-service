package interfaces

import (
	"context"
	"time"
)

// SSOStateCache defines the temporary caching contract for OAuth/OIDC state and nonce parameters to prevent CSRF.
type SSOStateCache interface {
	// StoreState caches the nonce/value associated with the OAuth/OIDC state for a specific TTL.
	StoreState(ctx context.Context, state string, nonce string, ttl time.Duration) error

	// GetState retrieves the cached nonce/value associated with the OAuth/OIDC state.
	// It returns ErrInvalidSSOState if the state is not found or has expired.
	GetState(ctx context.Context, state string) (string, error)

	// DeleteState invalidates/removes the cached state.
	DeleteState(ctx context.Context, state string) error
}
