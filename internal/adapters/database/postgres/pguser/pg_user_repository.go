package pguser

import (
	"context"
	"vexentra-api/internal/modules/user"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) user.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *user.User) error {
	model := fromUserEntity(u)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	// 💡 Sync กลับคืนสู่ Domain Entity (Logic Gap Fix)
	u.ID = model.ID
	u.CreatedAt = model.CreatedAt
	u.UpdatedAt = model.UpdatedAt

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return m.ToEntity(), nil
}

func AutoMigrate(db *gorm.DB) error {
	// ใส่ Model ทุกตัวที่ต้องการสร้างตารางตรงนี้ค่ะ
	return db.AutoMigrate(&userModel{})
}
