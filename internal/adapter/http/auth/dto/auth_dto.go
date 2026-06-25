// Package dto defines the data transfer objects for the authentication adapter.
package dto

// LoginRequest represents the payload for admin login.
type LoginRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	RememberMe bool   `json:"remember_me"`
}

// RefreshRequest represents the payload for rotating the access token.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LoginResponse represents the successful login response containing tokens.
type LoginResponse struct {
	AccessToken  string   `json:"access_token,omitempty"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	MFARequired  bool     `json:"mfa_required,omitempty"`
	MFAToken     string   `json:"mfa_token,omitempty"`
	MFAMethods   []string `json:"mfa_methods,omitempty"`
}

// MFAVerifyRequest represents the payload for verifying the MFA code.
type MFAVerifyRequest struct {
	MFAToken   string `json:"mfa_token" validate:"required"`
	Method     string `json:"method" validate:"required,oneof=totp webauthn"`
	Code       string `json:"code" validate:"required_if=Method totp"`
	RememberMe bool   `json:"remember_me"`
}

// PasswordResetRequest represents the payload for requesting a password reset.
type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// PasswordResetConfirmRequest represents the payload for confirming a password reset.
type PasswordResetConfirmRequest struct {
	Token                string `json:"token" validate:"required"`
	Password             string `json:"password" validate:"required"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password"`
}
