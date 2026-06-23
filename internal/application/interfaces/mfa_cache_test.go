package interfaces

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hros/admin-service/internal/domain"
	"github.com/stretchr/testify/assert"
)

type fakeMFACache struct {
	store map[string]*domain.AdminUser
}

func newFakeMFACache() *fakeMFACache {
	return &fakeMFACache{
		store: make(map[string]*domain.AdminUser),
	}
}

func (f *fakeMFACache) Store(ctx context.Context, token string, user *domain.AdminUser, _ time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.store[token] = user
	return nil
}

func (f *fakeMFACache) Get(ctx context.Context, token string) (*domain.AdminUser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	user, exists := f.store[token]
	if !exists {
		return nil, errors.New("MFA token has expired")
	}
	return user, nil
}

func (f *fakeMFACache) Delete(ctx context.Context, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	delete(f.store, token)
	return nil
}

func TestMFACache_Workflow(t *testing.T) {
	cache := newFakeMFACache()
	ctx := context.Background()
	token := "mfa_token_abc"
	user := &domain.AdminUser{ID: "admin_123", Email: "admin@hros.io"}

	// 1. Store
	err := cache.Store(ctx, token, user, 5*time.Minute)
	assert.NoError(t, err)

	// 2. Get
	retrieved, err := cache.Get(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Email, retrieved.Email)

	// 3. Delete
	err = cache.Delete(ctx, token)
	assert.NoError(t, err)

	// 4. Get after delete
	_, err = cache.Get(ctx, token)
	assert.Error(t, err)
	assert.Equal(t, "MFA token has expired", err.Error())
}

func TestMFACache_ContextCancellation(t *testing.T) {
	cache := newFakeMFACache()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	token := "mfa_token_abc"
	user := &domain.AdminUser{ID: "admin_123"}

	err := cache.Store(ctx, token, user, 5*time.Minute)
	assert.ErrorIs(t, err, context.Canceled)

	_, err = cache.Get(ctx, token)
	assert.ErrorIs(t, err, context.Canceled)

	err = cache.Delete(ctx, token)
	assert.ErrorIs(t, err, context.Canceled)
}
