package auth

import (
	"time"
)

// adminUserModel represents the GORM model for admin users.
type adminUserModel struct {
	ID           string    `gorm:"primaryKey;type:uuid"`
	Name         string    `gorm:"not null"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	RoleID       string    `gorm:"type:uuid;not null"`
	Status       string    `gorm:"not null"`
	MFAEnabled   bool      `gorm:"not null;default:false"`
	MFASecret    string    `gorm:"type:varchar(255)"`
	LastLoginAt  *time.Time
	FailCount    int       `gorm:"not null;default:0"`
	LockedUntil  *time.Time
	InvitedBy    *string   `gorm:"type:uuid"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

// TableName returns the table name for the adminUserModel.
func (adminUserModel) TableName() string {
	return "admin_users"
}
