package interfaces

import (
	"context"
	"testing"
	"time"

	domainErrors "github.com/hros/admin-service/internal/domain/errors"
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

func (f *fakeWebAuthnChallengeCache) StoreChallenge(ctx context.Context, email string, challenge []byte, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.store[email] = challenge
	f.ttls[email] = ttl
	return nil
}

func (f *fakeWebAuthnChallengeCache) GetChallenge(ctx context.Context, email string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	challenge, exists := f.store[email]
	if !exists {
		return nil, domainErrors.ErrChallengeNotFoundOrExpired
	}
	return challenge, nil
}

func (f *fakeWebAuthnChallengeCache) DeleteChallenge(ctx context.Context, email string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	delete(f.store, email)
	delete(f.ttls, email)
	return nil
}

func (f *fakeWebAuthnChallengeCache) VerifyAndConsumeChallenge(ctx context.Context, email string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	challenge, exists := f.store[email]
	if !exists {
		return nil, domainErrors.ErrChallengeNotFoundOrExpired
	}
	delete(f.store, email)
	delete(f.ttls, email)
	return challenge, nil
}

func TestWebAuthnChallengeCache_Workflow(t *testing.T) {
	cache := newFakeWebAuthnChallengeCache()
	ctx := context.Background()
	email := "admin@hros.com"
	challenge := []byte("cryptographic_challenge_payload")
	ttl := 60 * time.Second

	// 1. Store
	err := cache.StoreChallenge(ctx, email, challenge, ttl)
	assert.NoError(t, err)

	// 2. Get
	retrieved, err := cache.GetChallenge(ctx, email)
	assert.NoError(t, err)
	assert.Equal(t, challenge, retrieved)

	// 3. Verify and Consume
	consumed, err := cache.VerifyAndConsumeChallenge(ctx, email)
	assert.NoError(t, err)
	assert.Equal(t, challenge, consumed)

	// 4. Verify and Consume again (fails because it is deleted)
	_, err = cache.VerifyAndConsumeChallenge(ctx, email)
	assert.ErrorIs(t, err, domainErrors.ErrChallengeNotFoundOrExpired)

	// 5. Get after delete
	_, err = cache.GetChallenge(ctx, email)
	assert.ErrorIs(t, err, domainErrors.ErrChallengeNotFoundOrExpired)
}

func TestWebAuthnChallengeCache_ContextCancellation(t *testing.T) {
	cache := newFakeWebAuthnChallengeCache()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	email := "admin@hros.com"
	challenge := []byte("cryptographic_challenge_payload")
	ttl := 60 * time.Second

	err := cache.StoreChallenge(ctx, email, challenge, ttl)
	assert.ErrorIs(t, err, context.Canceled)

	_, err = cache.GetChallenge(ctx, email)
	assert.ErrorIs(t, err, context.Canceled)

	err = cache.DeleteChallenge(ctx, email)
	assert.ErrorIs(t, err, context.Canceled)

	_, err = cache.VerifyAndConsumeChallenge(ctx, email)
	assert.ErrorIs(t, err, context.Canceled)
}
