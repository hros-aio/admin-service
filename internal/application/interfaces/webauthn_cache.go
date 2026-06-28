package interfaces

import (
	"context"
)

// WebAuthnChallengeCache defines the transient storage contract for holding WebAuthn cryptographic challenges.
type WebAuthnChallengeCache interface {
	// StoreChallenge caches the challenge bytes associated with a given key (such as session token or user identifier).
	StoreChallenge(ctx context.Context, key string, challenge []byte) error

	// GetChallenge retrieves the cached challenge bytes for the given key.
	// Returns an error if the challenge does not exist or has expired.
	GetChallenge(ctx context.Context, key string) ([]byte, error)

	// DeleteChallenge removes the cached challenge for the given key to prevent reuse.
	DeleteChallenge(ctx context.Context, key string) error
}
