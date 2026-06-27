package auth

import (
	"testing"
	"time"

	"github.com/hros/admin-service/internal/domain"
	"github.com/stretchr/testify/assert"
)

// TestAdminUserMapperRoundTrip verifies that TotpSecret and WebauthnCredentials
// survive a full fromDomain → toDomain round-trip without loss or mutation.
// This test is deliberately kept pure (no DB) so it catches regressions in
// the mapping layer independently of persistence concerns.
func TestAdminUserMapperRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second) // truncate to avoid sub-second drift
	invitedBy := "invited-by-uuid"

	original := &domain.AdminUser{
		ID:                  "user-uuid",
		Name:                "Test Admin",
		Email:               "test@hros.com",
		PasswordHash:        "hashed-pw",
		RoleID:              "role-uuid",
		Status:              domain.AdminUserStatusActive,
		MFAEnabled:          true,
		TotpSecret:          "JBSWY3DPEHPK3PXP",
		WebauthnCredentials: []byte(`[{"id":"cred-1"}]`),
		FailCount:           3,
		InvitedBy:           &invitedBy,
		SSOIdentityID:       "sso-id-123",
		SSOProvider:         "google",
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	// Convert to the GORM model and immediately back to domain.
	model := fromDomain(original)
	roundTripped := model.toDomain()

	assert.Equal(t, original.TotpSecret, roundTripped.TotpSecret,
		"TotpSecret must survive the fromDomain→toDomain round-trip unchanged")
	assert.Equal(t, original.WebauthnCredentials, roundTripped.WebauthnCredentials,
		"WebauthnCredentials must survive the fromDomain→toDomain round-trip unchanged")
	assert.Equal(t, original.SSOIdentityID, roundTripped.SSOIdentityID,
		"SSOIdentityID must survive the fromDomain→toDomain round-trip unchanged")
	assert.Equal(t, original.SSOProvider, roundTripped.SSOProvider,
		"SSOProvider must survive the fromDomain→toDomain round-trip unchanged")

	// Sanity-check that the rest of the fields also round-trip correctly so that
	// future regressions in adjacent fields are caught by this test too.
	assert.Equal(t, original.ID, roundTripped.ID)
	assert.Equal(t, original.Email, roundTripped.Email)
	assert.Equal(t, original.MFAEnabled, roundTripped.MFAEnabled)
	assert.Equal(t, original.Status, roundTripped.Status)
	assert.Equal(t, original.FailCount, roundTripped.FailCount)
	assert.Equal(t, original.InvitedBy, roundTripped.InvitedBy)
}
