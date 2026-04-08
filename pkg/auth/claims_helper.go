// pkg/auth/claims_helper.go
package auth

import "github.com/gofiber/fiber/v3"

// GetClaims retrieves the *AccessClaims stored by AuthMiddleware from Fiber's Locals.
// Returns nil if the key is missing or the type assertion fails (e.g. called on a public route).
// Usage in handlers:
//
//	claims := auth.GetClaims(c)
//	if claims == nil { ... } // should not happen on protected routes
func GetClaims(c fiber.Ctx) *AccessClaims {
	val := c.Locals("user_claims")
	if val == nil {
		return nil
	}
	claims, _ := val.(*AccessClaims)
	return claims
}
