package server

import (
	"context"
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func LoggingMiddleware() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			// Common attributes for all log entries
			attrs := []slog.Attr{
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.String("method", c.Request().Method),
				slog.String("duration", time.Since(v.StartTime).String()),
			}

			// Set appropriate level and message based on error presence
			level := slog.LevelInfo
			message := "REQUEST"

			if v.Error != nil {
				level = slog.LevelError
				message = "REQUEST_ERROR"
				attrs = append(attrs, slog.String("err", v.Error.Error()))
			}

			slog.LogAttrs(context.Background(), level, message, attrs...)
			return nil
		},
	})
}
