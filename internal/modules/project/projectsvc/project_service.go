package projectsvc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"vexentra-api/internal/adapters/database/postgres/pgtx"
	"vexentra-api/internal/modules/project"
	"vexentra-api/internal/modules/user"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"
	"vexentra-api/pkg/wela"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CreateProjectInput is the narrowed surface the service exposes to handlers.
// Status is always created as 'draft' — promotion happens via Update/Close.
type CreateProjectInput struct {
	Name                      string
	ProjectKind               *project.ProjectKind
	Description               *string
	ClientPersonID            *uuid.UUID
	InitialLeadPersonID       *uuid.UUID
	ContractFinanceVisibility *project.FinanceVisibility
	ExpenseFinanceVisibility  *project.FinanceVisibility
	ClientNameRaw             *string
	ClientEmailRaw            *string
	ScheduledStartAt          *time.Time
	DeadlineAt                *time.Time
}

// UpdateProjectInput carries patchable fields. Status transitions and closure
// are handled separately by CloseProject / ReopenProject so each has its own
// invariant check.
type UpdateProjectInput struct {
	Name                      string
	ProjectKind               *project.ProjectKind
	Description               *string
	ClientPersonID            *uuid.UUID
	ContractFinanceVisibility *project.FinanceVisibility
	ExpenseFinanceVisibility  *project.FinanceVisibility
	ClientNameRaw             *string
	ClientEmailRaw            *string
	ScheduledStartAt          *time.Time
	DeadlineAt                *time.Time
	Status                    project.ProjectStatus // must not be 'closed' — use CloseProject
}

type CloseProjectInput struct {
	Reason   project.ProjectClosureReason
	ClosedAt *time.Time
}

type ProjectPaymentInstallmentInput struct {
	SortOrder           int
	Title               string
	Amount              decimal.Decimal
	PlannedDeliveryDate *time.Time
	PlannedReceiveDate  *time.Time
	Note                *string
}

type UpsertFinancialPlanInput struct {
	ContractAmount      decimal.Decimal
	RetentionAmount     decimal.Decimal
	PlannedDeliveryDate *time.Time
	PaymentNote         *string
	Installments        []ProjectPaymentInstallmentInput
}

type ProjectService interface {
	Create(ctx context.Context, caller user.Caller, in CreateProjectInput) (*project.Project, error)

	Get(ctx context.Context, caller user.Caller, id uuid.UUID) (*project.Project, error)
	GetByCode(ctx context.Context, caller user.Caller, code string) (*project.Project, error)
	List(ctx context.Context, caller user.Caller, f project.ProjectFilter, pg project.Pagination) ([]*project.Project, int64, error)
	ListStatuses(ctx context.Context, caller user.Caller, activeOnly bool) ([]project.ProjectStatusMeta, error)
	GetFinancialPlan(ctx context.Context, caller user.Caller, id uuid.UUID) (*project.ProjectFinancialPlan, error)

	Update(ctx context.Context, caller user.Caller, id uuid.UUID, in UpdateProjectInput) (*project.Project, error)
	Close(ctx context.Context, caller user.Caller, id uuid.UUID, in CloseProjectInput) (*project.Project, error)
	UpsertFinancialPlan(ctx context.Context, caller user.Caller, id uuid.UUID, in UpsertFinancialPlanInput) (*project.ProjectFinancialPlan, error)
	Delete(ctx context.Context, caller user.Caller, id uuid.UUID) error

	// CanAccessProject is exposed so other services (transactions, members)
	// reuse the same access rule: staff bypass, members must have an active
	// project_members row, creator always allowed on their own project.
	CanAccessProject(ctx context.Context, caller user.Caller, id uuid.UUID) (*project.Project, error)
	RequireContractFinanceRead(ctx context.Context, caller user.Caller, id uuid.UUID) (*project.Project, error)
	RequireExpenseFinanceRead(ctx context.Context, caller user.Caller, id uuid.UUID) (*project.Project, error)
}

type projectService struct {
	db          *gorm.DB
	projectRepo project.ProjectRepository
	memberRepo  project.ProjectMemberRepository
	codePrefix  string
	logger      logger.Logger
}

