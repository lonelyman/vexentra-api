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

// AdminCreateUserRequest is the request body for POST /api/v1/users (admin only).
type AdminCreateUserRequest struct {
	Email       string `json:"email"        validate:"required,email"                             vmsg:"required:กรุณาระบุอีเมล,email:รูปแบบอีเมลไม่ถูกต้อง"`
	Password    string `json:"password"     validate:"required,min=8"                             vmsg:"required:กรุณาระบุรหัสผ่าน,min:รหัสผ่านต้องมีอย่างน้อย 8 ตัวอักษร"`
	DisplayName string `json:"display_name" validate:"required"                                   vmsg:"required:กรุณาระบุชื่อ"`
	Role        string `json:"role"         validate:"omitempty,oneof=member manager admin"       vmsg:"oneof:role ต้องเป็น member | manager | admin"`
}

// AdminUpdateUserRequest is the request body for PATCH /api/v1/users/:id (admin only).
// Both fields are optional — send only the ones you want to change.
type AdminUpdateUserRequest struct {
	Role   string `json:"role"   validate:"omitempty,oneof=member manager admin"                    vmsg:"oneof:role ต้องเป็น member | manager | admin"`
	Status string `json:"status" validate:"omitempty,oneof=active banned pending_verification"    vmsg:"oneof:status ต้องเป็น active | banned | pending_verification"`
}
