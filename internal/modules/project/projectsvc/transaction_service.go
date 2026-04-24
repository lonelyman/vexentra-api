package projectsvc

import (
	"context"
	"errors"
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
	// ListForExport returns all rows for a project joined with category data, for CSV generation.
	ListForExport(ctx context.Context, caller user.Caller, projectID uuid.UUID) ([]*project.TransactionExportRow, error)
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
	cat, err := s.getUsableCategory(ctx, in.CategoryID)
	if err != nil {
		return nil, err
	}
	if cat.Type == txcategory.TransactionTypeExpense {
		if _, err := s.projectSvc.RequireExpenseFinanceRead(ctx, caller, projectID); err != nil {
			return nil, err
		}
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
	if err := s.ensureExpenseReadableByCategoryID(ctx, caller, projectID, t.CategoryID); err != nil {
		return nil, err
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
	items, total, err := s.txRepo.ListByProject(ctx, projectID, f, pg)
	if err != nil {
		return nil, 0, err
	}
	canReadExpense, err := s.canReadExpense(ctx, caller, projectID)
	if err != nil {
		return nil, 0, err
	}
	if canReadExpense {
		return items, total, nil
	}
	filtered, ferr := s.filterOutExpenseTransactions(ctx, items)
	if ferr != nil {
		return nil, 0, ferr
	}
	return filtered, int64(len(filtered)), nil
}

func (s *transactionService) Summary(ctx context.Context, caller user.Caller, projectID uuid.UUID) (*project.ProjectTotals, error) {
	if _, err := s.projectSvc.CanAccessProject(ctx, caller, projectID); err != nil {
		return nil, err
	}
	totals, err := s.txRepo.SumByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	canReadExpense, err := s.canReadExpense(ctx, caller, projectID)
	if err != nil {
		return nil, err
	}
	if canReadExpense {
		return totals, nil
	}
	return &project.ProjectTotals{
		Income:  totals.Income,
		Expense: decimal.Zero,
		Net:     totals.Income,
	}, nil
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
	cat, err := s.getUsableCategory(ctx, in.CategoryID)
	if err != nil {
		return nil, err
	}
	if cat.Type == txcategory.TransactionTypeExpense {
		if _, err := s.projectSvc.RequireExpenseFinanceRead(ctx, caller, projectID); err != nil {
			return nil, err
		}
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

func (s *transactionService) ListForExport(ctx context.Context, caller user.Caller, projectID uuid.UUID) ([]*project.TransactionExportRow, error) {
	if _, err := s.projectSvc.CanAccessProject(ctx, caller, projectID); err != nil {
		return nil, err
	}
	rows, err := s.txRepo.ListForExport(ctx, projectID)
	if err != nil {
		return nil, err
	}
	canReadExpense, err := s.canReadExpense(ctx, caller, projectID)
	if err != nil {
		return nil, err
	}
	if canReadExpense {
		return rows, nil
	}
	filtered := make([]*project.TransactionExportRow, 0, len(rows))
	for _, row := range rows {
		if row.CategoryType == string(txcategory.TransactionTypeExpense) {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered, nil
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
	if err := s.ensureExpenseReadableByCategoryID(ctx, caller, projectID, t.CategoryID); err != nil {
		return err
	}
	return s.txRepo.Delete(ctx, txID)
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// requireWriteAccess returns the project if caller may write transactions.
// Allowed: staff, current lead, or coordinator. Plain members are blocked.
func (s *transactionService) requireWriteAccess(ctx context.Context, caller user.Caller, projectID uuid.UUID) (*project.Project, error) {
	p, err := s.projectSvc.CanAccessProject(ctx, caller, projectID)
	if err != nil {
		return nil, err
	}
	if caller.IsStaff() {
		return p, nil
	}
	lead, err := s.memberRepo.GetActiveLead(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if lead != nil && lead.PersonID == caller.PersonID {
		return p, nil
	}
	isCoordinator, err := s.memberRepo.HasActiveRoleCode(ctx, projectID, caller.PersonID, projectCoordinatorRoleCode)
	if err != nil {
		return nil, err
	}
	if isCoordinator {
		return p, nil
	}
	return nil, custom_errors.New(403, custom_errors.ErrForbidden, "member ทั่วไปไม่สามารถแก้ไขธุรกรรมได้ (ต้องเป็นหัวหน้าทีม, coordinator หรือผู้ดูแลระบบ)")
}

func (s *transactionService) getUsableCategory(ctx context.Context, categoryID uuid.UUID) (*txcategory.TransactionCategory, error) {
	c, err := s.getCategoryByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if !c.IsActive {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "หมวดหมู่นี้ถูกปิดใช้งานแล้ว")
	}
	return c, nil
}

func (s *transactionService) getCategoryByID(ctx context.Context, categoryID uuid.UUID) (*txcategory.TransactionCategory, error) {
	c, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบหมวดหมู่ธุรกรรมนี้")
	}
	return c, nil
}

func (s *transactionService) ensureExpenseReadableByCategoryID(ctx context.Context, caller user.Caller, projectID, categoryID uuid.UUID) error {
	cat, err := s.getCategoryByID(ctx, categoryID)
	if err != nil {
		return err
	}
	if cat.Type != txcategory.TransactionTypeExpense {
		return nil
	}
	_, err = s.projectSvc.RequireExpenseFinanceRead(ctx, caller, projectID)
	return err
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

func (s *transactionService) canReadExpense(ctx context.Context, caller user.Caller, projectID uuid.UUID) (bool, error) {
	_, err := s.projectSvc.RequireExpenseFinanceRead(ctx, caller, projectID)
	if err == nil {
		return true, nil
	}
	var appErr *custom_errors.AppError
	if errors.As(err, &appErr) && appErr.Code == custom_errors.ErrForbidden {
		return false, nil
	}
	return false, err
}

func (s *transactionService) filterOutExpenseTransactions(ctx context.Context, items []*project.ProjectTransaction) ([]*project.ProjectTransaction, error) {
	if len(items) == 0 {
		return items, nil
	}
	typeCache := make(map[uuid.UUID]txcategory.TransactionType, len(items))
	filtered := make([]*project.ProjectTransaction, 0, len(items))
	for _, item := range items {
		tt, ok := typeCache[item.CategoryID]
		if !ok {
			cat, err := s.getCategoryByID(ctx, item.CategoryID)
			if err != nil {
				return nil, err
			}
			tt = cat.Type
			typeCache[item.CategoryID] = tt
		}
		if tt == txcategory.TransactionTypeExpense {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered, nil
}
