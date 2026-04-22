package txcategory

import (
	"context"

	"github.com/google/uuid"
)

// TransactionCategoryFilter narrows List queries. Zero value = return all non-deleted.
type TransactionCategoryFilter struct {
	Type           *TransactionType // nil = both income & expense
	ActiveOnly     bool             // true = only IsActive rows (for user-facing pickers)
	IncludeDeleted bool             // true = include soft-deleted rows (admin audit only)
}

type TransactionCategoryRepository interface {
	Create(ctx context.Context, c *TransactionCategory) error
	GetByID(ctx context.Context, id uuid.UUID) (*TransactionCategory, error)
	GetByCode(ctx context.Context, code string) (*TransactionCategory, error)

	// List returns categories ordered by (type ASC, sort_order ASC, name ASC).
	List(ctx context.Context, f TransactionCategoryFilter) ([]*TransactionCategory, error)

	Update(ctx context.Context, c *TransactionCategory) error

	// Delete soft-deletes a category. Service layer must refuse when IsSystem=true.
	Delete(ctx context.Context, id uuid.UUID) error
}
