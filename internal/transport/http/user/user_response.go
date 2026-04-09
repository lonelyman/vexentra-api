package userhdl

import (
	"time"
	"vexentra-api/internal/modules/user"
)

type UserResponse struct {
	ID              string     `json:"id"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	Status          string     `json:"status"`
	IsEmailVerified bool       `json:"is_email_verified"`
	LastLoginAt     *time.Time `json:"last_login_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

func NewUserResponse(u *user.User) UserResponse {
	return UserResponse{
		ID:              u.ID.String(),
		Username:        u.Username,
		Email:           u.Email,
		Status:          u.Status,
		IsEmailVerified: u.IsEmailVerified,
		LastLoginAt:     u.LastLoginAt,
		CreatedAt:       u.CreatedAt,
	}
}

// RegisterResponse is the response body for POST /users/register.
// Returns user profile and token pair so the client can immediately authenticate.
type RegisterResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}
