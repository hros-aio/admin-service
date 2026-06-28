package interfaces

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeWebAuthnChallengeCache struct {
	store map[string][]byte
	ttls  map[string]time.Duration
}

func newFakeWebAuthnChallengeCache() *fakeWebAuthnChallengeCache {
	return &fakeWebAuthnChallengeCache{
		store: make(map[string][]byte),
		ttls:  make(map[string]time.Duration),
	}
}

func (f *fakeWebAuthnChallengeCache) StoreChallenge(ctx context.Context, key string, challenge []byte, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.store[key] = challenge
	f.ttls[key] = ttl
	return nil
}

func (f *fakeWebAuthnChallengeCache) GetChallenge(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	challenge, exists := f.store[key]
	if !exists {
		return nil, errors.New("challenge not found or expired")
	}
	return challenge, nil
}

func (f *fakeWebAuthnChallengeCache) DeleteChallenge(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	delete(f.store, key)
	delete(f.ttls, key)
	return nil
}

func TestWebAuthnChallengeCache_Workflow(t *testing.T) {
	cache := newFakeWebAuthnChallengeCache()
	ctx := context.Background()
	key := "challenge_key_123"
	challenge := []byte("cryptographic_challenge_payload")
	ttl := 60 * time.Second

	// 1. Store
	err := cache.StoreChallenge(ctx, key, challenge, ttl)
	assert.NoError(t, err)

	// 2. Get
	retrieved, err := cache.GetChallenge(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, challenge, retrieved)

	// 3. Delete
	err = cache.DeleteChallenge(ctx, key)
	assert.NoError(t, err)

	// 4. Get after delete
	_, err = cache.GetChallenge(ctx, key)
	assert.Error(t, err)
}

func TestWebAuthnChallengeCache_ContextCancellation(t *testing.T) {
	cache := newFakeWebAuthnChallengeCache()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	key := "challenge_key_123"
	challenge := []byte("cryptographic_challenge_payload")
	ttl := 60 * time.Second

	err := cache.StoreChallenge(ctx, key, challenge, ttl)
	assert.ErrorIs(t, err, context.Canceled)

	_, err = cache.GetChallenge(ctx, key)
	assert.ErrorIs(t, err, context.Canceled)

	err = cache.DeleteChallenge(ctx, key)
	assert.ErrorIs(t, err, context.Canceled)
}
