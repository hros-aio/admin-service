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
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
