package http

import (
	"vexentra-api/internal/transport/http/presenter"
	userhdl "vexentra-api/internal/transport/http/user"

	// import handler ของ user
	"github.com/gofiber/fiber/v3"
)

// SetupRouter รับ UserHandler เข้ามาด้วย (Dependency Injection)
func SetupRouter(app *fiber.App, userHdl *userhdl.UserHandler) {

	app.Get("/test-list", func(c fiber.Ctx) error {
		products := []fiber.Map{{"id": 1, "name": "Vexentra Shirt"}}
		pg := presenter.NewOffsetPagination(100, 10, 0)
		return presenter.RenderList(c, products, pg)
	})

	// 2. เส้นทางจริงของ Module User
	api := app.Group("/api/v1")
	users := api.Group("/users")
	users.Post("/register", userHdl.Register)
}
