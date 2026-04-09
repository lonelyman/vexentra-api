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
	Profile *userhdl.ProfileHandler
	Auth    *authhdl.AuthHandler
	Health  *healthhdl.HealthHandler
	AuthSvc auth.AuthService
}

func SetupRouter(app *fiber.App, h Handlers) {
	// Health Check Routes — no version prefix, ops/infra concern
	health := app.Group("/health")
	health.Get("/", h.Health.Status)
	health.Get("/live", h.Health.Live)
	health.Get("/ready", h.Health.Ready)

	api := app.Group("/api/v1")

	// Public Routes
	api.Post("/users/register", h.User.Register)
	api.Post("/auth/login", h.Auth.Login)
	api.Post("/auth/refresh", h.Auth.RefreshToken)

	// Showcase — no login required; returns profile of the pre-configured showcase user
	api.Get("/showcase", h.Profile.GetShowcase)

	// Protected Routes
	protected := api.Group("/", middlewares.AuthMiddleware(h.AuthSvc))
	protected.Get("/me", h.User.GetProfile)
	protected.Get("/users", h.User.ListUsers)
	protected.Post("/auth/logout", h.Auth.Logout)

	// Profile & Portfolio — view any user's full profile (login required)
	protected.Get("/users/:id/profile", h.Profile.GetPublicProfile)

	// Self-service profile management
	protected.Put("/me/profile", h.Profile.UpsertProfile)

	protected.Post("/me/skills", h.Profile.AddSkill)
	protected.Delete("/me/skills/:skillID", h.Profile.RemoveSkill)

	protected.Post("/me/experiences", h.Profile.AddExperience)
	protected.Put("/me/experiences/:expID", h.Profile.UpdateExperience)
	protected.Delete("/me/experiences/:expID", h.Profile.RemoveExperience)

	protected.Post("/me/portfolio", h.Profile.AddPortfolioItem)
	protected.Put("/me/portfolio/:itemID", h.Profile.UpdatePortfolioItem)
	protected.Delete("/me/portfolio/:itemID", h.Profile.RemovePortfolioItem)
}
