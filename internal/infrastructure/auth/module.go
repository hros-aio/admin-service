package auth

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/hros/admin-service/internal/application/auth"
	"github.com/hros/admin-service/internal/application/usecase"
	"github.com/hros/admin-service/internal/config"
	"github.com/hros/admin-service/internal/infrastructure/cache"
	"go.uber.org/fx"
)

// Module is the Fx module for authentication infrastructure.
var Module = fx.Module(
	"auth-infra",
	fx.Provide(
		func() auth.PasswordHelper {
			return NewBcryptPasswordHelper(12)
		},
		func(cfg *config.Config) (auth.TokenProvider, error) {
			// In a real app, the private key would be loaded from config/secret
			// For now, we assume cfg.JWTPrivateKey contains the PEM encoded RS256 key
			block, _ := pem.Decode([]byte(cfg.JWTPrivateKey))
			if block == nil {
				// Fallback or error: For this implementation, we might need a dummy for bootstrap
				return nil, fmt.Errorf("failed to decode JWT private key PEM")
			}
			key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse JWT private key: %w", err)
			}
			return NewJWTTokenProvider(key, "hros-admin"), nil
		},
		NewSlogAuditLogger,
		NewDefaultSSOClient,
		cache.NewRedisSSOStateCache,
		ProvideSSOProviders,
	),
)

// ProvideSSOProviders parses and validates the SSO providers configuration.
func ProvideSSOProviders(cfg *config.Config) (map[string]usecase.SSOProviderConfig, error) {
	providers := make(map[string]usecase.SSOProviderConfig)

	hasClientID := cfg.SSOGoogleClientID != ""
	hasRedirectURL := cfg.SSOGoogleRedirectURL != ""
	hasAuthURL := cfg.SSOGoogleAuthURL != ""

	if hasClientID || hasRedirectURL || hasAuthURL {
		if !hasClientID || !hasRedirectURL || !hasAuthURL {
			return nil, fmt.Errorf("SSO Google provider configuration is incomplete: SSOGoogleClientID, SSOGoogleRedirectURL, and SSOGoogleAuthURL must all be set together")
		}
		providers["google"] = usecase.SSOProviderConfig{
			ClientID:    cfg.SSOGoogleClientID,
			RedirectURL: cfg.SSOGoogleRedirectURL,
			AuthURL:     cfg.SSOGoogleAuthURL,
			Scopes:      []string{"openid", "email", "profile"},
		}
	}
	return providers, nil
}
