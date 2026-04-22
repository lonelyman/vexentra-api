// pkg/auth/auth_service.go
package auth

import (
	"fmt"
	"time"
	"vexentra-api/internal/config"
	"vexentra-api/pkg/custom_errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ─────────────────────────────────────────────
//  Token Pair
// ─────────────────────────────────────────────

// TokenPair holds both Access and Refresh tokens as a single return value.
// Eliminates positional confusion from multi-return strings.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// ─────────────────────────────────────────────
//  Claims — Two separate structs, one per token type
//  (Prevents Token Confusion Attack per RFC 8725)
// ─────────────────────────────────────────────

// AccessClaims is embedded in short-lived Access Tokens only.
// sub (Subject) = userID per RFC 7519.
// PersonID identifies the Person record linked to this user — used by profile operations.
type AccessClaims struct {
	Role                 string `json:"role"`
	PersonID             string `json:"person_id"`  // UUID of the linked persons record
	TokenType            string `json:"token_type"` // always "access"
	jwt.RegisteredClaims        // carries: sub, iss, jti, iat, exp
}

// GetUserID is a convenience method to read sub without extra type work.
func (c *AccessClaims) GetUserID() string { return c.Subject }

// GetPersonID returns the linked Person UUID from the token claims.
func (c *AccessClaims) GetPersonID() string { return c.PersonID }

// RefreshClaims is embedded in long-lived Refresh Tokens only.
// Intentionally minimal — no Role (refresh endpoints don't authorize resources).
// DeviceID enables per-device revocation for multi-device logout support.
type RefreshClaims struct {
	PersonID             string `json:"person_id"`           // UUID of the linked persons record
	TokenType            string `json:"token_type"`          // always "refresh"
	DeviceID             string `json:"device_id,omitempty"` // optional; set for multi-device tracking
	jwt.RegisteredClaims        // carries: sub, iss, jti, iat, exp
}

// GetUserID is a convenience method to read sub without extra type work.
func (c *RefreshClaims) GetUserID() string { return c.Subject }

// GetPersonID returns the linked Person UUID from the refresh claims.
func (c *RefreshClaims) GetPersonID() string { return c.PersonID }

// ─────────────────────────────────────────────
//  Interface
// ─────────────────────────────────────────────

type AuthService interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, plainPassword string) error
	// GenerateTokenPair creates both tokens in one call.
	// personID is the linked Person UUID — embedded in the access token for profile operations.
	// deviceID is optional — pass it to enable per-device refresh token tracking.
	GenerateTokenPair(userID, personID, role string, deviceID ...string) (*TokenPair, error)
	ValidateAccessToken(tokenString string) (*AccessClaims, error)
	ValidateRefreshToken(tokenString string) (*RefreshClaims, error)
}

// ─────────────────────────────────────────────
//  Implementation
// ─────────────────────────────────────────────

type authService struct {
	cfg config.JWTConfig
}

// NewAuthService accepts the full JWTConfig so all JWT settings
// come from a single source of truth.
func NewAuthService(cfg config.JWTConfig) AuthService {
	return &authService{cfg: cfg}
}

func (s *authService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *authService) ComparePassword(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}

func (s *authService) GenerateTokenPair(userID, personID, role string, deviceID ...string) (*TokenPair, error) {
	now := time.Now()

	// --- Access Token ---
	accessClaims := &AccessClaims{
		Role:      role,
		PersonID:  personID,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), // jti — unique per token, enables future blacklisting
			Subject:   userID,
			Issuer:    s.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.AccessExpiry)),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).
		SignedString([]byte(s.cfg.AccessSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// --- Refresh Token ---
	dID := ""
	if len(deviceID) > 0 {
		dID = deviceID[0]
	}
	refreshClaims := &RefreshClaims{
		PersonID:  personID,
		TokenType: "refresh",
		DeviceID:  dID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), // jti — unique per token, enables future blacklisting
			Subject:   userID,
			Issuer:    s.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.RefreshExpiry)),
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).
		SignedString([]byte(s.cfg.RefreshSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) ValidateAccessToken(tokenString string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.AccessSecret), nil
	})
	if err != nil {
		return nil, custom_errors.New(401, custom_errors.ErrUnauthorized, "Access token ไม่ถูกต้องหรือหมดอายุ")
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid || claims.TokenType != "access" {
		return nil, custom_errors.New(401, custom_errors.ErrUnauthorized, "Access token ไม่สมบูรณ์")
	}
	return claims, nil
}

func (s *authService) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.RefreshSecret), nil
	})
	if err != nil {
		return nil, custom_errors.New(401, custom_errors.ErrUnauthorized, "Refresh token ไม่ถูกต้องหรือหมดอายุ")
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid || claims.TokenType != "refresh" {
		return nil, custom_errors.New(401, custom_errors.ErrUnauthorized, "Refresh token ไม่สมบูรณ์")
	}
	return claims, nil
}
