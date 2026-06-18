package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// BcryptPasswordHelper implements the PasswordHelper interface using bcrypt.
type BcryptPasswordHelper struct {
	cost int
}

// NewBcryptPasswordHelper creates a new BcryptPasswordHelper.
func NewBcryptPasswordHelper(cost int) *BcryptPasswordHelper {
	return &BcryptPasswordHelper{cost: cost}
}

// Hash generates a bcrypt hash from a password.
func (h *BcryptPasswordHelper) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	return string(bytes), err
}

// Compare compares a bcrypt hash with a password.
func (h *BcryptPasswordHelper) Compare(hashed, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}

// CompareDummy performs a dummy comparison against a fixed hash to prevent timing attacks.
func (h *BcryptPasswordHelper) CompareDummy(plain string) {
	// A typical bcrypt hash for comparison. This is not used for anything but timing.
	dummyHash := "$2a$12$R9h/NrC9.87.t.w4.4.4.4.4.4.4.4.4.4.4.4.4.4.4.4.4.4.4.4"
	_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(plain))
}
