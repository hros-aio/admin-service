package interfaces

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeWebAuthnChallengeCache struct {
	store map[string][]byte
}

func newFakeWebAuthnChallengeCache() *fakeWebAuthnChallengeCache {
	return &fakeWebAuthnChallengeCache{
		store: make(map[string][]byte),
	}
}

func (f *fakeWebAuthnChallengeCache) StoreChallenge(ctx context.Context, key string, challenge []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.store[key] = challenge
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
	return nil
}

func TestWebAuthnChallengeCache_Workflow(t *testing.T) {
	cache := newFakeWebAuthnChallengeCache()
	ctx := context.Background()
	key := "challenge_key_123"
	challenge := []byte("cryptographic_challenge_payload")

	// 1. Store
	err := cache.StoreChallenge(ctx, key, challenge)
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

	err := cache.StoreChallenge(ctx, key, challenge)
	assert.ErrorIs(t, err, context.Canceled)

	_, err = cache.GetChallenge(ctx, key)
	assert.ErrorIs(t, err, context.Canceled)

	err = cache.DeleteChallenge(ctx, key)
	assert.ErrorIs(t, err, context.Canceled)
}
