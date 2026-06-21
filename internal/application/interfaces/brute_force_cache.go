package interfaces

import (
	"context"
	"time"
)

// BruteForceCache defines the storage contract for tracking brute-force login attempts and lockout state.
type BruteForceCache interface {
	// IncrementFailedAttempts increments the failed login attempts counter for an email within a sliding window.
	// It returns the updated count of failed attempts.
	IncrementFailedAttempts(ctx context.Context, email string, window time.Duration) (int, error)

	// GetFailedAttempts returns the current failed login attempts counter for an email.
	GetFailedAttempts(ctx context.Context, email string) (int, error)

	// SetLockout sets a temporary lockout for an email for the specified duration.
	SetLockout(ctx context.Context, email string, duration time.Duration) error

	// IsLocked checks if the email is currently locked out.
	// It returns true and the lockout expiration time if locked, or false and a zero time if not locked.
	IsLocked(ctx context.Context, email string) (bool, time.Time, error)

	// Reset clears both the failed attempts counter and the lockout state for an email.
	Reset(ctx context.Context, email string) error
}
