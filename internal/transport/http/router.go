// internal/transport/http/router.go
package http

import (
	authhdl "vexentra-api/internal/transport/http/auth"
	healthhdl "vexentra-api/internal/transport/http/health"
	"vexentra-api/internal/transport/http/middlewares"
	userhdl "vexentra-api/internal/transport/http/user"
	"vexentra-api/pkg/auth"

	"github.com/gofiber/fiber/v3"
)

type Handlers struct {
	User    *userhdl.UserHandler
	Auth    *authhdl.AuthHandler
	Health  *healthhdl.HealthHandler
	AuthSvc auth.AuthService
}

func SetupRouter(app *fiber.App, h Handlers) {
	// Health Check Routes — no version prefix, ops/infra concern
	health := app.Group("/health")
	health.Get("/live", h.Health.Live)
	health.Get("/ready", h.Health.Ready)

	api := app.Group("/api/v1")

	// Public Routes
	api.Post("/users/register", h.User.Register)
	api.Post("/auth/login", h.Auth.Login)
	api.Post("/auth/refresh", h.Auth.RefreshToken)

	// Protected Routes
	protected := api.Group("/", middlewares.AuthMiddleware(h.AuthSvc))
	protected.Get("/me", h.User.GetProfile)
	protected.Post("/auth/logout", h.Auth.Logout)
}
