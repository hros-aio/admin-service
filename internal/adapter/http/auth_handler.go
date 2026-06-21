// Package http provides HTTP adapter handlers for Echo routing.
package http

import (
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

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	loginUC   *usecase.LoginUseCase
	logoutUC  *usecase.LogoutUseCase
	refreshUC *usecase.RefreshSessionUseCase
	validate  *validator.Validate
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	loginUC *usecase.LoginUseCase,
	logoutUC *usecase.LogoutUseCase,
	refreshUC *usecase.RefreshSessionUseCase,
) *AuthHandler {
	return &AuthHandler{
		loginUC:   loginUC,
		logoutUC:  logoutUC,
		refreshUC: refreshUC,
		validate:  validator.New(),
	}
}

// RegisterRoutes registers the authentication routes.
func RegisterRoutes(e *echo.Echo, h *AuthHandler) {
	e.POST("/v1/auth/login", h.Login)
	e.DELETE("/v1/auth/session", h.Logout)
	e.POST("/v1/auth/refresh", h.Refresh)
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

		resp := sharedErrors.NewErrorResponse("internal_error", "Internal server error", nil, traceID)
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp := dto.LoginResponse{
		AccessToken:  output.AccessToken,
		RefreshToken: output.RefreshToken,
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

		resp := sharedErrors.NewErrorResponse("internal_error", "Internal server error", nil, traceID)
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp := dto.LoginResponse{
		AccessToken:  output.AccessToken,
		RefreshToken: output.RefreshToken,
	}
	return c.JSON(http.StatusOK, resp)
}
