package interfaces

import (
	"context"
	"time"
)

// WebAuthnChallengeCache defines the transient storage contract for holding WebAuthn cryptographic challenges.
type WebAuthnChallengeCache interface {
	// StoreChallenge caches the challenge bytes associated with a ceremony or session-scoped email
	// to ensure replay-safety, expiring after the specified duration (TTL).
	StoreChallenge(ctx context.Context, email string, challenge []byte, ttl time.Duration) error

	// GetChallenge retrieves the cached challenge bytes for the given email.
	// Returns an error if the challenge does not exist or has expired.
	GetChallenge(ctx context.Context, email string) ([]byte, error)

	// DeleteChallenge removes the cached challenge for the given email to prevent reuse.
	DeleteChallenge(ctx context.Context, email string) error

	// VerifyAndConsumeChallenge atomically retrieves and deletes the cached challenge bytes for the given email.
	// Returns ErrChallengeNotFoundOrExpired if the challenge does not exist or has expired.
	VerifyAndConsumeChallenge(ctx context.Context, email string) ([]byte, error)
}
