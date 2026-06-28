package dto

import (
	"encoding/json"
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

func TestPasswordResetRequest_Validation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		request PasswordResetRequest
		isValid bool
	}{
		{
			name: "Valid request",
			request: PasswordResetRequest{
				Email: "admin@hros.com",
			},
			isValid: true,
		},
		{
			name:    "Missing email",
			request: PasswordResetRequest{},
			isValid: false,
		},
		{
			name: "Invalid email format",
			request: PasswordResetRequest{
				Email: "invalid-email",
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

func TestPasswordResetConfirmRequest_Validation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		request PasswordResetConfirmRequest
		isValid bool
	}{
		{
			name: "Valid request",
			request: PasswordResetConfirmRequest{
				Token:                "reset-token-123",
				Password:             "SecurePass1!",
				PasswordConfirmation: "SecurePass1!",
			},
			isValid: true,
		},
		{
			name: "Missing token",
			request: PasswordResetConfirmRequest{
				Password:             "SecurePass1!",
				PasswordConfirmation: "SecurePass1!",
			},
			isValid: false,
		},
		{
			name: "Missing password",
			request: PasswordResetConfirmRequest{
				Token:                "reset-token-123",
				PasswordConfirmation: "SecurePass1!",
			},
			isValid: false,
		},
		{
			name: "Missing password confirmation",
			request: PasswordResetConfirmRequest{
				Token:    "reset-token-123",
				Password: "SecurePass1!",
			},
			isValid: false,
		},
		{
			name: "Mismatched passwords",
			request: PasswordResetConfirmRequest{
				Token:                "reset-token-123",
				Password:             "SecurePass1!",
				PasswordConfirmation: "DifferentPass1!",
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

func TestPasswordResetDTOs_JSONMapping(t *testing.T) {
	t.Run("PasswordResetRequest JSON mapping", func(t *testing.T) {
		req := PasswordResetRequest{
			Email: "admin@hros.com",
		}
		data, err := json.Marshal(req)
		assert.NoError(t, err)
		assert.Contains(t, string(data), `"email":"admin@hros.com"`)

		var unmarshaled PasswordResetRequest
		err = json.Unmarshal(data, &unmarshaled)
		assert.NoError(t, err)
		assert.Equal(t, req.Email, unmarshaled.Email)
	})

	t.Run("PasswordResetConfirmRequest JSON mapping", func(t *testing.T) {
		req := PasswordResetConfirmRequest{
			Token:                "token_123",
			Password:             "SecurePass1!",
			PasswordConfirmation: "SecurePass1!",
		}
		data, err := json.Marshal(req)
		assert.NoError(t, err)
		assert.Contains(t, string(data), `"token":"token_123"`)
		assert.Contains(t, string(data), `"password":"SecurePass1!"`)
		assert.Contains(t, string(data), `"password_confirmation":"SecurePass1!"`)

		var unmarshaled PasswordResetConfirmRequest
		err = json.Unmarshal(data, &unmarshaled)
		assert.NoError(t, err)
		assert.Equal(t, req.Token, unmarshaled.Token)
		assert.Equal(t, req.Password, unmarshaled.Password)
		assert.Equal(t, req.PasswordConfirmation, unmarshaled.PasswordConfirmation)
	})
}

func TestAcceptInviteRequest_Validation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		request AcceptInviteRequest
		isValid bool
	}{
		{
			name: "Valid request",
			request: AcceptInviteRequest{
				Token:                "invite-token-123",
				Password:             "SecurePass1!",
				PasswordConfirmation: "SecurePass1!",
			},
			isValid: true,
		},
		{
			name: "Missing token",
			request: AcceptInviteRequest{
				Password:             "SecurePass1!",
				PasswordConfirmation: "SecurePass1!",
			},
			isValid: false,
		},
		{
			name: "Missing password",
			request: AcceptInviteRequest{
				Token:                "invite-token-123",
				PasswordConfirmation: "SecurePass1!",
			},
			isValid: false,
		},
		{
			name: "Missing password confirmation",
			request: AcceptInviteRequest{
				Token:    "invite-token-123",
				Password: "SecurePass1!",
			},
			isValid: false,
		},
		{
			name: "Mismatched passwords",
			request: AcceptInviteRequest{
				Token:                "invite-token-123",
				Password:             "SecurePass1!",
				PasswordConfirmation: "DifferentPass1!",
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

func TestAcceptInviteDTO_JSONMapping(t *testing.T) {
	req := AcceptInviteRequest{
		Token:                "token_123",
		Password:             "SecurePass1!",
		PasswordConfirmation: "SecurePass1!",
	}
	data, err := json.Marshal(req)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"token":"token_123"`)
	assert.Contains(t, string(data), `"password":"SecurePass1!"`)
	assert.Contains(t, string(data), `"password_confirmation":"SecurePass1!"`)

	var unmarshaled AcceptInviteRequest
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, req.Token, unmarshaled.Token)
	assert.Equal(t, req.Password, unmarshaled.Password)
	assert.Equal(t, req.PasswordConfirmation, unmarshaled.PasswordConfirmation)
}

func TestSSOInitiateRequest_Validation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		request SSOInitiateRequest
		isValid bool
	}{
		{
			name: "Valid request",
			request: SSOInitiateRequest{
				Provider: "google",
			},
			isValid: true,
		},
		{
			name:    "Missing provider",
			request: SSOInitiateRequest{},
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

func TestSSOCallbackRequest_Validation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		request SSOCallbackRequest
		isValid bool
	}{
		{
			name: "Valid request",
			request: SSOCallbackRequest{
				Code:  "auth_code_123",
				State: "state_abc",
			},
			isValid: true,
		},
		{
			name: "Missing code",
			request: SSOCallbackRequest{
				State: "state_abc",
			},
			isValid: false,
		},
		{
			name: "Missing state",
			request: SSOCallbackRequest{
				Code: "auth_code_123",
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

func TestBiometricChallengeRequest_Validation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		request BiometricChallengeRequest
		isValid bool
	}{
		{
			name: "Valid request",
			request: BiometricChallengeRequest{
				Email: "admin@hros.com",
			},
			isValid: true,
		},
		{
			name:    "Missing email",
			request: BiometricChallengeRequest{},
			isValid: false,
		},
		{
			name: "Invalid email format",
			request: BiometricChallengeRequest{
				Email: "invalid-email",
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

func TestBiometricVerifyRequest_Validation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		request BiometricVerifyRequest
		isValid bool
	}{
		{
			name: "Valid request",
			request: BiometricVerifyRequest{
				Email:             "admin@hros.com",
				CredentialID:      "cred_123",
				AuthenticatorData: "auth_data",
				ClientDataJSON:    "client_data",
				Signature:         "sig_123",
			},
			isValid: true,
		},
		{
			name: "Missing email",
			request: BiometricVerifyRequest{
				CredentialID:      "cred_123",
				AuthenticatorData: "auth_data",
				ClientDataJSON:    "client_data",
				Signature:         "sig_123",
			},
			isValid: false,
		},
		{
			name: "Invalid email",
			request: BiometricVerifyRequest{
				Email:             "not-an-email",
				CredentialID:      "cred_123",
				AuthenticatorData: "auth_data",
				ClientDataJSON:    "client_data",
				Signature:         "sig_123",
			},
			isValid: false,
		},
		{
			name: "Missing credential ID",
			request: BiometricVerifyRequest{
				Email:             "admin@hros.com",
				AuthenticatorData: "auth_data",
				ClientDataJSON:    "client_data",
				Signature:         "sig_123",
			},
			isValid: false,
		},
		{
			name: "Missing authenticator data",
			request: BiometricVerifyRequest{
				Email:          "admin@hros.com",
				CredentialID:   "cred_123",
				ClientDataJSON: "client_data",
				Signature:      "sig_123",
			},
			isValid: false,
		},
		{
			name: "Missing client data JSON",
			request: BiometricVerifyRequest{
				Email:             "admin@hros.com",
				CredentialID:      "cred_123",
				AuthenticatorData: "auth_data",
				Signature:         "sig_123",
			},
			isValid: false,
		},
		{
			name: "Missing signature",
			request: BiometricVerifyRequest{
				Email:             "admin@hros.com",
				CredentialID:      "cred_123",
				AuthenticatorData: "auth_data",
				ClientDataJSON:    "client_data",
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

func TestBiometricDTOs_JSONMapping(t *testing.T) {
	t.Run("BiometricChallengeRequest JSON mapping", func(t *testing.T) {
		req := BiometricChallengeRequest{
			Email: "admin@hros.com",
		}
		data, err := json.Marshal(req)
		assert.NoError(t, err)
		assert.Contains(t, string(data), `"email":"admin@hros.com"`)

		var unmarshaled BiometricChallengeRequest
		err = json.Unmarshal(data, &unmarshaled)
		assert.NoError(t, err)
		assert.Equal(t, req.Email, unmarshaled.Email)
	})

	t.Run("BiometricChallengeResponse JSON mapping", func(t *testing.T) {
		resp := BiometricChallengeResponse{
			Challenge: "challenge_123",
		}
		data, err := json.Marshal(resp)
		assert.NoError(t, err)
		assert.Contains(t, string(data), `"challenge":"challenge_123"`)

		var unmarshaled BiometricChallengeResponse
		err = json.Unmarshal(data, &unmarshaled)
		assert.NoError(t, err)
		assert.Equal(t, resp.Challenge, unmarshaled.Challenge)
	})

	t.Run("BiometricVerifyRequest JSON mapping", func(t *testing.T) {
		req := BiometricVerifyRequest{
			Email:             "admin@hros.com",
			CredentialID:      "cred_123",
			AuthenticatorData: "auth_data",
			ClientDataJSON:    "client_data",
			Signature:         "sig_123",
		}
		data, err := json.Marshal(req)
		assert.NoError(t, err)
		assert.Contains(t, string(data), `"email":"admin@hros.com"`)
		assert.Contains(t, string(data), `"credential_id":"cred_123"`)
		assert.Contains(t, string(data), `"authenticator_data":"auth_data"`)
		assert.Contains(t, string(data), `"client_data_json":"client_data"`)
		assert.Contains(t, string(data), `"signature":"sig_123"`)

		var unmarshaled BiometricVerifyRequest
		err = json.Unmarshal(data, &unmarshaled)
		assert.NoError(t, err)
		assert.Equal(t, req.Email, unmarshaled.Email)
		assert.Equal(t, req.CredentialID, unmarshaled.CredentialID)
		assert.Equal(t, req.AuthenticatorData, unmarshaled.AuthenticatorData)
		assert.Equal(t, req.ClientDataJSON, unmarshaled.ClientDataJSON)
		assert.Equal(t, req.Signature, unmarshaled.Signature)
	})
}
