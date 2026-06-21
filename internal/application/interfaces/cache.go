// Package interfaces defines application layer interfaces for external adapters and infrastructure.
package interfaces

import (
	"context"
	"time"
)

// TokenBlacklist defines the interface for blacklisting and checking revoked access and refresh tokens.
type TokenBlacklist interface {
	// Add blacklists a token with a specific time-to-live (TTL).
	Add(ctx context.Context, token string, ttl time.Duration) error
	// Exists checks if a token is currently blacklisted.
	Exists(ctx context.Context, token string) (bool, error)
}
