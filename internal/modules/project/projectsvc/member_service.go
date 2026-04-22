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
	Add(ctx context.Context, caller user.Caller, projectID, personID uuid.UUID) (*project.ProjectMember, error)
	List(ctx context.Context, caller user.Caller, projectID uuid.UUID) ([]*project.ProjectMember, error)
	Remove(ctx context.Context, caller user.Caller, projectID, memberID uuid.UUID) error
	TransferLead(ctx context.Context, caller user.Caller, projectID, toMemberID uuid.UUID) error
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
func (s *memberService) Add(ctx context.Context, caller user.Caller, projectID, personID uuid.UUID) (*project.ProjectMember, error) {
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
	lead, err := s.memberRepo.GetActiveLead(ctx, projectID)
	if err != nil {
		return err
	}
	if !caller.IsStaff() {
		if lead == nil || lead.PersonID != caller.PersonID {
			return custom_errors.New(403, custom_errors.ErrForbidden, "เฉพาะหัวหน้าทีมปัจจุบันหรือผู้ดูแลระบบเท่านั้นที่โอนสิทธิ์ได้")
		}
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

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// requireLeadOrStaff allows staff, the project creator, or the current project
// lead to perform membership mutations.
func (s *memberService) requireLeadOrStaff(ctx context.Context, caller user.Caller, projectID uuid.UUID) error {
	p, err := s.projectSvc.CanAccessProject(ctx, caller, projectID)
	if err != nil {
		return err
	}
	if caller.IsStaff() || p.CreatedByUserID == caller.UserID {
		return nil
	}
	lead, err := s.memberRepo.GetActiveLead(ctx, projectID)
	if err != nil {
		return err
	}
	if lead != nil && lead.PersonID == caller.PersonID {
		return nil
	}
	return custom_errors.New(403, custom_errors.ErrForbidden, "เฉพาะหัวหน้าทีมหรือผู้ดูแลระบบเท่านั้นที่จัดการสมาชิกได้")
}
