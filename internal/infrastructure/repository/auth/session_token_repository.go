// Package auth implements the authentication repository using GORM.
package auth

import (
	"context"

	"github.com/hros/admin-service/internal/domain"
	platformDB "github.com/hros/admin-service/internal/platform/database"
	"gorm.io/gorm"
)

// GormSessionTokenRepository implements domain.SessionTokenRepository using GORM.
type GormSessionTokenRepository struct {
	db *gorm.DB
}

// NewGormSessionTokenRepository creates a new GormSessionTokenRepository.
func NewGormSessionTokenRepository(db *gorm.DB) domain.SessionTokenRepository {
	return &GormSessionTokenRepository{db: db}
}

// Save persists a new session token.
func (r *GormSessionTokenRepository) Save(ctx context.Context, token *domain.SessionToken) error {
	db := platformDB.GetTx(ctx, r.db)
	model := fromSessionTokenDomain(token)
	return db.Create(model).Error
}

// FindByToken retrieves a session token by its refresh token value.
func (r *GormSessionTokenRepository) FindByToken(ctx context.Context, token string) (*domain.SessionToken, error) {
	db := platformDB.GetTx(ctx, r.db)
	var model sessionTokenModel
	if err := db.Where("refresh_token = ?", token).First(&model).Error; err != nil {
		return nil, err
	}
	return model.toDomain(), nil
}

// DeleteByToken removes a session token by its refresh token value.
func (r *GormSessionTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	db := platformDB.GetTx(ctx, r.db)
	return db.Where("refresh_token = ?", token).Delete(&sessionTokenModel{}).Error
}

// DeleteByAdminID removes all session tokens for a specific admin.
func (r *GormSessionTokenRepository) DeleteByAdminID(ctx context.Context, adminID string) error {
	db := platformDB.GetTx(ctx, r.db)
	return db.Where("admin_id = ?", adminID).Delete(&sessionTokenModel{}).Error
}

// Revoke updates a session token's revocation status.
func (r *GormSessionTokenRepository) Revoke(ctx context.Context, token string, reason string) error {
	db := platformDB.GetTx(ctx, r.db)
	// Implementation for Revoke if needed, otherwise this is enough for TSK-AUTH-005 scope
	return db.Model(&sessionTokenModel{}).
		Where("refresh_token = ?", token).
		Update("revoked_at", gorm.Expr("NOW()")).
		Update("revoke_reason", reason).Error
}
