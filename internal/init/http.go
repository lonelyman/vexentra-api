package init

import (
	"log/slog"
	"vexentra-api/internal/config"
	"vexentra-api/internal/transport/http/middlewares"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

// NewFiberServer รับ Config และ Logger มาเพื่อประกอบร่าง
func NewFiberServer(cfg *config.Config, l *slog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "Vexentra API",
		// ปรับแต่ง ErrorHandler ตรงนี้ได้ในอนาคตเพื่อให้ใช้ RenderError อัตโนมัติ
	})

	// 1. Recovery: ดักจับ Panic ไม่ให้ Server ล้ม (สำคัญที่สุด)
	app.Use(recover.New())

	// 2. Request ID: สร้าง ID ประจำตัวให้ทุก Request เพื่อการ Debug
	app.Use(requestid.New())

	// 3. CORS: อนุญาตให้ Frontend ติดต่อเข้ามา
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"}, // ในอนาคตดึงจาก cfg.App.AllowedOrigins
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	// 4. Structured Logger: ใช้ตัวที่เราสร้างไว้ในขั้นตอนที่ 2
	app.Use(middlewares.StructuredLogger(l))

	return app
}
