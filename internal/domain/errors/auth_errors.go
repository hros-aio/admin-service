// Package errors defines domain-level error types for authentication and authorization.
package errors

import "errors"

var (
	// ErrInvalidCredentials is returned when login fails due to wrong email or password.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrUserNotFound is returned when an admin user does not exist.
	ErrUserNotFound = errors.New("admin user not found")

	// ErrUserInactive is returned when a user account is not active.
	ErrUserInactive = errors.New("admin user account is not active")

	// ErrUserLocked is returned when a user account is locked due to too many failed attempts.
	ErrUserLocked = errors.New("admin user account is locked")

	// ErrAccountLocked is returned when a user account is temporarily locked due to brute-force protection.
	ErrAccountLocked = errors.New("account is temporarily locked")

	// ErrEmailAlreadyExists is returned when trying to create a user with an existing email.
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrTokenNotFound is returned when a session token does not exist.
	ErrTokenNotFound = errors.New("session token not found")

	// ErrTokenExpired is returned when a session token has expired.
	ErrTokenExpired = errors.New("session token has expired")

	// ErrTokenRevoked is returned when a session token has been revoked.
	ErrTokenRevoked = errors.New("session token has been revoked")

	// ErrInvalidRefreshToken is returned when a refresh token is invalid.
	ErrInvalidRefreshToken = errors.New("invalid refresh token")

	// ErrUnauthorized is returned when a request lacks valid authentication.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when a user does not have permission for an action.
	ErrForbidden = errors.New("forbidden")

	// ErrMFAInvalid is returned when the MFA verification fails.
	ErrMFAInvalid = errors.New("MFA verification failed")

	// ErrMFATokenExpired is returned when the short-lived MFA session token has expired.
	ErrMFATokenExpired = errors.New("MFA token has expired")

	// ErrTokenUsed is returned when a reset token has already been used.
	ErrTokenUsed = errors.New("reset token has already been used")

	// ErrPasswordWeak is returned when a new password does not meet complexity requirements.
	ErrPasswordWeak = errors.New("password does not meet complexity requirements")

	// ErrInviteExpired is returned when an invite token has expired.
	ErrInviteExpired = errors.New("invite token has expired")

	// ErrInviteUsed is returned when an invite token has already been used.
	ErrInviteUsed = errors.New("invite token has already been used")

	// ErrNoAccountLinked is returned when no admin account is linked to the SSO identity.
	ErrNoAccountLinked = errors.New("no admin account linked to this identity")

	// ErrInvalidSSOState is returned when the SSO state or nonce is invalid or expired.
	ErrInvalidSSOState = errors.New("invalid SSO state")

	// ErrIdentityConflict is returned when SSO identity and email resolve to different admin users.
	ErrIdentityConflict = errors.New("identity conflict: email and SSO ID map to different users")

	// ErrBiometricNotRegistered is returned when biometric login is attempted but no credentials are registered.
	ErrBiometricNotRegistered = errors.New("biometric credential not registered")

	// ErrInvalidBiometricSignature is returned when biometric cryptographic verification fails.
	ErrInvalidBiometricSignature = errors.New("invalid biometric signature")

	// ErrChallengeNotFoundOrExpired is returned when a biometric authentication challenge is not found or has expired.
	ErrChallengeNotFoundOrExpired = errors.New("cryptographic challenge not found or expired")
)
