// Package http provides HTTP adapter handlers for Echo routing.
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hros/admin-service/internal/adapter/http/auth/dto"
	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/hros/admin-service/internal/application/usecase"
	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/hros/admin-service/internal/domain/events"
	sharedErrors "github.com/hros/admin-service/internal/shared/errors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Define local mock structures for unit testing use case dependencies

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Save(ctx context.Context, u *domain.AdminUser) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockUserRepo) Update(ctx context.Context, u *domain.AdminUser) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*domain.AdminUser, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AdminUser), args.Error(1)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.AdminUser, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AdminUser), args.Error(1)
}

func (m *mockUserRepo) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

type mockSessionRepo struct{ mock.Mock }

func (m *mockSessionRepo) Save(ctx context.Context, t *domain.SessionToken) error {
	return m.Called(ctx, t).Error(0)
}

func (m *mockSessionRepo) FindByToken(ctx context.Context, t string) (*domain.SessionToken, error) {
	args := m.Called(ctx, t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SessionToken), args.Error(1)
}

func (m *mockSessionRepo) DeleteByToken(ctx context.Context, t string) error {
	return m.Called(ctx, t).Error(0)
}

func (m *mockSessionRepo) DeleteByAdminID(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockSessionRepo) Revoke(ctx context.Context, t string, r string) error {
	return m.Called(ctx, t, r).Error(0)
}

func (m *mockSessionRepo) UpdateToken(ctx context.Context, session *domain.SessionToken) error {
	return m.Called(ctx, session).Error(0)
}

type mockPasswordHelper struct{ mock.Mock }

func (m *mockPasswordHelper) Hash(_ string) (string, error) {
	return "", nil
}

func (m *mockPasswordHelper) Compare(h, p string) error {
	return m.Called(h, p).Error(0)
}

func (m *mockPasswordHelper) CompareDummy(p string) {
	m.Called(p)
}

type mockTokenProvider struct{ mock.Mock }

func (m *mockTokenProvider) GenerateAccessToken(ctx context.Context, u *domain.AdminUser, e time.Duration) (string, error) {
	args := m.Called(ctx, u, e)
	return args.String(0), args.Error(1)
}

