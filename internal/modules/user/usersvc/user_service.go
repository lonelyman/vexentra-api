package usersvc

import (
	"context"

	"golang.org/x/crypto/bcrypt"
	"vexentra-api/internal/modules/user"
)

type UserService interface {
	// 2. เปลี่ยนจาก *User เป็น *user.User
	Register(ctx context.Context, username, email, password, displayName string) (*user.User, error)
}

type userService struct {
	// 3. เปลี่ยนจาก UserRepository เป็น user.UserRepository
	repo user.UserRepository
}

func NewUserService(repo user.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(ctx context.Context, username, email, password, displayName string) (*user.User, error) {
	// 1. Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 2. สร้าง Entity (เรียกผ่าน user.User)
	newUser := &user.User{
		Username:    username,
		Email:       email,
		Password:    string(hashedPassword),
		DisplayName: displayName,
		IsActive:    true,
	}

	// 3. บันทึกลง Database ผ่าน Repo
	if err := s.repo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}
