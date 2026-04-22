package pgproject

import (
	"context"
	"errors"
	"strings"
	"time"

	"vexentra-api/internal/adapters/database/postgres/pgtx"
	"vexentra-api/internal/modules/project"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type projectRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

func NewProjectRepository(db *gorm.DB, l logger.Logger) project.ProjectRepository {
	if l == nil {
		l = logger.Get()
	}
	return &projectRepository{db: db, logger: l}
}

func (r *projectRepository) Create(ctx context.Context, p *project.Project) error {
	id, err := uuid.NewV7()
	if err != nil {
		return custom_errors.NewInternalError("ไม่สามารถสร้าง ID ได้")
	}
	p.ID = id

	m := fromProject(p)
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).Create(m).Error; err != nil {
		r.logger.Error("DB_CREATE_PROJECT_ERROR", err)
		return err
	}
	p.CreatedAt = m.CreatedAt
	p.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *projectRepository) GetByID(ctx context.Context, id uuid.UUID) (*project.Project, error) {
	var m projectModel
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("DB_GET_PROJECT_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *projectRepository) GetByCode(ctx context.Context, code string) (*project.Project, error) {
	var m projectModel
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).
		Where("LOWER(project_code) = LOWER(?)", code).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("DB_GET_PROJECT_BY_CODE_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *projectRepository) ListStatuses(ctx context.Context, activeOnly bool) ([]project.ProjectStatusMeta, error) {
	q := pgtx.DB(ctx, r.db).WithContext(ctx).Model(&projectStatusModel{})
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}

	var models []projectStatusModel
	if err := q.Order("sort_order ASC").Order("status ASC").Find(&models).Error; err != nil {
		r.logger.Error("DB_LIST_PROJECT_STATUSES_ERROR", err)
		return nil, err
	}

	items := make([]project.ProjectStatusMeta, len(models))
	for i := range models {
		items[i] = models[i].ToEntity()
	}
	return items, nil
}

func (r *projectRepository) GetFinancialPlan(ctx context.Context, projectID uuid.UUID) (*project.ProjectFinancialPlan, error) {
	db := pgtx.DB(ctx, r.db).WithContext(ctx)

	var planModel projectFinancialPlanModel
	err := db.First(&planModel, "project_id = ?", projectID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zero := decimal.Zero
			return &project.ProjectFinancialPlan{
				ProjectID:            projectID,
				ContractAmount:       zero,
				RetentionAmount:      zero,
				Installments:         []*project.ProjectPaymentInstallment{},
				InstallmentsTotal:    zero,
				NetReceivable:        zero,
				UnallocatedRemaining: zero,
			}, nil
		}
		r.logger.Error("DB_GET_PROJECT_FINANCIAL_PLAN_ERROR", err)
		return nil, err
	}

	var rows []projectPaymentInstallmentModel
	if err := db.
		Where("project_id = ?", projectID).
		Order("sort_order ASC, created_at ASC").
		Find(&rows).Error; err != nil {
		r.logger.Error("DB_LIST_PROJECT_INSTALLMENTS_ERROR", err)
		return nil, err
	}

	items := make([]*project.ProjectPaymentInstallment, len(rows))
	sum := decimal.Zero
	for i := range rows {
		items[i] = rows[i].ToEntity()
		sum = sum.Add(rows[i].Amount)
	}

	net := planModel.ContractAmount.Sub(planModel.RetentionAmount)
	return &project.ProjectFinancialPlan{
		ProjectID:            planModel.ProjectID,
		ContractAmount:       planModel.ContractAmount,
		RetentionAmount:      planModel.RetentionAmount,
		PlannedDeliveryDate:  planModel.PlannedDeliveryDate,
		PaymentNote:          planModel.PaymentNote,
		Installments:         items,
		InstallmentsTotal:    sum,
		NetReceivable:        net,
		UnallocatedRemaining: net.Sub(sum),
		CreatedAt:            planModel.CreatedAt,
		UpdatedAt:            planModel.UpdatedAt,
	}, nil
}

