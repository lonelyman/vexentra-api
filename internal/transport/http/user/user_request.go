package userhdl

// RegisterRequest ใช้รับข้อมูลจาก Frontend พร้อม Tag สำหรับการตรวจสอบ
type RegisterRequest struct {
	Email       string `json:"email"        validate:"required,email"            vmsg:"required:กรุณาระบุอีเมล,email:รูปแบบอีเมลไม่ถูกต้อง"`
	Password    string `json:"password"     validate:"required,strong_password" vmsg:"required:กรุณาระบุรหัสผ่าน"`
	RePassword  string `json:"re_password"  validate:"required,eqfield=Password" vmsg:"required:กรุณายืนยันรหัสผ่าน,eqfield:รหัสผ่านไม่ตรงกัน"`
	InviteToken string `json:"invite_token"` // optional — สำหรับ invite link flow
}

// ClaimPersonRequest ใช้ยืนยันการผูก Person ที่ระบบ suggest หลัง Register
type ClaimPersonRequest struct {
	PersonID string `json:"person_id" validate:"required,uuid" vmsg:"required:กรุณาระบุ person_id,uuid:รูปแบบ UUID ไม่ถูกต้อง"`
}

// ChangePasswordRequest is the request body for PUT /api/v1/me/password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"               vmsg:"required:กรุณาระบุรหัสผ่านปัจจุบัน"`
	NewPassword     string `json:"new_password"     validate:"required,strong_password" vmsg:"required:กรุณาระบุรหัสผ่านใหม่"`
}
