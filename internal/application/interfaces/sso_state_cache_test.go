package interfaces

import (
	"context"
	"testing"
	"time"

	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/stretchr/testify/assert"
)

type fakeSSOStateCache struct {
	store map[string]string
}

func newFakeSSOStateCache() *fakeSSOStateCache {
	return &fakeSSOStateCache{
		store: make(map[string]string),
	}
}

func (f *fakeSSOStateCache) StoreState(ctx context.Context, state string, nonce string, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.store[state] = nonce
	return nil
}

func (f *fakeSSOStateCache) GetState(ctx context.Context, state string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	nonce, exists := f.store[state]
	if !exists {
		return "", domainErrors.ErrInvalidSSOState
	}
	return nonce, nil
}

func (f *fakeSSOStateCache) DeleteState(ctx context.Context, state string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	delete(f.store, state)
	return nil
}

func TestSSOStateCache_Workflow(t *testing.T) {
	cache := newFakeSSOStateCache()
	ctx := context.Background()
	state := "oauth_state_123"
	nonce := "oidc_nonce_456"

	// 1. Store State
	err := cache.StoreState(ctx, state, nonce, 5*time.Minute)
	assert.NoError(t, err)

	// 2. Get State
	retrieved, err := cache.GetState(ctx, state)
	assert.NoError(t, err)
	assert.Equal(t, nonce, retrieved)

	// 3. Delete State
	err = cache.DeleteState(ctx, state)
	assert.NoError(t, err)

	// 4. Get after delete should return ErrInvalidSSOState
	_, err = cache.GetState(ctx, state)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domainErrors.ErrInvalidSSOState)
}

func TestSSOStateCache_ContextCancellation(t *testing.T) {
	cache := newFakeSSOStateCache()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	state := "oauth_state_123"
	nonce := "oidc_nonce_456"

	err := cache.StoreState(ctx, state, nonce, 5*time.Minute)
	assert.ErrorIs(t, err, context.Canceled)

	_, err = cache.GetState(ctx, state)
	assert.ErrorIs(t, err, context.Canceled)

	err = cache.DeleteState(ctx, state)
	assert.ErrorIs(t, err, context.Canceled)
}