const projectCoordinatorRoleCode = "coordinator"

func NewProjectService(
	db *gorm.DB,
	projectRepo project.ProjectRepository,
	memberRepo project.ProjectMemberRepository,
	codePrefix string,
	l logger.Logger,
) ProjectService {
	if l == nil {
		l = logger.Get()
	}
	return &projectService{
		db:          db,
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
		codePrefix:  strings.ToUpper(codePrefix),
		logger:      l,
	}
}

// -----------------------------------------------------------------------------
// Commands
// -----------------------------------------------------------------------------

// Create inserts a new Project (status='draft'). Initial lead is optional:
// if InitialLeadPersonID is provided, that person is added as the first lead.
// This keeps creator/audit identity separate from team responsibility.
func (s *projectService) Create(ctx context.Context, caller user.Caller, in CreateProjectInput) (*project.Project, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "ชื่อโปรเจกต์ห้ามว่าง")
	}

	contractVisibility := project.FinanceVisibilityAllMembers
	if in.ContractFinanceVisibility != nil {
		contractVisibility = *in.ContractFinanceVisibility
	}
	expenseVisibility := project.FinanceVisibilityAllMembers
	if in.ExpenseFinanceVisibility != nil {
		expenseVisibility = *in.ExpenseFinanceVisibility
	}
	if !isValidFinanceVisibility(contractVisibility) || !isValidFinanceVisibility(expenseVisibility) {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "ค่า finance visibility ไม่ถูกต้อง")
	}
	projectKind := project.ProjectKindClientDelivery
	if in.ProjectKind != nil {
		projectKind = *in.ProjectKind
	}
	if !isValidProjectKind(projectKind) {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "project_kind ไม่ถูกต้อง")
	}

	var created *project.Project
	err := pgtx.Run(ctx, s.db, func(ctx context.Context) error {
		seq, err := s.projectRepo.NextCodeSeq(ctx)
		if err != nil {
			return err
		}
		code := fmt.Sprintf("%s-%d-%04d", s.codePrefix, wela.NowUTC().Year(), seq)

		p := &project.Project{
			ProjectCode:               code,
			Name:                      name,
			Description:               in.Description,
			Kind:                      projectKind,
			Status:                    project.ProjectStatusDraft,
			ContractFinanceVisibility: contractVisibility,
			ExpenseFinanceVisibility:  expenseVisibility,
			ClientPersonID:            in.ClientPersonID,
			ClientNameRaw:             in.ClientNameRaw,
			ClientEmailRaw:            in.ClientEmailRaw,
			ScheduledStartAt:          in.ScheduledStartAt,
			DeadlineAt:                in.DeadlineAt,
			CreatedByUserID:           caller.UserID,
		}
		if err := s.projectRepo.Create(ctx, p); err != nil {
			return err
		}

		if in.InitialLeadPersonID != nil {
			lead := &project.ProjectMember{
				ProjectID:     p.ID,
				PersonID:      *in.InitialLeadPersonID,
				IsLead:        true,
				AddedByUserID: caller.UserID,
			}
			if err := s.memberRepo.Add(ctx, lead); err != nil {
				return err
			}
		}

		created = p
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// Update allows editing mutable fields + non-terminal status transitions.
// Closing is routed through Close() which enforces the closed_at+reason bijection.
func (s *projectService) Update(ctx context.Context, caller user.Caller, id uuid.UUID, in UpdateProjectInput) (*project.Project, error) {
	p, err := s.requireModifyAccess(ctx, caller, id)
	if err != nil {
		return nil, err
	}
	if p.Status == project.ProjectStatusClosed && !caller.IsStaff() {
		return nil, custom_errors.New(409, "PROJECT_CLOSED", "โปรเจกต์ปิดแล้ว ไม่สามารถแก้ไขได้ — ใช้ Reopen ก่อน")
	}
	if in.Status == project.ProjectStatusClosed {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "ปิดโปรเจกต์ต้องใช้ endpoint close (ระบุ closure_reason)")
	}
	if !isValidOpenStatus(in.Status) {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "สถานะโปรเจกต์ไม่ถูกต้อง")
	}

	// Promote to active/on_hold/closed requires a client — mirrors DB CHECK.
	nextKind := p.Kind
	if in.ProjectKind != nil {
		if !isValidProjectKind(*in.ProjectKind) {
			return nil, custom_errors.New(400, custom_errors.ErrValidation, "project_kind ไม่ถูกต้อง")
		}
		nextKind = *in.ProjectKind
	}
	if requiresClient(in.Status, nextKind) && in.ClientPersonID == nil {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "ต้องระบุลูกค้า (client_person_id) ก่อนเปลี่ยนสถานะเป็น active/on_hold")
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "ชื่อโปรเจกต์ห้ามว่าง")
	}

	// Stamp activated_at the first time the project enters 'active'.
	if in.Status == project.ProjectStatusActive && p.ActivatedAt == nil {
		lead, leadErr := s.memberRepo.GetActiveLead(ctx, p.ID)
		if leadErr != nil {
			return nil, leadErr
		}
		if lead == nil {
			return nil, custom_errors.New(400, custom_errors.ErrValidation, "ก่อนเปลี่ยนเป็น active ต้องกำหนดหัวหน้าทีมก่อน")
		}
		now := wela.NowUTC()
		p.ActivatedAt = &now
	}

	p.Name = name
	p.Kind = nextKind
	p.Description = in.Description
	p.Status = in.Status
	if in.ContractFinanceVisibility != nil {
		if !isValidFinanceVisibility(*in.ContractFinanceVisibility) {
			return nil, custom_errors.New(400, custom_errors.ErrValidation, "contract_finance_visibility ไม่ถูกต้อง")
		}
		p.ContractFinanceVisibility = *in.ContractFinanceVisibility
	}
	if in.ExpenseFinanceVisibility != nil {
		if !isValidFinanceVisibility(*in.ExpenseFinanceVisibility) {
			return nil, custom_errors.New(400, custom_errors.ErrValidation, "expense_finance_visibility ไม่ถูกต้อง")
		}
		p.ExpenseFinanceVisibility = *in.ExpenseFinanceVisibility
	}
	if in.Status != project.ProjectStatusClosed {
		// Reopen path: when moving from terminal status back to open statuses,
		// closed markers must be cleared to satisfy DB consistency checks.
		p.ClosedAt = nil
		p.ClosureReason = nil
	}
	p.ClientPersonID = in.ClientPersonID
	p.ClientNameRaw = in.ClientNameRaw
	p.ClientEmailRaw = in.ClientEmailRaw
	p.ScheduledStartAt = in.ScheduledStartAt
	p.DeadlineAt = in.DeadlineAt

	if err := s.projectRepo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Close transitions a project to 'closed' and stamps closed_at + closure_reason.
