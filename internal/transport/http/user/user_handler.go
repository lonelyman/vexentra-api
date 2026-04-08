package userhdl

import (
	"vexentra-api/internal/modules/user/usersvc"
	"vexentra-api/internal/transport/http/presenter"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger" // ✅ Import Interface ของเรามา

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	svc      usersvc.UserService
	validate *validator.Validate
	logger   logger.Logger // ✅ เพิ่มฟิลด์นี้เข้าไป
}

// NewUserHandler ต้องรับ logger เข้ามาด้วย
func NewUserHandler(svc usersvc.UserService, l logger.Logger) *UserHandler {
	if l == nil { l = logger.Get() }
	return &UserHandler{
		svc:      svc,
		validate: validator.New(),
		logger:   l, // ✅ เก็บไว้ใช้งาน
	}
}

func (h *UserHandler) Register(c fiber.Ctx) error {
	req := new(RegisterRequest)

	// --- [A] ใช้ Info เพื่อบอกว่ามีคนเริ่มยิง API เข้ามา ---
	h.logger.Info("Attempting to register new user")

	// 1. Bind JSON
	if err := c.Bind().Body(req); err != nil {
		h.logger.Error("Failed to bind user request", err) // --- [B] ใช้ Error บันทึกปัญหา ---
		return presenter.RenderError(c, custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง"))
	}

	// --- [C] ใช้ Dump เพื่อดูข้อมูลที่ส่งมา (ดีมากตอน Debug) ---
	h.logger.Dump("User Registration Payload", req)

	// 2. Validate Input
	if err := h.validate.Struct(req); err != nil {
		// --- [D] ใช้ Warn สำหรับความผิดพลาดที่เกิดจาก User (ไม่ถึงกับ Error ระบบ) ---
		h.logger.Warn("Validation failed for user registration", "details", err.Error())
		return presenter.RenderError(c, custom_errors.New(400, "VALIDATION_FAILED", err.Error()))
	}

	// 3. เรียก Service (สมมติว่าผ่าน)
	// h.svc.Register(...)

	// --- [E] ใช้ Success เมื่อทุกอย่างจบลงอย่างสวยงาม ---
	h.logger.Success("User registered successfully", "username", req.Username)

	return presenter.RenderItem(c, "User registration successful", fiber.StatusCreated)
}
