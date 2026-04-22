package projectsvc

import (
	"context"
	"strings"
	"time"

	"vexentra-api/internal/modules/project"
	"vexentra-api/internal/modules/txcategory"
	"vexentra-api/internal/modules/user"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateTransactionInput struct {
	CategoryID   uuid.UUID
	Amount       decimal.Decimal
	CurrencyCode string // defaults to THB if empty
	Note         *string
	OccurredAt   time.Time
}

type UpdateTransactionInput struct {
	CategoryID   uuid.UUID
	Amount       decimal.Decimal
	CurrencyCode string
	Note         *string
	OccurredAt   time.Time
}

type TransactionService interface {
	Create(ctx context.Context, caller user.Caller, projectID uuid.UUID, in CreateTransactionInput) (*project.ProjectTransaction, error)
	Get(ctx context.Context, caller user.Caller, projectID, txID uuid.UUID) (*project.ProjectTransaction, error)
	List(ctx context.Context, caller user.Caller, projectID uuid.UUID, f project.TransactionFilter, pg project.Pagination) ([]*project.ProjectTransaction, int64, error)
	Summary(ctx context.Context, caller user.Caller, projectID uuid.UUID) (*project.ProjectTotals, error)
	Update(ctx context.Context, caller user.Caller, projectID, txID uuid.UUID, in UpdateTransactionInput) (*project.ProjectTransaction, error)
	Delete(ctx context.Context, caller user.Caller, projectID, txID uuid.UUID) error
}

type transactionService struct {
	projectSvc   ProjectService
	memberRepo   project.ProjectMemberRepository
	txRepo       project.ProjectTransactionRepository
	categoryRepo txcategory.TransactionCategoryRepository
	logger       logger.Logger
}

func NewTransactionService(
	projectSvc ProjectService,
	memberRepo project.ProjectMemberRepository,
	txRepo project.ProjectTransactionRepository,
	categoryRepo txcategory.TransactionCategoryRepository,
	l logger.Logger,
) TransactionService {
	if l == nil {
		l = logger.Get()
	}
	return &transactionService{
		projectSvc:   projectSvc,
		memberRepo:   memberRepo,
		txRepo:       txRepo,
		categoryRepo: categoryRepo,
		logger:       l,
	}
}

// Create records a new income/expense line. Blocked when the project is closed —
// closed projects are historical record and must not accept new financial mutations.
func (s *transactionService) Create(ctx context.Context, caller user.Caller, projectID uuid.UUID, in CreateTransactionInput) (*project.ProjectTransaction, error) {
	p, err := s.requireWriteAccess(ctx, caller, projectID)
	if err != nil {
		return nil, err
	}
	if p.Status == project.ProjectStatusClosed {
		return nil, custom_errors.New(409, "PROJECT_CLOSED", "โปรเจกต์ปิดแล้ว ไม่สามารถบันทึกรายรับ/รายจ่ายได้")
	}

	if err := s.validateAmount(in.Amount); err != nil {
		return nil, err
	}
	currency, err := normalizeCurrency(in.CurrencyCode)
	if err != nil {
		return nil, err
	}
	if err := s.assertCategoryUsable(ctx, in.CategoryID); err != nil {
		return nil, err
	}

	t := &project.ProjectTransaction{
		ProjectID:       projectID,
		CategoryID:      in.CategoryID,
		Amount:          in.Amount,
		CurrencyCode:    currency,
		Note:            in.Note,
		OccurredAt:      in.OccurredAt,
		CreatedByUserID: caller.UserID,
	}
	if err := s.txRepo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *transactionService) Get(ctx context.Context, caller user.Caller, projectID, txID uuid.UUID) (*project.ProjectTransaction, error) {
	if _, err := s.projectSvc.CanAccessProject(ctx, caller, projectID); err != nil {
		return nil, err
	}
	t, err := s.txRepo.GetByID(ctx, txID)
	if err != nil {
		return nil, err
	}
	if t == nil || t.ProjectID != projectID {
		return nil, custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบรายการธุรกรรมนี้")
	}
	return t, nil
}

func (s *transactionService) List(ctx context.Context, caller user.Caller, projectID uuid.UUID, f project.TransactionFilter, pg project.Pagination) ([]*project.ProjectTransaction, int64, error) {
	if _, err := s.projectSvc.CanAccessProject(ctx, caller, projectID); err != nil {
		return nil, 0, err
	}
	if pg.Limit <= 0 || pg.Limit > 200 {
		pg.Limit = 50
	}
	if pg.Offset < 0 {
		pg.Offset = 0
	}
	return s.txRepo.ListByProject(ctx, projectID, f, pg)
}

func (s *transactionService) Summary(ctx context.Context, caller user.Caller, projectID uuid.UUID) (*project.ProjectTotals, error) {
	if _, err := s.projectSvc.CanAccessProject(ctx, caller, projectID); err != nil {
		return nil, err
	}
	return s.txRepo.SumByProject(ctx, projectID)
}

func (s *transactionService) Update(ctx context.Context, caller user.Caller, projectID, txID uuid.UUID, in UpdateTransactionInput) (*project.ProjectTransaction, error) {
	p, err := s.requireWriteAccess(ctx, caller, projectID)
	if err != nil {
		return nil, err
	}
	if p.Status == project.ProjectStatusClosed {
		return nil, custom_errors.New(409, "PROJECT_CLOSED", "โปรเจกต์ปิดแล้ว ไม่สามารถแก้ไขธุรกรรมได้")
	}

	t, err := s.txRepo.GetByID(ctx, txID)
	if err != nil {
		return nil, err
	}
	if t == nil || t.ProjectID != projectID {
		return nil, custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบรายการธุรกรรมนี้")
	}

	if err := s.validateAmount(in.Amount); err != nil {
		return nil, err
	}
	currency, err := normalizeCurrency(in.CurrencyCode)
	if err != nil {
		return nil, err
	}
	if err := s.assertCategoryUsable(ctx, in.CategoryID); err != nil {
		return nil, err
	}

	t.CategoryID = in.CategoryID
	t.Amount = in.Amount
	t.CurrencyCode = currency
	t.Note = in.Note
	t.OccurredAt = in.OccurredAt

	if err := s.txRepo.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *transactionService) Delete(ctx context.Context, caller user.Caller, projectID, txID uuid.UUID) error {
	p, err := s.requireWriteAccess(ctx, caller, projectID)
	if err != nil {
		return err
	}
	if p.Status == project.ProjectStatusClosed {
		return custom_errors.New(409, "PROJECT_CLOSED", "โปรเจกต์ปิดแล้ว ไม่สามารถลบธุรกรรมได้")
	}
	t, err := s.txRepo.GetByID(ctx, txID)
	if err != nil {
		return err
	}
	if t == nil || t.ProjectID != projectID {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบรายการธุรกรรมนี้")
	}
	return s.txRepo.Delete(ctx, txID)
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// requireWriteAccess returns the project if the caller may write transactions.
// Staff, creator, project lead, and any active member are all permitted —
// transactions are day-to-day logging, not governance.
func (s *transactionService) requireWriteAccess(ctx context.Context, caller user.Caller, projectID uuid.UUID) (*project.Project, error) {
	p, err := s.projectSvc.CanAccessProject(ctx, caller, projectID)
	if err != nil {
		return nil, err
	}
	// CanAccessProject already enforces staff/creator/active-member gate; that
	// is exactly the set allowed to write transactions.
	return p, nil
}

func (s *transactionService) assertCategoryUsable(ctx context.Context, categoryID uuid.UUID) error {
	c, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return err
	}
	if c == nil {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบหมวดหมู่ธุรกรรมนี้")
	}
	if !c.IsActive {
		return custom_errors.New(400, custom_errors.ErrValidation, "หมวดหมู่นี้ถูกปิดใช้งานแล้ว")
	}
	return nil
}

func (s *transactionService) validateAmount(a decimal.Decimal) error {
	if a.IsNegative() {
		return custom_errors.New(400, custom_errors.ErrValidation, "จำนวนเงินต้องไม่เป็นลบ")
	}
	// NUMERIC(15,2) — max absolute value 9999999999999.99 (13 digits before decimal).
	if a.GreaterThanOrEqual(decimal.NewFromInt(10_000_000_000_000)) {
		return custom_errors.New(400, custom_errors.ErrValidation, "จำนวนเงินเกินค่าที่รองรับ")
	}
	return nil
}

func normalizeCurrency(code string) (string, error) {
	c := strings.ToUpper(strings.TrimSpace(code))
	if c == "" {
		return "THB", nil
	}
	if len(c) != 3 {
		return "", custom_errors.New(400, custom_errors.ErrValidation, "รหัสสกุลเงินต้องเป็นตัวอักษร 3 ตัว (ISO 4217)")
	}
	for _, r := range c {
		if r < 'A' || r > 'Z' {
			return "", custom_errors.New(400, custom_errors.ErrValidation, "รหัสสกุลเงินต้องเป็นตัวพิมพ์ใหญ่ A-Z")
		}
	}
	return c, nil
}
