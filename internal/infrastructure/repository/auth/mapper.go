package auth

import (
	"github.com/hros/admin-service/internal/domain"
)

// toDomain converts the GORM model to a domain entity.
func (m adminUserModel) toDomain() *domain.AdminUser {
	return &domain.AdminUser{
		ID:           m.ID,
		Name:         m.Name,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		RoleID:       m.RoleID,
		Status:       domain.AdminUserStatus(m.Status),
		MFAEnabled:   m.MFAEnabled,
		MFASecret:    m.MFASecret,
		LastLoginAt:  m.LastLoginAt,
		FailCount:    m.FailCount,
		LockedUntil:  m.LockedUntil,
		InvitedBy:    m.InvitedBy,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// fromDomain converts a domain entity to a GORM model.
func fromDomain(u *domain.AdminUser) *adminUserModel {
	return &adminUserModel{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		RoleID:       u.RoleID,
		Status:       string(u.Status),
		MFAEnabled:   u.MFAEnabled,
		MFASecret:    u.MFASecret,
		LastLoginAt:  u.LastLoginAt,
		FailCount:    u.FailCount,
		LockedUntil:  u.LockedUntil,
		InvitedBy:    u.InvitedBy,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