// Idempotent guard: re-closing an already-closed project returns 409.
func (s *projectService) Close(ctx context.Context, caller user.Caller, id uuid.UUID, in CloseProjectInput) (*project.Project, error) {
	p, err := s.requireModifyAccess(ctx, caller, id)
	if err != nil {
		return nil, err
	}
	if p.Status == project.ProjectStatusClosed {
		return nil, custom_errors.New(409, "PROJECT_ALREADY_CLOSED", "โปรเจกต์ปิดไปแล้ว")
	}
	if !isValidClosureReason(in.Reason) {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "closure_reason ไม่ถูกต้อง")
	}
	if in.ClosedAt == nil {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "ต้องระบุวันที่ปิดโครงการ (closed_at)")
	}
	if p.Kind == project.ProjectKindClientDelivery && p.ClientPersonID == nil {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "โปรเจกต์ต้องมีลูกค้าก่อนปิดงาน")
	}

	closedAt := in.ClosedAt.UTC()
	reason := in.Reason
	p.Status = project.ProjectStatusClosed
	p.ClosureReason = &reason
	p.ClosedAt = &closedAt

	if err := s.projectRepo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Delete soft-deletes a project.
// Permission model follows governance writes: staff, lead, coordinator.
func (s *projectService) Delete(ctx context.Context, caller user.Caller, id uuid.UUID) error {
	p, err := s.requireModifyAccess(ctx, caller, id)
	if err != nil {
		return err
	}
	if p.Status == project.ProjectStatusClosed && !caller.IsStaff() {
		return custom_errors.New(409, "PROJECT_CLOSED", "โปรเจกต์ปิดแล้ว ไม่สามารถลบได้")
	}
	return s.projectRepo.Delete(ctx, id)
}

