// Package interfaces defines application layer interfaces for external adapters and infrastructure.
package interfaces

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// fakeTokenBlacklist implements TokenBlacklist interface in-memory for testing.
type fakeTokenBlacklist struct {
	tokens map[string]time.Time
}

func newFakeTokenBlacklist() *fakeTokenBlacklist {
	return &fakeTokenBlacklist{
		tokens: make(map[string]time.Time),
	}
}

func (f *fakeTokenBlacklist) Add(ctx context.Context, token string, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.tokens[token] = time.Now().Add(ttl)
	return nil
}

func (f *fakeTokenBlacklist) Exists(ctx context.Context, token string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	expiry, exists := f.tokens[token]
	if !exists {
		return false, nil
	}
	if time.Now().After(expiry) {
		delete(f.tokens, token)
		return false, nil
	}
	return true, nil
}

func TestTokenBlacklist_Success(t *testing.T) {
	blacklist := newFakeTokenBlacklist()
	ctx := context.Background()

	token := "test-token"
	ttl := 10 * time.Minute

	// Verify not blacklisted initially
	exists, err := blacklist.Exists(ctx, token)
	assert.NoError(t, err)
	assert.False(t, exists)

	// Add to blacklist
	err = blacklist.Add(ctx, token, ttl)
	assert.NoError(t, err)

	// Verify blacklisted
	exists, err = blacklist.Exists(ctx, token)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestTokenBlacklist_Expiration(t *testing.T) {
	blacklist := newFakeTokenBlacklist()
	ctx := context.Background()

	token := "expired-token"
	ttl := -1 * time.Second // already expired

	err := blacklist.Add(ctx, token, ttl)
	assert.NoError(t, err)

	// Verify it does not exist (expired)
	exists, err := blacklist.Exists(ctx, token)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestTokenBlacklist_ContextCancellation(t *testing.T) {
	blacklist := newFakeTokenBlacklist()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel context immediately

	token := "token-cancelled"
	ttl := 5 * time.Minute

	// Add should return context canceled error
	err := blacklist.Add(ctx, token, ttl)
	assert.ErrorIs(t, err, context.Canceled)

	// Exists should return context canceled error
	_, err = blacklist.Exists(ctx, token)
	assert.ErrorIs(t, err, context.Canceled)
}
