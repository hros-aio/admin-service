package auth

import (
	"context"
	"errors"
	"time"

	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	platformDB "github.com/hros/admin-service/internal/platform/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormInviteTokenRepository implements domain.InviteTokenRepository using GORM.
type GormInviteTokenRepository struct {
	db *gorm.DB
}

// NewGormInviteTokenRepository creates a new GormInviteTokenRepository.
func NewGormInviteTokenRepository(db *gorm.DB) domain.InviteTokenRepository {
	return &GormInviteTokenRepository{db: db}
}

// Save persists a new invite token.
func (r *GormInviteTokenRepository) Save(ctx context.Context, token *domain.InviteToken) error {
	db := platformDB.GetTx(ctx, r.db)
	model := fromInviteTokenDomain(token)
	return db.Create(model).Error
}

// FindByToken retrieves an invite token by its token string.
func (r *GormInviteTokenRepository) FindByToken(ctx context.Context, token string) (*domain.InviteToken, error) {
	db := platformDB.GetTx(ctx, r.db)
	var model inviteTokenModel
	if err := db.Where("token = ?", token).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrTokenNotFound
		}
		return nil, err
	}
	return model.toDomain(), nil
}

// Update updates an existing invite token.
func (r *GormInviteTokenRepository) Update(ctx context.Context, token *domain.InviteToken) error {
	db := platformDB.GetTx(ctx, r.db)
	model := fromInviteTokenDomain(token)
	return db.Save(model).Error
}

// Consume atomically and conditionally marks an invite token as used only if it is still unused and unexpired.
func (r *GormInviteTokenRepository) Consume(ctx context.Context, token string) (*domain.InviteToken, error) {
	db := platformDB.GetTx(ctx, r.db)
	var model inviteTokenModel
	now := time.Now()

	err := db.Transaction(func(tx *gorm.DB) error {
		// SELECT FOR UPDATE to atomically lock the row
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("token = ?", token).First(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainErrors.ErrTokenNotFound
			}
			return err
		}

		if model.ConsumedAt != nil {
			return domainErrors.ErrInviteUsed
		}

		// Since now.After(model.ExpiresAt) or now.Equal(model.ExpiresAt) means expired
		if !now.Before(model.ExpiresAt) {
			return domainErrors.ErrInviteExpired
		}

		model.ConsumedAt = &now
		if err := tx.Save(&model).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return model.toDomain(), nil
}
