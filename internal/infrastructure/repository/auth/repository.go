package auth

import (
	"context"
	"errors"

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
