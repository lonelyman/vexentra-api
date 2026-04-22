package project

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TransactionExportRow is a flattened view used only for CSV export —
// it joins category name and type so the handler can stream rows without extra lookups.
type TransactionExportRow struct {
	OccurredAt   time.Time
	CategoryName string
	CategoryType string // "income" | "expense"
	Amount       decimal.Decimal
	CurrencyCode string
	Note         *string
}

// TransactionFilter narrows ListByProject queries. All fields optional.
type TransactionFilter struct {
	CategoryIDs []uuid.UUID // OR-match; empty = all categories
	OccurredGTE *time.Time  // inclusive lower bound
	OccurredLT  *time.Time  // exclusive upper bound
}

// ProjectTotals is the running sum of a project's financials, computed at the DB layer.
// Income/Expense are always non-negative; Net = Income - Expense.
type ProjectTotals struct {
	Income  decimal.Decimal
	Expense decimal.Decimal
	Net     decimal.Decimal
}

type ProjectTransactionRepository interface {
	Create(ctx context.Context, t *ProjectTransaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*ProjectTransaction, error)

	// ListByProject returns transactions of a project ordered by occurred_at DESC.
	ListByProject(ctx context.Context, projectID uuid.UUID, f TransactionFilter, pg Pagination) (items []*ProjectTransaction, total int64, err error)

	// SumByProject aggregates income and expense totals for a project, excluding soft-deleted rows.
	// Intended for dashboards — avoids pulling the full transaction list into service memory.
	SumByProject(ctx context.Context, projectID uuid.UUID) (*ProjectTotals, error)

	// ListForExport returns all transactions for a project joined with category data,
	// ordered by occurred_at ASC. Intended for CSV/report generation only — no pagination.
	ListForExport(ctx context.Context, projectID uuid.UUID) ([]*TransactionExportRow, error)

	Update(ctx context.Context, t *ProjectTransaction) error
	Delete(ctx context.Context, id uuid.UUID) error // soft delete
}
