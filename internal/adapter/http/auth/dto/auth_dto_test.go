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
