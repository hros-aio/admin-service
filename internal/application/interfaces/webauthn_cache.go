package interfaces

import (
	"context"
	"time"
)

// WebAuthnChallengeCache defines the transient storage contract for holding WebAuthn cryptographic challenges.
type WebAuthnChallengeCache interface {
	// StoreChallenge caches the challenge bytes associated with a ceremony or session-scoped key
	// to ensure replay-safety, expiring after the specified duration (TTL).
	StoreChallenge(ctx context.Context, key string, challenge []byte, ttl time.Duration) error

	// GetChallenge retrieves the cached challenge bytes for the given key.
	// Returns an error if the challenge does not exist or has expired.
	GetChallenge(ctx context.Context, key string) ([]byte, error)

	// DeleteChallenge removes the cached challenge for the given key to prevent reuse.
	DeleteChallenge(ctx context.Context, key string) error

	// VerifyAndConsumeChallenge atomically retrieves and deletes the cached challenge bytes for the given key.
	// Returns ErrChallengeNotFoundOrExpired if the challenge does not exist or has expired.
	VerifyAndConsumeChallenge(ctx context.Context, key string) ([]byte, error)
}
