package txcategorysvc

import (
	"context"
	"regexp"
	"strings"

	"vexentra-api/internal/modules/txcategory"
	"vexentra-api/internal/modules/user"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
)

type CreateCategoryInput struct {
	Code      string
	Name      string
	Type      txcategory.TransactionType
	IconKey   *string
	IsActive  bool
	SortOrder int
}

type UpdateCategoryInput struct {
	Name      string
	IconKey   *string
	IsActive  bool
	SortOrder int
}

type TransactionCategoryService interface {
	Create(ctx context.Context, caller user.Caller, in CreateCategoryInput) (*txcategory.TransactionCategory, error)
	Get(ctx context.Context, caller user.Caller, id uuid.UUID) (*txcategory.TransactionCategory, error)
	List(ctx context.Context, caller user.Caller, f txcategory.TransactionCategoryFilter) ([]*txcategory.TransactionCategory, error)
	Update(ctx context.Context, caller user.Caller, id uuid.UUID, in UpdateCategoryInput) (*txcategory.TransactionCategory, error)
	Delete(ctx context.Context, caller user.Caller, id uuid.UUID) error
}

type transactionCategoryService struct {
	repo   txcategory.TransactionCategoryRepository
	logger logger.Logger
}

func NewTransactionCategoryService(repo txcategory.TransactionCategoryRepository, l logger.Logger) TransactionCategoryService {
	if l == nil {
		l = logger.Get()
	}
	return &transactionCategoryService{repo: repo, logger: l}
}

var codePattern = regexp.MustCompile(`^[a-z0-9_]+$`)

// Create adds a new (non-system) transaction category. Admin-only.
// IsSystem is always false — system categories arrive via migration seed, not runtime.
func (s *transactionCategoryService) Create(ctx context.Context, caller user.Caller, in CreateCategoryInput) (*txcategory.TransactionCategory, error) {
	if !caller.IsAdmin() {
		return nil, custom_errors.New(403, custom_errors.ErrForbidden, "เฉพาะผู้ดูแลระบบเท่านั้นที่จัดการหมวดหมู่ธุรกรรมได้")
	}

	code := strings.TrimSpace(strings.ToLower(in.Code))
	if !codePattern.MatchString(code) {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "รหัสหมวดหมู่ต้องเป็นตัวพิมพ์เล็ก ตัวเลข หรือ _ เท่านั้น")
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "ต้องระบุชื่อหมวดหมู่")
	}
	if err := validateType(in.Type); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, custom_errors.New(409, "CATEGORY_CODE_EXISTS", "รหัสหมวดหมู่นี้มีอยู่แล้ว")
	}

	c := &txcategory.TransactionCategory{
		Code:      code,
		Name:      name,
		Type:      in.Type,
		IconKey:   in.IconKey,
		IsSystem:  false,
		IsActive:  in.IsActive,
		SortOrder: in.SortOrder,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Get returns a single category. Any authenticated caller may read.
func (s *transactionCategoryService) Get(ctx context.Context, caller user.Caller, id uuid.UUID) (*txcategory.TransactionCategory, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบหมวดหมู่ธุรกรรมนี้")
	}
	return c, nil
}

// List returns categories. Non-admins cannot see soft-deleted rows regardless of filter.
func (s *transactionCategoryService) List(ctx context.Context, caller user.Caller, f txcategory.TransactionCategoryFilter) ([]*txcategory.TransactionCategory, error) {
	if f.IncludeDeleted && !caller.IsAdmin() {
		f.IncludeDeleted = false
	}
	return s.repo.List(ctx, f)
}

// Update edits mutable fields. System categories allow only is_active/sort_order/icon/name edits —
// which is exactly the field set exposed by UpdateCategoryInput, so no extra branching is needed.
// Code and Type are immutable post-creation.
func (s *transactionCategoryService) Update(ctx context.Context, caller user.Caller, id uuid.UUID, in UpdateCategoryInput) (*txcategory.TransactionCategory, error) {
	if !caller.IsAdmin() {
		return nil, custom_errors.New(403, custom_errors.ErrForbidden, "เฉพาะผู้ดูแลระบบเท่านั้นที่จัดการหมวดหมู่ธุรกรรมได้")
	}
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบหมวดหมู่ธุรกรรมนี้")
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "ต้องระบุชื่อหมวดหมู่")
	}
	c.Name = name
	c.IconKey = in.IconKey
	c.IsActive = in.IsActive
	c.SortOrder = in.SortOrder

	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Delete soft-deletes a user-created category. System categories are protected —
// admins should toggle IsActive=false instead.
func (s *transactionCategoryService) Delete(ctx context.Context, caller user.Caller, id uuid.UUID) error {
	if !caller.IsAdmin() {
		return custom_errors.New(403, custom_errors.ErrForbidden, "เฉพาะผู้ดูแลระบบเท่านั้นที่จัดการหมวดหมู่ธุรกรรมได้")
	}
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if c == nil {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบหมวดหมู่ธุรกรรมนี้")
	}
	if c.IsSystem {
		return custom_errors.New(409, "CATEGORY_IS_SYSTEM", "ไม่สามารถลบหมวดหมู่ระบบได้ — ให้ปิดการใช้งานแทน")
	}
	return s.repo.Delete(ctx, id)
}

func validateType(t txcategory.TransactionType) error {
	switch t {
	case txcategory.TransactionTypeIncome, txcategory.TransactionTypeExpense:
		return nil
	default:
		return custom_errors.New(400, custom_errors.ErrValidation, "ประเภทหมวดหมู่ต้องเป็น income หรือ expense")
	}
}
