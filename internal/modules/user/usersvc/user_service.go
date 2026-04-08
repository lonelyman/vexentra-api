package usersvc

import (
	"context"

	"vexentra-api/internal/modules/user"
	"vexentra-api/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, username, email, password, displayName string) (*user.User, error)
}

type userService struct {
	repo   user.UserRepository
	logger logger.Logger
}

// NewUserService รับ Logger เป็นตัวสุดท้ายเสมอตามกฎของนายท่าน
func NewUserService(repo user.UserRepository, l logger.Logger) UserService {
	if l == nil {
		l = logger.Get() // Hybrid: ถ้าเป็น nil ให้ดึงจาก Global
	}
	return &userService{
		repo:   repo,
		logger: l,
	}
}

func (s *userService) Register(ctx context.Context, username, email, password, displayName string) (*user.User, error) {
	// 1. บันทึกข้อมูลการเริ่มต้นทำงาน
	s.logger.Info("Starting user registration process", "username", username, "email", email)

	// 2. Hash Password (Security First)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password during registration", err)
		return nil, err
	}

	// 3. สร้าง Entity (Mapping ข้อมูลจาก Input ไปยัง User Entity)
	newUser := &user.User{
		Username:    username,
		Email:       email,
		Password:    string(hashedPassword),
		DisplayName: displayName,
		IsActive:    true, // ตั้งค่าเริ่มต้นที่ชั้น Service
	}

	// 4. บันทึกลง Database ผ่าน Repository
	if err := s.repo.Create(ctx, newUser); err != nil {
		s.logger.Error("Database persistence failed for user", err, "username", username)
		return nil, err
	}

	// 5. บันทึกความสำเร็จ
	s.logger.Success("User registered successfully in system", "username", username)

	return newUser, nil
}
