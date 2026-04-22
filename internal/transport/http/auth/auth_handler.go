package authhdl

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

type AuthHandler struct {
	userSvc  usersvc.UserService
	authSvc  auth.AuthService
	validate *validator.Validate
	isDev    bool
	logger   logger.Logger
}

func NewAuthHandler(userSvc usersvc.UserService, authSvc auth.AuthService, env string, l logger.Logger) *AuthHandler {
	if l == nil {
		l = logger.Get()
	}
	return &AuthHandler{
		userSvc:  userSvc,
		authSvc:  authSvc,
		validate: validation.New(),
		isDev:    env != "production",
		logger:   l,
	}
}

// Login godoc
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c fiber.Ctx) error {
	req := new(LoginRequest)

	if err := c.Bind().Body(req); err != nil {
		h.logger.Error("Failed to bind login request", err)
		return presenter.RenderError(c, custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง"))
	}

	if vResult := validation.Validate(h.validate, req); !vResult.IsValid {
		h.logger.Warn("Login validation failed", "errors", vResult.Errors)
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", vResult.Errors))
	}

	result, err := h.userSvc.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return err
	}

	h.logger.Success("User logged in", "userID", result.User.ID)
	return presenter.RenderItem(c, LoginResponse{
		UserID:       result.User.ID.String(),
		Email:        result.User.Email,
		Role:         result.User.Role,
		AccessToken:  result.TokenPair.AccessToken,
		RefreshToken: result.TokenPair.RefreshToken,
	})
}

// RefreshToken godoc
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c fiber.Ctx) error {
	req := new(RefreshTokenRequest)

	if err := c.Bind().Body(req); err != nil {
		h.logger.Error("Failed to bind refresh request", err)
		return presenter.RenderError(c, custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง"))
	}

	if vResult := validation.Validate(h.validate, req); !vResult.IsValid {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", vResult.Errors))
	}

	// Validate the refresh token
	claims, err := h.authSvc.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.Warn("Invalid refresh token", "err", err)
		return custom_errors.New(401, custom_errors.ErrUnauthorized, "Refresh Token ไม่ถูกต้องหรือหมดอายุ")
	}

	// Issue a new token pair with the same userID + personID
	tokenPair, err := h.authSvc.GenerateTokenPair(claims.GetUserID(), claims.GetPersonID(), "user")
	if err != nil {
		h.logger.Error("Failed to generate token pair on refresh", err)
		return custom_errors.NewInternalError("ไม่สามารถออก Token ใหม่ได้")
	}

	h.logger.Info("Token refreshed", "userID", claims.GetUserID())
	return presenter.RenderItem(c, TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	})
}

// Logout godoc
// POST /api/v1/auth/logout — requires valid Access Token (protected route)
//
// NOTE: This is currently stateless logout — the client must discard both tokens.
// For server-side token revocation, store the Access Token's jti (claims.ID)
// in Redis with TTL = remaining token lifetime, then check it in AuthMiddleware.
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	claims := auth.GetClaims(c)
	if claims == nil {
		return custom_errors.New(401, custom_errors.ErrUnauthorized, "ไม่พบข้อมูล Token")
	}

	h.logger.Info("User logged out", "userID", claims.GetUserID(), "jti", claims.ID)

	// TODO: blacklist claims.ID (jti) in Redis for server-side revocation
	// ttl := time.Until(claims.ExpiresAt.Time)
	// redisClient.Set(ctx, "blacklist:"+claims.ID, 1, ttl)

	return presenter.RenderItem(c, fiber.Map{"message": "ออกจากระบบสำเร็จ"})
}

// VerifyEmail godoc
// GET /api/v1/auth/verify-email?token=<token>
func (h *AuthHandler) VerifyEmail(c fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "กรุณาระบุ token"))
	}
	if err := h.userSvc.VerifyEmail(c.Context(), token); err != nil {
		return presenter.RenderError(c, err)
	}
	h.logger.Info("Email verified successfully")
	return presenter.RenderItem(c, fiber.Map{"message": "ยืนยันอีเมลสำเร็จ"})
}

// ResendVerifyEmail godoc
// POST /api/v1/auth/resend-verify — requires valid Access Token
func (h *AuthHandler) ResendVerifyEmail(c fiber.Ctx) error {
	claims := auth.GetClaims(c)
	if claims == nil {
		return custom_errors.New(401, custom_errors.ErrUnauthorized, "ไม่พบข้อมูล Token")
	}
	userID, err := uuid.Parse(claims.GetUserID())
	if err != nil {
		return custom_errors.New(401, custom_errors.ErrUnauthorized, "Token มี UserID ไม่ถูกต้อง")
	}
	token, svcErr := h.userSvc.ResendVerifyEmail(c.Context(), userID)
	if svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}
	h.logger.Info("Verification email resent", "userID", userID)
	resp := fiber.Map{"message": "ส่งอีเมลยืนยันใหม่แล้ว"}
	if h.isDev {
		resp["token"] = token // dev only: ใช้ทดสอบเป็น email flow แทน email service จริง
	}
	return presenter.RenderItem(c, resp)
}

// ForgotPassword godoc
// POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c fiber.Ctx) error {
	req := new(ForgotPasswordRequest)
	if err := c.Bind().Body(req); err != nil {
		return presenter.RenderError(c, custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง"))
	}
	if vResult := validation.Validate(h.validate, req); !vResult.IsValid {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", vResult.Errors))
	}
	token, svcErr := h.userSvc.ForgotPassword(c.Context(), req.Email)
	if svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}
	h.logger.Info("Forgot password requested", "email", req.Email)
	// Always return success to prevent user enumeration
	resp := fiber.Map{"message": "หากอีเมลนี้มีในระบบ คุณจะได้รับลิงก์รีเซ็ตรหัสผ่าน"}
	if h.isDev && token != "" {
		resp["token"] = token // dev only: ใช้ทดสอบเป็น reset flow แทน email service จริง
	}
	return presenter.RenderItem(c, resp)
}

// ResetPassword godoc
// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c fiber.Ctx) error {
	req := new(ResetPasswordRequest)
	if err := c.Bind().Body(req); err != nil {
		return presenter.RenderError(c, custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง"))
	}
	if vResult := validation.Validate(h.validate, req); !vResult.IsValid {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", vResult.Errors))
	}
	if err := h.userSvc.ResetPassword(c.Context(), req.Token, req.NewPassword); err != nil {
		return presenter.RenderError(c, err)
	}
	h.logger.Info("Password reset successfully")
	return presenter.RenderItem(c, fiber.Map{"message": "รีเซ็ตรหัสผ่านสำเร็จ กรุณา login ใหม่"})
}
