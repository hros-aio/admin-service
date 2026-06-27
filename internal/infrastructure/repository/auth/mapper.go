// Package auth implements the authentication repository using GORM.
package auth

import (
	"github.com/hros/admin-service/internal/domain"
)

// toDomain converts the GORM model to a domain entity.
func (m adminUserModel) toDomain() *domain.AdminUser {
	return &domain.AdminUser{
		ID:                  m.ID,
		Name:                m.Name,
		Email:               m.Email,
		PasswordHash:        m.PasswordHash,
		RoleID:              m.RoleID,
		Status:              domain.AdminUserStatus(m.Status),
		MFAEnabled:          m.MFAEnabled,
		TotpSecret:          m.TotpSecret,
		WebauthnCredentials: m.WebauthnCredentials,
		LastLoginAt:         m.LastLoginAt,
		FailCount:           m.FailCount,
		LockedUntil:         m.LockedUntil,
		InvitedBy:           m.InvitedBy,
		SSOIdentityID:       m.SSOIdentityID,
		SSOProvider:         m.SSOProvider,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}

// fromDomain converts a domain entity to a GORM model.
func fromDomain(u *domain.AdminUser) *adminUserModel {
	return &adminUserModel{
		ID:                  u.ID,
		Name:                u.Name,
		Email:               u.Email,
		PasswordHash:        u.PasswordHash,
		RoleID:              u.RoleID,
		Status:              string(u.Status),
		MFAEnabled:          u.MFAEnabled,
		TotpSecret:          u.TotpSecret,
		WebauthnCredentials: u.WebauthnCredentials,
		LastLoginAt:         u.LastLoginAt,
		FailCount:           u.FailCount,
		LockedUntil:         u.LockedUntil,
		InvitedBy:           u.InvitedBy,
		SSOIdentityID:       u.SSOIdentityID,
		SSOProvider:         u.SSOProvider,
		CreatedAt:           u.CreatedAt,
		UpdatedAt:           u.UpdatedAt,
	}
}

// toSessionTokenDomain converts the GORM model to a domain entity.
func (m sessionTokenModel) toDomain() *domain.SessionToken {
	var reason string
	if m.RevokeReason != nil {
		reason = *m.RevokeReason
	}
	return &domain.SessionToken{
		ID:           m.ID,
		AdminID:      m.AdminID,
		RefreshToken: m.RefreshToken,
		ExpiresAt:    m.ExpiresAt,
		IsPersistent: m.IsPersistent,
		IPAddress:    m.IPAddress,
		UserAgent:    m.UserAgent,
		CreatedAt:    m.CreatedAt,
		RevokedAt:    m.RevokedAt,
		RevokeReason: reason,
	}
}

// fromSessionTokenDomain converts a domain entity to a GORM model.
func fromSessionTokenDomain(t *domain.SessionToken) *sessionTokenModel {
	var reason *string
	if t.RevokeReason != "" {
		reason = &t.RevokeReason
	}
	return &sessionTokenModel{
		ID:           t.ID,
		AdminID:      t.AdminID,
		RefreshToken: t.RefreshToken,
		ExpiresAt:    t.ExpiresAt,
		IsPersistent: t.IsPersistent,
		IPAddress:    t.IPAddress,
		UserAgent:    t.UserAgent,
		CreatedAt:    t.CreatedAt,
		RevokedAt:    t.RevokedAt,
		RevokeReason: reason,
	}
}

// toInviteTokenDomain converts the GORM model to a domain entity.
func (m inviteTokenModel) toDomain() *domain.InviteToken {
	return &domain.InviteToken{
		ID:        m.ID,
		AdminID:   m.AdminID,
		Token:     m.Token,
		ExpiresAt: m.ExpiresAt,
		UsedAt:    m.ConsumedAt,
		CreatedBy: m.InviterID,
		CreatedAt: m.CreatedAt,
	}
}

// fromInviteTokenDomain converts a domain entity to a GORM model.
func fromInviteTokenDomain(t *domain.InviteToken) *inviteTokenModel {
	return &inviteTokenModel{
		ID:         t.ID,
		AdminID:    t.AdminID,
		Token:      t.Token,
		InviterID:  t.CreatedBy,
		ExpiresAt:  t.ExpiresAt,
		ConsumedAt: t.UsedAt,
		CreatedAt:  t.CreatedAt,
	}
}
