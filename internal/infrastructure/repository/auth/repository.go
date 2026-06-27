package auth

import (
	"context"
	"errors"
	"time"

	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	platformDB "github.com/hros/admin-service/internal/platform/database"
	"gorm.io/gorm"
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
