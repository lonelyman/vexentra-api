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

// ResendVerifyEmailRequest is the request body for POST /api/v1/auth/resend-verify.
type ResendVerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email" vmsg:"required:กรุณาระบุอีเมล,email:รูปแบบอีเมลไม่ถูกต้อง"`
}

// ForgotPasswordRequest is the request body for POST /api/v1/auth/forgot-password.
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email" vmsg:"required:กรุณาระบุอีเมล,email:รูปแบบอีเมลไม่ถูกต้อง"`
}

// ResetPasswordRequest is the request body for POST /api/v1/auth/reset-password.
type ResetPasswordRequest struct {
	Token       string `json:"token"        validate:"required"                  vmsg:"required:กรุณาระบุ token"`
	NewPassword string `json:"new_password" validate:"required,strong_password" vmsg:"required:กรุณาระบุรหัสผ่านใหม่"`
}
