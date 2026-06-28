package dto

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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
	_ = validate.RegisterValidation("notblank", func(fl validator.FieldLevel) bool {
		return len(strings.TrimSpace(fl.Field().String())) > 0
	})

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
		{
			name: "Whitespace-only credential ID",
			request: BiometricVerifyRequest{
				Email:             "admin@hros.com",
				CredentialID:      "   ",
				AuthenticatorData: "auth_data",
				ClientDataJSON:    "client_data",
				Signature:         "sig_123",
			},
			isValid: false,
		},
		{
			name: "Whitespace-only authenticator data",
			request: BiometricVerifyRequest{
				Email:             "admin@hros.com",
				CredentialID:      "cred_123",
				AuthenticatorData: "\t",
				ClientDataJSON:    "client_data",
				Signature:         "sig_123",
			},
			isValid: false,
		},
		{
			name: "Whitespace-only client data JSON",
			request: BiometricVerifyRequest{
				Email:             "admin@hros.com",
				CredentialID:      "cred_123",
				AuthenticatorData: "auth_data",
				ClientDataJSON:    "\n",
				Signature:         "sig_123",
			},
			isValid: false,
		},
		{
			name: "Whitespace-only signature",
			request: BiometricVerifyRequest{
				Email:             "admin@hros.com",
				CredentialID:      "cred_123",
				AuthenticatorData: "auth_data",
				ClientDataJSON:    "client_data",
				Signature:         " \n\t ",
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

func TestBiometricOpenAPIRegression(t *testing.T) {
	// Locate openapi.yaml
	// Try paths up to project root
	var filePath string
	paths := []string{
		"../../../../../api/openapi.yaml",
		"../../../../api/openapi.yaml",
		"../../../api/openapi.yaml",
		"../../api/openapi.yaml",
		"./api/openapi.yaml",
		"api/openapi.yaml",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			filePath = p
			break
		}
	}
	require.NotEmpty(t, filePath, "openapi.yaml file not found in paths")

	data, err := os.ReadFile(filePath)
	require.NoError(t, err)

	var doc map[string]any
	err = yaml.Unmarshal(data, &doc)
	require.NoError(t, err)

	// Assert paths exist
	pathsMap, ok := doc["paths"].(map[string]any)
	require.True(t, ok, "paths section missing or invalid")

	// Challenge endpoint assertions
	challengePath, ok := pathsMap["/v1/auth/biometric/challenge"].(map[string]any)
	require.True(t, ok, "/v1/auth/biometric/challenge path missing")
	challengePost, ok := challengePath["post"].(map[string]any)
	require.True(t, ok, "POST method missing for challenge endpoint")
	assert.Equal(t, "biometricChallenge", challengePost["operationId"])

	// Verify request body for challenge
	challengeReqBody, ok := challengePost["requestBody"].(map[string]any)
	require.True(t, ok, "requestBody missing for challenge endpoint")
	challengeReqContent, ok := challengeReqBody["content"].(map[string]any)
	require.True(t, ok, "content missing in requestBody for challenge endpoint")
	challengeReqJSON, ok := challengeReqContent["application/json"].(map[string]any)
	require.True(t, ok, "application/json missing in requestBody content for challenge endpoint")
	challengeReqSchema, ok := challengeReqJSON["schema"].(map[string]any)
	require.True(t, ok, "schema missing in requestBody application/json for challenge endpoint")
	assert.Equal(t, "#/components/schemas/BiometricChallengeRequest", challengeReqSchema["$ref"])

	// Verify responses for challenge
	challengeResponses, ok := challengePost["responses"].(map[string]any)
	require.True(t, ok, "responses missing for challenge endpoint")

	challengeResp200, ok := challengeResponses["200"].(map[string]any)
	require.True(t, ok, "200 response missing for challenge endpoint")
	challengeResp200Content, ok := challengeResp200["content"].(map[string]any)
	require.True(t, ok, "200 response content missing for challenge endpoint")
	challengeResp200JSON, ok := challengeResp200Content["application/json"].(map[string]any)
	require.True(t, ok, "200 response application/json missing for challenge endpoint")
	challengeResp200Schema, ok := challengeResp200JSON["schema"].(map[string]any)
	require.True(t, ok, "200 response schema missing for challenge endpoint")
	assert.Equal(t, "#/components/schemas/BiometricChallengeResponse", challengeResp200Schema["$ref"])

	assert.Equal(t, map[string]any{"$ref": "#/components/responses/BadRequest"}, challengeResponses["400"])
	assert.Equal(t, map[string]any{"$ref": "#/components/responses/InternalServerError"}, challengeResponses["500"])

	// Verify endpoint assertions
	verifyPath, ok := pathsMap["/v1/auth/biometric/verify"].(map[string]any)
	require.True(t, ok, "/v1/auth/biometric/verify path missing")
	verifyPost, ok := verifyPath["post"].(map[string]any)
	require.True(t, ok, "POST method missing for verify endpoint")
	assert.Equal(t, "biometricVerify", verifyPost["operationId"])

	// Verify request body for verify
	verifyReqBody, ok := verifyPost["requestBody"].(map[string]any)
	require.True(t, ok, "requestBody missing for verify endpoint")
	verifyReqContent, ok := verifyReqBody["content"].(map[string]any)
	require.True(t, ok, "content missing in requestBody for verify endpoint")
	verifyReqJSON, ok := verifyReqContent["application/json"].(map[string]any)
	require.True(t, ok, "application/json missing in requestBody content for verify endpoint")
	verifyReqSchema, ok := verifyReqJSON["schema"].(map[string]any)
	require.True(t, ok, "schema missing in requestBody application/json for verify endpoint")
	assert.Equal(t, "#/components/schemas/BiometricVerifyRequest", verifyReqSchema["$ref"])

	// Verify responses for verify
	verifyResponses, ok := verifyPost["responses"].(map[string]any)
	require.True(t, ok, "responses missing for verify endpoint")

	verifyResp200, ok := verifyResponses["200"].(map[string]any)
	require.True(t, ok, "200 response missing for verify endpoint")
	verifyResp200Content, ok := verifyResp200["content"].(map[string]any)
	require.True(t, ok, "200 response content missing for verify endpoint")
	verifyResp200JSON, ok := verifyResp200Content["application/json"].(map[string]any)
	require.True(t, ok, "200 response application/json missing for verify endpoint")
	verifyResp200Schema, ok := verifyResp200JSON["schema"].(map[string]any)
	require.True(t, ok, "200 response schema missing for verify endpoint")
	assert.Equal(t, "#/components/schemas/LoginResponse", verifyResp200Schema["$ref"])

	assert.Equal(t, map[string]any{"$ref": "#/components/responses/BadRequest"}, verifyResponses["400"])
	assert.Equal(t, map[string]any{"$ref": "#/components/responses/Unauthorized"}, verifyResponses["401"])
	assert.Equal(t, map[string]any{"$ref": "#/components/responses/InternalServerError"}, verifyResponses["500"])

	// Assert schemas exist in components
	components, ok := doc["components"].(map[string]any)
	require.True(t, ok, "components section missing")
	schemas, ok := components["schemas"].(map[string]any)
	require.True(t, ok, "components.schemas missing")

	_, ok = schemas["BiometricChallengeRequest"].(map[string]any)
	assert.True(t, ok, "BiometricChallengeRequest schema missing")

	_, ok = schemas["BiometricChallengeResponse"].(map[string]any)
	assert.True(t, ok, "BiometricChallengeResponse schema missing")

	_, ok = schemas["BiometricVerifyRequest"].(map[string]any)
	assert.True(t, ok, "BiometricVerifyRequest schema missing")

	_, ok = schemas["LoginResponse"].(map[string]any)
	assert.True(t, ok, "LoginResponse schema missing")
}
