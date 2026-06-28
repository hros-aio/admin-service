// Package http provides HTTP adapter handlers for Echo routing.
package http

import (
	"context"
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

// AcceptInviteExecutor is the handler-layer interface satisfied by *usecase.AcceptInviteUseCase.
// Declaring it here keeps the handler decoupled from the concrete use-case type and
// allows straightforward mock injection in unit tests.
type AcceptInviteExecutor interface {
	Execute(ctx context.Context, input usecase.AcceptInviteInput) error
}

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	loginUC                *usecase.LoginUseCase
	logoutUC               *usecase.LogoutUseCase
	refreshUC              *usecase.RefreshSessionUseCase
	verifyMfaUC            *usecase.VerifyMFAUseCase
	requestPasswordResetUC *usecase.RequestPasswordResetUseCase
	confirmPasswordResetUC *usecase.ConfirmPasswordResetUseCase
	acceptInviteUC         AcceptInviteExecutor
	validate               *validator.Validate
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	loginUC *usecase.LoginUseCase,
	logoutUC *usecase.LogoutUseCase,
	refreshUC *usecase.RefreshSessionUseCase,
	verifyMfaUC *usecase.VerifyMFAUseCase,
	requestPasswordResetUC *usecase.RequestPasswordResetUseCase,
	confirmPasswordResetUC *usecase.ConfirmPasswordResetUseCase,
	acceptInviteUC AcceptInviteExecutor,
) *AuthHandler {
	v := validator.New()
	_ = v.RegisterValidation("notblank", func(fl validator.FieldLevel) bool {
		return len(strings.TrimSpace(fl.Field().String())) > 0
	})
	return &AuthHandler{
		loginUC:                loginUC,
		logoutUC:               logoutUC,
		refreshUC:              refreshUC,
		verifyMfaUC:            verifyMfaUC,
		requestPasswordResetUC: requestPasswordResetUC,
		confirmPasswordResetUC: confirmPasswordResetUC,
		acceptInviteUC:         acceptInviteUC,
		validate:               v,
	}
}

// RegisterRoutes registers the authentication routes.
func RegisterRoutes(e *echo.Echo, h *AuthHandler, sso *AuthSSOHandler) {
	e.POST("/v1/auth/login", h.Login)
	e.DELETE("/v1/auth/session", h.Logout)
	e.POST("/v1/auth/refresh", h.Refresh)
	e.POST("/v1/auth/mfa/verify", h.VerifyMFA)
	e.POST("/v1/auth/password-reset/request", h.RequestPasswordReset)
	e.POST("/v1/auth/password-reset/confirm", h.ConfirmPasswordReset)
	e.POST("/v1/auth/accept-invite", h.AcceptInvite)

	// SSO routes
	e.GET("/v1/auth/sso/initiate", sso.Initiate)
	e.GET("/v1/auth/sso/callback", sso.Callback)
}

// Login handles the admin login request.
func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
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

	input := usecase.LoginInput{
		Email:      req.Email,
		Password:   req.Password,
		RememberMe: req.RememberMe,
		IPAddress:  c.RealIP(),
		UserAgent:  c.Request().UserAgent(),
	}

	output, err := h.loginUC.Execute(c.Request().Context(), input)
	if err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)

		if errors.Is(err, domainErrors.ErrInvalidCredentials) {
			resp := sharedErrors.NewErrorResponse("unauthorized", "Invalid email or password", nil, traceID)
			return c.JSON(http.StatusUnauthorized, resp)
		}
		if errors.Is(err, domainErrors.ErrUserInactive) {
			resp := sharedErrors.NewErrorResponse("forbidden", "Account is deactivated", nil, traceID)
			return c.JSON(http.StatusForbidden, resp)
		}
		if errors.Is(err, domainErrors.ErrUserLocked) {
			resp := sharedErrors.NewErrorResponse("forbidden", "Account is locked", nil, traceID)
			return c.JSON(http.StatusForbidden, resp)
		}
		if errors.Is(err, domainErrors.ErrAccountLocked) {
			resp := sharedErrors.NewErrorResponse("ACCOUNT_LOCKED", "Account is temporarily locked", nil, traceID)
			return c.JSON(http.StatusUnauthorized, resp)
		}

		resp := sharedErrors.NewErrorResponse("internal_error", "Internal server error", nil, traceID)
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp := dto.LoginResponse{
		AccessToken:  output.AccessToken,
		RefreshToken: output.RefreshToken,
		MFARequired:  output.MFARequired,
		MFAToken:     output.MFAToken,
		MFAMethods:   output.MFAMethods,
	}
	return c.JSON(http.StatusOK, resp)
}

