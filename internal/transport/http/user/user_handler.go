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

	result, err := h.svc.Register(c.Context(), req.Username, req.Email, req.Password, req.DisplayName)
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
