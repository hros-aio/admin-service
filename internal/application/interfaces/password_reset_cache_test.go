package interfaces

import (
	"context"
	"testing"
	"time"

	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/stretchr/testify/assert"
)

type fakePasswordResetCache struct {
	store map[string]string
}

func newFakePasswordResetCache() *fakePasswordResetCache {
	return &fakePasswordResetCache{
		store: make(map[string]string),
	}
}

func (f *fakePasswordResetCache) StoreToken(ctx context.Context, token string, adminID string, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.store[token] = adminID
	return nil
}

func (f *fakePasswordResetCache) GetAdminID(ctx context.Context, token string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	adminID, exists := f.store[token]
	if !exists {
		return "", domainErrors.ErrTokenExpired
	}
	return adminID, nil
}

func (f *fakePasswordResetCache) DeleteToken(ctx context.Context, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	delete(f.store, token)
	return nil
}

func TestPasswordResetCache_Workflow(t *testing.T) {
	cache := newFakePasswordResetCache()
	ctx := context.Background()
	token := "reset_token_abc"
	adminID := "admin_123"

	// 1. Store
	err := cache.StoreToken(ctx, token, adminID, 60*time.Minute)
	assert.NoError(t, err)

	// 2. Get
	retrieved, err := cache.GetAdminID(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, adminID, retrieved)

	// 3. Delete
	err = cache.DeleteToken(ctx, token)
	assert.NoError(t, err)

	// 4. Get after delete
	_, err = cache.GetAdminID(ctx, token)
	assert.Error(t, err)
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

	_, err = cache.GetAdminID(ctx, token)
	assert.ErrorIs(t, err, context.Canceled)

	err = cache.DeleteToken(ctx, token)
	assert.ErrorIs(t, err, context.Canceled)
}