// Logout handles the admin logout request.
func (h *AuthHandler) Logout(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		resp := sharedErrors.NewErrorResponse("unauthorized", "Missing or invalid authorization header", nil, traceID)
		return c.JSON(http.StatusUnauthorized, resp)
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		resp := sharedErrors.NewErrorResponse("unauthorized", "Token is empty", nil, traceID)
		return c.JSON(http.StatusUnauthorized, resp)
	}

	var refreshToken string
	var accessToken string

	if _, hasHeader := c.Request().Header["X-Refresh-Token"]; hasHeader {
		xRefreshToken := strings.TrimSpace(c.Request().Header.Get("X-Refresh-Token"))
		if xRefreshToken == "" {
			traceID := c.Response().Header().Get(echo.HeaderXRequestID)
			resp := sharedErrors.NewErrorResponse("bad_request", "X-Refresh-Token header is empty or malformed", nil, traceID)
			return c.JSON(http.StatusBadRequest, resp)
		}
		refreshToken = xRefreshToken
		accessToken = token
	} else {
		refreshToken = token
	}

	input := usecase.LogoutInput{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}

	err := h.logoutUC.Execute(c.Request().Context(), input)
	if err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		resp := sharedErrors.NewErrorResponse("internal_error", "Internal server error", nil, traceID)
		return c.JSON(http.StatusInternalServerError, resp)
	}

	return c.NoContent(http.StatusNoContent)
}

// Refresh handles the admin session token refresh request.
func (h *AuthHandler) Refresh(c echo.Context) error {
	var req dto.RefreshRequest
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

	input := usecase.RefreshInput{
		RefreshToken: req.RefreshToken,
	}

	output, err := h.refreshUC.Execute(c.Request().Context(), input)
	if err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)

		if errors.Is(err, domainErrors.ErrInvalidRefreshToken) || errors.Is(err, domainErrors.ErrTokenExpired) {
			resp := sharedErrors.NewErrorResponse("unauthorized", "Invalid or expired refresh token", nil, traceID)
			return c.JSON(http.StatusUnauthorized, resp)
		}
		if errors.Is(err, domainErrors.ErrUserInactive) {
			resp := sharedErrors.NewErrorResponse("forbidden", "Account is deactivated", nil, traceID)
			return c.JSON(http.StatusForbidden, resp)
		}
		if errors.Is(err, domainErrors.ErrUserLocked) {
			resp := sharedErrors.NewErrorResponse("forbidden", "Account is locked", nil, traceID)
			return c.JSON(http.StatusForbidden, resp)
		}
		if errors.Is(err, domainErrors.ErrAccountLocked) {
			resp := sharedErrors.NewErrorResponse("ACCOUNT_LOCKED", "Account is temporarily locked", nil, traceID)
			return c.JSON(http.StatusUnauthorized, resp)
		}

		resp := sharedErrors.NewErrorResponse("internal_error", "Internal server error", nil, traceID)
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp := dto.LoginResponse{
		AccessToken:  output.AccessToken,
		RefreshToken: output.RefreshToken,
	}
	return c.JSON(http.StatusOK, resp)
}

// VerifyMFA handles the MFA verification request.
func (h *AuthHandler) VerifyMFA(c echo.Context) error {
	var req dto.MFAVerifyRequest
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

	input := usecase.VerifyMFAInput{
		MFAToken:   req.MFAToken,
		Method:     req.Method,
		Code:       req.Code,
		RememberMe: req.RememberMe,
		IPAddress:  c.RealIP(),
		UserAgent:  c.Request().UserAgent(),
	}

	output, err := h.verifyMfaUC.Execute(c.Request().Context(), input)
	if err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)

		if errors.Is(err, domainErrors.ErrMFAInvalid) {
			resp := sharedErrors.NewErrorResponse("MFA_INVALID", "Invalid MFA verification code", nil, traceID)
			return c.JSON(http.StatusUnauthorized, resp)
		}
		if errors.Is(err, domainErrors.ErrMFATokenExpired) {
			resp := sharedErrors.NewErrorResponse("MFA_TOKEN_EXPIRED", "MFA token has expired", nil, traceID)
			return c.JSON(http.StatusUnauthorized, resp)
		}
		if errors.Is(err, domainErrors.ErrUserInactive) {
			resp := sharedErrors.NewErrorResponse("forbidden", "Account is deactivated", nil, traceID)
			return c.JSON(http.StatusForbidden, resp)
		}
		if errors.Is(err, domainErrors.ErrUserLocked) {
			resp := sharedErrors.NewErrorResponse("forbidden", "Account is locked", nil, traceID)
			return c.JSON(http.StatusForbidden, resp)
		}

		resp := sharedErrors.NewErrorResponse("internal_error", "Internal server error", nil, traceID)
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp := dto.LoginResponse{
		AccessToken:  output.AccessToken,
		RefreshToken: output.RefreshToken,
	}
	return c.JSON(http.StatusOK, resp)
}

