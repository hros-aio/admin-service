// Package auth implements the authentication repository using GORM.
package auth

import (
	"time"
)

// adminUserModel represents the GORM model for admin users.
type adminUserModel struct {
	ID                  string `gorm:"primaryKey;type:uuid"`
	Name                string `gorm:"not null"`
	Email               string `gorm:"uniqueIndex;not null"`
	PasswordHash        string `gorm:"not null"`
	RoleID              string `gorm:"type:uuid;not null"`
	Status              string `gorm:"not null"`
	MFAEnabled          bool   `gorm:"not null;default:false"`
	TotpSecret          string `gorm:"column:totp_secret;type:varchar(255)"`
	WebauthnCredentials []byte `gorm:"column:webauthn_credentials;type:jsonb"`
	LastLoginAt         *time.Time
	FailCount           int `gorm:"not null;default:0"`
	LockedUntil         *time.Time
	InvitedBy           *string   `gorm:"type:uuid"`
	CreatedAt           time.Time `gorm:"not null"`
	UpdatedAt           time.Time `gorm:"not null"`
}

// TableName returns the table name for the adminUserModel.
func (adminUserModel) TableName() string {
	return "admin_users"
}

// sessionTokenModel represents the GORM model for session tokens.
type sessionTokenModel struct {
	ID           string     `gorm:"primaryKey;type:uuid"`
	AdminID      string     `gorm:"type:uuid;not null;index"`
	RefreshToken string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	ExpiresAt    time.Time  `gorm:"not null"`
	IsPersistent bool       `gorm:"not null;default:false"`
	IPAddress    string     `gorm:"type:varchar(45)"`
	UserAgent    string     `gorm:"type:text"`
	CreatedAt    time.Time  `gorm:"not null"`
	RevokedAt    *time.Time `gorm:"default:null"`
	RevokeReason *string    `gorm:"type:varchar(100)"`
}

// TableName returns the table name for the sessionTokenModel.
func (sessionTokenModel) TableName() string {
	return "session_tokens"
}
