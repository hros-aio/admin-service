// Package http provides HTTP adapter handlers for Echo routing.
package http

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/hros/admin-service/internal/adapter/http/auth/dto"
	"github.com/hros/admin-service/internal/application/usecase"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	sharedErrors "github.com/hros/admin-service/internal/shared/errors"
	"github.com/labstack/echo/v4"
)

// GenerateChallengeExecutor defines the handler-layer contract for biometric challenge generation.
type GenerateChallengeExecutor interface {
	Execute(ctx context.Context, input usecase.GenerateBiometricChallengeInput) (*usecase.GenerateBiometricChallengeOutput, error)
}

// VerifyBiometricExecutor defines the handler-layer contract for biometric signature verification.
type VerifyBiometricExecutor interface {
	Execute(ctx context.Context, input usecase.VerifyBiometricInput) (*usecase.LoginOutput, error)
}

// AuthBiometricHandler handles biometric-related HTTP requests.
type AuthBiometricHandler struct {
	generateChallengeUC GenerateChallengeExecutor
	verifyBiometricUC   VerifyBiometricExecutor
	validate            *validator.Validate
}

// NewAuthBiometricHandler creates a new AuthBiometricHandler.
func NewAuthBiometricHandler(
	generateChallengeUC GenerateChallengeExecutor,
	verifyBiometricUC VerifyBiometricExecutor,
) *AuthBiometricHandler {
	v := validator.New()
	_ = v.RegisterValidation("notblank", func(fl validator.FieldLevel) bool {
		return len(strings.TrimSpace(fl.Field().String())) > 0
	})
	return &AuthBiometricHandler{
		generateChallengeUC: generateChallengeUC,
		verifyBiometricUC:   verifyBiometricUC,
		validate:            v,
	}
}

// Challenge handles POST /v1/auth/biometric/challenge.
func (h *AuthBiometricHandler) Challenge(c echo.Context) error {
	var req dto.BiometricChallengeRequest
	if err := c.Bind(&req); err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		resp := sharedErrors.NewErrorResponse("bad_request", "Invalid request body", err.Error(), traceID)
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := h.validate.Struct(&req); err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		resp := sharedErrors.NewErrorResponse("validation_error", "Request validation failed", err.Error(), traceID)
		return c.JSON(http.StatusBadRequest, resp)
	}

	input := usecase.GenerateBiometricChallengeInput{
		Email: req.Email,
	}

	out, err := h.generateChallengeUC.Execute(c.Request().Context(), input)
	if err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		if errors.Is(err, domainErrors.ErrBiometricNotRegistered) {
			resp := sharedErrors.NewErrorResponse("unauthorized", "Biometric authentication not registered", err.Error(), traceID)
			return c.JSON(http.StatusUnauthorized, resp)
		}
		resp := sharedErrors.NewErrorResponse("internal_error", "Failed to generate biometric challenge", err.Error(), traceID)
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp := dto.BiometricChallengeResponse{
		Challenge:    out.Challenge,
		CredentialID: out.CredentialID,
	}
	return c.JSON(http.StatusOK, resp)
}

// Verify handles POST /v1/auth/biometric/verify.
func (h *AuthBiometricHandler) Verify(c echo.Context) error {
	var req dto.BiometricVerifyRequest
	if err := c.Bind(&req); err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		resp := sharedErrors.NewErrorResponse("bad_request", "Invalid request body", err.Error(), traceID)
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := h.validate.Struct(&req); err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		resp := sharedErrors.NewErrorResponse("validation_error", "Request validation failed", err.Error(), traceID)
		return c.JSON(http.StatusBadRequest, resp)
	}

	// Helper to decode standard and url-safe base64 formats
	decodeBase64 := func(in string) ([]byte, error) {
		trimmed := strings.TrimSpace(in)
		if dec, err := base64.RawURLEncoding.DecodeString(trimmed); err == nil {
			return dec, nil
		}
		if dec, err := base64.URLEncoding.DecodeString(trimmed); err == nil {
			return dec, nil
		}
		if dec, err := base64.RawStdEncoding.DecodeString(trimmed); err == nil {
			return dec, nil
		}
		return base64.StdEncoding.DecodeString(trimmed)
	}

	clientDataJSON, err := decodeBase64(req.ClientDataJSON)
	if err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		resp := sharedErrors.NewErrorResponse("bad_request", "Invalid base64 in client_data_json", err.Error(), traceID)
		return c.JSON(http.StatusBadRequest, resp)
	}

	authenticatorData, err := decodeBase64(req.AuthenticatorData)
	if err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		resp := sharedErrors.NewErrorResponse("bad_request", "Invalid base64 in authenticator_data", err.Error(), traceID)
		return c.JSON(http.StatusBadRequest, resp)
	}

	signature, err := decodeBase64(req.Signature)
	if err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		resp := sharedErrors.NewErrorResponse("bad_request", "Invalid base64 in signature", err.Error(), traceID)
		return c.JSON(http.StatusBadRequest, resp)
	}

	input := usecase.VerifyBiometricInput{
		Email:             req.Email,
		CredentialID:      req.CredentialID,
		ClientDataJSON:    clientDataJSON,
		AuthenticatorData: authenticatorData,
		Signature:         signature,
		RememberMe:        req.RememberMe,
		IPAddress:         c.RealIP(),
		UserAgent:         c.Request().UserAgent(),
	}

	out, err := h.verifyBiometricUC.Execute(c.Request().Context(), input)
	if err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		if errors.Is(err, domainErrors.ErrBiometricNotRegistered) ||
			errors.Is(err, domainErrors.ErrInvalidBiometricSignature) ||
			errors.Is(err, domainErrors.ErrChallengeNotFoundOrExpired) ||
			errors.Is(err, domainErrors.ErrUserInactive) ||
			errors.Is(err, domainErrors.ErrUserLocked) {
			resp := sharedErrors.NewErrorResponse("unauthorized", "Biometric verification failed", err.Error(), traceID)
			return c.JSON(http.StatusUnauthorized, resp)
		}
		resp := sharedErrors.NewErrorResponse("internal_error", "Verification failed", err.Error(), traceID)
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp := dto.LoginResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}
	return c.JSON(http.StatusOK, resp)
}
