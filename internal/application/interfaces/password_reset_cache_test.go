package interfaces_test

import (
	"context"
	"testing"
	"time"

	"github.com/hros/admin-service/internal/application/interfaces"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/stretchr/testify/assert"
)

type fakePasswordResetCache struct {
	store map[string]string
	used  map[string]bool
}

func newFakePasswordResetCache() *fakePasswordResetCache {
	return &fakePasswordResetCache{
		store: make(map[string]string),
		used:  make(map[string]bool),
	}
}

func (f *fakePasswordResetCache) StoreToken(ctx context.Context, token string, adminID string, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.store[token] = adminID
	return nil
}

func (f *fakePasswordResetCache) ConsumeToken(ctx context.Context, token string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if f.used[token] {
		return "", domainErrors.ErrTokenUsed
	}
	adminID, exists := f.store[token]
	if !exists {
		return "", domainErrors.ErrTokenExpired
	}
	f.used[token] = true
	delete(f.store, token)
	return adminID, nil
}

func TestPasswordResetCache_Workflow(t *testing.T) {
	var cache interfaces.PasswordResetCache = newFakePasswordResetCache()
	ctx := context.Background()
	token := "reset_token_abc"
	adminID := "admin_123"

	// 1. Store
	err := cache.StoreToken(ctx, token, adminID, 60*time.Minute)
	assert.NoError(t, err)

	// 2. Consume
	retrieved, err := cache.ConsumeToken(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, adminID, retrieved)

	// 3. Consume again -> ErrTokenUsed
	_, err = cache.ConsumeToken(ctx, token)
	assert.ErrorIs(t, err, domainErrors.ErrTokenUsed)

	// 4. Consume nonexistent -> ErrTokenExpired
	_, err = cache.ConsumeToken(ctx, "nonexistent")
	assert.ErrorIs(t, err, domainErrors.ErrTokenExpired)
}

func TestPasswordResetCache_ContextCancellation(t *testing.T) {
	cache := newFakePasswordResetCache()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	token := "reset_token_abc"
	adminID := "admin_123"

	err := cache.StoreToken(ctx, token, adminID, 60*time.Minute)
	assert.ErrorIs(t, err, context.Canceled)

	_, err = cache.ConsumeToken(ctx, token)
	assert.ErrorIs(t, err, context.Canceled)
}
