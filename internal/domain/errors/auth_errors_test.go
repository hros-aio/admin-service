package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthErrors(t *testing.T) {
	assert.Equal(t, "invalid credentials", ErrInvalidCredentials.Error())
	assert.Equal(t, "admin user not found", ErrUserNotFound.Error())
	assert.Equal(t, "admin user account is not active", ErrUserInactive.Error())
	assert.Equal(t, "admin user account is locked", ErrUserLocked.Error())
	assert.Equal(t, "account is temporarily locked", ErrAccountLocked.Error())
	assert.Equal(t, "email already exists", ErrEmailAlreadyExists.Error())
	assert.Equal(t, "session token not found", ErrTokenNotFound.Error())
	assert.Equal(t, "session token has expired", ErrTokenExpired.Error())
	assert.Equal(t, "session token has been revoked", ErrTokenRevoked.Error())
	assert.Equal(t, "invalid refresh token", ErrInvalidRefreshToken.Error())
	assert.Equal(t, "unauthorized", ErrUnauthorized.Error())
	assert.Equal(t, "forbidden", ErrForbidden.Error())
	assert.Equal(t, "MFA verification failed", ErrMFAInvalid.Error())
	assert.Equal(t, "MFA token has expired", ErrMFATokenExpired.Error())
	assert.Equal(t, "reset token has already been used", ErrTokenUsed.Error())
	assert.Equal(t, "password does not meet complexity requirements", ErrPasswordWeak.Error())
	assert.Equal(t, "invite token has expired", ErrInviteExpired.Error())
	assert.Equal(t, "invite token has already been used", ErrInviteUsed.Error())
	assert.Equal(t, "no admin account linked to this identity", ErrNoAccountLinked.Error())
	assert.Equal(t, "invalid SSO state", ErrInvalidSSOState.Error())
}
