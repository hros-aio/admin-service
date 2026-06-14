// Package database provides tools for database connection and transaction management.
package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hros/admin-service/internal/config"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewDatabase initializes the GORM database connection.
func NewDatabase(cfg *config.Config, _ *slog.Logger, lc fx.Lifecycle) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DBURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		},
	})

	return db, nil
}
