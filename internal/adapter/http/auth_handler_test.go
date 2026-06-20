// Package http provides HTTP adapter handlers for Echo routing.
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hros/admin-service/internal/adapter/http/auth/dto"
	"github.com/hros/admin-service/internal/application/usecase"
	"github.com/hros/admin-service/internal/domain"
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

func TestAuthHandler_Login(t *testing.T) {
	e := echo.New()

	t.Run("Successful Login", func(t *testing.T) {
		mockUserRepo := new(mockUserRepo)
		mockSessionRepo := new(mockSessionRepo)
		mockPasswordHelper := new(mockPasswordHelper)
		mockTokenProvider := new(mockTokenProvider)
		mockAuditLogger := new(mockAuditLogger)

		loginUC := usecase.NewLoginUseCase(mockUserRepo, mockSessionRepo, mockPasswordHelper, mockTokenProvider, mockAuditLogger)
		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockAuditLogger)
		handler := NewAuthHandler(loginUC, logoutUC)

		reqBody := dto.LoginRequest{
			Email:    "admin@hros.com",
			Password: "password123",
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
		mockSessionRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
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

		loginUC := usecase.NewLoginUseCase(mockUserRepo, mockSessionRepo, mockPasswordHelper, mockTokenProvider, mockAuditLogger)
		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockAuditLogger)
		handler := NewAuthHandler(loginUC, logoutUC)

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

		loginUC := usecase.NewLoginUseCase(mockUserRepo, mockSessionRepo, mockPasswordHelper, mockTokenProvider, mockAuditLogger)
		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockAuditLogger)
		handler := NewAuthHandler(loginUC, logoutUC)

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

		loginUC := usecase.NewLoginUseCase(mockUserRepo, mockSessionRepo, mockPasswordHelper, mockTokenProvider, mockAuditLogger)
		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockAuditLogger)
		handler := NewAuthHandler(loginUC, logoutUC)

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
		loginUC := usecase.NewLoginUseCase(nil, nil, nil, nil, nil)
		handler := NewAuthHandler(loginUC, nil)

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
		mockAuditLogger := new(mockAuditLogger)

		logoutUC := usecase.NewLogoutUseCase(mockSessionRepo, mockAuditLogger)
		handler := NewAuthHandler(nil, logoutUC)

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

	t.Run("Missing Authorization Header", func(t *testing.T) {
		handler := NewAuthHandler(nil, nil)

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
		handler := NewAuthHandler(nil, nil)

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
}