func (r *projectRepository) UpsertFinancialPlan(ctx context.Context, plan *project.ProjectFinancialPlan) error {
	db := pgtx.DB(ctx, r.db).WithContext(ctx)
	now := time.Now().UTC()

	planModel := &projectFinancialPlanModel{
		ProjectID:           plan.ProjectID,
		ContractAmount:      plan.ContractAmount,
		RetentionAmount:     plan.RetentionAmount,
		PlannedDeliveryDate: plan.PlannedDeliveryDate,
		PaymentNote:         plan.PaymentNote,
		UpdatedAt:           now,
	}

	if err := db.
		Model(&projectFinancialPlanModel{}).
		Where("project_id = ?", plan.ProjectID).
		Updates(map[string]any{
			"contract_amount":       planModel.ContractAmount,
			"retention_amount":      planModel.RetentionAmount,
			"planned_delivery_date": planModel.PlannedDeliveryDate,
			"payment_note":          planModel.PaymentNote,
			"updated_at":            planModel.UpdatedAt,
		}).Error; err != nil {
		r.logger.Error("DB_UPDATE_PROJECT_FINANCIAL_PLAN_ERROR", err)
		return err
	}

	res := db.Model(&projectFinancialPlanModel{}).Where("project_id = ?", plan.ProjectID).Select("project_id").First(&projectFinancialPlanModel{})
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		planModel.CreatedAt = now
		if err := db.Create(planModel).Error; err != nil {
			r.logger.Error("DB_CREATE_PROJECT_FINANCIAL_PLAN_ERROR", err)
			return err
		}
	} else if res.Error != nil {
		r.logger.Error("DB_CHECK_PROJECT_FINANCIAL_PLAN_ERROR", res.Error)
		return res.Error
	}

	if err := db.Where("project_id = ?", plan.ProjectID).Delete(&projectPaymentInstallmentModel{}).Error; err != nil {
		r.logger.Error("DB_DELETE_PROJECT_INSTALLMENTS_ERROR", err)
		return err
	}

	if len(plan.Installments) == 0 {
		return nil
	}

	rows := make([]projectPaymentInstallmentModel, len(plan.Installments))
	for i := range plan.Installments {
		item := plan.Installments[i]
		rows[i] = projectPaymentInstallmentModel{
			ID:                  item.ID,
			ProjectID:           plan.ProjectID,
			SortOrder:           item.SortOrder,
			Title:               item.Title,
			Amount:              item.Amount,
			PlannedDeliveryDate: item.PlannedDeliveryDate,
			PlannedReceiveDate:  item.PlannedReceiveDate,
			Note:                item.Note,
			CreatedAt:           now,
			UpdatedAt:           now,
		}
	}

	if err := db.Create(&rows).Error; err != nil {
		r.logger.Error("DB_CREATE_PROJECT_INSTALLMENTS_ERROR", err)
		return err
	}
	return nil
}

func (r *projectRepository) List(ctx context.Context, f project.ProjectFilter, pg project.Pagination) ([]*project.Project, int64, error) {
	q := pgtx.DB(ctx, r.db).WithContext(ctx).Model(&projectModel{})

	if len(f.Statuses) > 0 {
		statuses := make([]string, len(f.Statuses))
		for i, s := range f.Statuses {
			statuses[i] = string(s)
		}
		q = q.Where("status IN ?", statuses)
	}
	if f.CreatedByUserID != nil {
		q = q.Where("created_by_user_id = ?", *f.CreatedByUserID)
	}
	if f.ClientPersonID != nil {
		q = q.Where("client_person_id = ?", *f.ClientPersonID)
	}
	if f.MemberPersonID != nil {
		// EXISTS avoids JOIN row duplication
		q = q.Where(`EXISTS (
			SELECT 1 FROM project_members pm
			WHERE pm.project_id = projects.id
			  AND pm.person_id = ?
			  AND pm.deleted_at IS NULL
		)`, *f.MemberPersonID)
	}
	if s := strings.TrimSpace(f.Search); s != "" {
		pattern := "%" + strings.ToLower(s) + "%"
		q = q.Where("LOWER(name) LIKE ? OR LOWER(project_code) LIKE ?", pattern, pattern)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		r.logger.Error("DB_COUNT_PROJECTS_ERROR", err)
		return nil, 0, err
	}

	var models []projectModel
	if err := q.Order("updated_at DESC").
		Limit(pg.Limit).Offset(pg.Offset).
		Find(&models).Error; err != nil {
		r.logger.Error("DB_LIST_PROJECTS_ERROR", err)
		return nil, 0, err
	}

	items := make([]*project.Project, len(models))
	for i := range models {
		items[i] = models[i].ToEntity()
	}
	return items, total, nil
}

func (r *projectRepository) Update(ctx context.Context, p *project.Project) error {
	m := fromProject(p)

	result := pgtx.DB(ctx, r.db).WithContext(ctx).
		Model(&projectModel{}).
		Where("id = ?", p.ID).
		Updates(map[string]any{
			"name":               m.Name,
			"description":        m.Description,
			"status":             m.Status,
			"closure_reason":     m.ClosureReason,
			"client_person_id":   m.ClientPersonID,
			"client_name_raw":    m.ClientNameRaw,
			"client_email_raw":   m.ClientEmailRaw,
			"scheduled_start_at": m.ScheduledStartAt,
			"deadline_at":        m.DeadlineAt,
			"activated_at":       m.ActivatedAt,
			"closed_at":          m.ClosedAt,
		})
	if result.Error != nil {
		r.logger.Error("DB_UPDATE_PROJECT_ERROR", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบโปรเจกต์นี้")
	}
	return nil
}

func (r *projectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// GORM soft delete — automatically sets deleted_at when DeletedAt is gorm.DeletedAt
	result := pgtx.DB(ctx, r.db).WithContext(ctx).Delete(&projectModel{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("DB_DELETE_PROJECT_ERROR", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบโปรเจกต์นี้")
	}
	return nil
}

func (r *projectRepository) NextCodeSeq(ctx context.Context) (int64, error) {
	var seq int64
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).
		Raw("SELECT nextval('project_code_seq')").Scan(&seq).Error; err != nil {
		r.logger.Error("DB_NEXT_PROJECT_CODE_SEQ_ERROR", err)
		return 0, err
	}
	return seq, nil
}
