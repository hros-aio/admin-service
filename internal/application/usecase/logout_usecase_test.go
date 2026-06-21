package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTokenBlacklist struct{ mock.Mock }

func (m *mockTokenBlacklist) Add(ctx context.Context, token string, ttl time.Duration) error {
	return m.Called(ctx, token, ttl).Error(0)
}

func (m *mockTokenBlacklist) Exists(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

const (
	testJTIToken      = "test-jti-123"
	validRefreshToken = "valid-refresh-token"
)

func TestLogoutUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	// Helper to generate a dummy JWT
	generateDummyJWT := func(jti string, exp int64) string {
		claims := jwt.MapClaims{}
		if jti != "" {
			claims["jti"] = jti
		}
		if exp > 0 {
			claims["exp"] = float64(exp)
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte("dummy-key"))
		return tokenString
	}

	t.Run("Success_NoAccessToken", func(t *testing.T) {
		sessionRepo := new(mockSessionRepo)
		blacklist := new(mockTokenBlacklist)
		audit := new(mockAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, blacklist, audit)

		token := validRefreshToken
		input := LogoutInput{RefreshToken: token}

		sessionRepo.On("DeleteByToken", ctx, token).Return(nil).Once()
		audit.On("LogLogoutSuccess", ctx, token).Return().Once()

		err := uc.Execute(ctx, input)

		assert.NoError(t, err)
		sessionRepo.AssertExpectations(t)
		blacklist.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("Success_WithAccessToken", func(t *testing.T) {
		sessionRepo := new(mockSessionRepo)
		blacklist := new(mockTokenBlacklist)
		audit := new(mockAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, blacklist, audit)

		refreshToken := validRefreshToken
		jti := testJTIToken
		exp := time.Now().Add(10 * time.Minute).Unix()
		accessToken := generateDummyJWT(jti, exp)

		input := LogoutInput{
			RefreshToken: refreshToken,
			AccessToken:  accessToken,
		}

		blacklist.On("Add", ctx, jti, mock.MatchedBy(func(ttl time.Duration) bool {
			return ttl > 9*time.Minute && ttl < 11*time.Minute
		})).Return(nil).Once()

		sessionRepo.On("DeleteByToken", ctx, refreshToken).Return(nil).Once()
		audit.On("LogLogoutSuccess", ctx, refreshToken).Return().Once()

		err := uc.Execute(ctx, input)

		assert.NoError(t, err)
		sessionRepo.AssertExpectations(t)
		blacklist.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("Success_ExpiredAccessToken", func(t *testing.T) {
		sessionRepo := new(mockSessionRepo)
		blacklist := new(mockTokenBlacklist)
		audit := new(mockAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, blacklist, audit)

		refreshToken := validRefreshToken
		jti := testJTIToken
		exp := time.Now().Add(-10 * time.Minute).Unix() // Expired 10 minutes ago
		accessToken := generateDummyJWT(jti, exp)

		input := LogoutInput{
			RefreshToken: refreshToken,
			AccessToken:  accessToken,
		}

		// blacklist.Add should NOT be called since ttl <= 0

		sessionRepo.On("DeleteByToken", ctx, refreshToken).Return(nil).Once()
		audit.On("LogLogoutSuccess", ctx, refreshToken).Return().Once()

		err := uc.Execute(ctx, input)

		assert.NoError(t, err)
		sessionRepo.AssertExpectations(t)
		blacklist.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("Success_BlacklistAddError_Proceeds", func(t *testing.T) {
		sessionRepo := new(mockSessionRepo)
		blacklist := new(mockTokenBlacklist)
		audit := new(mockAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, blacklist, audit)

		refreshToken := validRefreshToken
		jti := testJTIToken
		exp := time.Now().Add(10 * time.Minute).Unix()
		accessToken := generateDummyJWT(jti, exp)

		input := LogoutInput{
			RefreshToken: refreshToken,
			AccessToken:  accessToken,
		}

		blacklist.On("Add", ctx, jti, mock.Anything).Return(errors.New("redis connection refused")).Once()
		sessionRepo.On("DeleteByToken", ctx, refreshToken).Return(nil).Once()
		audit.On("LogLogoutSuccess", ctx, refreshToken).Return().Once()

		err := uc.Execute(ctx, input)

		// Usecase should log warning but not fail
		assert.NoError(t, err)
		sessionRepo.AssertExpectations(t)
		blacklist.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("Success_MalformedAccessToken_Proceeds", func(t *testing.T) {
		sessionRepo := new(mockSessionRepo)
		blacklist := new(mockTokenBlacklist)
		audit := new(mockAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, blacklist, audit)

		refreshToken := validRefreshToken
		input := LogoutInput{
			RefreshToken: refreshToken,
			AccessToken:  "malformed-token-string",
		}

		// blacklist.Add should NOT be called

		sessionRepo.On("DeleteByToken", ctx, refreshToken).Return(nil).Once()
		audit.On("LogLogoutSuccess", ctx, refreshToken).Return().Once()

		err := uc.Execute(ctx, input)

		assert.NoError(t, err)
		sessionRepo.AssertExpectations(t)
		blacklist.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("EmptyRefreshToken", func(t *testing.T) {
		sessionRepo := new(mockSessionRepo)
		blacklist := new(mockTokenBlacklist)
		audit := new(mockAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, blacklist, audit)

		input := LogoutInput{RefreshToken: ""}

		err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "refresh token cannot be empty")
	})

	t.Run("RepositoryError", func(t *testing.T) {
		sessionRepo := new(mockSessionRepo)
		blacklist := new(mockTokenBlacklist)
		audit := new(mockAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, blacklist, audit)

		token := "some-token"
		input := LogoutInput{RefreshToken: token}
		dbErr := errors.New("database connection lost")

		sessionRepo.On("DeleteByToken", ctx, token).Return(dbErr).Once()

		err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.ErrorIs(t, err, dbErr)
		sessionRepo.AssertExpectations(t)
		blacklist.AssertExpectations(t)
		audit.AssertNotCalled(t, "LogLogoutSuccess", mock.Anything, mock.Anything)
	})
}
