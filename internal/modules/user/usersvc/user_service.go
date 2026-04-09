package usersvc

import (
	"context"
	"strings"
	"time"

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

// ListUsersOffsetResult bundles users + total count for offset pagination.
type ListUsersOffsetResult struct {
	Users []*user.User
	Total int64
}

// ListUsersCursorResult bundles users + cursor metadata.
type ListUsersCursorResult struct {
	Users      []*user.User
	HasMore    bool
	NextCursor string // ID of last item; empty when HasMore=false
}

type UserService interface {
	Register(ctx context.Context, email, password string) (*RegisterResult, error)
	Login(ctx context.Context, email, password string) (*RegisterResult, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*user.User, error)
	ListUsersOffset(ctx context.Context, limit, offset int) (*ListUsersOffsetResult, error)
	ListUsersCursor(ctx context.Context, afterID uuid.UUID, limit int) (*ListUsersCursorResult, error)
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

func (s *userService) Register(ctx context.Context, email, password string) (*RegisterResult, error) {
	// auto-generate username จาก local-part ของ email (ก่อน @)
	username := strings.SplitN(email, "@", 2)[0]

	s.logger.Info("Starting user registration", "username", username, "email", email)

	// 1. ตรวจ email ซ้ำ
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		s.logger.Warn("Registration rejected: email already in use", "email", email)
		return nil, custom_errors.New(409, custom_errors.ErrAlreadyExists, "อีเมลนี้ถูกใช้งานแล้ว")
	}

	// 3. Hash password
	hashedPassword, err := s.authSvc.HashPassword(password)
	if err != nil {
		s.logger.Error("Failed to hash password", err)
		return nil, custom_errors.NewInternalError("ไม่สามารถประมวลผล password ได้")
	}

	// 4. สร้าง User (สถานะ pending_verification รอยืนยันอีเมล)
	newUser := &user.User{
		Username: username,
		Email:    email,
		Status:   user.UserStatusPendingVerification,
	}

	if err := s.repo.Create(ctx, newUser); err != nil {
		s.logger.Error("Failed to persist new user", err, "username", username)
		return nil, err
	}

	// 5. สร้าง UserAuth แยก (local provider)
	newAuth := &user.UserAuth{
		UserID:     newUser.ID,
		Provider:   user.AuthProviderLocal,
		ProviderID: email, // สำหรับ local ใช้ email เป็น identifier
		Secret:     hashedPassword,
	}
	if err := s.repo.CreateAuth(ctx, newAuth); err != nil {
		s.logger.Error("Failed to persist user auth", err, "userID", newUser.ID)
		return nil, err
	}

	// 6. ออก token pair (auto-login หลังสมัคร)
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

	// 1. ดึง user + local auth ในครั้งเดียว
	u, localAuth, err := s.repo.GetByEmailWithLocalAuth(ctx, email)
	if err != nil {
		return nil, err
	}
	// Generic error — ป้องกัน user enumeration
	if u == nil || localAuth == nil {
		s.logger.Warn("Login failed: user or local auth not found", "email", email)
		return nil, custom_errors.New(401, custom_errors.ErrUnauthorized, "อีเมลหรือรหัสผ่านไม่ถูกต้อง")
	}

	// 2. ตรวจสอบสถานะบัญชี
	if u.Status == user.UserStatusBanned {
		s.logger.Warn("Login failed: account banned", "userID", u.ID)
		return nil, custom_errors.New(403, custom_errors.ErrForbidden, "บัญชีนี้ถูกระงับการใช้งาน")
	}

	// 3. ตรวจสอบ password
	if err := s.authSvc.ComparePassword(localAuth.Secret, password); err != nil {
		s.logger.Warn("Login failed: wrong password", "email", email)
		return nil, custom_errors.New(401, custom_errors.ErrUnauthorized, "อีเมลหรือรหัสผ่านไม่ถูกต้อง")
	}

	// 4. ออก token pair
	tokenPair, err := s.authSvc.GenerateTokenPair(u.ID.String(), "user")
	if err != nil {
		s.logger.Error("Failed to generate token pair on login", err, "userID", u.ID)
		return nil, custom_errors.NewInternalError("ไม่สามารถออก Token ได้")
	}

	// 5. บันทึกเวลา login ล่าสุด (best-effort — ไม่ block login ถ้าล้มเหลว)
	now := time.Now()
	if err := s.repo.UpdateLastLogin(ctx, u.ID, now); err != nil {
		s.logger.Warn("Failed to update last_login_at", "userID", u.ID)
	}
	u.LastLoginAt = &now

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

func (s *userService) ListUsersOffset(ctx context.Context, limit, offset int) (*ListUsersOffsetResult, error) {
	s.logger.Info("Listing users (offset)", "limit", limit, "offset", offset)

	users, total, err := s.repo.ListOffset(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list users (offset)", err)
		return nil, err
	}

	return &ListUsersOffsetResult{Users: users, Total: total}, nil
}

func (s *userService) ListUsersCursor(ctx context.Context, afterID uuid.UUID, limit int) (*ListUsersCursorResult, error) {
	s.logger.Info("Listing users (cursor)", "afterID", afterID, "limit", limit)

	// Fetch limit+1 to detect whether more pages exist
	users, err := s.repo.ListAfterCursor(ctx, afterID, limit+1)
	if err != nil {
		s.logger.Error("Failed to list users (cursor)", err)
		return nil, err
	}

	hasMore := len(users) > limit
	if hasMore {
		users = users[:limit] // trim the extra item
	}

	nextCursor := ""
	if hasMore && len(users) > 0 {
		nextCursor = users[len(users)-1].ID.String()
	}

	return &ListUsersCursorResult{
		Users:      users,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}
