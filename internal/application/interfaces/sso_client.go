package interfaces

import "context"

// SSOUserProfile contains profile data retrieved from the Identity Provider.
type SSOUserProfile struct {
	Email      string
	IdentityID string
	Provider   string
}

// SSOClient defines the interface for interacting with Identity Providers to exchange codes.
type SSOClient interface {
	// ExchangeCode exchanges an authorization code for the user profile details.
	// Contract: If err == nil, the returned *SSOUserProfile pointer MUST be non-nil.
	ExchangeCode(ctx context.Context, provider string, code string) (*SSOUserProfile, error)
}
