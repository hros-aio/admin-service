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

// AuthSSOHandler handles single sign-on HTTP requests.
type AuthSSOHandler struct {
	initiateUC *usecase.InitiateSSOUseCase
	callbackUC *usecase.CallbackSSOUseCase
	validate   *validator.Validate
}

// NewAuthSSOHandler creates a new AuthSSOHandler.
func NewAuthSSOHandler(
	initiateUC *usecase.InitiateSSOUseCase,
	callbackUC *usecase.CallbackSSOUseCase,
) *AuthSSOHandler {
	v := validator.New()
	_ = v.RegisterValidation("notblank", func(fl validator.FieldLevel) bool {
		return len(strings.TrimSpace(fl.Field().String())) > 0
	})
	return &AuthSSOHandler{
		initiateUC: initiateUC,
		callbackUC: callbackUC,
		validate:   v,
	}
}

// Initiate handles GET /auth/sso/initiate.
func (h *AuthSSOHandler) Initiate(c echo.Context) error {
	var req dto.SSOInitiateRequest
	if err := c.Bind(&req); err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		resp := sharedErrors.NewErrorResponse("bad_request", "Invalid request query parameters", err.Error(), traceID)
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := h.validate.Struct(&req); err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		resp := sharedErrors.NewErrorResponse("validation_error", "Request validation failed", err.Error(), traceID)
		return c.JSON(http.StatusBadRequest, resp)
	}

	input := usecase.InitiateSSOInput{
		Provider: req.Provider,
	}

	output, err := h.initiateUC.Execute(c.Request().Context(), input)
	if err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		if strings.Contains(err.Error(), "provider cannot be empty") || strings.Contains(err.Error(), "unsupported or unconfigured provider") {
			resp := sharedErrors.NewErrorResponse("bad_request", err.Error(), nil, traceID)
			return c.JSON(http.StatusBadRequest, resp)
		}
		resp := sharedErrors.NewErrorResponse("internal_error", "Internal server error", nil, traceID)
		return c.JSON(http.StatusInternalServerError, resp)
	}

	return c.Redirect(http.StatusFound, output.RedirectURL)
}

// Callback handles GET /auth/sso/callback.
func (h *AuthSSOHandler) Callback(c echo.Context) error {
	var req dto.SSOCallbackRequest
	if err := c.Bind(&req); err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		resp := sharedErrors.NewErrorResponse("bad_request", "Invalid request query parameters", err.Error(), traceID)
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := h.validate.Struct(&req); err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)
		resp := sharedErrors.NewErrorResponse("validation_error", "Request validation failed", err.Error(), traceID)
		return c.JSON(http.StatusBadRequest, resp)
	}

	// Resolve provider from query param if available, otherwise default to "google"
	provider := c.QueryParam("provider")
	if provider == "" {
		provider = "google"
	}

	input := usecase.CallbackSSOInput{
		Provider:  provider,
		Code:      req.Code,
		State:     req.State,
		IPAddress: c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	}

	output, err := h.callbackUC.Execute(c.Request().Context(), input)
	if err != nil {
		traceID := c.Response().Header().Get(echo.HeaderXRequestID)

		if errors.Is(err, domainErrors.ErrInvalidSSOState) {
			resp := sharedErrors.NewErrorResponse("INVALID_SSO_STATE", "Invalid or expired SSO state parameter", nil, traceID)
			return c.JSON(http.StatusBadRequest, resp)
		}
		if errors.Is(err, domainErrors.ErrNoAccountLinked) {
			resp := sharedErrors.NewErrorResponse("NO_ACCOUNT_LINKED", "No admin account linked to this identity", nil, traceID)
			return c.JSON(http.StatusUnauthorized, resp)
		}
		if errors.Is(err, domainErrors.ErrIdentityConflict) {
			resp := sharedErrors.NewErrorResponse("identity_conflict", err.Error(), nil, traceID)
			return c.JSON(http.StatusConflict, resp)
		}

		resp := sharedErrors.NewErrorResponse("internal_error", "Internal server error", nil, traceID)
		return c.JSON(http.StatusInternalServerError, resp)
	}

	// Set HTTP-only refresh cookie
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    output.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	// If Accept header prefers HTML/text or redirect param is set, redirect to dashboard.
	// Otherwise, return JSON per spec.
	acceptHeader := c.Request().Header.Get("Accept")
	if strings.Contains(acceptHeader, "text/html") {
		return c.Redirect(http.StatusFound, "/dashboard")
	}

	resp := dto.LoginResponse{
		AccessToken:  output.AccessToken,
		RefreshToken: "", // Keep refresh token exclusively in the cookie
	}
	return c.JSON(http.StatusOK, resp)
}
