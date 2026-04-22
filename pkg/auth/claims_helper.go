// pkg/auth/claims_helper.go
package auth

import (
	"vexentra-api/internal/modules/user"
	"vexentra-api/pkg/custom_errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

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

// GetCaller converts the JWT claims stored by AuthMiddleware into a user.Caller —
// the permission-check primitive used by every service in the project module.
//
// Role is read from the access token (stateless — Option A): a role demotion
// propagates on the next refresh, not mid-token. This is the deliberate trade-off
// chosen for Phase A6. If a handler needs the absolute latest role (e.g. admin
// panel), it should call userSvc.GetProfile directly and not rely on this helper.
//
// Returns a 401 AppError if claims are missing or the ID fields are malformed —
// those cases indicate the middleware stack is misconfigured, not a client error.
func GetCaller(c fiber.Ctx) (user.Caller, error) {
	claims := GetClaims(c)
	if claims == nil {
		return user.Caller{}, custom_errors.New(401, custom_errors.ErrUnauthorized, "ไม่พบข้อมูล Token")
	}
	userID, err := uuid.Parse(claims.GetUserID())
	if err != nil {
		return user.Caller{}, custom_errors.New(401, custom_errors.ErrUnauthorized, "Token มี UserID ไม่ถูกต้อง")
	}
	personID, err := uuid.Parse(claims.GetPersonID())
	if err != nil {
		return user.Caller{}, custom_errors.New(401, custom_errors.ErrUnauthorized, "Token มี PersonID ไม่ถูกต้อง")
	}
	return user.Caller{
		UserID:   userID,
		PersonID: personID,
		Role:     claims.Role,
	}, nil
}
