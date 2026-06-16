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

	// ErrEmailAlreadyExists is returned when trying to create a user with an existing email.
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrTokenNotFound is returned when a session token does not exist.
	ErrTokenNotFound = errors.New("session token not found")

	// ErrTokenExpired is returned when a session token has expired.
	ErrTokenExpired = errors.New("session token has expired")

	// ErrTokenRevoked is returned when a session token has been revoked.
	ErrTokenRevoked = errors.New("session token has been revoked")

	// ErrUnauthorized is returned when a request lacks valid authentication.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when a user does not have permission for an action.
	ErrForbidden = errors.New("forbidden")
)