// -----------------------------------------------------------------------------
// Queries
// -----------------------------------------------------------------------------

func (s *projectService) Get(ctx context.Context, caller user.Caller, id uuid.UUID) (*project.Project, error) {
	return s.CanAccessProject(ctx, caller, id)
}

func (s *projectService) GetByCode(ctx context.Context, caller user.Caller, code string) (*project.Project, error) {
	p, err := s.projectRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบโปรเจกต์นี้")
	}
	// Same access rules as CanAccessProject — reuse loaded project to avoid second DB hit
	if caller.IsStaff() || p.CreatedByUserID == caller.UserID {
		return p, nil
	}
	m, err := s.memberRepo.GetActiveByProjectAndPerson(ctx, p.ID, caller.PersonID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, custom_errors.New(403, custom_errors.ErrForbidden, "ไม่มีสิทธิ์เข้าถึงโปรเจกต์นี้")
	}
	return p, nil
}

// List restricts non-staff callers to projects they are an active member of.
// Staff see everything; filter fields remain respected for both.
func (s *projectService) List(ctx context.Context, caller user.Caller, f project.ProjectFilter, pg project.Pagination) ([]*project.Project, int64, error) {
	if !caller.IsStaff() {
		// Force scope to caller's active memberships. If the handler already
		// set MemberPersonID to something else, refuse rather than silently
		// broaden — no peeking at other people's rosters.
		if f.MemberPersonID != nil && *f.MemberPersonID != caller.PersonID {
			return nil, 0, custom_errors.New(403, custom_errors.ErrForbidden, "ไม่สามารถดูโปรเจกต์ของคนอื่นได้")
		}
		pid := caller.PersonID
		f.MemberPersonID = &pid
	}
	if pg.Limit <= 0 || pg.Limit > 100 {
		pg.Limit = 20
	}
	if pg.Offset < 0 {
		pg.Offset = 0
	}
	return s.projectRepo.List(ctx, f, pg)
}

func (s *projectService) ListStatuses(ctx context.Context, caller user.Caller, activeOnly bool) ([]project.ProjectStatusMeta, error) {
	// Any authenticated user can read project status master data.
	_ = caller
	return s.projectRepo.ListStatuses(ctx, activeOnly)
}

func (s *projectService) GetFinancialPlan(ctx context.Context, caller user.Caller, id uuid.UUID) (*project.ProjectFinancialPlan, error) {
	if _, err := s.RequireContractFinanceRead(ctx, caller, id); err != nil {
		return nil, err
	}
	return s.projectRepo.GetFinancialPlan(ctx, id)
}

