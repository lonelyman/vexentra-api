package userhdl

import (
	"vexentra-api/internal/modules/user/usersvc"
	"vexentra-api/internal/transport/http/presenter"
	"vexentra-api/pkg/custom_errors"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	svc usersvc.UserService
}

func NewUserHandler(svc usersvc.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

type RegisterRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

func (h *UserHandler) Register(c fiber.Ctx) error {
	// 1. ประกาศและผูกข้อมูลจาก Request Body (จุดที่หายไปตะกี้ค่ะ!)
	req := new(RegisterRequest)
	if err := c.Bind().Body(req); err != nil {
		appErr := custom_errors.New(
			400,
			custom_errors.ErrInvalidFormat,
			"ข้อมูลที่ส่งมาไม่ถูกต้องตามรูปแบบ",
			err.Error(),
		)
		return presenter.RenderError(c, appErr)
	}

	// 2. ส่งข้อมูลที่ Bind แล้วให้ Service ประมวลผล
	usr, err := h.svc.Register(
		c.Context(),
		req.Username,
		req.Email,
		req.Password,
		req.DisplayName,
	)

	if err != nil {
		return presenter.RenderError(c, err)
	}

	// 3. แปลงเป็น DTO และส่งออกผ่าน "data":{} เท่านั้น!
	return presenter.RenderItem(c, NewUserResponse(usr), fiber.StatusCreated)
}
