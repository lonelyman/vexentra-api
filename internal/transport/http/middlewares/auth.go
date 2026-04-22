// internal/transport/http/middlewares/auth.go
package middlewares

import (
	"strings"
	"vexentra-api/pkg/auth"
	"vexentra-api/pkg/custom_errors"

	"github.com/gofiber/fiber/v3"
)

func RoleMiddleware(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c fiber.Ctx) error {
		claims, ok := c.Locals("user_claims").(*auth.AccessClaims)
		if !ok || claims == nil {
			return custom_errors.New(401, custom_errors.ErrUnauthorized, "กรุณาเข้าสู่ระบบ")
		}
		if _, permitted := allowed[claims.Role]; !permitted {
			return custom_errors.New(403, "FORBIDDEN", "คุณไม่มีสิทธิ์เข้าถึงส่วนนี้")
		}
		return c.Next()
	}
}

func AuthMiddleware(authSvc auth.AuthService) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return custom_errors.New(401, custom_errors.ErrUnauthorized, "กรุณาส่ง Authorization Header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return custom_errors.New(401, "INVALID_TOKEN_FORMAT", "รูปแบบ Token ต้องเป็น Bearer [token]")
		}

		claims, err := authSvc.ValidateAccessToken(parts[1])
		if err != nil {
			return err
		}

		// เก็บ *AccessClaims ลง Locals — ดึงออกไปใช้ใน handler ด้วย auth.GetClaims(c)
		c.Locals("user_claims", claims)

		return c.Next()
	}
}
