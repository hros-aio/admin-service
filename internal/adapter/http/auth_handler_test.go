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
	"github.com/hros/admin-service/internal/application/usecase"
	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/hros/admin-service/internal/domain/events"
	sharedErrors "github.com/hros/admin-service/internal/shared/errors"
	"github.com/labstack/echo/v4"
	"github.com/pquerna/otp/totp"
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

func (m *mockUserRepo) GetRoleCodeByID(ctx context.Context, roleID string) (string, error) {
	args := m.Called(ctx, roleID)
	return args.String(0), args.Error(1)
}

func (m *mockUserRepo) UpdatePassword(ctx context.Context, id string, newHash string) error {
	return m.Called(ctx, id, newHash).Error(0)
}

func (m *mockUserRepo) ActivateAccount(ctx context.Context, adminID string, newHash string) error {
	return m.Called(ctx, adminID, newHash).Error(0)
}

type mockMFACache struct{ mock.Mock }

func (m *mockMFACache) StoreToken(ctx context.Context, token string, adminID string) error {
	return m.Called(ctx, token, adminID).Error(0)
}
func (m *mockMFACache) GetAdminID(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}
func (m *mockMFACache) DeleteToken(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
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

func (m *mockSessionRepo) DeleteAllByAdminID(ctx context.Context, adminID string) error {
	return m.Called(ctx, adminID).Error(0)
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

func (m *mockAuditLogger) LogMFAChallengeIssued(ctx context.Context, userID string, email string) {
	m.Called(ctx, userID, email)
}

func (m *mockAuditLogger) LogMFASuccess(ctx context.Context, userID string, email string) {
	m.Called(ctx, userID, email)
}

func (m *mockAuditLogger) LogMFAFailed(ctx context.Context, email string, reason string) {
	m.Called(ctx, email, reason)
}

func (m *mockAuditLogger) LogPasswordResetRequested(ctx context.Context, event events.PasswordResetRequestedEvent) {
	m.Called(ctx, event)
}

func (m *mockAuditLogger) LogPasswordResetCompleted(ctx context.Context, event events.PasswordResetCompletedEvent) {
	m.Called(ctx, event)
}

func (m *mockAuditLogger) LogInviteAccepted(ctx context.Context, event events.InviteAcceptedEvent) {
	m.Called(ctx, event)
}

func (m *mockAuditLogger) LogAdminActivated(ctx context.Context, event events.AdminActivatedEvent) {
	m.Called(ctx, event)
}

type mockPasswordResetCache struct {
	mock.Mock
}

func (m *mockPasswordResetCache) StoreToken(ctx context.Context, token string, adminID string, ttl time.Duration) error {
	return m.Called(ctx, token, adminID, ttl).Error(0)
}

func (m *mockPasswordResetCache) ConsumeToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func (m *mockPasswordResetCache) DeleteToken(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}

type mockPasswordResetNotifier struct {
	mock.Mock
}

func (m *mockPasswordResetNotifier) PublishPasswordResetEmail(ctx context.Context, event events.EmailSendEvent) error {
	return m.Called(ctx, event).Error(0)
}

// mockAcceptInviteUseCase is a testify mock for AcceptInviteUseCase used in handler tests.
type mockAcceptInviteUseCase struct{ mock.Mock }

func (m *mockAcceptInviteUseCase) Execute(ctx context.Context, input usecase.AcceptInviteInput) error {
	return m.Called(ctx, input).Error(0)
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
	password *mockPasswordHelper,
	tokens *mockTokenProvider,
	audit *mockAuditLogger,
) *usecase.LoginUseCase {
	if mRepo, ok := userRepo.(*mockUserRepo); ok {
		mRepo.On("GetRoleCodeByID", mock.Anything, mock.Anything).Return("STANDARD_ADMIN", nil).Maybe()
	}
	return usecase.NewLoginUseCase(
		userRepo,
		sessionRepo,
		password,
		tokens,
		audit,
		&nopBruteForceCache{},
		&nopLockoutNotifier{},
		&mockMFACache{},
		slog.Default(),
	)
}

type mockBruteForceCache struct{ mock.Mock }

func (m *mockBruteForceCache) IncrementFailedAttempts(ctx context.Context, email string, window time.Duration) (int, error) {
	args := m.Called(ctx, email, window)
	return args.Int(0), args.Error(1)
}

func (m *mockBruteForceCache) GetFailedAttempts(ctx context.Context, email string) (int, error) {
	args := m.Called(ctx, email)
	return args.Int(0), args.Error(1)
}

func (m *mockBruteForceCache) SetLockout(ctx context.Context, email string, duration time.Duration) error {
	args := m.Called(ctx, email, duration)
	return args.Error(0)
}

func (m *mockBruteForceCache) IsLocked(ctx context.Context, email string) (bool, time.Time, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Get(1).(time.Time), args.Error(2)
}

func (m *mockBruteForceCache) Reset(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
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

		loginUC := newLoginUCForTest(mockUserRepo, mockSessionRepo, mockPasswordHelper, mockTokenProvider, mockAuditLogger)
		mockTokenBlacklist := new(mockTokenBlacklist)
		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockTokenBlacklist, mockAuditLogger)
		handler := NewAuthHandler(loginUC, logoutUC, nil, nil, nil, nil, nil)

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

	t.Run("Super Admin MFA Challenge", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockPasswordHelper := new(mockPasswordHelper)
		mockTokenProvider := new(mockTokenProvider)
		mockAuditLogger := new(mockAuditLogger)
		mockMFACache := new(mockMFACache)

		loginUC := usecase.NewLoginUseCase(
			mockUserRepo,
			mockSessionRepo,
			mockPasswordHelper,
			mockTokenProvider,
			mockAuditLogger,
			&nopBruteForceCache{},
			&nopLockoutNotifier{},
			mockMFACache,
			slog.Default(),
		)
		mockTokenBlacklist := new(mockTokenBlacklist)
		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockTokenBlacklist, mockAuditLogger)
		handler := NewAuthHandler(loginUC, logoutUC, nil, nil, nil, nil, nil)

		reqBody := dto.LoginRequest{
			Email:      "superadmin@hros.com",
			Password:   "password123",
			RememberMe: true,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		user := &domain.AdminUser{
			ID:           "super-admin-id-456",
			Email:        "superadmin@hros.com",
			PasswordHash: "hashed-pwd",
			Status:       "active",
			RoleID:       "super-admin-role-id",
		}

		mockUserRepo.On("FindByEmail", mock.Anything, reqBody.Email).Return(user, nil)
		mockPasswordHelper.On("Compare", user.PasswordHash, reqBody.Password).Return(nil)
		mockUserRepo.On("GetRoleCodeByID", mock.Anything, user.RoleID).Return("SUPER_ADMIN", nil)
		mockMFACache.On("StoreToken", mock.Anything, mock.Anything, user.ID).Return(nil)
		mockAuditLogger.On("LogMFAChallengeIssued", mock.Anything, user.ID, user.Email).Return()

		err := handler.Login(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp dto.LoginResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.MFARequired)
		assert.Regexp(t, "^[0-9a-fA-F]{64}$", resp.MFAToken)
		assert.Equal(t, []string{"totp", "webauthn"}, resp.MFAMethods)
		assert.Empty(t, resp.AccessToken)
		assert.Empty(t, resp.RefreshToken)
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockPasswordHelper := new(mockPasswordHelper)
		mockTokenProvider := new(mockTokenProvider)
		mockAuditLogger := new(mockAuditLogger)

		loginUC := newLoginUCForTest(mockUserRepo, mockSessionRepo, mockPasswordHelper, mockTokenProvider, mockAuditLogger)
		mockTokenBlacklist := new(mockTokenBlacklist)
		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockTokenBlacklist, mockAuditLogger)
		handler := NewAuthHandler(loginUC, logoutUC, nil, nil, nil, nil, nil)

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

		loginUC := newLoginUCForTest(mockUserRepo, mockSessionRepo, mockPasswordHelper, mockTokenProvider, mockAuditLogger)
		mockTokenBlacklist := new(mockTokenBlacklist)
		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockTokenBlacklist, mockAuditLogger)
		handler := NewAuthHandler(loginUC, logoutUC, nil, nil, nil, nil, nil)

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

		loginUC := newLoginUCForTest(mockUserRepo, mockSessionRepo, mockPasswordHelper, mockTokenProvider, mockAuditLogger)
		mockTokenBlacklist := new(mockTokenBlacklist)
		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockTokenBlacklist, mockAuditLogger)
		handler := NewAuthHandler(loginUC, logoutUC, nil, nil, nil, nil, nil)

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
		loginUC := newLoginUCForTest(nil, nil, nil, nil, nil)
		handler := NewAuthHandler(loginUC, nil, nil, nil, nil, nil, nil)

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

	t.Run("Temporary Lockout (Brute Force)", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockPasswordHelper := new(mockPasswordHelper)
		mockTokenProvider := new(mockTokenProvider)
		mockAuditLogger := new(mockAuditLogger)
		mockBruteForce := new(mockBruteForceCache)

		mockUserRepo.On("GetRoleCodeByID", mock.Anything, mock.Anything).Return("STANDARD_ADMIN", nil).Maybe()

		loginUC := usecase.NewLoginUseCase(
			mockUserRepo,
			mockSessionRepo,
			mockPasswordHelper,
			mockTokenProvider,
			mockAuditLogger,
			mockBruteForce,
			&nopLockoutNotifier{},
			&mockMFACache{},
			slog.Default(),
		)
		mockTokenBlacklist := new(mockTokenBlacklist)
		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockTokenBlacklist, mockAuditLogger)
		handler := NewAuthHandler(loginUC, logoutUC, nil, nil, nil, nil, nil)

		reqBody := dto.LoginRequest{
			Email:    "locked-bf@hros.com",
			Password: "password123",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockBruteForce.On("IsLocked", mock.Anything, "locked-bf@hros.com").Return(true, time.Now().Add(30*time.Minute), nil)

		err := handler.Login(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "ACCOUNT_LOCKED", errorResp.Code)
		assert.Equal(t, "Account is temporarily locked", errorResp.Message)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	e := echo.New()

	t.Run("Successful Logout", func(t *testing.T) {
		mockSessionRepo := new(mockSessionRepo)
		mockTokenBlacklist := new(mockTokenBlacklist)
		mockAuditLogger := new(mockAuditLogger)

		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockTokenBlacklist, mockAuditLogger)
		handler := NewAuthHandler(nil, logoutUC, nil, nil, nil, nil, nil)

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
		handler := NewAuthHandler(nil, logoutUC, nil, nil, nil, nil, nil)

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
		handler := NewAuthHandler(nil, nil, nil, nil, nil, nil, nil)

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
		handler := NewAuthHandler(nil, nil, nil, nil, nil, nil, nil)

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
		handler := NewAuthHandler(nil, nil, nil, nil, nil, nil, nil)

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
		handler := NewAuthHandler(nil, logoutUC, nil, nil, nil, nil, nil)

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
		handler := NewAuthHandler(nil, nil, nil, nil, nil, nil, nil)

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
		handler := NewAuthHandler(nil, nil, refreshUC, nil, nil, nil, nil)

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
		handler := NewAuthHandler(nil, nil, nil, nil, nil, nil, nil)

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
		handler := NewAuthHandler(nil, nil, refreshUC, nil, nil, nil, nil)

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
		handler := NewAuthHandler(nil, nil, refreshUC, nil, nil, nil, nil)

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
		handler := NewAuthHandler(nil, nil, refreshUC, nil, nil, nil, nil)

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
		handler := NewAuthHandler(nil, nil, refreshUC, nil, nil, nil, nil)

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
		handler := NewAuthHandler(nil, nil, refreshUC, nil, nil, nil, nil)

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

	t.Run("Temporary Lockout (Brute Force)", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockTokenProvider := new(mockTokenProvider)
		mockAuditLogger := new(mockAuditLogger)

		refreshUC := usecase.NewRefreshSessionUseCase(mockUserRepo, mockSessionRepo, mockTokenProvider, mockAuditLogger)
		handler := NewAuthHandler(nil, nil, refreshUC, nil, nil, nil, nil)

		reqBody := dto.RefreshRequest{
			RefreshToken: "valid-token",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		session := &domain.SessionToken{
			ID:           "session-id",
			AdminID:      "user-123",
			RefreshToken: "valid-token",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}

		mockSessionRepo.On("FindByToken", mock.Anything, "valid-token").Return(session, nil).Once()
		mockUserRepo.On("FindByID", mock.Anything, "user-123").Return((*domain.AdminUser)(nil), domainErrors.ErrAccountLocked).Once()

		err := handler.Refresh(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "ACCOUNT_LOCKED", errorResp.Code)
		assert.Equal(t, "Account is temporarily locked", errorResp.Message)
	})
}

func TestAuthHandler_VerifyMFA(t *testing.T) {
	e := echo.New()

	t.Run("Successful MFA Verification", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockTokenProvider := new(mockTokenProvider)
		mockAuditLogger := new(mockAuditLogger)
		mockMFACache := new(mockMFACache)

		verifyMfaUC := usecase.NewVerifyMFAUseCase(
			mockUserRepo,
			mockSessionRepo,
			mockTokenProvider,
			mockAuditLogger,
			mockMFACache,
			slog.Default(),
		)
		handler := NewAuthHandler(nil, nil, nil, verifyMfaUC, nil, nil, nil)

		reqBody := dto.MFAVerifyRequest{
			MFAToken: "valid-mfa-token",
			Method:   "totp",
			Code:     "123456",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/verify", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		user := &domain.AdminUser{
			ID:         "admin-id-123",
			Email:      "admin@hros.com",
			Name:       "Admin",
			Status:     "active",
			TotpSecret: "secret",
		}

		mockMFACache.On("GetAdminID", mock.Anything, "valid-mfa-token").Return("admin-id-123", nil).Once()
		mockUserRepo.On("FindByID", mock.Anything, "admin-id-123").Return(user, nil).Once()

		totpSecret := "JBSWY3DPEHPK3PXP" // Valid Base32
		user.TotpSecret = totpSecret
		validCode, err := totp.GenerateCode(totpSecret, time.Now())
		assert.NoError(t, err)
		reqBody.Code = validCode
		bodyBytes, _ = json.Marshal(reqBody)
		req = httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/verify", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)

		mockAuditLogger.On("LogMFASuccess", mock.Anything, "admin-id-123", "admin@hros.com").Return().Once()
		mockTokenProvider.On("GenerateAccessToken", mock.Anything, user, 15*time.Minute).Return("access-token-xyz", nil).Once()
		mockTokenProvider.On("GenerateRefreshToken", mock.Anything).Return("refresh-token-abc", nil).Once()
		mockSessionRepo.On("Save", mock.Anything, mock.Anything).Return(nil).Once()
		mockMFACache.On("DeleteToken", mock.Anything, "valid-mfa-token").Return(nil).Once()

		err = handler.VerifyMFA(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp dto.LoginResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "access-token-xyz", resp.AccessToken)
		assert.Equal(t, "refresh-token-abc", resp.RefreshToken)
	})

	t.Run("Validation Failure", func(t *testing.T) {
		handler := NewAuthHandler(nil, nil, nil, nil, nil, nil, nil)

		reqBody := dto.MFAVerifyRequest{
			MFAToken: "", // Empty token triggers validation failure
			Method:   "totp",
			Code:     "",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/verify", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.VerifyMFA(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "validation_error", errorResp.Code)
	})

	t.Run("Invalid MFA Token / Expired", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockTokenProvider := new(mockTokenProvider)
		mockAuditLogger := new(mockAuditLogger)
		mockMFACache := new(mockMFACache)

		verifyMfaUC := usecase.NewVerifyMFAUseCase(
			mockUserRepo,
			mockSessionRepo,
			mockTokenProvider,
			mockAuditLogger,
			mockMFACache,
			slog.Default(),
		)
		handler := NewAuthHandler(nil, nil, nil, verifyMfaUC, nil, nil, nil)

		reqBody := dto.MFAVerifyRequest{
			MFAToken: "expired-token",
			Method:   "totp",
			Code:     "123456",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/verify", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockMFACache.On("GetAdminID", mock.Anything, "expired-token").Return("", domainErrors.ErrMFATokenExpired).Once()

		err := handler.VerifyMFA(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "MFA_TOKEN_EXPIRED", errorResp.Code)
	})

	t.Run("Invalid Code", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockTokenProvider := new(mockTokenProvider)
		mockAuditLogger := new(mockAuditLogger)
		mockMFACache := new(mockMFACache)

		verifyMfaUC := usecase.NewVerifyMFAUseCase(
			mockUserRepo,
			mockSessionRepo,
			mockTokenProvider,
			mockAuditLogger,
			mockMFACache,
			slog.Default(),
		)
		handler := NewAuthHandler(nil, nil, nil, verifyMfaUC, nil, nil, nil)

		reqBody := dto.MFAVerifyRequest{
			MFAToken: "valid-mfa-token",
			Method:   "totp",
			Code:     "000000", // Incorrect code
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/verify", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		user := &domain.AdminUser{
			ID:         "admin-id-123",
			Email:      "admin@hros.com",
			Name:       "Admin",
			Status:     "active",
			TotpSecret: "JBSWY3DPEHPK3PXP",
		}

		mockMFACache.On("GetAdminID", mock.Anything, "valid-mfa-token").Return("admin-id-123", nil).Once()
		mockUserRepo.On("FindByID", mock.Anything, "admin-id-123").Return(user, nil).Once()
		mockAuditLogger.On("LogMFAFailed", mock.Anything, "admin@hros.com", "invalid TOTP code").Return().Once()

		err := handler.VerifyMFA(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "MFA_INVALID", errorResp.Code)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockTokenProvider := new(mockTokenProvider)
		mockAuditLogger := new(mockAuditLogger)
		mockMFACache := new(mockMFACache)

		verifyMfaUC := usecase.NewVerifyMFAUseCase(
			mockUserRepo,
			mockSessionRepo,
			mockTokenProvider,
			mockAuditLogger,
			mockMFACache,
			slog.Default(),
		)
		handler := NewAuthHandler(nil, nil, nil, verifyMfaUC, nil, nil, nil)

		reqBody := dto.MFAVerifyRequest{
			MFAToken: "valid-mfa-token",
			Method:   "totp",
			Code:     "123456",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/verify", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockMFACache.On("GetAdminID", mock.Anything, "valid-mfa-token").Return("", errors.New("db disconnect")).Once()

		err := handler.VerifyMFA(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "internal_error", errorResp.Code)
	})

	t.Run("Forbidden - User Inactive", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockTokenProvider := new(mockTokenProvider)
		mockAuditLogger := new(mockAuditLogger)
		mockMFACache := new(mockMFACache)

		verifyMfaUC := usecase.NewVerifyMFAUseCase(
			mockUserRepo,
			mockSessionRepo,
			mockTokenProvider,
			mockAuditLogger,
			mockMFACache,
			slog.Default(),
		)
		handler := NewAuthHandler(nil, nil, nil, verifyMfaUC, nil, nil, nil)

		reqBody := dto.MFAVerifyRequest{
			MFAToken: "valid-mfa-token",
			Method:   "totp",
			Code:     "123456",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/verify", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockMFACache.On("GetAdminID", mock.Anything, "valid-mfa-token").Return("admin-id-123", nil).Once()
		mockUserRepo.On("FindByID", mock.Anything, "admin-id-123").Return((*domain.AdminUser)(nil), domainErrors.ErrUserInactive).Once()

		err := handler.VerifyMFA(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "forbidden", errorResp.Code)
	})

	t.Run("Forbidden - User Locked", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockTokenProvider := new(mockTokenProvider)
		mockAuditLogger := new(mockAuditLogger)
		mockMFACache := new(mockMFACache)

		verifyMfaUC := usecase.NewVerifyMFAUseCase(
			mockUserRepo,
			mockSessionRepo,
			mockTokenProvider,
			mockAuditLogger,
			mockMFACache,
			slog.Default(),
		)
		handler := NewAuthHandler(nil, nil, nil, verifyMfaUC, nil, nil, nil)

		reqBody := dto.MFAVerifyRequest{
			MFAToken: "valid-mfa-token",
			Method:   "totp",
			Code:     "123456",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/mfa/verify", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockMFACache.On("GetAdminID", mock.Anything, "valid-mfa-token").Return("admin-id-123", nil).Once()
		mockUserRepo.On("FindByID", mock.Anything, "admin-id-123").Return((*domain.AdminUser)(nil), domainErrors.ErrUserLocked).Once()

		err := handler.VerifyMFA(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "forbidden", errorResp.Code)
	})
}

func TestAuthHandler_RequestPasswordReset(t *testing.T) {
	e := echo.New()

	t.Run("Successful Request - Registered Email", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockCache := new(mockPasswordResetCache)
		mockAudit := new(mockAuditLogger)
		mockNotifier := new(mockPasswordResetNotifier)

		uc := usecase.NewRequestPasswordResetUseCase(mockUserRepo, mockCache, mockAudit, mockNotifier)
		handler := NewAuthHandler(nil, nil, nil, nil, uc, nil, nil)

		reqBody := dto.PasswordResetRequest{
			Email: "admin@example.com",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/password-reset/request", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		user := &domain.AdminUser{ID: "admin-123", Email: "admin@example.com"}
		mockUserRepo.On("FindByEmail", mock.Anything, "admin@example.com").Return(user, nil).Once()
		mockCache.On("StoreToken", mock.Anything, mock.Anything, "admin-123", 60*time.Minute).Return(nil).Once()
		mockAudit.On("LogPasswordResetRequested", mock.Anything, mock.Anything).Once()
		mockNotifier.On("PublishPasswordResetEmail", mock.Anything, mock.Anything).Return(nil).Once()

		err := handler.RequestPasswordReset(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "If an account exists for that email, a reset link has been sent.", resp["message"])
	})

	t.Run("Validation Failure - Empty Email", func(t *testing.T) {
		handler := NewAuthHandler(nil, nil, nil, nil, nil, nil, nil)

		reqBody := dto.PasswordResetRequest{
			Email: "",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/password-reset/request", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.RequestPasswordReset(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "validation_error", errorResp.Code)
	})

	t.Run("UseCase Internal Error - Returns 500", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockCache := new(mockPasswordResetCache)
		mockAudit := new(mockAuditLogger)
		mockNotifier := new(mockPasswordResetNotifier)

		uc := usecase.NewRequestPasswordResetUseCase(mockUserRepo, mockCache, mockAudit, mockNotifier)
		handler := NewAuthHandler(nil, nil, nil, nil, uc, nil, nil)

		reqBody := dto.PasswordResetRequest{
			Email: "admin@example.com",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/password-reset/request", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUserRepo.On("FindByEmail", mock.Anything, "admin@example.com").Return(nil, errors.New("database down")).Once()

		err := handler.RequestPasswordReset(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "internal_error", errorResp.Code)
	})
}

func TestAuthHandler_ConfirmPasswordReset(t *testing.T) {
	e := echo.New()

	t.Run("Successful Confirm", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockCache := new(mockPasswordResetCache)
		mockAudit := new(mockAuditLogger)

		uc := usecase.NewConfirmPasswordResetUseCase(mockUserRepo, mockSessionRepo, mockCache, mockAudit)
		handler := NewAuthHandler(nil, nil, nil, nil, nil, uc, nil)

		reqBody := dto.PasswordResetConfirmRequest{
			Token:                "valid-token",
			Password:             "StrongPass1!",
			PasswordConfirmation: "StrongPass1!",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/password-reset/confirm", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockCache.On("ConsumeToken", mock.Anything, "valid-token").Return("admin-123", nil).Once()
		mockUserRepo.On("FindByID", mock.Anything, "admin-123").Return(&domain.AdminUser{Email: "admin@example.com"}, nil).Once()
		mockUserRepo.On("UpdatePassword", mock.Anything, "admin-123", mock.Anything).Return(nil).Once()
		mockSessionRepo.On("DeleteAllByAdminID", mock.Anything, "admin-123").Return(nil).Once()
		mockAudit.On("LogPasswordResetCompleted", mock.Anything, mock.Anything).Once()

		err := handler.ConfirmPasswordReset(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Password updated successfully.", resp["message"])
	})

	t.Run("Validation Failure - Password Mismatch", func(t *testing.T) {
		handler := NewAuthHandler(nil, nil, nil, nil, nil, nil, nil)

		reqBody := dto.PasswordResetConfirmRequest{
			Token:                "valid-token",
			Password:             "StrongPass1!",
			PasswordConfirmation: "DifferentPass2!",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/password-reset/confirm", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.ConfirmPasswordReset(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "validation_error", errorResp.Code)
	})

	t.Run("Token Expired - Returns 400", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockCache := new(mockPasswordResetCache)
		mockAudit := new(mockAuditLogger)

		uc := usecase.NewConfirmPasswordResetUseCase(mockUserRepo, mockSessionRepo, mockCache, mockAudit)
		handler := NewAuthHandler(nil, nil, nil, nil, nil, uc, nil)

		reqBody := dto.PasswordResetConfirmRequest{
			Token:                "expired-token",
			Password:             "StrongPass1!",
			PasswordConfirmation: "StrongPass1!",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/password-reset/confirm", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockCache.On("ConsumeToken", mock.Anything, "expired-token").Return("", domainErrors.ErrTokenExpired).Once()

		err := handler.ConfirmPasswordReset(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "TOKEN_EXPIRED", errorResp.Code)
	})

	t.Run("Token Already Used - Returns 400", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockCache := new(mockPasswordResetCache)
		mockAudit := new(mockAuditLogger)

		uc := usecase.NewConfirmPasswordResetUseCase(mockUserRepo, mockSessionRepo, mockCache, mockAudit)
		handler := NewAuthHandler(nil, nil, nil, nil, nil, uc, nil)

		reqBody := dto.PasswordResetConfirmRequest{
			Token:                "used-token",
			Password:             "StrongPass1!",
			PasswordConfirmation: "StrongPass1!",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/password-reset/confirm", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockCache.On("ConsumeToken", mock.Anything, "used-token").Return("", domainErrors.ErrTokenUsed).Once()

		err := handler.ConfirmPasswordReset(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "TOKEN_USED", errorResp.Code)
	})

	t.Run("Weak Password - Returns 422", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockCache := new(mockPasswordResetCache)
		mockAudit := new(mockAuditLogger)

		uc := usecase.NewConfirmPasswordResetUseCase(mockUserRepo, mockSessionRepo, mockCache, mockAudit)
		handler := NewAuthHandler(nil, nil, nil, nil, nil, uc, nil)

		reqBody := dto.PasswordResetConfirmRequest{
			Token:                "valid-token",
			Password:             "weak",
			PasswordConfirmation: "weak",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/password-reset/confirm", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.ConfirmPasswordReset(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "PASSWORD_WEAK", errorResp.Code)
	})

	t.Run("Internal Error - Returns 500", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockCache := new(mockPasswordResetCache)
		mockAudit := new(mockAuditLogger)

		uc := usecase.NewConfirmPasswordResetUseCase(mockUserRepo, mockSessionRepo, mockCache, mockAudit)
		handler := NewAuthHandler(nil, nil, nil, nil, nil, uc, nil)

		reqBody := dto.PasswordResetConfirmRequest{
			Token:                "valid-token",
			Password:             "StrongPass1!",
			PasswordConfirmation: "StrongPass1!",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/password-reset/confirm", bytes.NewBuffer(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockCache.On("ConsumeToken", mock.Anything, "valid-token").Return("", errors.New("cache error")).Once()

		err := handler.ConfirmPasswordReset(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "internal_error", errorResp.Code)
	})
}

func TestAuthHandler_AcceptInvite(t *testing.T) {
	e := echo.New()

	newHandler := func(uc *mockAcceptInviteUseCase) *AuthHandler {
		return NewAuthHandler(nil, nil, nil, nil, nil, nil, uc)
	}

	t.Run("Success", func(t *testing.T) {
		mockUC := new(mockAcceptInviteUseCase)
		mockUC.On("Execute", mock.Anything, mock.MatchedBy(func(input usecase.AcceptInviteInput) bool {
			return input.Token == "valid-token" && input.Password == "SecurePassword123!"
		})).Return(nil)
		handler := newHandler(mockUC)

		reqBody := dto.AcceptInviteRequest{
			Token:                "valid-token",
			Password:             "SecurePassword123!",
			PasswordConfirmation: "SecurePassword123!",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/accept-invite", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.AcceptInvite(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Account activated successfully.", resp["message"])
		mockUC.AssertExpectations(t)
	})

	t.Run("Bind Failure", func(t *testing.T) {
		handler := newHandler(nil)

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/accept-invite", bytes.NewReader([]byte(`{invalid-json`)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.AcceptInvite(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "bad_request", errorResp.Code)
	})

	t.Run("Validation Failure", func(t *testing.T) {
		handler := newHandler(nil)

		reqBody := dto.AcceptInviteRequest{
			Token:                "",
			Password:             "SecurePassword123!",
			PasswordConfirmation: "Mismatched!",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/accept-invite", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.AcceptInvite(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "validation_error", errorResp.Code)
	})

	t.Run("ErrInviteExpired maps to 400 INVITE_EXPIRED", func(t *testing.T) {
		mockUC := new(mockAcceptInviteUseCase)
		mockUC.On("Execute", mock.Anything, mock.Anything).Return(domainErrors.ErrInviteExpired)
		handler := newHandler(mockUC)

		reqBody := dto.AcceptInviteRequest{
			Token:                "expired-token",
			Password:             "SecurePassword123!",
			PasswordConfirmation: "SecurePassword123!",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/accept-invite", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.AcceptInvite(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "INVITE_EXPIRED", errorResp.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("ErrInviteUsed maps to 400 INVITE_USED", func(t *testing.T) {
		mockUC := new(mockAcceptInviteUseCase)
		mockUC.On("Execute", mock.Anything, mock.Anything).Return(domainErrors.ErrInviteUsed)
		handler := newHandler(mockUC)

		reqBody := dto.AcceptInviteRequest{
			Token:                "used-token",
			Password:             "SecurePassword123!",
			PasswordConfirmation: "SecurePassword123!",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/accept-invite", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.AcceptInvite(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "INVITE_USED", errorResp.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("ErrPasswordWeak maps to 422 PASSWORD_WEAK", func(t *testing.T) {
		mockUC := new(mockAcceptInviteUseCase)
		mockUC.On("Execute", mock.Anything, mock.Anything).Return(domainErrors.ErrPasswordWeak)
		handler := newHandler(mockUC)

		reqBody := dto.AcceptInviteRequest{
			Token:                "valid-token",
			Password:             "weak",
			PasswordConfirmation: "weak",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/accept-invite", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.AcceptInvite(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "PASSWORD_WEAK", errorResp.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("Internal Error maps to 500", func(t *testing.T) {
		mockUC := new(mockAcceptInviteUseCase)
		mockUC.On("Execute", mock.Anything, mock.Anything).Return(errors.New("unexpected db failure"))
		handler := newHandler(mockUC)

		reqBody := dto.AcceptInviteRequest{
			Token:                "valid-token",
			Password:             "SecurePassword123!",
			PasswordConfirmation: "SecurePassword123!",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/accept-invite", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.AcceptInvite(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "internal_error", errorResp.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("Nil UseCase returns 500 without panic", func(t *testing.T) {
		// Passing nil explicitly exercises the nil-guard added in the handler.
		// Without the guard this would panic; with it, callers get a controlled 500.
		handler := NewAuthHandler(nil, nil, nil, nil, nil, nil, nil)

		reqBody := dto.AcceptInviteRequest{
			Token:                "valid-token",
			Password:             "SecurePassword123!",
			PasswordConfirmation: "SecurePassword123!",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/accept-invite", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.AcceptInvite(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		var errorResp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "internal_error", errorResp.Code)
	})
}
