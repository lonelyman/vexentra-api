package pguser

import (
	"context"
	"errors"
	"strings"
	"time"
	"vexentra-api/internal/adapters/database/postgres/pgtx"
	"vexentra-api/internal/modules/user"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

func NewUserRepository(db *gorm.DB, l logger.Logger) user.UserRepository {
	if l == nil {
		l = logger.Get()
	}
	return &userRepository{
		db:     db,
		logger: l,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
//  User CRUD
// ─────────────────────────────────────────────────────────────────────────────

func (r *userRepository) Create(ctx context.Context, u *user.User) error {
	id, err := uuid.NewV7()
	if err != nil {
		r.logger.Error("UUID_GENERATE_ERROR", err)
		return custom_errors.NewInternalError("ไม่สามารถสร้าง ID ได้")
	}
	u.ID = id
	model := fromUserEntity(u)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.Error("DB_CREATE_ERROR", err)
		return err
	}

	u.CreatedAt = model.CreatedAt
	u.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom_errors.NewNotFoundError(custom_errors.ErrNotFound, "user not found")
		}
		r.logger.Error("DB_GET_BY_ID_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *userRepository) GetByPersonID(ctx context.Context, personID uuid.UUID) (*user.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).Where("person_id = ?", personID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("DB_GET_BY_PERSON_ID_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("DB_GET_BY_EMAIL_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("DB_GET_BY_USERNAME_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

// ─────────────────────────────────────────────────────────────────────────────
//  สถานะ & การ tracking
// ─────────────────────────────────────────────────────────────────────────────

func (r *userRepository) UpdateStatus(ctx context.Context, userID uuid.UUID, status string) error {
	if err := r.db.WithContext(ctx).Model(&userModel{}).
		Where("id = ?", userID).
		Update("status", status).Error; err != nil {
		r.logger.Error("DB_UPDATE_STATUS_ERROR", err)
		return err
	}
	return nil
}

func (r *userRepository) UpdateRole(ctx context.Context, userID uuid.UUID, role string) error {
	if err := r.db.WithContext(ctx).Model(&userModel{}).
		Where("id = ?", userID).
		Update("role", role).Error; err != nil {
		r.logger.Error("DB_UPDATE_ROLE_ERROR", err)
		return err
	}
	return nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, t time.Time) error {
	if err := r.db.WithContext(ctx).Model(&userModel{}).
		Where("id = ?", userID).
		Update("last_login_at", t).Error; err != nil {
		r.logger.Error("DB_UPDATE_LAST_LOGIN_ERROR", err)
		return err
	}
	return nil
}

func (r *userRepository) UpdatePersonID(ctx context.Context, userID, personID uuid.UUID) error {
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).Model(&userModel{}).
		Where("id = ?", userID).
		Update("person_id", personID).Error; err != nil {
		r.logger.Error("DB_UPDATE_PERSON_ID_ERROR", err)
		return custom_errors.NewInternalError("ไม่สามารถอัปเดต person_id ได้")
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
//  Email Verification
// ─────────────────────────────────────────────────────────────────────────────

func (r *userRepository) SetEmailVerificationToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	if err := r.db.WithContext(ctx).Model(&userModel{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"email_verification_token":            token,
			"email_verification_token_expires_at": expiresAt,
		}).Error; err != nil {
		r.logger.Error("DB_SET_VERIFY_TOKEN_ERROR", err)
		return err
	}
	return nil
}

func (r *userRepository) GetByEmailVerificationToken(ctx context.Context, token string) (*user.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).
		Where("email_verification_token = ?", token).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("DB_GET_BY_VERIFY_TOKEN_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *userRepository) SetEmailVerified(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&userModel{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"is_email_verified":                   true,
			"email_verification_token":            nil,
			"email_verification_token_expires_at": nil,
			"status":                              user.UserStatusActive,
		}).Error; err != nil {
		r.logger.Error("DB_SET_EMAIL_VERIFIED_ERROR", err)
		return err
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
//  Password Reset
// ─────────────────────────────────────────────────────────────────────────────

func (r *userRepository) SetPasswordResetToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	if err := r.db.WithContext(ctx).Model(&userModel{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"password_reset_token":            token,
			"password_reset_token_expires_at": expiresAt,
		}).Error; err != nil {
		r.logger.Error("DB_SET_RESET_TOKEN_ERROR", err)
		return err
	}
	return nil
}

func (r *userRepository) GetByPasswordResetToken(ctx context.Context, token string) (*user.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).
		Where("password_reset_token = ?", token).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("DB_GET_BY_RESET_TOKEN_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *userRepository) ClearPasswordResetToken(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&userModel{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"password_reset_token":            nil,
			"password_reset_token_expires_at": nil,
		}).Error; err != nil {
		r.logger.Error("DB_CLEAR_RESET_TOKEN_ERROR", err)
		return err
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
//  UserAuth (multi-provider)
// ─────────────────────────────────────────────────────────────────────────────

func (r *userRepository) CreateAuth(ctx context.Context, a *user.UserAuth) error {
	id, err := uuid.NewV7()
	if err != nil {
		r.logger.Error("UUID_GENERATE_ERROR", err)
		return custom_errors.NewInternalError("ไม่สามารถสร้าง ID ได้")
	}
	a.ID = id
	model := fromUserAuthEntity(a)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.Error("DB_CREATE_AUTH_ERROR", err)
		return err
	}
	return nil
}

func (r *userRepository) GetAuthByUserAndProvider(ctx context.Context, userID uuid.UUID, provider string) (*user.UserAuth, error) {
	var m userAuthModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, provider).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("DB_GET_AUTH_BY_PROVIDER_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

// GetByEmailWithLocalAuth ดึง user + local auth ในครั้งเดียว — ใช้ตอน Login
func (r *userRepository) GetByEmailWithLocalAuth(ctx context.Context, email string) (*user.User, *user.UserAuth, error) {
	var u userModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil
		}
		r.logger.Error("DB_GET_BY_EMAIL_ERROR", err)
		return nil, nil, err
	}

	var a userAuthModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", u.ID, user.AuthProviderLocal).
		First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return u.ToEntity(), nil, nil
		}
		r.logger.Error("DB_GET_LOCAL_AUTH_ERROR", err)
		return nil, nil, err
	}

	return u.ToEntity(), a.ToEntity(), nil
}

