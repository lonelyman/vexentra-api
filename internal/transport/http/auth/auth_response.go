package authhdl

// TokenResponse is the response body for /auth/refresh.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// LoginResponse is the response body for /auth/login.
// Returns user identity fields alongside token pair.
type LoginResponse struct {
	UserID              string  `json:"user_id"`
	Email               string  `json:"email"`
	Role                string  `json:"role"`
	ForcePasswordChange bool    `json:"force_password_change"`
	PasswordChangedAt   *string `json:"password_changed_at,omitempty"`
	AccessToken         string  `json:"access_token"`
	RefreshToken        string  `json:"refresh_token"`
}
