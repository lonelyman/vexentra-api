package pguser

import (
	"context"
	"errors"
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

func (r *userRepository) Create(ctx context.Context, u *user.User) error {
	// UUID v7: timestamp-prefixed → B-tree index friendly, monotonically increasing
	id, err := uuid.NewV7()
	if err != nil {
		r.logger.Error("UUID_GENERATE_ERROR", err)
		return custom_errors.NewInternalError("ไม่สามารถสร้าง ID ได้")
	}
	u.ID = id
	model := fromUserEntity(u)

	// ตรวจสอบ Error ที่เกิดจาก Database โดยตรง
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.Error("DB_CREATE_ERROR", err) // Log จุดที่เกิดปัญหาจริง
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

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // nil, nil = ไม่พบ แต่ไม่ใช่ error
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
			return nil, nil // nil, nil = ไม่พบ แต่ไม่ใช่ error
		}
		r.logger.Error("DB_GET_BY_USERNAME_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

func AutoMigrate(db *gorm.DB) error {
	// ใส่ Model ทุกตัวที่ต้องการสร้างตารางตรงนี้ค่ะ
	return db.AutoMigrate(&userModel{})
}
