package txcategory

import (
	"time"

	"github.com/google/uuid"
)

// TransactionType classifies a transaction category as income or expense.
// Mirrors the PostgreSQL enum `transaction_type`.
type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

// TransactionCategory is master data for classifying ProjectTransactions.
// System categories (IsSystem=true) are seeded by migration and cannot be deleted —
// only their IsActive flag and display fields may be edited by admins.
// User-created categories are fully editable/deletable.
//
// Code convention: lowercase snake_case, [a-z0-9_]+ (DB CHECK enforced).
type TransactionCategory struct {
	ID        uuid.UUID
	Code      string
	Name      string
	Type      TransactionType
	IconKey   *string
	IsSystem  bool
	IsActive  bool
	SortOrder int

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
