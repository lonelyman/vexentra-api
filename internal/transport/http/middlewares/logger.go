package middlewares

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
)

func StructuredLogger(l *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		// 1. ให้ Request ไปทำงานต่อจนเสร็จก่อน
		err := c.Next()

		// 2. ดึงค่า Request ID ที่ Middleware 'requestid' สร้างไว้ให้ใน Response Header
		// แก้จาก c.RespHeader เป็น c.GetRespHeader
		requestID := c.GetRespHeader("X-Request-ID")

		l.Info("http_request",
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("latency", time.Since(start)),
			slog.String("ip", c.IP()),
		)

		return err
	}
}
