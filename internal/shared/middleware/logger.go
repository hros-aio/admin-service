// Package middleware provides Echo middleware components.
package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

// Logger returns a middleware that logs HTTP requests using slog.
func Logger(log *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			stop := time.Now()

			req := c.Request()
			res := c.Response()

			log.InfoContext(req.Context(), "request",
				slog.String("remote_ip", c.RealIP()),
				slog.String("host", req.Host),
				slog.String("method", req.Method),
				slog.String("uri", req.RequestURI),
				slog.Int("status", res.Status),
				slog.Int64("size", res.Size),
				slog.String("user_agent", req.UserAgent()),
				slog.Duration("latency", stop.Sub(start)),
				slog.String("request_id", res.Header().Get(echo.HeaderXRequestID)),
			)

			return nil
		}
	}
}
