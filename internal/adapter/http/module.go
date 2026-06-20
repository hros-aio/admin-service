// Package http provides HTTP adapter handlers for Echo routing.
package http

import (
	"go.uber.org/fx"
)

// Module is the Fx module for HTTP adapters.
var Module = fx.Module("http-adapter",
	fx.Provide(
		NewAuthHandler,
	),
	fx.Invoke(
		RegisterRoutes,
	),
)