func (m *mockTokenProvider) GenerateRefreshToken(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

type mockAuditLogger struct{ mock.Mock }

func (m *mockAuditLogger) LogLoginSuccess(ctx context.Context, id, email string) {
	m.Called(ctx, id, email)
}

func (m *mockAuditLogger) LogLoginFailed(ctx context.Context, email, reason string) {
	m.Called(ctx, email, reason)
}

func (m *mockAuditLogger) LogLogoutSuccess(ctx context.Context, token string) {
	m.Called(ctx, token)
}

func (m *mockAuditLogger) LogSessionRefreshed(ctx context.Context, userID string) {
	m.Called(ctx, userID)
}

func (m *mockAuditLogger) LogAccountLocked(ctx context.Context, email string) {
	m.Called(ctx, email)
}

// nopBruteForceCache is a no-op implementation of interfaces.BruteForceCache
// used in handler tests where brute-force state is not under test.
type nopBruteForceCache struct{}

func (n *nopBruteForceCache) IncrementFailedAttempts(_ context.Context, _ string, _ time.Duration) (int, error) {
	return 0, nil
}
func (n *nopBruteForceCache) GetFailedAttempts(_ context.Context, _ string) (int, error) {
	return 0, nil
}
func (n *nopBruteForceCache) SetLockout(_ context.Context, _ string, _ time.Duration) error {
	return nil
}
func (n *nopBruteForceCache) IsLocked(_ context.Context, _ string) (bool, time.Time, error) {
	return false, time.Time{}, nil
}
func (n *nopBruteForceCache) Reset(_ context.Context, _ string) error { return nil }

// nopLockoutNotifier is a no-op implementation of interfaces.LockoutNotifier.
type nopLockoutNotifier struct{}

func (n *nopLockoutNotifier) PublishLockoutEmail(_ context.Context, _ events.EmailSendEvent) error {
	return nil
}

// newLoginUCForTest constructs a LoginUseCase with no-op brute-force dependencies
// for handler-level tests that only exercise routing and error mapping.
func newLoginUCForTest(
	userRepo domain.AdminUserRepository,
	sessionRepo domain.SessionTokenRepository,
	pwHelper interfaces.BruteForceCache, // unused overload — kept for signature compat
	_ interface{}, // unused
	password *mockPasswordHelper,
	tokens *mockTokenProvider,
	audit *mockAuditLogger,
) *usecase.LoginUseCase {
	return usecase.NewLoginUseCase(
		userRepo,
		sessionRepo,
		password,
		tokens,
		audit,
		&nopBruteForceCache{},
		&nopLockoutNotifier{},
		slog.Default(),
	)
}

type mockTokenBlacklist struct{ mock.Mock }

func (m *mockTokenBlacklist) Add(ctx context.Context, token string, ttl time.Duration) error {
	return m.Called(ctx, token, ttl).Error(0)
}

func (m *mockTokenBlacklist) Exists(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

func TestAuthHandler_Login(t *testing.T) {
	e := echo.New()

	t.Run("Successful Login", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockPasswordHelper := new(mockPasswordHelper)
		mockTokenProvider := new(mockTokenProvider)
		mockAuditLogger := new(mockAuditLogger)

		loginUC := usecase.NewLoginUseCase(mockUserRepo, mockSessionRepo, mockPasswordHelper, mockTokenProvider, mockAuditLogger, &nopBruteForceCache{}, &nopLockoutNotifier{}, slog.Default())
		mockTokenBlacklist := new(mockTokenBlacklist)
		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockTokenBlacklist, mockAuditLogger)
		handler := NewAuthHandler(loginUC, logoutUC, nil)

		reqBody := dto.LoginRequest{
			Email:      "admin@hros.com",
			Password:   "password123",
			RememberMe: true,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		user := &domain.AdminUser{
			ID:           "admin-id-123",
			Email:        "admin@hros.com",
			PasswordHash: "hashed-pwd",
			Status:       "active",
		}

		mockUserRepo.On("FindByEmail", mock.Anything, reqBody.Email).Return(user, nil)
		mockPasswordHelper.On("Compare", user.PasswordHash, reqBody.Password).Return(nil)
		mockTokenProvider.On("GenerateAccessToken", mock.Anything, user, 15*time.Minute).Return("access-token-xyz", nil)
		mockTokenProvider.On("GenerateRefreshToken", mock.Anything).Return("refresh-token-abc", nil)
		mockSessionRepo.On("Save", mock.Anything, mock.MatchedBy(func(s *domain.SessionToken) bool {
			return s.IsPersistent == true
		})).Return(nil)
		mockAuditLogger.On("LogLoginSuccess", mock.Anything, user.ID, user.Email).Return()

		err := handler.Login(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp dto.LoginResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "access-token-xyz", resp.AccessToken)
		assert.Equal(t, "refresh-token-abc", resp.RefreshToken)
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockPasswordHelper := new(mockPasswordHelper)
		mockTokenProvider := new(mockTokenProvider)
		mockAuditLogger := new(mockAuditLogger)

		loginUC := usecase.NewLoginUseCase(mockUserRepo, mockSessionRepo, mockPasswordHelper, mockTokenProvider, mockAuditLogger, &nopBruteForceCache{}, &nopLockoutNotifier{}, slog.Default())
		mockTokenBlacklist := new(mockTokenBlacklist)
		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockTokenBlacklist, mockAuditLogger)
		handler := NewAuthHandler(loginUC, logoutUC, nil)

		reqBody := dto.LoginRequest{
			Email:    "admin@hros.com",
			Password: "wrong-password",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		user := &domain.AdminUser{
			ID:           "admin-id-123",
			Email:        "admin@hros.com",
			PasswordHash: "hashed-pwd",
			Status:       "active",
		}

		mockUserRepo.On("FindByEmail", mock.Anything, reqBody.Email).Return(user, nil)
		mockPasswordHelper.On("Compare", user.PasswordHash, reqBody.Password).Return(errors.New("invalid password"))
		mockAuditLogger.On("LogLoginFailed", mock.Anything, user.Email, "invalid password").Return()

		err := handler.Login(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "unauthorized", errorResp.Code)
		assert.Equal(t, "Invalid email or password", errorResp.Message)
	})

	t.Run("Deactivated Account", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockPasswordHelper := new(mockPasswordHelper)
		mockTokenProvider := new(mockTokenProvider)
		mockAuditLogger := new(mockAuditLogger)

		loginUC := usecase.NewLoginUseCase(mockUserRepo, mockSessionRepo, mockPasswordHelper, mockTokenProvider, mockAuditLogger, &nopBruteForceCache{}, &nopLockoutNotifier{}, slog.Default())
		mockTokenBlacklist := new(mockTokenBlacklist)
		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockTokenBlacklist, mockAuditLogger)
		handler := NewAuthHandler(loginUC, logoutUC, nil)

		reqBody := dto.LoginRequest{
			Email:    "inactive@hros.com",
			Password: "password123",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		user := &domain.AdminUser{
			ID:           "admin-id-123",
			Email:        "inactive@hros.com",
			PasswordHash: "hashed-pwd",
			Status:       "inactive",
		}

		mockUserRepo.On("FindByEmail", mock.Anything, reqBody.Email).Return(user, nil)
		mockPasswordHelper.On("CompareDummy", reqBody.Password).Return()
		mockAuditLogger.On("LogLoginFailed", mock.Anything, user.Email, "user inactive").Return()

		err := handler.Login(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "forbidden", errorResp.Code)
		assert.Equal(t, "Account is deactivated", errorResp.Message)
	})

	t.Run("Locked Account", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockPasswordHelper := new(mockPasswordHelper)
		mockTokenProvider := new(mockTokenProvider)
		mockAuditLogger := new(mockAuditLogger)

		loginUC := usecase.NewLoginUseCase(mockUserRepo, mockSessionRepo, mockPasswordHelper, mockTokenProvider, mockAuditLogger, &nopBruteForceCache{}, &nopLockoutNotifier{}, slog.Default())
		mockTokenBlacklist := new(mockTokenBlacklist)
		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockTokenBlacklist, mockAuditLogger)
		handler := NewAuthHandler(loginUC, logoutUC, nil)

		reqBody := dto.LoginRequest{
			Email:    "locked@hros.com",
			Password: "password123",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		user := &domain.AdminUser{
			ID:           "admin-id-123",
			Email:        "locked@hros.com",
			PasswordHash: "hashed-pwd",
			Status:       "active",
			LockedUntil:  func(t time.Time) *time.Time { return &t }(time.Now().Add(5 * time.Minute)),
		}

		mockUserRepo.On("FindByEmail", mock.Anything, reqBody.Email).Return(user, nil)
		mockPasswordHelper.On("CompareDummy", reqBody.Password).Return()
		mockAuditLogger.On("LogLoginFailed", mock.Anything, user.Email, "user locked").Return()

		err := handler.Login(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "forbidden", errorResp.Code)
		assert.Equal(t, "Account is locked", errorResp.Message)
	})

	t.Run("Validation Failure", func(t *testing.T) {
		loginUC := usecase.NewLoginUseCase(nil, nil, nil, nil, nil, &nopBruteForceCache{}, &nopLockoutNotifier{}, slog.Default())
		handler := NewAuthHandler(loginUC, nil, nil)

		reqBody := dto.LoginRequest{
			Email: "invalid-email",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Login(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "validation_error", errorResp.Code)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	e := echo.New()

	t.Run("Successful Logout", func(t *testing.T) {
		mockSessionRepo := new(mockSessionRepo)
		mockTokenBlacklist := new(mockTokenBlacklist)
		mockAuditLogger := new(mockAuditLogger)

		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockTokenBlacklist, mockAuditLogger)
		handler := NewAuthHandler(nil, logoutUC, nil)

		req := httptest.NewRequest(http.MethodDelete, "/v1/auth/session", nil)
		req.Header.Set("Authorization", "Bearer valid-refresh-token")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockSessionRepo.On("DeleteByToken", mock.Anything, "valid-refresh-token").Return(nil)
		mockAuditLogger.On("LogLogoutSuccess", mock.Anything, "valid-refresh-token").Return()

		err := handler.Logout(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("Successful Logout with Access Token Blacklisting", func(t *testing.T) {
		mockSessionRepo := new(mockSessionRepo)
		mockTokenBlacklist := new(mockTokenBlacklist)
		mockAuditLogger := new(mockAuditLogger)

		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockTokenBlacklist, mockAuditLogger)
		handler := NewAuthHandler(nil, logoutUC, nil)

		req := httptest.NewRequest(http.MethodDelete, "/v1/auth/session", nil)
		xToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"jti": "test-jti-uuid",
			"exp": float64(time.Now().Add(10 * time.Minute).Unix()),
		})
		accessTokenStr, err := xToken.SignedString([]byte("secret"))
		assert.NoError(t, err)

		req.Header.Set("Authorization", "Bearer "+accessTokenStr)
		req.Header.Set("X-Refresh-Token", "valid-refresh-token")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockSessionRepo.On("DeleteByToken", mock.Anything, "valid-refresh-token").Return(nil)
		mockAuditLogger.On("LogLogoutSuccess", mock.Anything, "valid-refresh-token").Return()
		mockTokenBlacklist.On("Add", mock.Anything, "test-jti-uuid", mock.Anything).Return(nil)

		err = handler.Logout(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("Malformed X-Refresh-Token Returns 400", func(t *testing.T) {
		handler := NewAuthHandler(nil, nil, nil)

		invalidTokens := []string{"", "   ", "\t", "\n"}
		for _, token := range invalidTokens {
			req := httptest.NewRequest(http.MethodDelete, "/v1/auth/session", nil)
			req.Header.Set("Authorization", "Bearer valid-access-token")
			req.Header.Set("X-Refresh-Token", token)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.Logout(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected BadRequest for token %q", token)

			var errorResp sharedErrors.ErrorResponse
			err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
			assert.NoError(t, err)
			assert.Equal(t, "bad_request", errorResp.Code)
		}
	})

	t.Run("Missing Authorization Header", func(t *testing.T) {
		handler := NewAuthHandler(nil, nil, nil)

		req := httptest.NewRequest(http.MethodDelete, "/v1/auth/session", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Logout(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "unauthorized", errorResp.Code)
	})

	t.Run("Malformed Authorization Header", func(t *testing.T) {
		handler := NewAuthHandler(nil, nil, nil)

		req := httptest.NewRequest(http.MethodDelete, "/v1/auth/session", nil)
		req.Header.Set("Authorization", "InvalidHeaderFormat")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Logout(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "unauthorized", errorResp.Code)
	})

	t.Run("UseCase Error Returns 500", func(t *testing.T) {
		mockSessionRepo := new(mockSessionRepo)
		mockTokenBlacklist := new(mockTokenBlacklist)
		mockAuditLogger := new(mockAuditLogger)

		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockTokenBlacklist, mockAuditLogger)
		handler := NewAuthHandler(nil, logoutUC, nil)

		req := httptest.NewRequest(http.MethodDelete, "/v1/auth/session", nil)
		req.Header.Set("Authorization", "Bearer valid-refresh-token")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockSessionRepo.On("DeleteByToken", mock.Anything, "valid-refresh-token").Return(errors.New("db error"))

		err := handler.Logout(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "internal_error", errorResp.Code)
	})

	t.Run("Empty Token After Bearer Prefix", func(t *testing.T) {
		handler := NewAuthHandler(nil, nil, nil)

		req := httptest.NewRequest(http.MethodDelete, "/v1/auth/session", nil)
		req.Header.Set("Authorization", "Bearer ")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Logout(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "unauthorized", errorResp.Code)
	})
}

func TestAuthHandler_Refresh(t *testing.T) {
	e := echo.New()

	t.Run("Successful Refresh", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockTokenProvider := new(mockTokenProvider)
		mockAuditLogger := new(mockAuditLogger)

		refreshUC := usecase.NewRefreshSessionUseCase(mockUserRepo, mockSessionRepo, mockTokenProvider, mockAuditLogger)
		handler := NewAuthHandler(nil, nil, refreshUC)

		reqBody := dto.RefreshRequest{
			RefreshToken: "valid-refresh-token",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		session := &domain.SessionToken{
			ID:           "session-id",
			AdminID:      "user-123",
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}
		user := &domain.AdminUser{
			ID:     "user-123",
			Status: domain.AdminUserStatusActive,
		}

		mockSessionRepo.On("FindByToken", mock.Anything, "valid-refresh-token").Return(session, nil).Once()
		mockUserRepo.On("FindByID", mock.Anything, "user-123").Return(user, nil).Once()
		mockTokenProvider.On("GenerateAccessToken", mock.Anything, user, 15*time.Minute).Return("new-access-token", nil).Once()
		mockTokenProvider.On("GenerateRefreshToken", mock.Anything).Return("new-refresh-token", nil).Once()
		mockSessionRepo.On("UpdateToken", mock.Anything, mock.Anything).Return(nil).Once()
		mockAuditLogger.On("LogSessionRefreshed", mock.Anything, "user-123").Return().Once()

		err := handler.Refresh(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp dto.LoginResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "new-access-token", resp.AccessToken)
		assert.Equal(t, "new-refresh-token", resp.RefreshToken)
	})

	t.Run("Validation Failure", func(t *testing.T) {
		handler := NewAuthHandler(nil, nil, nil)

		reqBody := dto.RefreshRequest{
			RefreshToken: "",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Refresh(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "validation_error", errorResp.Code)
	})

	t.Run("Invalid Refresh Token Returns 401", func(t *testing.T) {
		mockSessionRepo := new(mockSessionRepo)
		refreshUC := usecase.NewRefreshSessionUseCase(nil, mockSessionRepo, nil, nil)
		handler := NewAuthHandler(nil, nil, refreshUC)

		reqBody := dto.RefreshRequest{
			RefreshToken: "invalid-token",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockSessionRepo.On("FindByToken", mock.Anything, "invalid-token").Return((*domain.SessionToken)(nil), domainErrors.ErrInvalidRefreshToken).Once()

		err := handler.Refresh(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "unauthorized", errorResp.Code)
	})

	t.Run("Inactive User Returns 403", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		refreshUC := usecase.NewRefreshSessionUseCase(mockUserRepo, mockSessionRepo, nil, nil)
		handler := NewAuthHandler(nil, nil, refreshUC)

		reqBody := dto.RefreshRequest{
			RefreshToken: "valid-token-inactive-user",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		session := &domain.SessionToken{
			ID:           "session-id",
			AdminID:      "user-123",
			RefreshToken: "valid-token-inactive-user",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}
		user := &domain.AdminUser{
			ID:     "user-123",
			Status: "inactive",
		}

		mockSessionRepo.On("FindByToken", mock.Anything, "valid-token-inactive-user").Return(session, nil).Once()
		mockUserRepo.On("FindByID", mock.Anything, "user-123").Return(user, nil).Once()

		err := handler.Refresh(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "forbidden", errorResp.Code)
	})

	t.Run("Locked User Returns 403", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		refreshUC := usecase.NewRefreshSessionUseCase(mockUserRepo, mockSessionRepo, nil, nil)
		handler := NewAuthHandler(nil, nil, refreshUC)

		reqBody := dto.RefreshRequest{
			RefreshToken: "valid-token-locked-user",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		session := &domain.SessionToken{
			ID:           "session-id",
			AdminID:      "user-123",
			RefreshToken: "valid-token-locked-user",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}
		lockedTime := time.Now().Add(1 * time.Hour)
		user := &domain.AdminUser{
			ID:          "user-123",
			Status:      domain.AdminUserStatusActive,
			LockedUntil: &lockedTime,
		}

		mockSessionRepo.On("FindByToken", mock.Anything, "valid-token-locked-user").Return(session, nil).Once()
		mockUserRepo.On("FindByID", mock.Anything, "user-123").Return(user, nil).Once()

		err := handler.Refresh(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "forbidden", errorResp.Code)
	})

	t.Run("Expired Token Returns 401", func(t *testing.T) {
		mockSessionRepo := new(mockSessionRepo)
		refreshUC := usecase.NewRefreshSessionUseCase(nil, mockSessionRepo, nil, nil)
		handler := NewAuthHandler(nil, nil, refreshUC)

		reqBody := dto.RefreshRequest{
			RefreshToken: "expired-token",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		session := &domain.SessionToken{
			ID:           "session-id",
			AdminID:      "user-123",
			RefreshToken: "expired-token",
			ExpiresAt:    time.Now().Add(-1 * time.Hour),
		}

		mockSessionRepo.On("FindByToken", mock.Anything, "expired-token").Return(session, nil).Once()

		err := handler.Refresh(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "unauthorized", errorResp.Code)
	})

	t.Run("Internal Error Returns 500", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		refreshUC := usecase.NewRefreshSessionUseCase(mockUserRepo, mockSessionRepo, nil, nil)
		handler := NewAuthHandler(nil, nil, refreshUC)

		reqBody := dto.RefreshRequest{
			RefreshToken: "valid-token-db-err",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		session := &domain.SessionToken{
			ID:           "session-id",
			AdminID:      "user-123",
			RefreshToken: "valid-token-db-err",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}

		mockSessionRepo.On("FindByToken", mock.Anything, "valid-token-db-err").Return(session, nil).Once()
		mockUserRepo.On("FindByID", mock.Anything, "user-123").Return((*domain.AdminUser)(nil), errors.New("db error")).Once()

		err := handler.Refresh(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "internal_error", errorResp.Code)
	})
}
