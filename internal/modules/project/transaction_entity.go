package project

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProjectTransaction records a single income/expense line item against a Project.
// The income/expense classification is not stored here — it is derived from the
// linked TransactionCategory.Type (see migration 20260422000003).
//
// Amount is modeled with shopspring/decimal to match PostgreSQL NUMERIC(15,2) without
// precision loss. DB CHECK guarantees amount >= 0 and currency_code matches ^[A-Z]{3}$.
type ProjectTransaction struct {
	ID         uuid.UUID
	ProjectID  uuid.UUID
	CategoryID uuid.UUID

	Amount       decimal.Decimal
	CurrencyCode string // ISO 4217, 3 uppercase letters (default 'THB' in DB)
	Note         *string
	OccurredAt   time.Time

	CreatedByUserID uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}