func (s *projectService) UpsertFinancialPlan(ctx context.Context, caller user.Caller, id uuid.UUID, in UpsertFinancialPlanInput) (*project.ProjectFinancialPlan, error) {
	p, err := s.requireModifyAccess(ctx, caller, id)
	if err != nil {
		return nil, err
	}
	if p.Status == project.ProjectStatusClosed && !caller.IsStaff() {
		return nil, custom_errors.New(409, "PROJECT_CLOSED", "โปรเจกต์ปิดแล้ว ไม่สามารถแก้ไขแผนค่าจ้างได้")
	}
	if p.Kind == project.ProjectKindInternalContinuous {
		if len(in.Installments) > 0 {
			return nil, custom_errors.New(400, custom_errors.ErrValidation, "โปรเจกต์ภายในไม่รองรับการแบ่งงวดชำระ")
		}
		if in.PlannedDeliveryDate != nil {
			return nil, custom_errors.New(400, custom_errors.ErrValidation, "โปรเจกต์ภายในไม่บังคับวันส่งมอบ")
		}
	}

	if in.ContractAmount.IsNegative() {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "ค่าจ้างตามสัญญาต้องไม่เป็นค่าติดลบ")
	}
	if in.RetentionAmount.IsNegative() {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "เงินประกันต้องไม่เป็นค่าติดลบ")
	}
	if in.RetentionAmount.GreaterThan(in.ContractAmount) {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "เงินประกันต้องไม่เกินค่าจ้างตามสัญญา")
	}

	netReceivable := in.ContractAmount.Sub(in.RetentionAmount)
	installments := make([]*project.ProjectPaymentInstallment, 0, len(in.Installments))
	sumInstallments := decimal.Zero

	for i := range in.Installments {
		item := in.Installments[i]
		title := strings.TrimSpace(item.Title)
		if title == "" {
			return nil, custom_errors.New(400, custom_errors.ErrValidation, "ชื่องวดห้ามว่าง")
		}
		if item.Amount.LessThanOrEqual(decimal.Zero) {
			return nil, custom_errors.New(400, custom_errors.ErrValidation, "จำนวนเงินในงวดต้องมากกว่า 0")
		}
		sortOrder := item.SortOrder
		if sortOrder <= 0 {
			sortOrder = i + 1
		}
		id, idErr := uuid.NewV7()
		if idErr != nil {
			return nil, custom_errors.NewInternalError("ไม่สามารถสร้างรหัสงวดชำระได้")
		}

		sumInstallments = sumInstallments.Add(item.Amount)
		installments = append(installments, &project.ProjectPaymentInstallment{
			ID:                  id,
			ProjectID:           p.ID,
			SortOrder:           sortOrder,
			Title:               title,
			Amount:              item.Amount,
			PlannedDeliveryDate: item.PlannedDeliveryDate,
			PlannedReceiveDate:  item.PlannedReceiveDate,
			Note:                item.Note,
		})
	}

	if sumInstallments.GreaterThan(netReceivable) {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "ยอดรวมงวดชำระต้องไม่เกินยอดสุทธิหลังหักเงินประกัน")
	}

	plan := &project.ProjectFinancialPlan{
		ProjectID:            p.ID,
		ContractAmount:       in.ContractAmount,
		RetentionAmount:      in.RetentionAmount,
		PlannedDeliveryDate:  in.PlannedDeliveryDate,
		PaymentNote:          in.PaymentNote,
		Installments:         installments,
		InstallmentsTotal:    sumInstallments,
		NetReceivable:        netReceivable,
		UnallocatedRemaining: netReceivable.Sub(sumInstallments),
	}

	if err := s.projectRepo.UpsertFinancialPlan(ctx, plan); err != nil {
		return nil, err
	}
	return s.projectRepo.GetFinancialPlan(ctx, p.ID)
}

// CanAccessProject returns the project if caller is allowed to read it, else
// 403/404. It is also used by Member/Transaction services — keeps the rule in one place.
func (s *projectService) CanAccessProject(ctx context.Context, caller user.Caller, id uuid.UUID) (*project.Project, error) {
	p, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบโปรเจกต์นี้")
	}
	if caller.IsStaff() || p.CreatedByUserID == caller.UserID {
		return p, nil
	}
	m, err := s.memberRepo.GetActiveByProjectAndPerson(ctx, id, caller.PersonID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, custom_errors.New(403, custom_errors.ErrForbidden, "ไม่มีสิทธิ์เข้าถึงโปรเจกต์นี้")
	}
	return p, nil
}

func (s *projectService) RequireContractFinanceRead(ctx context.Context, caller user.Caller, id uuid.UUID) (*project.Project, error) {
	p, err := s.CanAccessProject(ctx, caller, id)
	if err != nil {
		return nil, err
	}
	ok, err := s.canReadByVisibility(ctx, caller, id, p.ContractFinanceVisibility)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, custom_errors.New(403, custom_errors.ErrForbidden, "ไม่มีสิทธิ์ดูข้อมูลส่วนว่าจ้างของโครงการนี้")
	}
	return p, nil
}

