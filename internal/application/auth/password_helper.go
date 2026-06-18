package auth

// PasswordHelper defines the interface for password hashing and verification.
type PasswordHelper interface {
	// Hash generates a hash from a plain text password.
	Hash(password string) (string, error)
	// Compare verifies a plain text password against a hash.
	Compare(hashed, plain string) error
	// CompareDummy performs a dummy comparison to prevent timing attacks.
	CompareDummy(plain string)
}
