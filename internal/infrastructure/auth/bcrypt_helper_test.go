package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestBcryptPasswordHelper(t *testing.T) {
	h := NewBcryptPasswordHelper(bcrypt.MinCost)

	password := "secure123"

	t.Run("HashAndCompare", func(t *testing.T) {
		hashed, err := h.Hash(password)
		assert.NoError(t, err)
		assert.NotEmpty(t, hashed)

		err = h.Compare(hashed, password)
		assert.NoError(t, err)

		err = h.Compare(hashed, "wrong")
		assert.Error(t, err)
	})

	t.Run("CompareDummy", func(t *testing.T) {
		// Just ensure it doesn't panic
		assert.NotPanics(t, func() {
			h.CompareDummy(password)
		})
	})
}
