package init

import (
	"vexentra-api/internal/config"
	"vexentra-api/internal/transport/http/middlewares"
	"vexentra-api/pkg/logger" // ✅ Import Interface

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

// ✅ แก้ไข: รับ logger.Logger (Interface)
func NewFiberServer(cfg *config.Config, l logger.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "Vexentra API",
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	// ✅ ใช้ GetSlog() ส่งให้ Middleware
	app.Use(middlewares.StructuredLogger(l.GetSlog()))

	return app
}
