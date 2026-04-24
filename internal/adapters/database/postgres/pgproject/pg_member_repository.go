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
	items := []*project.ProjectMember{m.ToEntity()}
	if err := r.attachRoles(ctx, items); err != nil {
		return nil, err
	}
	return items[0], nil
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
	if err := r.attachRoles(ctx, result); err != nil {
		return nil, err
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

func (r *projectMemberRepository) ListRoleMaster(ctx context.Context, activeOnly bool) ([]*project.ProjectRole, error) {
	query := pgtx.DB(ctx, r.db).WithContext(ctx).Model(&projectRoleMasterModel{})
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	var rows []projectRoleMasterModel
	if err := query.Order("sort_order ASC, name_th ASC").Find(&rows).Error; err != nil {
		r.logger.Error("DB_LIST_PROJECT_ROLE_MASTER_ERROR", err)
		return nil, err
	}
	items := make([]*project.ProjectRole, len(rows))
	for i := range rows {
		items[i] = rows[i].ToEntity()
	}
	return items, nil
}

func (r *projectMemberRepository) CountActiveRoleMasterByIDs(ctx context.Context, roleIDs []uuid.UUID) (int64, error) {
	if len(roleIDs) == 0 {
		return 0, nil
	}
	var count int64
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).
		Model(&projectRoleMasterModel{}).
		Where("id IN ? AND is_active = ?", roleIDs, true).
		Count(&count).Error; err != nil {
		r.logger.Error("DB_COUNT_PROJECT_ROLE_MASTER_ERROR", err)
		return 0, err
	}
	return count, nil
}

func (r *projectMemberRepository) ReplaceMemberRoles(ctx context.Context, memberID, assignedByUserID uuid.UUID, roleIDs []uuid.UUID, primaryRoleID *uuid.UUID) error {
	uniqueRoleIDs := make([]uuid.UUID, 0, len(roleIDs))
	seen := make(map[uuid.UUID]struct{}, len(roleIDs))
	for _, id := range roleIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueRoleIDs = append(uniqueRoleIDs, id)
	}

	return pgtx.DB(ctx, r.db).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&projectMemberRoleAssignmentModel{}).
			Where("project_member_id = ? AND deleted_at IS NULL", memberID).
			Updates(map[string]any{
				"deleted_at": gorm.Expr("now()"),
				"updated_at": gorm.Expr("now()"),
			}).Error; err != nil {
			r.logger.Error("DB_CLEAR_MEMBER_ROLES_ERROR", err)
			return err
		}

		if len(uniqueRoleIDs) == 0 {
			return nil
		}

		rows := make([]projectMemberRoleAssignmentModel, 0, len(uniqueRoleIDs))
		for _, roleID := range uniqueRoleIDs {
			id, err := uuid.NewV7()
			if err != nil {
				return custom_errors.NewInternalError("ไม่สามารถสร้าง ID ได้")
			}
			isPrimary := primaryRoleID != nil && *primaryRoleID == roleID
			rows = append(rows, projectMemberRoleAssignmentModel{
				ID:               id,
				ProjectMemberID:  memberID,
				RoleID:           roleID,
				IsPrimary:        isPrimary,
				AssignedByUserID: assignedByUserID,
			})
		}

		if err := tx.Create(&rows).Error; err != nil {
			r.logger.Error("DB_INSERT_MEMBER_ROLES_ERROR", err)
			return err
		}
		return nil
	})
}

func (r *projectMemberRepository) HasActiveRoleCode(ctx context.Context, projectID, personID uuid.UUID, roleCode string) (bool, error) {
	var count int64
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).
		Table("project_member_role_assignments AS pmra").
		Joins("JOIN project_members AS pm ON pm.id = pmra.project_member_id AND pm.deleted_at IS NULL").
		Joins("JOIN project_role_master AS prm ON prm.id = pmra.role_id AND prm.deleted_at IS NULL").
		Where("pm.project_id = ? AND pm.person_id = ?", projectID, personID).
		Where("pmra.deleted_at IS NULL AND prm.is_active = ? AND prm.code = ?", true, roleCode).
		Count(&count).Error; err != nil {
		r.logger.Error("DB_CHECK_MEMBER_ROLE_CODE_ERROR", err)
		return false, err
	}
	return count > 0, nil
}

type memberRoleJoinRow struct {
	AssignmentID uuid.UUID `gorm:"column:assignment_id"`
	MemberID     uuid.UUID `gorm:"column:member_id"`
	RoleID       uuid.UUID `gorm:"column:role_id"`
	Code         string    `gorm:"column:code"`
	NameTH       string    `gorm:"column:name_th"`
	NameEN       string    `gorm:"column:name_en"`
	IsPrimary    bool      `gorm:"column:is_primary"`
}

func (r *projectMemberRepository) attachRoles(ctx context.Context, members []*project.ProjectMember) error {
	if len(members) == 0 {
		return nil
	}
	memberIDs := make([]uuid.UUID, 0, len(members))
	for _, m := range members {
		memberIDs = append(memberIDs, m.ID)
	}

	var rows []memberRoleJoinRow
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).
		Table("project_member_role_assignments AS pmra").
		Select(
			"pmra.id AS assignment_id",
			"pmra.project_member_id AS member_id",
			"pmra.role_id AS role_id",
			"prm.code AS code",
			"prm.name_th AS name_th",
			"prm.name_en AS name_en",
			"pmra.is_primary AS is_primary",
		).
		Joins("JOIN project_role_master AS prm ON prm.id = pmra.role_id AND prm.deleted_at IS NULL").
		Where("pmra.project_member_id IN ? AND pmra.deleted_at IS NULL", memberIDs).
		Order("pmra.project_member_id ASC, pmra.is_primary DESC, prm.sort_order ASC, pmra.created_at ASC").
		Find(&rows).Error; err != nil {
		r.logger.Error("DB_LIST_MEMBER_ROLE_ASSIGNMENTS_ERROR", err)
		return err
	}

	byMember := make(map[uuid.UUID][]project.ProjectMemberRole, len(memberIDs))
	for _, row := range rows {
		byMember[row.MemberID] = append(byMember[row.MemberID], project.ProjectMemberRole{
			AssignmentID: row.AssignmentID,
			RoleID:       row.RoleID,
			Code:         row.Code,
			NameTH:       row.NameTH,
			NameEN:       row.NameEN,
			IsPrimary:    row.IsPrimary,
		})
	}
	for i := range members {
		members[i].Roles = byMember[members[i].ID]
	}
	return nil
}
