package usersvc

import (
	"context"

	"vexentra-api/internal/modules/user"
	"vexentra-api/pkg/auth"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
)

// RegisterResult bundles the created user and issued token pair.
// Avoids multi-return positional confusion at the call site.
type RegisterResult struct {
	User      *user.User
	TokenPair *auth.TokenPair
}

type UserService interface {
	Register(ctx context.Context, username, email, password, displayName string) (*RegisterResult, error)
	Login(ctx context.Context, email, password string) (*RegisterResult, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*user.User, error)
}

type userService struct {
	repo    user.UserRepository
	authSvc auth.AuthService
	logger  logger.Logger
}

func NewUserService(repo user.UserRepository, authSvc auth.AuthService, l logger.Logger) UserService {
	if l == nil {
		l = logger.Get()
	}
	return &userService{
		repo:    repo,
		authSvc: authSvc,
		logger:  l,
	}
}

func (s *userService) Register(ctx context.Context, username, email, password, displayName string) (*RegisterResult, error) {
	s.logger.Info("Starting user registration", "username", username, "email", email)

	// 1. Duplicate email check
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		s.logger.Warn("Registration rejected: email already in use", "email", email)
		return nil, custom_errors.New(409, custom_errors.ErrAlreadyExists, "อีเมลนี้ถูกใช้งานแล้ว")
	}

	// 2. Duplicate username check
	existingByUsername, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existingByUsername != nil {
		s.logger.Warn("Registration rejected: username already in use", "username", username)
		return nil, custom_errors.New(409, custom_errors.ErrAlreadyExists, "Username นี้ถูกใช้งานแล้ว")
	}

	// 2. Hash password via AuthService (single source of bcrypt logic)
	hashedPassword, err := s.authSvc.HashPassword(password)
	if err != nil {
		s.logger.Error("Failed to hash password", err)
		return nil, custom_errors.NewInternalError("ไม่สามารถประมวลผล password ได้")
	}

	// 3. Build domain entity
	newUser := &user.User{
		Username:    username,
		Email:       email,
		Password:    hashedPassword,
		DisplayName: displayName,
		IsActive:    true,
	}

	// 4. Persist
	if err := s.repo.Create(ctx, newUser); err != nil {
		s.logger.Error("Failed to persist new user", err, "username", username)
		return nil, err
	}

	// 5. Issue token pair immediately (auto-login after register)
	tokenPair, err := s.authSvc.GenerateTokenPair(newUser.ID.String(), "user")
	if err != nil {
		s.logger.Error("Failed to generate token pair", err, "userID", newUser.ID)
		return nil, custom_errors.NewInternalError("ไม่สามารถออก Token ได้")
	}

	s.logger.Success("User registered successfully", "userID", newUser.ID)

	return &RegisterResult{
		User:      newUser,
		TokenPair: tokenPair,
	}, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (*RegisterResult, error) {
	s.logger.Info("Login attempt", "email", email)

	// 1. Find user by email
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	// Return generic error — do not reveal whether email exists (prevents user enumeration)
	if u == nil {
		s.logger.Warn("Login failed: email not found", "email", email)
		return nil, custom_errors.New(401, custom_errors.ErrUnauthorized, "อีเมลหรือรหัสผ่านไม่ถูกต้อง")
	}

	// 2. Verify password
	if err := s.authSvc.ComparePassword(u.Password, password); err != nil {
		s.logger.Warn("Login failed: wrong password", "email", email)
		return nil, custom_errors.New(401, custom_errors.ErrUnauthorized, "อีเมลหรือรหัสผ่านไม่ถูกต้อง")
	}

	// 3. Issue token pair
	tokenPair, err := s.authSvc.GenerateTokenPair(u.ID.String(), "user")
	if err != nil {
		s.logger.Error("Failed to generate token pair on login", err, "userID", u.ID)
		return nil, custom_errors.NewInternalError("ไม่สามารถออก Token ได้")
	}

	s.logger.Success("User logged in successfully", "userID", u.ID)
	return &RegisterResult{User: u, TokenPair: tokenPair}, nil
}

func (s *userService) GetProfile(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	s.logger.Info("Fetching user profile", "userID", userID)

	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to fetch user profile", err, "userID", userID)
		return nil, err
	}

	return u, nil
}
