package userhdl

import (
	"time"
	"vexentra-api/internal/modules/user"
)

type UserResponse struct {
	ID              string     `json:"id"`
	PersonID        string     `json:"person_id"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	Role            string     `json:"role"`
	Status          string     `json:"status"`
	IsEmailVerified bool       `json:"is_email_verified"`
	LastLoginAt     *time.Time `json:"last_login_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

func NewUserResponse(u *user.User) UserResponse {
	return UserResponse{
		ID:              u.ID.String(),
		PersonID:        u.PersonID.String(),
		Username:        u.Username,
		Email:           u.Email,
		Role:            u.Role,
		Status:          u.Status,
		IsEmailVerified: u.IsEmailVerified,
		LastLoginAt:     u.LastLoginAt,
		CreatedAt:       u.CreatedAt,
	}
}

// RegisterResponse is the response body for POST /users/register.
// Returns user profile and token pair so the client can immediately authenticate.
// ClaimSuggestion มีเมื่อ email ตรงกับ InviteEmail ของ Person — frontend แสดง dialog ถามก่อน claim
type RegisterResponse struct {
	User            UserResponse             `json:"user"`
	AccessToken     string                   `json:"access_token"`
	RefreshToken    string                   `json:"refresh_token"`
	ClaimSuggestion *ClaimSuggestionResponse `json:"claim_suggestion,omitempty"`
}

// ClaimSuggestionResponse ส่งกลับเพื่อให้ frontend แสดง dialog ถามว่าต้องการ claim ไหม
type ClaimSuggestionResponse struct {
	PersonID string `json:"person_id"`
	Name     string `json:"name"`
}
