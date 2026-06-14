// Package migration provides database migration utilities.
package migration

import (
	"log/slog"

	"gorm.io/gorm"
)

// Migrator handles database migrations.
type Migrator struct {
	db  *gorm.DB
	log *slog.Logger
}

// NewMigrator creates a new Migrator.
func NewMigrator(db *gorm.DB, log *slog.Logger) *Migrator {
	return &Migrator{
		db:  db,
		log: log,
	}
}

// Run executes the migrations.
func (m *Migrator) Run(models ...interface{}) error {
	m.log.Info("running migrations")
	if err := m.db.AutoMigrate(models...); err != nil {
		m.log.Error("migration failed", slog.Any("error", err))
		return err
	}
	m.log.Info("migrations completed successfully")
	return nil
}
