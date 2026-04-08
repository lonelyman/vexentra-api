// internal/bootstrap/http.go
package bootstrap

import (
	"errors"
	"vexentra-api/internal/config"
	"vexentra-api/internal/transport/http/middlewares"
	"vexentra-api/internal/transport/http/presenter"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func NewFiberServer(cfg *config.Config, l logger.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "Vexentra API",
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var appErr *custom_errors.AppError

			if errors.As(err, &appErr) {
				return presenter.RenderError(c, appErr)
			}
			return presenter.RenderError(c, custom_errors.New(code, custom_errors.ErrInternal, err.Error()))
		},
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(middlewares.StructuredLogger(l))
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.App.CORSAllowedOrigins,
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	}))

	return app
}
