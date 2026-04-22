package pgproject

import (
	"context"
	"errors"

	"vexentra-api/internal/adapters/database/postgres/pgtx"
	"vexentra-api/internal/modules/project"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectMemberRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

func NewProjectMemberRepository(db *gorm.DB, l logger.Logger) project.ProjectMemberRepository {
	if l == nil {
		l = logger.Get()
	}
	return &projectMemberRepository{db: db, logger: l}
}

func (r *projectMemberRepository) Add(ctx context.Context, m *project.ProjectMember) error {
	id, err := uuid.NewV7()
	if err != nil {
		return custom_errors.NewInternalError("ไม่สามารถสร้าง ID ได้")
	}
	m.ID = id

	model := fromMember(m)
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).Create(model).Error; err != nil {
		r.logger.Error("DB_ADD_PROJECT_MEMBER_ERROR", err)
		return err
	}
	m.CreatedAt = model.CreatedAt
	m.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *projectMemberRepository) GetByID(ctx context.Context, id uuid.UUID) (*project.ProjectMember, error) {
	var m projectMemberModel
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("DB_GET_PROJECT_MEMBER_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *projectMemberRepository) GetActiveByProjectAndPerson(ctx context.Context, projectID, personID uuid.UUID) (*project.ProjectMember, error) {
	var m projectMemberModel
	// GORM auto-injects `deleted_at IS NULL` via gorm.DeletedAt, but we spell it out
	// here for clarity on the intent of this lookup.
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).
		Where("project_id = ? AND person_id = ?", projectID, personID).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("DB_GET_ACTIVE_MEMBER_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *projectMemberRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*project.ProjectMember, error) {
	var models []projectMemberModel
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at ASC").
		Find(&models).Error; err != nil {
		r.logger.Error("DB_LIST_PROJECT_MEMBERS_ERROR", err)
		return nil, err
	}
	result := make([]*project.ProjectMember, len(models))
	for i := range models {
		result[i] = models[i].ToEntity()
	}
	return result, nil
}

func (r *projectMemberRepository) GetActiveLead(ctx context.Context, projectID uuid.UUID) (*project.ProjectMember, error) {
	var m projectMemberModel
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).
		Where("project_id = ? AND is_lead = ?", projectID, true).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("DB_GET_ACTIVE_LEAD_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

// TransferLead flips is_lead atomically: strips the flag from any current active
// lead of the same project, then sets it on toMemberID. Wrapped in a transaction so
// the one-lead partial unique index is never momentarily violated.
func (r *projectMemberRepository) TransferLead(ctx context.Context, projectID, toMemberID uuid.UUID) error {
	return pgtx.DB(ctx, r.db).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Clear any existing active lead in this project (may be 0 rows if none).
		if err := tx.Model(&projectMemberModel{}).
			Where("project_id = ? AND is_lead = ? AND id <> ?", projectID, true, toMemberID).
			Update("is_lead", false).Error; err != nil {
			r.logger.Error("DB_TRANSFER_LEAD_CLEAR_ERROR", err)
			return err
		}

		result := tx.Model(&projectMemberModel{}).
			Where("id = ? AND project_id = ?", toMemberID, projectID).
			Update("is_lead", true)
		if result.Error != nil {
			r.logger.Error("DB_TRANSFER_LEAD_SET_ERROR", result.Error)
			return result.Error
		}
		if result.RowsAffected == 0 {
			return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบสมาชิกที่ต้องการโอนสิทธิ์หัวหน้าทีม")
		}
		return nil
	})
}

func (r *projectMemberRepository) Remove(ctx context.Context, id uuid.UUID) error {
	result := pgtx.DB(ctx, r.db).WithContext(ctx).Delete(&projectMemberModel{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("DB_REMOVE_PROJECT_MEMBER_ERROR", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบสมาชิกนี้")
	}
	return nil
}
