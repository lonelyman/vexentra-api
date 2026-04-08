package userhdl

import (
	"time"
	"vexentra-api/internal/modules/user"
)

type UserResponse struct {
	ID          string    `json:"id"` // UUID as string
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewUserResponse(u *user.User) UserResponse {
	return UserResponse{
		ID:          u.ID.String(),
		Username:    u.Username,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		CreatedAt:   u.CreatedAt,
	}
}

// RegisterResponse is the response body for POST /users/register.
// Returns user profile and token pair so the client can immediately authenticate.
type RegisterResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}
