// internal/transport/http/router.go
package http

import (
	authhdl "vexentra-api/internal/transport/http/auth"
	dashboardhdl "vexentra-api/internal/transport/http/dashboard"
	healthhdl "vexentra-api/internal/transport/http/health"
	"vexentra-api/internal/transport/http/middlewares"
	projecthdl "vexentra-api/internal/transport/http/project"
	socialplatformhdl "vexentra-api/internal/transport/http/socialplatform"
	txcategoryhdl "vexentra-api/internal/transport/http/txcategory"
	userhdl "vexentra-api/internal/transport/http/user"
	"vexentra-api/pkg/auth"

	"github.com/gofiber/fiber/v3"
)

type Handlers struct {
	User           *userhdl.UserHandler
	Profile        *userhdl.ProfileHandler
	SocialPlatform *socialplatformhdl.SocialPlatformHandler
	Auth           *authhdl.AuthHandler
	Health         *healthhdl.HealthHandler
	Project        *projecthdl.ProjectHandler
	Member         *projecthdl.MemberHandler
	Transaction    *projecthdl.TransactionHandler
	TxCategory     *txcategoryhdl.CategoryHandler
	Dashboard      *dashboardhdl.DashboardHandler
	AuthSvc        auth.AuthService
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
	api.Get("/auth/verify-email", h.Auth.VerifyEmail)
	api.Post("/auth/forgot-password", h.Auth.ForgotPassword)
	api.Post("/auth/reset-password", h.Auth.ResetPassword)

	// Showcase — no login required; returns profile of the pre-configured showcase user
	api.Get("/showcase", h.Profile.GetShowcase)

	// Social Platforms — master data (public read)
	api.Get("/social-platforms", h.SocialPlatform.List)

	// Protected Routes
	protected := api.Group("/", middlewares.AuthMiddleware(h.AuthSvc))
	protected.Get("/me", h.User.GetProfile)
	protected.Get("/users", middlewares.RoleMiddleware("admin"), h.User.ListUsers)
	protected.Post("/auth/logout", h.Auth.Logout)
	protected.Post("/auth/resend-verify", h.Auth.ResendVerifyEmail)
	protected.Put("/me/password", h.User.ChangePassword)
	protected.Post("/me/claim-person", h.User.ClaimPerson) // ยืนยัน claim Person ที่ระบบ suggest

	// Profile & Portfolio — view any user's full profile (login required)
	protected.Get("/me/profile", h.Profile.GetMyProfile)
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

	protected.Put("/me/social-links/:platformID", h.Profile.UpsertSocialLink)
	protected.Delete("/me/social-links/:linkID", h.Profile.DeleteSocialLink)
	protected.Get("/me/social-links", h.Profile.GetSocialLinks)

	// Social Platforms — master data (admin write)
	protected.Post("/social-platforms", middlewares.RoleMiddleware("admin"), h.SocialPlatform.Create)
	protected.Put("/social-platforms/:id", middlewares.RoleMiddleware("admin"), h.SocialPlatform.Update)
	protected.Delete("/social-platforms/:id", middlewares.RoleMiddleware("admin"), h.SocialPlatform.Delete)

	// ───────── Project Management ─────────
	// Per-action permission (staff / creator / lead / member) is enforced inside the
	// service layer via user.Caller — no RoleMiddleware guard needed on these routes.
	protected.Post("/projects", h.Project.Create)
	protected.Get("/projects", h.Project.List)
	protected.Get("/projects/by-code/:code", h.Project.GetByCode)
	protected.Get("/projects/:id", h.Project.Get)
	protected.Put("/projects/:id", h.Project.Update)
	protected.Post("/projects/:id/close", h.Project.Close)
	protected.Delete("/projects/:id", h.Project.Delete)

	// Members
	protected.Post("/projects/:id/members", h.Member.Add)
	protected.Get("/projects/:id/members", h.Member.List)
	protected.Delete("/projects/:id/members/:memberID", h.Member.Remove)
	protected.Post("/projects/:id/transfer-lead", h.Member.TransferLead)

	// Transactions — all write paths return 409 PROJECT_CLOSED once project is closed
	protected.Post("/projects/:id/transactions", h.Transaction.Create)
	protected.Get("/projects/:id/transactions", h.Transaction.List)
	protected.Get("/projects/:id/transactions/summary", h.Transaction.Summary)
	protected.Get("/projects/:id/transactions/export", h.Transaction.ExportCSV)
	protected.Get("/projects/:id/transactions/:txID", h.Transaction.Get)
	protected.Put("/projects/:id/transactions/:txID", h.Transaction.Update)
	protected.Delete("/projects/:id/transactions/:txID", h.Transaction.Delete)

	// Dashboard — aggregate stats scoped to caller's accessible projects
	protected.Get("/dashboard/stats", h.Dashboard.GetStats)

	// Transaction Categories — read open to any authenticated user, writes admin-only
	protected.Get("/tx-categories", h.TxCategory.List)
	protected.Get("/tx-categories/:id", h.TxCategory.Get)
	protected.Post("/tx-categories", middlewares.RoleMiddleware("admin"), h.TxCategory.Create)
	protected.Put("/tx-categories/:id", middlewares.RoleMiddleware("admin"), h.TxCategory.Update)
	protected.Delete("/tx-categories/:id", middlewares.RoleMiddleware("admin"), h.TxCategory.Delete)
}
