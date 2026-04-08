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
)

type AuthHandler struct {
	userSvc  usersvc.UserService
	authSvc  auth.AuthService
	validate *validator.Validate
	logger   logger.Logger
}

func NewAuthHandler(userSvc usersvc.UserService, authSvc auth.AuthService, l logger.Logger) *AuthHandler {
	if l == nil {
		l = logger.Get()
	}
	return &AuthHandler{
		userSvc:  userSvc,
		authSvc:  authSvc,
		validate: validation.New(),
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
		DisplayName:  result.User.DisplayName,
		Role:         "user",
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

	// Issue a new token pair with the same userID
	tokenPair, err := h.authSvc.GenerateTokenPair(claims.GetUserID(), "user")
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
