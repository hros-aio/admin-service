package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	platformDB "github.com/hros/admin-service/internal/platform/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormAdminUserRepository implements domain.AdminUserRepository using GORM.
type GormAdminUserRepository struct {
	db *gorm.DB
}

// NewGormAdminUserRepository creates a new GormAdminUserRepository.
func NewGormAdminUserRepository(db *gorm.DB) domain.AdminUserRepository {
	return &GormAdminUserRepository{db: db}
}

// Save persists a new admin user.
func (r *GormAdminUserRepository) Save(ctx context.Context, user *domain.AdminUser) error {
	db := platformDB.GetTx(ctx, r.db)
	model := fromDomain(user)
	return db.Create(model).Error
}

// Update updates an existing admin user.
func (r *GormAdminUserRepository) Update(ctx context.Context, user *domain.AdminUser) error {
	db := platformDB.GetTx(ctx, r.db)
	model := fromDomain(user)
	return db.Save(model).Error
}

// FindByID retrieves an admin user by ID.
func (r *GormAdminUserRepository) FindByID(ctx context.Context, id string) (*domain.AdminUser, error) {
	db := platformDB.GetTx(ctx, r.db)
	var model adminUserModel
	if err := db.First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrUserNotFound
		}
		return nil, err
	}
	return model.toDomain(), nil
}

// FindByEmail retrieves an admin user by email.
func (r *GormAdminUserRepository) FindByEmail(ctx context.Context, email string) (*domain.AdminUser, error) {
	db := platformDB.GetTx(ctx, r.db)
	var model adminUserModel
	if err := db.Where("email = ?", email).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrUserNotFound
		}
		return nil, err
	}
	return model.toDomain(), nil
}

// Delete removes an admin user.
func (r *GormAdminUserRepository) Delete(ctx context.Context, id string) error {
	db := platformDB.GetTx(ctx, r.db)
	return db.Delete(&adminUserModel{}, "id = ?", id).Error
}

// GetRoleCodeByID retrieves the role code (e.g. "SUPER_ADMIN", "STANDARD_ADMIN") for the given role ID.
func (r *GormAdminUserRepository) GetRoleCodeByID(ctx context.Context, roleID string) (string, error) {
	db := platformDB.GetTx(ctx, r.db)
	var result struct {
		Name string
	}
	err := db.Table("roles").Select("name").Where("id = ?", roleID).First(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", domainErrors.ErrUserNotFound
		}
		return "", err
	}
	// Map the display name to an immutable role code.
	switch result.Name {
	case "Super Admin":
		return "SUPER_ADMIN", nil
	case "Standard Admin":
		return "STANDARD_ADMIN", nil
	default:
		return "STANDARD_ADMIN", nil
	}
}

// UpdatePassword updates only the password hash of the admin user.
func (r *GormAdminUserRepository) UpdatePassword(ctx context.Context, id string, newHash string) error {
	db := platformDB.GetTx(ctx, r.db)
	result := db.Model(&adminUserModel{}).
		Where("id = ?", id).
		Update("password_hash", newHash)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainErrors.ErrUserNotFound
	}
	return nil
}

// ActivateAccount securely updates the password hash and changes the status from pending to active in one atomic GORM execution.
func (r *GormAdminUserRepository) ActivateAccount(ctx context.Context, adminID string, newHash string) error {
	db := platformDB.GetTx(ctx, r.db)
	result := db.Model(&adminUserModel{}).
		Where("id = ? AND status = ?", adminID, "pending").
		Updates(map[string]interface{}{
			"password_hash": newHash,
			"status":        "active",
			"updated_at":    time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainErrors.ErrUserNotFound
	}
	return nil
}

// FindByEmailOrSSO retrieves an admin user by email OR SSO Identity ID/Provider.
func (r *GormAdminUserRepository) FindByEmailOrSSO(ctx context.Context, email string, ssoProvider string, ssoID string) (*domain.AdminUser, error) {
	db := platformDB.GetTx(ctx, r.db)

	var emailUser *domain.AdminUser
	var ssoUser *domain.AdminUser

	// 1. Resolve by email if not blank
	if email != "" {
		var emailModel adminUserModel
		err := db.Where("email = ?", email).First(&emailModel).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if err == nil {
			emailUser = emailModel.toDomain()
		}
	}

	// 2. Resolve by SSO if provider and ssoID are not blank
	if ssoProvider != "" && ssoID != "" {
		var ssoModel adminUserModel
		err := db.Where("sso_provider = ? AND sso_identity_id = ?", ssoProvider, ssoID).First(&ssoModel).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if err == nil {
			ssoUser = ssoModel.toDomain()
		}
	}

	// 3. Compare and return results
	if emailUser != nil && ssoUser != nil {
		if emailUser.ID != ssoUser.ID {
			return nil, domainErrors.ErrIdentityConflict
		}
		return emailUser, nil
	}

	if emailUser != nil {
		return emailUser, nil
	}

	if ssoUser != nil {
		return ssoUser, nil
	}

	// If neither resolved, return ErrUserNotFound
	return nil, domainErrors.ErrUserNotFound
}

type repoWebAuthnCredential struct {
	ID        string `json:"id"`
	PublicKey string `json:"public_key"`
	SignCount uint32 `json:"sign_count"`
}

// UpdateWebAuthnSignCount updates the signature count of the matched credential inside the webauthn_credentials JSONB column monotonically.
func (r *GormAdminUserRepository) UpdateWebAuthnSignCount(ctx context.Context, adminID string, credentialID string, newCount uint32) error {
	db := platformDB.GetTx(ctx, r.db)

	return db.Transaction(func(tx *gorm.DB) error {
		var user adminUserModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", adminID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainErrors.ErrUserNotFound
			}
			return err
		}

		trimmedCreds := bytes.TrimSpace(user.WebauthnCredentials)
		if len(trimmedCreds) == 0 {
			return domainErrors.ErrBiometricNotRegistered
		}

		isArray := trimmedCreds[0] == '['
		var creds []repoWebAuthnCredential
		switch trimmedCreds[0] {
		case '[':
			if err := json.Unmarshal(trimmedCreds, &creds); err != nil {
				return err
			}
		case '{':
			var singleCred repoWebAuthnCredential
			if err := json.Unmarshal(trimmedCreds, &singleCred); err != nil {
				return err
			}
			creds = []repoWebAuthnCredential{singleCred}
		default:
			return domainErrors.ErrBiometricNotRegistered
		}

		found := false
		for i := range creds {
			if creds[i].ID == credentialID {
				if newCount > creds[i].SignCount {
					creds[i].SignCount = newCount
				}
				found = true
				break
			}
		}

		if !found {
			return domainErrors.ErrBiometricNotRegistered
		}

		var updatedBytes []byte
		var err error
		if isArray {
			updatedBytes, err = json.Marshal(creds)
		} else {
			updatedBytes, err = json.Marshal(creds[0])
		}
		if err != nil {
			return err
		}

		if err := tx.Model(&adminUserModel{}).Where("id = ?", adminID).Update("webauthn_credentials", updatedBytes).Error; err != nil {
			return err
		}

		return nil
	})
}
