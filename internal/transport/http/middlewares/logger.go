// vexentra-api/internal/transport/http/middlewares/logger.go
package middlewares

import (
	"time"
	"vexentra-api/pkg/logger"

	"log/slog"

	"github.com/gofiber/fiber/v3"
)

func StructuredLogger(l logger.Logger) fiber.Handler {
	sl := l.GetSlog()
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		sl.Info("http_request",
			slog.String("request_id", c.GetRespHeader("X-Request-ID")),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("latency", time.Since(start)),
			slog.String("ip", c.IP()),
		)

		return err
	}
}
