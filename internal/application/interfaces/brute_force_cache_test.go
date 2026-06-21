package interfaces

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeBruteForceCache struct {
	attempts map[string]int
	lockouts map[string]time.Time
}

func newFakeBruteForceCache() *fakeBruteForceCache {
	return &fakeBruteForceCache{
		attempts: make(map[string]int),
		lockouts: make(map[string]time.Time),
	}
}

func (f *fakeBruteForceCache) IncrementFailedAttempts(ctx context.Context, email string, _ time.Duration) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	f.attempts[email]++
	return f.attempts[email], nil
}

func (f *fakeBruteForceCache) GetFailedAttempts(ctx context.Context, email string) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return f.attempts[email], nil
}

func (f *fakeBruteForceCache) SetLockout(ctx context.Context, email string, duration time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.lockouts[email] = time.Now().Add(duration)
	return nil
}

func (f *fakeBruteForceCache) IsLocked(ctx context.Context, email string) (bool, time.Time, error) {
	if err := ctx.Err(); err != nil {
		return false, time.Time{}, err
	}
	expiry, exists := f.lockouts[email]
	if !exists {
		return false, time.Time{}, nil
	}
	if time.Now().After(expiry) {
		delete(f.lockouts, email)
		return false, time.Time{}, nil
	}
	return true, expiry, nil
}

func (f *fakeBruteForceCache) Reset(ctx context.Context, email string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	delete(f.attempts, email)
	delete(f.lockouts, email)
	return nil
}

func TestBruteForceCache_Workflow(t *testing.T) {
	cache := newFakeBruteForceCache()
	ctx := context.Background()
	email := "test@hros.io"

	// 1. Initial state check
	attempts, err := cache.GetFailedAttempts(ctx, email)
	assert.NoError(t, err)
	assert.Equal(t, 0, attempts)

	locked, _, err := cache.IsLocked(ctx, email)
	assert.NoError(t, err)
	assert.False(t, locked)

	// 2. Increment attempts
	attempts, err = cache.IncrementFailedAttempts(ctx, email, 15*time.Minute)
	assert.NoError(t, err)
	assert.Equal(t, 1, attempts)

	attempts, err = cache.GetFailedAttempts(ctx, email)
	assert.NoError(t, err)
	assert.Equal(t, 1, attempts)

	// 3. Lockout behavior
	err = cache.SetLockout(ctx, email, 30*time.Minute)
	assert.NoError(t, err)

	locked, expiry, err := cache.IsLocked(ctx, email)
	assert.NoError(t, err)
	assert.True(t, locked)
	assert.True(t, expiry.After(time.Now()))

	// 4. Reset behavior
	err = cache.Reset(ctx, email)
	assert.NoError(t, err)

	attempts, err = cache.GetFailedAttempts(ctx, email)
	assert.NoError(t, err)
	assert.Equal(t, 0, attempts)

	locked, _, err = cache.IsLocked(ctx, email)
	assert.NoError(t, err)
	assert.False(t, locked)
}

func TestBruteForceCache_ContextCancellation(t *testing.T) {
	cache := newFakeBruteForceCache()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	email := "test@hros.io"

	_, err := cache.IncrementFailedAttempts(ctx, email, 15*time.Minute)
	assert.ErrorIs(t, err, context.Canceled)

	_, err = cache.GetFailedAttempts(ctx, email)
	assert.ErrorIs(t, err, context.Canceled)

	err = cache.SetLockout(ctx, email, 30*time.Minute)
	assert.ErrorIs(t, err, context.Canceled)

	_, _, err = cache.IsLocked(ctx, email)
	assert.ErrorIs(t, err, context.Canceled)

	err = cache.Reset(ctx, email)
	assert.ErrorIs(t, err, context.Canceled)
}
