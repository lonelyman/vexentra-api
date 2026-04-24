package projectsvc

import (
	"context"

	"vexentra-api/internal/modules/project"
	"vexentra-api/internal/modules/user"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
)

type MemberService interface {
	Add(ctx context.Context, caller user.Caller, projectID, personID uuid.UUID, roleIDs []uuid.UUID, primaryRoleID *uuid.UUID) (*project.ProjectMember, error)
	List(ctx context.Context, caller user.Caller, projectID uuid.UUID) ([]*project.ProjectMember, error)
	Remove(ctx context.Context, caller user.Caller, projectID, memberID uuid.UUID) error
	TransferLead(ctx context.Context, caller user.Caller, projectID, toMemberID uuid.UUID) error
	ListRoleMaster(ctx context.Context, caller user.Caller) ([]*project.ProjectRole, error)
	SetMemberRoles(ctx context.Context, caller user.Caller, projectID, memberID uuid.UUID, roleIDs []uuid.UUID, primaryRoleID *uuid.UUID) error
}

type memberService struct {
	projectSvc ProjectService
	memberRepo project.ProjectMemberRepository
	logger     logger.Logger
}

func NewMemberService(projectSvc ProjectService, memberRepo project.ProjectMemberRepository, l logger.Logger) MemberService {
	if l == nil {
		l = logger.Get()
	}
	return &memberService{projectSvc: projectSvc, memberRepo: memberRepo, logger: l}
}

// Add enrolls a person as a project member (is_lead=false by default).
// Only staff, project creator, or the current lead may add members.
func (s *memberService) Add(ctx context.Context, caller user.Caller, projectID, personID uuid.UUID, roleIDs []uuid.UUID, primaryRoleID *uuid.UUID) (*project.ProjectMember, error) {
	if err := s.requireLeadOrStaff(ctx, caller, projectID); err != nil {
		return nil, err
	}

	// Duplicate-add guard (also protected by partial unique index, but better to
	// fail cleanly than rely on a 500-mapped constraint error).
	existing, err := s.memberRepo.GetActiveByProjectAndPerson(ctx, projectID, personID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, custom_errors.New(409, "MEMBER_ALREADY_EXISTS", "บุคคลนี้เป็นสมาชิกของโปรเจกต์อยู่แล้ว")
	}

	m := &project.ProjectMember{
		ProjectID:     projectID,
		PersonID:      personID,
		IsLead:        false,
		AddedByUserID: caller.UserID,
	}
	if err := s.memberRepo.Add(ctx, m); err != nil {
		return nil, err
	}
	roleIDs = uniqueUUIDs(roleIDs)
	if err := s.validateMemberRoles(ctx, roleIDs, primaryRoleID); err != nil {
		return nil, err
	}
	if err := s.memberRepo.ReplaceMemberRoles(ctx, m.ID, caller.UserID, roleIDs, primaryRoleID); err != nil {
		return nil, err
	}
	latest, err := s.memberRepo.GetByID(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	if latest != nil {
		return latest, nil
	}
	return m, nil
}

func (s *memberService) List(ctx context.Context, caller user.Caller, projectID uuid.UUID) ([]*project.ProjectMember, error) {
	if _, err := s.projectSvc.CanAccessProject(ctx, caller, projectID); err != nil {
		return nil, err
	}
	return s.memberRepo.ListByProject(ctx, projectID)
}

// Remove soft-deletes a membership. Refuses to remove the sole active lead —
// caller must TransferLead to someone else first.
func (s *memberService) Remove(ctx context.Context, caller user.Caller, projectID, memberID uuid.UUID) error {
	if err := s.requireLeadOrStaff(ctx, caller, projectID); err != nil {
		return err
	}

	target, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return err
	}
	if target == nil || target.ProjectID != projectID {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบสมาชิกในโปรเจกต์นี้")
	}

	if target.IsLead {
		return custom_errors.New(409, "CANNOT_REMOVE_LEAD", "ไม่สามารถลบหัวหน้าทีมคนเดียวของโปรเจกต์ได้ — ต้องโอนสิทธิ์ให้คนอื่นก่อน")
	}

	return s.memberRepo.Remove(ctx, memberID)
}