func (s *projectService) RequireExpenseFinanceRead(ctx context.Context, caller user.Caller, id uuid.UUID) (*project.Project, error) {
	p, err := s.CanAccessProject(ctx, caller, id)
	if err != nil {
		return nil, err
	}
	ok, err := s.canReadByVisibility(ctx, caller, id, p.ExpenseFinanceVisibility)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, custom_errors.New(403, custom_errors.ErrForbidden, "ไม่มีสิทธิ์ดูข้อมูลส่วนค่าใช้จ่ายของโครงการนี้")
	}
	return p, nil
}

// -----------------------------------------------------------------------------
// Internal helpers
// -----------------------------------------------------------------------------

// requireModifyAccess permits staff, current lead, or assigned coordinator.
// Plain members are read-only on project governance actions.
func (s *projectService) requireModifyAccess(ctx context.Context, caller user.Caller, id uuid.UUID) (*project.Project, error) {
	p, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบโปรเจกต์นี้")
	}
	if caller.IsStaff() {
		return p, nil
	}
	lead, err := s.memberRepo.GetActiveLead(ctx, id)
	if err != nil {
		return nil, err
	}
	if lead != nil && lead.PersonID == caller.PersonID {
		return p, nil
	}
	isCoordinator, err := s.memberRepo.HasActiveRoleCode(ctx, id, caller.PersonID, projectCoordinatorRoleCode)
	if err != nil {
		return nil, err
	}
	if isCoordinator {
		return p, nil
	}
	return nil, custom_errors.New(403, custom_errors.ErrForbidden, "เฉพาะหัวหน้าทีม, coordinator หรือผู้ดูแลระบบเท่านั้นที่แก้ไขโปรเจกต์ได้")
}

func isValidOpenStatus(s project.ProjectStatus) bool {
	switch s {
	case project.ProjectStatusDraft,
		project.ProjectStatusPlanned,
		project.ProjectStatusBidding,
		project.ProjectStatusActive,
		project.ProjectStatusOnHold:
		return true
	}
	return false
}

func requiresClient(s project.ProjectStatus, kind project.ProjectKind) bool {
	if kind == project.ProjectKindInternalContinuous {
		return false
	}
	switch s {
	case project.ProjectStatusActive, project.ProjectStatusOnHold:
		return true
	}
	return false
}

func isValidProjectKind(k project.ProjectKind) bool {
	switch k {
	case project.ProjectKindClientDelivery, project.ProjectKindInternalContinuous:
		return true
	}
	return false
}

func isValidClosureReason(r project.ProjectClosureReason) bool {
	switch r {
	case project.ClosureReasonWonCompleted,
		project.ClosureReasonLostToCompetitor,
		project.ClosureReasonClientDeclined,
		project.ClosureReasonWeWithdrew,
		project.ClosureReasonClientAbandoned,
		project.ClosureReasonBudgetCut,
		project.ClosureReasonCancelledInternal:
		return true
	}
	return false
}

func isValidFinanceVisibility(v project.FinanceVisibility) bool {
	switch v {
	case project.FinanceVisibilityAllMembers,
		project.FinanceVisibilityLeadCoordinatorStaff,
		project.FinanceVisibilityStaffOnly:
		return true
	}
	return false
}

func (s *projectService) canReadByVisibility(ctx context.Context, caller user.Caller, projectID uuid.UUID, visibility project.FinanceVisibility) (bool, error) {
	if caller.IsStaff() {
		return true, nil
	}
	switch visibility {
	case project.FinanceVisibilityAllMembers:
		return true, nil
	case project.FinanceVisibilityLeadCoordinatorStaff:
		lead, err := s.memberRepo.GetActiveLead(ctx, projectID)
		if err != nil {
			return false, err
		}
		if lead != nil && lead.PersonID == caller.PersonID {
			return true, nil
		}
		isCoordinator, err := s.memberRepo.HasActiveRoleCode(ctx, projectID, caller.PersonID, projectCoordinatorRoleCode)
		if err != nil {
			return false, err
		}
		return isCoordinator, nil
	case project.FinanceVisibilityStaffOnly:
		return false, nil
	default:
		return false, nil
	}
}
