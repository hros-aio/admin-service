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

func (f *fakeSSOStateCache) VerifyAndConsumeState(ctx context.Context, state string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	nonce, exists := f.store[state]
	if !exists {
		return "", domainErrors.ErrInvalidSSOState
	}
	delete(f.store, state)
	return nonce, nil
}

func TestSSOStateCache_Workflow(t *testing.T) {
	cache := newFakeSSOStateCache()
	ctx := context.Background()
	state := "oauth_state_123"
	nonce := "oidc_nonce_456"

	// 1. Store State
	err := cache.StoreState(ctx, state, nonce, 5*time.Minute)
	assert.NoError(t, err)

	// 2. Verify and Consume State
	retrieved, err := cache.VerifyAndConsumeState(ctx, state)
	assert.NoError(t, err)
	assert.Equal(t, nonce, retrieved)

	// 3. Second consume should fail with ErrInvalidSSOState
	_, err = cache.VerifyAndConsumeState(ctx, state)
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

	_, err = cache.VerifyAndConsumeState(ctx, state)
	assert.ErrorIs(t, err, context.Canceled)
}
