// Package domain defines the core domain models and repository interfaces.
package domain

import (
	"context"
	"time"
)

// AdminUserStatus represents the status of an admin user account.
type AdminUserStatus string

// AdminUserStatus constants define the possible states of an admin user.
const (
	AdminUserStatusActive   AdminUserStatus = "active"
	AdminUserStatusInactive AdminUserStatus = "inactive"
	AdminUserStatusPending  AdminUserStatus = "pending"
)

// AdminUser represents an HROS administrator.
type AdminUser struct {
	ID                  string
	Name                string
	Email               string
	PasswordHash        string
	RoleID              string
	Status              AdminUserStatus
	MFAEnabled          bool
	MFASecret           string
	TotpSecret          string
	WebauthnCredentials []byte
	LastLoginAt         *time.Time
	FailCount           int
	LockedUntil         *time.Time
	InvitedBy           *string
	SSOIdentityID       string
	SSOProvider         string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// Role represents a set of permissions.
type Role struct {
	ID           string
	Name         string
	Description  string
	IsSystemRole bool
	Permissions  []RolePermission
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// RolePermission defines access to specific modules.
type RolePermission struct {
	ID         string
	RoleID     string
	Module     string
	CanView    bool
	CanCreate  bool
	CanUpdate  bool
	CanDelete  bool
	CanApprove bool
	CanExport  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// IsActive checks if the admin user account is active.
func (u *AdminUser) IsActive() bool {
	return u.Status == AdminUserStatusActive
}

// IsLocked checks if the admin user account is currently locked.
func (u *AdminUser) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return u.LockedUntil.After(time.Now())
}

// AdminUserRepository defines the interface for persisting and retrieving admin users.
type AdminUserRepository interface {
	Save(ctx context.Context, user *AdminUser) error
	Update(ctx context.Context, user *AdminUser) error
	FindByID(ctx context.Context, id string) (*AdminUser, error)
	FindByEmail(ctx context.Context, email string) (*AdminUser, error)
	Delete(ctx context.Context, id string) error
	GetRoleCodeByID(ctx context.Context, roleID string) (string, error)
	UpdatePassword(ctx context.Context, id string, newHash string) error
	ActivateAccount(ctx context.Context, adminID string, newHash string) error
	FindByEmailOrSSO(ctx context.Context, email string, ssoProvider string, ssoID string) (*AdminUser, error)
	UpdateWebAuthnSignCount(ctx context.Context, adminID string, credentialID string, newCount uint32) error
}