// RequestPasswordReset handles the request to trigger a password reset.
func (h *AuthHandler) RequestPasswordReset(c echo.Context) error {
	var req dto.PasswordResetRequest
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

	input := usecase.RequestPasswordResetInput{
		Email:     req.Email,
		IPAddress: c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	}

	if err := h.requestPasswordResetUC.Execute(c.Request().Context(), input); err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		resp := sharedErrors.NewErrorResponse("internal_error", "Internal server error", nil, traceID)
		return c.JSON(http.StatusInternalServerError, resp)
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "If an account exists for that email, a reset link has been sent."})
}

// ConfirmPasswordReset handles the request to confirm the password reset.
func (h *AuthHandler) ConfirmPasswordReset(c echo.Context) error {
	var req dto.PasswordResetConfirmRequest
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

	input := usecase.ConfirmPasswordResetInput{
		Token:     req.Token,
		Password:  req.Password,
		IPAddress: c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	}

	if err := h.confirmPasswordResetUC.Execute(c.Request().Context(), input); err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)

		if errors.Is(err, domainErrors.ErrTokenExpired) {
			resp := sharedErrors.NewErrorResponse("TOKEN_EXPIRED", "Reset token has expired", nil, traceID)
			return c.JSON(http.StatusBadRequest, resp)
		}
		if errors.Is(err, domainErrors.ErrTokenUsed) {
			resp := sharedErrors.NewErrorResponse("TOKEN_USED", "Reset token has already been used", nil, traceID)
			return c.JSON(http.StatusBadRequest, resp)
		}
		if errors.Is(err, domainErrors.ErrPasswordWeak) {
			resp := sharedErrors.NewErrorResponse("PASSWORD_WEAK", "Password does not meet complexity requirements", nil, traceID)
			return c.JSON(http.StatusUnprocessableEntity, resp)
		}

		resp := sharedErrors.NewErrorResponse("internal_error", "Internal server error", nil, traceID)
		return c.JSON(http.StatusInternalServerError, resp)
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Password updated successfully."})
}

// AcceptInvite handles the request to accept an administrator invitation.
func (h *AuthHandler) AcceptInvite(c echo.Context) error {
	var req dto.AcceptInviteRequest
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

	input := usecase.AcceptInviteInput{
		Token:     req.Token,
		Password:  req.Password,
		IPAddress: c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	}

	if h.acceptInviteUC == nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		resp := sharedErrors.NewErrorResponse("internal_error", "Internal server error", nil, traceID)
		return c.JSON(http.StatusInternalServerError, resp)
	}

	if err := h.acceptInviteUC.Execute(c.Request().Context(), input); err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)

		if errors.Is(err, domainErrors.ErrInviteExpired) {
			resp := sharedErrors.NewErrorResponse("INVITE_EXPIRED", "Invite token has expired", nil, traceID)
			return c.JSON(http.StatusBadRequest, resp)
		}
		if errors.Is(err, domainErrors.ErrInviteUsed) {
			resp := sharedErrors.NewErrorResponse("INVITE_USED", "Invite token has already been used", nil, traceID)
			return c.JSON(http.StatusBadRequest, resp)
		}
		if errors.Is(err, domainErrors.ErrPasswordWeak) {
			resp := sharedErrors.NewErrorResponse("PASSWORD_WEAK", "Password does not meet complexity requirements", nil, traceID)
			return c.JSON(http.StatusUnprocessableEntity, resp)
		}

		resp := sharedErrors.NewErrorResponse("internal_error", "Internal server error", nil, traceID)
		return c.JSON(http.StatusInternalServerError, resp)
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Account activated successfully."})
}
