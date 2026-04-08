package authhdl

// LoginRequest is the request body for POST /api/v1/auth/login.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email" vmsg:"required:กรุณาระบุอีเมล,email:รูปแบบอีเมลไม่ถูกต้อง"`
	Password string `json:"password" validate:"required"       vmsg:"required:กรุณาระบุรหัสผ่าน"`
}

// RefreshTokenRequest is the request body for POST /api/v1/auth/refresh.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" vmsg:"required:กรุณาระบุ Refresh Token"`
}
