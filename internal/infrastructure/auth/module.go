package auth

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"

	"github.com/hros/admin-service/internal/application/auth"
	"github.com/hros/admin-service/internal/config"
	authDomain "github.com/hros/admin-service/internal/domain/auth"
	"go.uber.org/fx"
)

// Module is the Fx module for authentication infrastructure.
var Module = fx.Module("auth-infra",
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
		func(log *slog.Logger) authDomain.AuditLogger {
			return NewSlogAuditLogger(log)
		},
		NewDefaultSSOClient,
	),
)
