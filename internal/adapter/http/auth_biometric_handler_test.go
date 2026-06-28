package http

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hros/admin-service/internal/adapter/http/auth/dto"
	"github.com/hros/admin-service/internal/application/usecase"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	sharedErrors "github.com/hros/admin-service/internal/shared/errors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockGenerateChallengeExecutor struct {
	mock.Mock
}

func (m *mockGenerateChallengeExecutor) Execute(ctx context.Context, input usecase.GenerateBiometricChallengeInput) (*usecase.GenerateBiometricChallengeOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.GenerateBiometricChallengeOutput), args.Error(1)
}

type mockVerifyBiometricExecutor struct {
	mock.Mock
}

func (m *mockVerifyBiometricExecutor) Execute(ctx context.Context, input usecase.VerifyBiometricInput) (*usecase.LoginOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.LoginOutput), args.Error(1)
}

func TestAuthBiometricHandler_Challenge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		e := echo.New()
		reqBody, _ := json.Marshal(dto.BiometricChallengeRequest{
			Email: "admin@hros.com",
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/biometric/challenge", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUC := new(mockGenerateChallengeExecutor)
		mockUC.On("Execute", mock.Anything, usecase.GenerateBiometricChallengeInput{
			Email: "admin@hros.com",
		}).Return(&usecase.GenerateBiometricChallengeOutput{
			Challenge:    "base64_encoded_challenge",
			CredentialID: "cred_id_123",
		}, nil).Once()

		handler := NewAuthBiometricHandler(mockUC, nil)
		err := handler.Challenge(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp dto.BiometricChallengeResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "base64_encoded_challenge", resp.Challenge)
		assert.Equal(t, "cred_id_123", resp.CredentialID)
		mockUC.AssertExpectations(t)
	})

	t.Run("bind error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/biometric/challenge", bytes.NewReader([]byte("{invalid-json")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := NewAuthBiometricHandler(nil, nil)
		err := handler.Challenge(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var resp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "bad_request", resp.Code)
	})

	t.Run("validation error", func(t *testing.T) {
		e := echo.New()
		reqBody, _ := json.Marshal(dto.BiometricChallengeRequest{
			Email: "invalid-email",
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/biometric/challenge", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := NewAuthBiometricHandler(nil, nil)
		err := handler.Challenge(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var resp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "validation_error", resp.Code)
	})

	t.Run("biometric not registered - indistinguishable failure", func(t *testing.T) {
		e := echo.New()
		reqBody, _ := json.Marshal(dto.BiometricChallengeRequest{
			Email: "admin@hros.com",
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/biometric/challenge", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUC := new(mockGenerateChallengeExecutor)
		mockUC.On("Execute", mock.Anything, usecase.GenerateBiometricChallengeInput{
			Email: "admin@hros.com",
		}).Return((*usecase.GenerateBiometricChallengeOutput)(nil), domainErrors.ErrBiometricNotRegistered).Once()

		handler := NewAuthBiometricHandler(mockUC, nil)
		err := handler.Challenge(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		var resp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "internal_error", resp.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("internal server error", func(t *testing.T) {
		e := echo.New()
		reqBody, _ := json.Marshal(dto.BiometricChallengeRequest{
			Email: "admin@hros.com",
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/biometric/challenge", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUC := new(mockGenerateChallengeExecutor)
		mockUC.On("Execute", mock.Anything, usecase.GenerateBiometricChallengeInput{
			Email: "admin@hros.com",
		}).Return((*usecase.GenerateBiometricChallengeOutput)(nil), errors.New("db failure")).Once()

		handler := NewAuthBiometricHandler(mockUC, nil)
		err := handler.Challenge(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestAuthBiometricHandler_Verify(t *testing.T) {
	clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(`{"type":"webauthn.get"}`))
	authDataB64 := base64.RawURLEncoding.EncodeToString([]byte("auth_data_bytes"))
	sigB64 := base64.RawURLEncoding.EncodeToString([]byte("signature_bytes"))

	t.Run("success", func(t *testing.T) {
		e := echo.New()
		reqBody, _ := json.Marshal(dto.BiometricVerifyRequest{
			Email:             "admin@hros.com",
			CredentialID:      "cred-id",
			ClientDataJSON:    clientDataB64,
			AuthenticatorData: authDataB64,
			Signature:         sigB64,
			RememberMe:        true,
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/biometric/verify", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUC := new(mockVerifyBiometricExecutor)
		mockUC.On("Execute", mock.Anything, usecase.VerifyBiometricInput{
			Email:             "admin@hros.com",
			CredentialID:      "cred-id",
			ClientDataJSON:    []byte(`{"type":"webauthn.get"}`),
			AuthenticatorData: []byte("auth_data_bytes"),
			Signature:         []byte("signature_bytes"),
			RememberMe:        true,
			IPAddress:         c.RealIP(),
			UserAgent:         c.Request().UserAgent(),
		}).Return(&usecase.LoginOutput{
			AccessToken:  "access-token-123",
			RefreshToken: "refresh-token-123",
		}, nil).Once()

		handler := NewAuthBiometricHandler(nil, mockUC)
		err := handler.Verify(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp dto.LoginResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "access-token-123", resp.AccessToken)
		assert.Equal(t, "refresh-token-123", resp.RefreshToken)
		mockUC.AssertExpectations(t)
	})

	t.Run("bind error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/biometric/verify", bytes.NewReader([]byte("{invalid-json")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := NewAuthBiometricHandler(nil, nil)
		err := handler.Verify(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("validation error", func(t *testing.T) {
		e := echo.New()
		reqBody, _ := json.Marshal(dto.BiometricVerifyRequest{
			Email: "admin@hros.com",
			// credential ID missing
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/biometric/verify", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := NewAuthBiometricHandler(nil, nil)
		err := handler.Verify(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid client_data_json base64", func(t *testing.T) {
		e := echo.New()
		reqBody, _ := json.Marshal(dto.BiometricVerifyRequest{
			Email:             "admin@hros.com",
			CredentialID:      "cred-id",
			ClientDataJSON:    "invalid base64 @!!",
			AuthenticatorData: authDataB64,
			Signature:         sigB64,
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/biometric/verify", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := NewAuthBiometricHandler(nil, nil)
		err := handler.Verify(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid authenticator_data base64", func(t *testing.T) {
		e := echo.New()
		reqBody, _ := json.Marshal(dto.BiometricVerifyRequest{
			Email:             "admin@hros.com",
			CredentialID:      "cred-id",
			ClientDataJSON:    clientDataB64,
			AuthenticatorData: "invalid base64 @!!",
			Signature:         sigB64,
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/biometric/verify", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := NewAuthBiometricHandler(nil, nil)
		err := handler.Verify(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid signature base64", func(t *testing.T) {
		e := echo.New()
		reqBody, _ := json.Marshal(dto.BiometricVerifyRequest{
			Email:             "admin@hros.com",
			CredentialID:      "cred-id",
			ClientDataJSON:    clientDataB64,
			AuthenticatorData: authDataB64,
			Signature:         "invalid base64 @!!",
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/biometric/verify", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := NewAuthBiometricHandler(nil, nil)
		err := handler.Verify(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("verification error status code mappings (unauthorized)", func(t *testing.T) {
		testErrors := []error{
			domainErrors.ErrBiometricNotRegistered,
			domainErrors.ErrInvalidBiometricSignature,
			domainErrors.ErrChallengeNotFoundOrExpired,
			domainErrors.ErrUserInactive,
			domainErrors.ErrUserLocked,
		}

		for _, testErr := range testErrors {
			t.Run(testErr.Error(), func(t *testing.T) {
				e := echo.New()
				reqBody, _ := json.Marshal(dto.BiometricVerifyRequest{
					Email:             "admin@hros.com",
					CredentialID:      "cred-id",
					ClientDataJSON:    clientDataB64,
					AuthenticatorData: authDataB64,
					Signature:         sigB64,
				})
				req := httptest.NewRequest(http.MethodPost, "/v1/auth/biometric/verify", bytes.NewReader(reqBody))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				mockUC := new(mockVerifyBiometricExecutor)
				mockUC.On("Execute", mock.Anything, mock.Anything).Return((*usecase.LoginOutput)(nil), testErr).Once()

				handler := NewAuthBiometricHandler(nil, mockUC)
				err := handler.Verify(c)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusUnauthorized, rec.Code)

				var resp sharedErrors.ErrorResponse
				err = json.Unmarshal(rec.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "unauthorized", resp.Code)
				mockUC.AssertExpectations(t)
			})
		}
	})

	t.Run("internal server error during verification", func(t *testing.T) {
		e := echo.New()
		reqBody, _ := json.Marshal(dto.BiometricVerifyRequest{
			Email:             "admin@hros.com",
			CredentialID:      "cred-id",
			ClientDataJSON:    clientDataB64,
			AuthenticatorData: authDataB64,
			Signature:         sigB64,
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/biometric/verify", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUC := new(mockVerifyBiometricExecutor)
		mockUC.On("Execute", mock.Anything, mock.Anything).Return((*usecase.LoginOutput)(nil), errors.New("unexpected context timeout")).Once()

		handler := NewAuthBiometricHandler(nil, mockUC)
		err := handler.Verify(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		mockUC.AssertExpectations(t)
	})
}