// TransferLead hands over is_lead to another active member of the same project.
// Only the current lead or staff may trigger the handover.
func (s *memberService) TransferLead(ctx context.Context, caller user.Caller, projectID, toMemberID uuid.UUID) error {
	_, err := s.projectSvc.CanAccessProject(ctx, caller, projectID)
	if err != nil {
		return err
	}

	lead, err := s.memberRepo.GetActiveLead(ctx, projectID)
	if err != nil {
		return err
	}

	isCoordinator, err := s.memberRepo.HasActiveRoleCode(ctx, projectID, caller.PersonID, projectCoordinatorRoleCode)
	if err != nil {
		return err
	}
	allowed := caller.IsStaff() || isCoordinator || (lead != nil && lead.PersonID == caller.PersonID)
	if !allowed {
		return custom_errors.New(403, custom_errors.ErrForbidden, "เฉพาะหัวหน้าทีม, coordinator หรือผู้ดูแลระบบเท่านั้นที่โอนสิทธิ์ได้")
	}

	target, err := s.memberRepo.GetByID(ctx, toMemberID)
	if err != nil {
		return err
	}
	if target == nil || target.ProjectID != projectID {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบสมาชิกที่ต้องการโอนสิทธิ์ในโปรเจกต์นี้")
	}
	if target.DeletedAt != nil {
		return custom_errors.New(409, custom_errors.ErrValidation, "สมาชิกเป้าหมายถูกนำออกจากโปรเจกต์ไปแล้ว")
	}
	if lead != nil && lead.ID == toMemberID {
		return nil // no-op: already the lead
	}

	return s.memberRepo.TransferLead(ctx, projectID, toMemberID)
}

func (s *memberService) ListRoleMaster(ctx context.Context, _ user.Caller) ([]*project.ProjectRole, error) {
	return s.memberRepo.ListRoleMaster(ctx, true)
}

func (s *memberService) SetMemberRoles(ctx context.Context, caller user.Caller, projectID, memberID uuid.UUID, roleIDs []uuid.UUID, primaryRoleID *uuid.UUID) error {
	if err := s.requireLeadOrStaff(ctx, caller, projectID); err != nil {
		return err
	}

	target, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return err
	}
	if target == nil || target.ProjectID != projectID {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบสมาชิกในโปรเจกต์นี้")
	}
	if target.DeletedAt != nil {
		return custom_errors.New(409, custom_errors.ErrValidation, "สมาชิกถูกนำออกจากโปรเจกต์แล้ว")
	}

	roleIDs = uniqueUUIDs(roleIDs)
	if err := s.validateMemberRoles(ctx, roleIDs, primaryRoleID); err != nil {
		return err
	}
	return s.memberRepo.ReplaceMemberRoles(ctx, memberID, caller.UserID, roleIDs, primaryRoleID)
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// requireLeadOrStaff allows staff, current lead, or coordinator to manage members.
func (s *memberService) requireLeadOrStaff(ctx context.Context, caller user.Caller, projectID uuid.UUID) error {
	_, err := s.projectSvc.CanAccessProject(ctx, caller, projectID)
	if err != nil {
		return err
	}
	if caller.IsStaff() {
		return nil
	}
	lead, err := s.memberRepo.GetActiveLead(ctx, projectID)
	if err != nil {
		return err
	}
	if lead != nil && lead.PersonID == caller.PersonID {
		return nil
	}
	isCoordinator, err := s.memberRepo.HasActiveRoleCode(ctx, projectID, caller.PersonID, projectCoordinatorRoleCode)
	if err != nil {
		return err
	}
	if isCoordinator {
		return nil
	}
	return custom_errors.New(403, custom_errors.ErrForbidden, "เฉพาะหัวหน้าทีม, coordinator หรือผู้ดูแลระบบเท่านั้นที่จัดการสมาชิกได้")
}

func (s *memberService) validateMemberRoles(ctx context.Context, roleIDs []uuid.UUID, primaryRoleID *uuid.UUID) error {
	roleIDs = uniqueUUIDs(roleIDs)
	if len(roleIDs) == 0 {
		if primaryRoleID != nil {
			return custom_errors.New(400, custom_errors.ErrValidation, "primary_role_id ต้องว่างเมื่อยังไม่กำหนด role")
		}
		return nil
	}
	if primaryRoleID != nil {
		found := false
		for _, id := range roleIDs {
			if id == *primaryRoleID {
				found = true
				break
			}
		}
		if !found {
			return custom_errors.New(400, custom_errors.ErrValidation, "primary_role_id ต้องอยู่ใน role_ids")
		}
	}
	count, err := s.memberRepo.CountActiveRoleMasterByIDs(ctx, roleIDs)
	if err != nil {
		return err
	}
	if count != int64(len(roleIDs)) {
		return custom_errors.New(400, custom_errors.ErrValidation, "role_ids มีค่าที่ไม่ถูกต้องหรือไม่ active")
	}
	return nil
}

func uniqueUUIDs(ids []uuid.UUID) []uuid.UUID {
	if len(ids) <= 1 {
		return ids
	}
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
