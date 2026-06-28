// Package http provides HTTP adapter handlers for Echo routing.
package http

import (
	"github.com/hros/admin-service/internal/application/usecase"
	"go.uber.org/fx"
)

// AcceptInviteParams contains dependencies for AcceptInviteExecutor,
// marking AcceptInviteUseCase as optional to support integration tests
// that do not bootstrap accept-invite repositories/publishers.
type AcceptInviteParams struct {
	fx.In
	UseCase *usecase.AcceptInviteUseCase `optional:"true"`
}

// Module is the Fx module for HTTP adapters.
var Module = fx.Module("http-adapter",
	fx.Provide(
		NewAuthHandler,
		NewAuthSSOHandler,
		func(p AcceptInviteParams) AcceptInviteExecutor {
			if p.UseCase == nil {
				return nil
			}
			return p.UseCase
		},
	),
	fx.Invoke(
		RegisterRoutes,
	),
)
