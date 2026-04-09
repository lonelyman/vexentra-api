package pgsocialplatform

import (
	"context"
	"errors"

	"vexentra-api/internal/modules/socialplatform"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type socialPlatformRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

func NewSocialPlatformRepository(db *gorm.DB, l logger.Logger) socialplatform.SocialPlatformRepository {
	return &socialPlatformRepository{db: db, logger: l}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&socialPlatformModel{})
}

func ResetTable(db *gorm.DB) error {
	if err := db.Migrator().DropTable(&socialPlatformModel{}); err != nil {
		return err
	}
	return AutoMigrate(db)
}

func (r *socialPlatformRepository) List(ctx context.Context) ([]*socialplatform.SocialPlatform, error) {
	var models []socialPlatformModel
	if err := r.db.WithContext(ctx).
		Order("sort_order ASC, name ASC").
		Find(&models).Error; err != nil {
		r.logger.Error("DB_LIST_SOCIAL_PLATFORMS_ERROR", err)
		return nil, err
	}
	result := make([]*socialplatform.SocialPlatform, len(models))
	for i := range models {
		result[i] = models[i].ToEntity()
	}
	return result, nil
}

func (r *socialPlatformRepository) GetByID(ctx context.Context, id uuid.UUID) (*socialplatform.SocialPlatform, error) {
	var m socialPlatformModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("DB_GET_SOCIAL_PLATFORM_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *socialPlatformRepository) GetByKey(ctx context.Context, key string) (*socialplatform.SocialPlatform, error) {
	var m socialPlatformModel
	if err := r.db.WithContext(ctx).Where("key = ?", key).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("DB_GET_SOCIAL_PLATFORM_BY_KEY_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *socialPlatformRepository) Create(ctx context.Context, p *socialplatform.SocialPlatform) error {
	id, err := uuid.NewV7()
	if err != nil {
		return custom_errors.NewInternalError("ไม่สามารถสร้าง ID ได้")
	}
	p.ID = id
	m := &socialPlatformModel{
		ID:        p.ID,
		Key:       p.Key,
		Name:      p.Name,
		IconURL:   p.IconURL,
		SortOrder: p.SortOrder,
		IsActive:  p.IsActive,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		r.logger.Error("DB_CREATE_SOCIAL_PLATFORM_ERROR", err)
		return err
	}
	p.CreatedAt = m.CreatedAt
	p.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *socialPlatformRepository) Update(ctx context.Context, p *socialplatform.SocialPlatform) error {
	result := r.db.WithContext(ctx).Model(&socialPlatformModel{}).
		Where("id = ?", p.ID).
		Updates(map[string]any{
			"name":       p.Name,
			"icon_url":   p.IconURL,
			"sort_order": p.SortOrder,
			"is_active":  p.IsActive,
		})
	if result.Error != nil {
		r.logger.Error("DB_UPDATE_SOCIAL_PLATFORM_ERROR", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบ platform นี้")
	}
	return nil
}

func (r *socialPlatformRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&socialPlatformModel{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("DB_DELETE_SOCIAL_PLATFORM_ERROR", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบ platform นี้")
	}
	return nil
}
