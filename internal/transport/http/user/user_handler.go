package userhdl

import (
	"vexentra-api/internal/modules/user/usersvc"
	"vexentra-api/internal/transport/http/presenter"
	"vexentra-api/pkg/auth"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"
	"vexentra-api/pkg/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type UserHandler struct {
	svc      usersvc.UserService
	validate *validator.Validate
	logger   logger.Logger
}

func NewUserHandler(svc usersvc.UserService, l logger.Logger) *UserHandler {
	if l == nil {
		l = logger.Get()
	}
	return &UserHandler{
		svc:      svc,
		validate: validation.New(),
		logger:   l,
	}
}

func (h *UserHandler) Register(c fiber.Ctx) error {
	req := new(RegisterRequest)

	if err := c.Bind().Body(req); err != nil {
		h.logger.Error("Failed to bind register request", err)
		return presenter.RenderError(c, custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง"))
	}

	if vResult := validation.Validate(h.validate, req); !vResult.IsValid {
		h.logger.Warn("Validation failed", "errors", vResult.Errors)
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", vResult.Errors))
	}

	result, err := h.svc.Register(c.Context(), req.Email, req.Password)
	if err != nil {
		return err
	}

	h.logger.Success("User registered successfully", "userID", result.User.ID)
	return presenter.RenderItem(c, RegisterResponse{
		User:         NewUserResponse(result.User),
		AccessToken:  result.TokenPair.AccessToken,
		RefreshToken: result.TokenPair.RefreshToken,
	}, fiber.StatusCreated)
}

func (h *UserHandler) GetProfile(c fiber.Ctx) error {
	claims := auth.GetClaims(c)
	if claims == nil {
		return custom_errors.New(401, custom_errors.ErrUnauthorized, "ไม่พบข้อมูล Token")
	}

	// sub is stored as UUID string — parse it back
	userID, err := uuid.Parse(claims.GetUserID())
	if err != nil {
		return custom_errors.New(401, custom_errors.ErrUnauthorized, "Token มี UserID ไม่ถูกต้อง")
	}

	u, err := h.svc.GetProfile(c.Context(), userID)
	if err != nil {
		return err
	}

	h.logger.Info("Profile retrieved", "userID", userID)
	return presenter.RenderItem(c, NewUserResponse(u))
}

// ListUsers godoc
// GET /api/v1/users?page=1&limit=10           → offset-based pagination
// GET /api/v1/users?cursor=<uuid>&limit=20    → cursor-based pagination
//
// Mode is selected by the presence of the "cursor" query param.
// If "cursor" is present (even empty string won't trigger — must use ?cursor=<value>),
// cursor mode is used. Otherwise, offset mode is used.
func (h *UserHandler) ListUsers(c fiber.Ctx) error {
	if _, exists := c.Queries()["cursor"]; exists {
		return h.listUsersCursor(c)
	}
	return h.listUsersOffset(c)
}

func (h *UserHandler) listUsersOffset(c fiber.Ctx) error {
	q := presenter.ParseOffsetQuery(c)

	result, err := h.svc.ListUsersOffset(c.Context(), q.Limit, q.Offset)
	if err != nil {
		return err
	}

	items := make([]UserResponse, len(result.Users))
	for i, u := range result.Users {
		items[i] = NewUserResponse(u)
	}

	pg := presenter.NewOffsetPagination(int(result.Total), q.Limit, q.Offset)
	h.logger.Info("Listed users (offset)", "page", q.Page, "total", result.Total)
	return presenter.RenderList(c, items, pg)
}

func (h *UserHandler) listUsersCursor(c fiber.Ctx) error {
	q := presenter.ParseCursorQuery(c)

	afterID := uuid.Nil
	if q.Cursor != "" {
		var err error
		afterID, err = uuid.Parse(q.Cursor)
		if err != nil {
			return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "cursor ไม่ถูกต้อง"))
		}
	}

	result, err := h.svc.ListUsersCursor(c.Context(), afterID, q.Limit)
	if err != nil {
		return err
	}

	items := make([]UserResponse, len(result.Users))
	for i, u := range result.Users {
		items[i] = NewUserResponse(u)
	}

	pg := presenter.NewCursorPagination(result.NextCursor, result.HasMore, q.Limit)
	h.logger.Info("Listed users (cursor)", "cursor", q.Cursor, "count", len(items))
	return presenter.RenderList(c, items, pg)
}

// ChangePassword — PUT /api/v1/me/password
func (h *UserHandler) ChangePassword(c fiber.Ctx) error {
	claims := auth.GetClaims(c)
	if claims == nil {
		return custom_errors.New(401, custom_errors.ErrUnauthorized, "ไม่พบข้อมูล Token")
	}
	userID, err := uuid.Parse(claims.GetUserID())
	if err != nil {
		return custom_errors.New(401, custom_errors.ErrUnauthorized, "Token มี UserID ไม่ถูกต้อง")
	}

	req := new(ChangePasswordRequest)
	if bindErr := c.Bind().Body(req); bindErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง"))
	}
	if vResult := validation.Validate(h.validate, req); !vResult.IsValid {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", vResult.Errors))
	}

	if svcErr := h.svc.ChangePassword(c.Context(), userID, req.CurrentPassword, req.NewPassword); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	h.logger.Info("Password changed", "userID", userID)
	return presenter.RenderItem(c, fiber.Map{"message": "เปลี่ยนรหัสผ่านสำเร็จ"})
}
