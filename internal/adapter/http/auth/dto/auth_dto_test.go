package dto

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestLoginRequest_Validation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		request LoginRequest
		isValid bool
	}{
		{
			name: "Valid request",
			request: LoginRequest{
				Email:    "admin@hros.com",
				Password: "password123",
			},
			isValid: true,
		},
		{
			name: "Missing email",
			request: LoginRequest{
				Password: "password123",
			},
			isValid: false,
		},
		{
			name: "Invalid email format",
			request: LoginRequest{
				Email:    "not-an-email",
				Password: "password123",
			},
			isValid: false,
		},
		{
			name: "Missing password",
			request: LoginRequest{
				Email: "admin@hros.com",
			},
			isValid: false,
		},
		{
			name: "Valid request with remember_me true",
			request: LoginRequest{
				Email:      "admin@hros.com",
				Password:   "password123",
				RememberMe: true,
			},
			isValid: true,
		},
		{
			name: "Valid request with remember_me false",
			request: LoginRequest{
				Email:      "admin@hros.com",
				Password:   "password123",
				RememberMe: false,
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.request)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestRefreshRequest_Validation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		request RefreshRequest
		isValid bool
	}{
		{
			name: "Valid request",
			request: RefreshRequest{
				RefreshToken: "def456_valid_token",
			},
			isValid: true,
		},
		{
			name:    "Missing refresh token",
			request: RefreshRequest{},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.request)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestMFAVerifyRequest_Validation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		request MFAVerifyRequest
		isValid bool
	}{
		{
			name: "Valid TOTP request",
			request: MFAVerifyRequest{
				MFAToken: "mfa_token_123",
				Method:   "totp",
				Code:     "123456",
			},
			isValid: true,
		},
		{
			name: "Valid WebAuthn request (code optional)",
			request: MFAVerifyRequest{
				MFAToken: "mfa_token_123",
				Method:   "webauthn",
				Code:     "",
			},
			isValid: true,
		},
		{
			name: "Missing token",
			request: MFAVerifyRequest{
				Method: "totp",
				Code:   "123456",
			},
			isValid: false,
		},
		{
			name: "Missing method",
			request: MFAVerifyRequest{
				MFAToken: "mfa_token_123",
				Code:     "123456",
			},
			isValid: false,
		},
		{
			name: "Invalid method",
			request: MFAVerifyRequest{
				MFAToken: "mfa_token_123",
				Method:   "sms",
				Code:     "123456",
			},
			isValid: false,
		},
		{
			name: "Missing code for TOTP",
			request: MFAVerifyRequest{
				MFAToken: "mfa_token_123",
				Method:   "totp",
				Code:     "",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.request)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
