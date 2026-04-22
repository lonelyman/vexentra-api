package pgtxcategory

import (
	"context"
	"errors"
	"strings"

	"vexentra-api/internal/adapters/database/postgres/pgtx"
	"vexentra-api/internal/modules/txcategory"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type txCategoryRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

func NewTransactionCategoryRepository(db *gorm.DB, l logger.Logger) txcategory.TransactionCategoryRepository {
	if l == nil {
		l = logger.Get()
	}
	return &txCategoryRepository{db: db, logger: l}
}

func (r *txCategoryRepository) Create(ctx context.Context, c *txcategory.TransactionCategory) error {
	id, err := uuid.NewV7()
	if err != nil {
		return custom_errors.NewInternalError("ไม่สามารถสร้าง ID ได้")
	}
	c.ID = id
	// DB CHECK constraint enforces lowercase [a-z0-9_]+, but normalize defensively.
	c.Code = strings.ToLower(strings.TrimSpace(c.Code))

	m := fromCategory(c)
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).Create(m).Error; err != nil {
		r.logger.Error("DB_CREATE_TX_CATEGORY_ERROR", err)
		return err
	}
	c.CreatedAt = m.CreatedAt
	c.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *txCategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*txcategory.TransactionCategory, error) {
	var m txCategoryModel
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("DB_GET_TX_CATEGORY_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *txCategoryRepository) GetByCode(ctx context.Context, code string) (*txcategory.TransactionCategory, error) {
	var m txCategoryModel
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).
		Where("code = ?", strings.ToLower(strings.TrimSpace(code))).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("DB_GET_TX_CATEGORY_BY_CODE_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *txCategoryRepository) List(ctx context.Context, f txcategory.TransactionCategoryFilter) ([]*txcategory.TransactionCategory, error) {
	q := pgtx.DB(ctx, r.db).WithContext(ctx).Model(&txCategoryModel{})
	if f.IncludeDeleted {
		q = q.Unscoped()
	}
	if f.Type != nil {
		q = q.Where("type = ?", string(*f.Type))
	}
	if f.ActiveOnly {
		q = q.Where("is_active = ?", true)
	}

	var models []txCategoryModel
	if err := q.Order("type ASC, sort_order ASC, name ASC").Find(&models).Error; err != nil {
		r.logger.Error("DB_LIST_TX_CATEGORIES_ERROR", err)
		return nil, err
	}
	result := make([]*txcategory.TransactionCategory, len(models))
	for i := range models {
		result[i] = models[i].ToEntity()
	}
	return result, nil
}

func (r *txCategoryRepository) Update(ctx context.Context, c *txcategory.TransactionCategory) error {
	result := pgtx.DB(ctx, r.db).WithContext(ctx).
		Model(&txCategoryModel{}).
		Where("id = ?", c.ID).
		Updates(map[string]any{
			"name":       c.Name,
			"icon_key":   c.IconKey,
			"is_active":  c.IsActive,
			"sort_order": c.SortOrder,
			// code/type/is_system are immutable post-creation — service layer enforces this.
		})
	if result.Error != nil {
		r.logger.Error("DB_UPDATE_TX_CATEGORY_ERROR", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบหมวดหมู่นี้")
	}
	return nil
}

func (r *txCategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := pgtx.DB(ctx, r.db).WithContext(ctx).Delete(&txCategoryModel{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("DB_DELETE_TX_CATEGORY_ERROR", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบหมวดหมู่นี้")
	}
	return nil
}
