package userhdl

// RegisterRequest ใช้รับข้อมูลจาก Frontend พร้อม Tag สำหรับการตรวจสอบ
type RegisterRequest struct {
	Username    string `json:"username"     validate:"required,min=4,max=20"     vmsg:"required:กรุณาระบุ Username,min:Username ต้องมีอย่างน้อย 4 ตัวอักษร,max:Username ต้องไม่เกิน 20 ตัวอักษร"`
	DisplayName string `json:"display_name" validate:"required,min=2,max=50"     vmsg:"required:กรุณาระบุชื่อที่แสดง,min:ชื่อที่แสดงต้องมีอย่างน้อย 2 ตัวอักษร"`
	Email       string `json:"email"        validate:"required,email"            vmsg:"required:กรุณาระบุอีเมล,email:รูปแบบอีเมลไม่ถูกต้อง"`
	Password    string `json:"password"     validate:"required,strong_password" vmsg:"required:กรุณาระบุรหัสผ่าน"`
}
