package interfaces

import (
	"context"
	"testing"

	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/stretchr/testify/assert"
)

type fakeMFACache struct {
	store map[string]string
}

func newFakeMFACache() *fakeMFACache {
	return &fakeMFACache{
		store: make(map[string]string),
	}
}

func (f *fakeMFACache) StoreToken(ctx context.Context, mfaToken string, adminID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.store[mfaToken] = adminID
	return nil
}

func (f *fakeMFACache) GetAdminID(ctx context.Context, mfaToken string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	adminID, exists := f.store[mfaToken]
	if !exists {
		return "", domainErrors.ErrMFATokenExpired
	}
	return adminID, nil
}

func (f *fakeMFACache) DeleteToken(ctx context.Context, mfaToken string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	delete(f.store, mfaToken)
	return nil
}

func TestMFACache_Workflow(t *testing.T) {
	cache := newFakeMFACache()
	ctx := context.Background()
	token := "mfa_token_abc"
	adminID := "admin_123"

	// 1. Store
	err := cache.StoreToken(ctx, token, adminID)
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
	assert.ErrorIs(t, err, domainErrors.ErrMFATokenExpired)
}

func TestMFACache_ContextCancellation(t *testing.T) {
	cache := newFakeMFACache()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	token := "mfa_token_abc"
	adminID := "admin_123"

	err := cache.StoreToken(ctx, token, adminID)
	assert.ErrorIs(t, err, context.Canceled)

	_, err = cache.GetAdminID(ctx, token)
	assert.ErrorIs(t, err, context.Canceled)

	err = cache.DeleteToken(ctx, token)
	assert.ErrorIs(t, err, context.Canceled)
}