func (r *userRepository) UpdateLocalAuthSecret(ctx context.Context, userID uuid.UUID, secret string) error {
	if err := r.db.WithContext(ctx).Model(&userAuthModel{}).
		Where("user_id = ? AND provider = ?", userID, user.AuthProviderLocal).
		Update("secret", secret).Error; err != nil {
		r.logger.Error("DB_UPDATE_LOCAL_SECRET_ERROR", err)
		return err
	}
	return nil
}

func (r *userRepository) UpdateAuthRefreshToken(ctx context.Context, authID uuid.UUID, token *string) error {
	if err := r.db.WithContext(ctx).Model(&userAuthModel{}).
		Where("id = ?", authID).
		Update("refresh_token", token).Error; err != nil {
		r.logger.Error("DB_UPDATE_REFRESH_TOKEN_ERROR", err)
		return err
	}
	return nil
}

func (r *userRepository) SetForcePasswordChange(ctx context.Context, userID uuid.UUID, required bool) error {
	if err := r.db.WithContext(ctx).Model(&userModel{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"force_password_change": required,
			"password_changed_at":   nil,
		}).Error; err != nil {
		r.logger.Error("DB_SET_FORCE_PASSWORD_CHANGE_ERROR", err)
		return err
	}
	return nil
}

func (r *userRepository) MarkPasswordChanged(ctx context.Context, userID uuid.UUID, changedAt time.Time) error {
	if err := r.db.WithContext(ctx).Model(&userModel{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"password_changed_at":             changedAt,
			"force_password_change":           false,
			"password_reset_token":            nil,
			"password_reset_token_expires_at": nil,
		}).Error; err != nil {
		r.logger.Error("DB_MARK_PASSWORD_CHANGED_ERROR", err)
		return err
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
//  Pagination
// ─────────────────────────────────────────────────────────────────────────────

func (r *userRepository) ListOffset(ctx context.Context, limit, offset int, search, status string) ([]*user.User, int64, error) {
	var models []userModel
	var total int64

	base := r.db.WithContext(ctx).Model(&userModel{})
	base = applyUserListFilter(base, search, status)

	if err := base.Count(&total).Error; err != nil {
		r.logger.Error("DB_LIST_COUNT_ERROR", err)
		return nil, 0, err
	}

	if err := base.Order("created_at DESC, id DESC").
		Limit(limit).Offset(offset).
		Find(&models).Error; err != nil {
		r.logger.Error("DB_LIST_OFFSET_ERROR", err)
		return nil, 0, err
	}

	users := make([]*user.User, len(models))
	for i := range models {
		users[i] = models[i].ToEntity()
	}
	return users, total, nil
}

func (r *userRepository) ListAfterCursor(ctx context.Context, afterID uuid.UUID, limit int, search, status string) ([]*user.User, error) {
	var models []userModel

	q := r.db.WithContext(ctx).Model(&userModel{}).Order("users.id ASC").Limit(limit)
	q = applyUserListFilter(q, search, status)
	if afterID != uuid.Nil {
		q = q.Where("users.id > ?", afterID)
	}

	if err := q.Find(&models).Error; err != nil {
		r.logger.Error("DB_LIST_CURSOR_ERROR", err)
		return nil, err
	}

	users := make([]*user.User, len(models))
	for i := range models {
		users[i] = models[i].ToEntity()
	}
	return users, nil
}

func applyUserListFilter(q *gorm.DB, search, status string) *gorm.DB {
	search = strings.TrimSpace(search)
	status = strings.TrimSpace(strings.ToLower(status))

	if status != "" {
		q = q.Where("users.status = ?", status)
	}
	if search != "" {
		pattern := "%" + search + "%"
		q = q.Joins("LEFT JOIN profiles ON profiles.person_id = users.person_id").
			Where(
				`users.username ILIKE ?
OR users.email ILIKE ?
OR COALESCE(profiles.display_name, '') ILIKE ?
OR COALESCE(profiles.headline, '') ILIKE ?`,
				pattern, pattern, pattern, pattern,
			)
	}
	return q
}
