package pgproject

import (
	"context"
	"errors"

	"vexentra-api/internal/adapters/database/postgres/pgtx"
	"vexentra-api/internal/modules/project"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type projectTransactionRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

func NewProjectTransactionRepository(db *gorm.DB, l logger.Logger) project.ProjectTransactionRepository {
	if l == nil {
		l = logger.Get()
	}
	return &projectTransactionRepository{db: db, logger: l}
}

func (r *projectTransactionRepository) Create(ctx context.Context, t *project.ProjectTransaction) error {
	id, err := uuid.NewV7()
	if err != nil {
		return custom_errors.NewInternalError("ไม่สามารถสร้าง ID ได้")
	}
	t.ID = id

	m := fromTransaction(t)
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).Create(m).Error; err != nil {
		r.logger.Error("DB_CREATE_PROJECT_TRANSACTION_ERROR", err)
		return err
	}
	t.CreatedAt = m.CreatedAt
	t.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *projectTransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*project.ProjectTransaction, error) {
	var m projectTransactionModel
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("DB_GET_PROJECT_TRANSACTION_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *projectTransactionRepository) ListByProject(ctx context.Context, projectID uuid.UUID, f project.TransactionFilter, pg project.Pagination) ([]*project.ProjectTransaction, int64, error) {
	q := pgtx.DB(ctx, r.db).WithContext(ctx).
		Model(&projectTransactionModel{}).
		Where("project_id = ?", projectID)

	if len(f.CategoryIDs) > 0 {
		q = q.Where("category_id IN ?", f.CategoryIDs)
	}
	if f.OccurredGTE != nil {
		q = q.Where("occurred_at >= ?", *f.OccurredGTE)
	}
	if f.OccurredLT != nil {
		q = q.Where("occurred_at < ?", *f.OccurredLT)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		r.logger.Error("DB_COUNT_PROJECT_TRANSACTIONS_ERROR", err)
		return nil, 0, err
	}

	var models []projectTransactionModel
	if err := q.Order("occurred_at DESC").
		Limit(pg.Limit).Offset(pg.Offset).
		Find(&models).Error; err != nil {
		r.logger.Error("DB_LIST_PROJECT_TRANSACTIONS_ERROR", err)
		return nil, 0, err
	}

	items := make([]*project.ProjectTransaction, len(models))
	for i := range models {
		items[i] = models[i].ToEntity()
	}
	return items, total, nil
}

// SumByProject aggregates income/expense totals at the DB layer via a JOIN to
// transaction_categories.type. Soft-deleted rows are excluded on both sides.
func (r *projectTransactionRepository) SumByProject(ctx context.Context, projectID uuid.UUID) (*project.ProjectTotals, error) {
	var row struct {
		Income  decimal.Decimal
		Expense decimal.Decimal
	}
	const query = `
		SELECT
			COALESCE(SUM(CASE WHEN tc.type = 'income'  THEN pt.amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN tc.type = 'expense' THEN pt.amount ELSE 0 END), 0) AS expense
		FROM project_transactions pt
		JOIN transaction_categories tc ON tc.id = pt.category_id
		WHERE pt.project_id = ?
		  AND pt.deleted_at IS NULL
		  AND tc.deleted_at IS NULL
	`
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).
		Raw(query, projectID).Scan(&row).Error; err != nil {
		r.logger.Error("DB_SUM_PROJECT_TRANSACTIONS_ERROR", err)
		return nil, err
	}
	return &project.ProjectTotals{
		Income:  row.Income,
		Expense: row.Expense,
		Net:     row.Income.Sub(row.Expense),
	}, nil
}

func (r *projectTransactionRepository) Update(ctx context.Context, t *project.ProjectTransaction) error {
	result := pgtx.DB(ctx, r.db).WithContext(ctx).
		Model(&projectTransactionModel{}).
		Where("id = ?", t.ID).
		Updates(map[string]any{
			"category_id":   t.CategoryID,
			"amount":        t.Amount,
			"currency_code": t.CurrencyCode,
			"note":          t.Note,
			"occurred_at":   t.OccurredAt,
		})
	if result.Error != nil {
		r.logger.Error("DB_UPDATE_PROJECT_TRANSACTION_ERROR", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบรายการธุรกรรมนี้")
	}
	return nil
}

func (r *projectTransactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := pgtx.DB(ctx, r.db).WithContext(ctx).Delete(&projectTransactionModel{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("DB_DELETE_PROJECT_TRANSACTION_ERROR", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบรายการธุรกรรมนี้")
	}
	return nil
}
