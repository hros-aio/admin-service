package auth

import (
	"context"
	"errors"

	"github.com/hros/admin-service/internal/application/interfaces"
)

// DefaultSSOClient is a stub/noop implementation of interfaces.SSOClient.
type DefaultSSOClient struct{}

// NewDefaultSSOClient creates a new DefaultSSOClient.
func NewDefaultSSOClient() interfaces.SSOClient {
	return &DefaultSSOClient{}
}

// ExchangeCode implements the interfaces.SSOClient interface.
func (c *DefaultSSOClient) ExchangeCode(ctx context.Context, provider string, code string) (*interfaces.SSOUserProfile, error) {
	return nil, errors.New("sso client code exchange not implemented")
}
