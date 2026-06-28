package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockChallengeCache struct {
	mock.Mock
}

func (m *mockChallengeCache) StoreChallenge(ctx context.Context, email string, challenge []byte, ttl time.Duration) error {
	args := m.Called(ctx, email, challenge, ttl)
	return args.Error(0)
}

func (m *mockChallengeCache) GetChallenge(ctx context.Context, email string) ([]byte, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockChallengeCache) DeleteChallenge(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *mockChallengeCache) VerifyAndConsumeChallenge(ctx context.Context, email string) ([]byte, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func TestGenerateBiometricChallengeUseCase(t *testing.T) {
	email := "admin@example.com"

	t.Run("success single-object credential format", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewGenerateBiometricChallengeUseCase(ur, cache)

		user := &domain.AdminUser{
			ID:                  "user-uuid",
			Email:               email,
			WebauthnCredentials: []byte(`{"id":"cred_123", "public_key":"pubkey", "sign_count": 0}`),
		}

		ur.On("FindByEmail", mock.Anything, email).Return(user, nil).Once()
		cache.On("StoreChallenge", mock.Anything, email, mock.Anything, 5*time.Minute).Return(nil).Once()

		out, err := uc.Execute(context.Background(), GenerateBiometricChallengeInput{Email: email})

		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.NotEmpty(t, out.Challenge)
		assert.Equal(t, "cred_123", out.CredentialID)
		ur.AssertExpectations(t)
		cache.AssertExpectations(t)
	})

	t.Run("success array credential format", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewGenerateBiometricChallengeUseCase(ur, cache)

		user := &domain.AdminUser{
			ID:                  "user-uuid",
			Email:               email,
			WebauthnCredentials: []byte(`[{"id":"cred_array", "public_key":"pubkey", "sign_count": 5}]`),
		}

		ur.On("FindByEmail", mock.Anything, email).Return(user, nil).Once()
		cache.On("StoreChallenge", mock.Anything, email, mock.Anything, 5*time.Minute).Return(nil).Once()

		out, err := uc.Execute(context.Background(), GenerateBiometricChallengeInput{Email: email})

		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.NotEmpty(t, out.Challenge)
		assert.Equal(t, "cred_array", out.CredentialID)
		ur.AssertExpectations(t)
		cache.AssertExpectations(t)
	})

	t.Run("empty email error", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewGenerateBiometricChallengeUseCase(ur, cache)

		out, err := uc.Execute(context.Background(), GenerateBiometricChallengeInput{Email: ""})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email cannot be empty")
		assert.Nil(t, out)
	})

	t.Run("user not found returns biometric not registered", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewGenerateBiometricChallengeUseCase(ur, cache)

		ur.On("FindByEmail", mock.Anything, email).Return((*domain.AdminUser)(nil), domainErrors.ErrUserNotFound).Once()

		out, err := uc.Execute(context.Background(), GenerateBiometricChallengeInput{Email: email})

		assert.ErrorIs(t, err, domainErrors.ErrBiometricNotRegistered)
		assert.Nil(t, out)
		ur.AssertExpectations(t)
	})

	t.Run("user repository general error propagates", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewGenerateBiometricChallengeUseCase(ur, cache)

		expectedErr := errors.New("db crash")
		ur.On("FindByEmail", mock.Anything, email).Return((*domain.AdminUser)(nil), expectedErr).Once()

		out, err := uc.Execute(context.Background(), GenerateBiometricChallengeInput{Email: email})

		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, out)
		ur.AssertExpectations(t)
	})

	t.Run("empty webauthn credentials returns biometric not registered", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewGenerateBiometricChallengeUseCase(ur, cache)

		user := &domain.AdminUser{
			ID:                  "user-uuid",
			Email:               email,
			WebauthnCredentials: nil,
		}

		ur.On("FindByEmail", mock.Anything, email).Return(user, nil).Once()

		out, err := uc.Execute(context.Background(), GenerateBiometricChallengeInput{Email: email})

		assert.ErrorIs(t, err, domainErrors.ErrBiometricNotRegistered)
		assert.Nil(t, out)
		ur.AssertExpectations(t)
	})

	t.Run("malformed json credentials returns error", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewGenerateBiometricChallengeUseCase(ur, cache)

		user := &domain.AdminUser{
			ID:                  "user-uuid",
			Email:               email,
			WebauthnCredentials: []byte(`{"id":`),
		}

		ur.On("FindByEmail", mock.Anything, email).Return(user, nil).Once()

		out, err := uc.Execute(context.Background(), GenerateBiometricChallengeInput{Email: email})

		assert.Error(t, err)
		assert.Nil(t, out)
		ur.AssertExpectations(t)
	})

	t.Run("unsupported first json char returns biometric not registered", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewGenerateBiometricChallengeUseCase(ur, cache)

		user := &domain.AdminUser{
			ID:                  "user-uuid",
			Email:               email,
			WebauthnCredentials: []byte(`"invalid"`),
		}

		ur.On("FindByEmail", mock.Anything, email).Return(user, nil).Once()

		out, err := uc.Execute(context.Background(), GenerateBiometricChallengeInput{Email: email})

		assert.ErrorIs(t, err, domainErrors.ErrBiometricNotRegistered)
		assert.Nil(t, out)
		ur.AssertExpectations(t)
	})

	t.Run("empty parsed credentials returns biometric not registered", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewGenerateBiometricChallengeUseCase(ur, cache)

		user := &domain.AdminUser{
			ID:                  "user-uuid",
			Email:               email,
			WebauthnCredentials: []byte(`[]`),
		}

		ur.On("FindByEmail", mock.Anything, email).Return(user, nil).Once()

		out, err := uc.Execute(context.Background(), GenerateBiometricChallengeInput{Email: email})

		assert.ErrorIs(t, err, domainErrors.ErrBiometricNotRegistered)
		assert.Nil(t, out)
		ur.AssertExpectations(t)
	})

	t.Run("parsed credential with empty ID returns biometric not registered", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewGenerateBiometricChallengeUseCase(ur, cache)

		user := &domain.AdminUser{
			ID:                  "user-uuid",
			Email:               email,
			WebauthnCredentials: []byte(`{"id":""}`),
		}

		ur.On("FindByEmail", mock.Anything, email).Return(user, nil).Once()

		out, err := uc.Execute(context.Background(), GenerateBiometricChallengeInput{Email: email})

		assert.ErrorIs(t, err, domainErrors.ErrBiometricNotRegistered)
		assert.Nil(t, out)
		ur.AssertExpectations(t)
	})

	t.Run("cache storage error propagates", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewGenerateBiometricChallengeUseCase(ur, cache)

		user := &domain.AdminUser{
			ID:                  "user-uuid",
			Email:               email,
			WebauthnCredentials: []byte(`{"id":"cred_123"}`),
		}

		cacheErr := errors.New("redis error")
		ur.On("FindByEmail", mock.Anything, email).Return(user, nil).Once()
		cache.On("StoreChallenge", mock.Anything, email, mock.Anything, 5*time.Minute).Return(cacheErr).Once()

		out, err := uc.Execute(context.Background(), GenerateBiometricChallengeInput{Email: email})

		assert.ErrorIs(t, err, cacheErr)
		assert.Nil(t, out)
		ur.AssertExpectations(t)
		cache.AssertExpectations(t)
	})

	t.Run("success with leading and trailing whitespace in credentials", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewGenerateBiometricChallengeUseCase(ur, cache)

		user := &domain.AdminUser{
			ID:                  "user-uuid",
			Email:               email,
			WebauthnCredentials: []byte("  \n\t [{\"id\":\"cred_whitespace\", \"public_key\":\"pubkey\", \"sign_count\": 0}] \n  "),
		}

		ur.On("FindByEmail", mock.Anything, email).Return(user, nil).Once()
		cache.On("StoreChallenge", mock.Anything, email, mock.Anything, 5*time.Minute).Return(nil).Once()

		out, err := uc.Execute(context.Background(), GenerateBiometricChallengeInput{Email: email})

		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.NotEmpty(t, out.Challenge)
		assert.Equal(t, "cred_whitespace", out.CredentialID)
		ur.AssertExpectations(t)
		cache.AssertExpectations(t)
	})
}
