package projecthdl

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ----- Project -----

type CreateProjectRequest struct {
	Name             string     `json:"name"              validate:"required,min=1,max=200"`
	Description      *string    `json:"description"`
	ClientPersonID   *string    `json:"client_person_id"  validate:"omitempty,uuid"`
	ClientNameRaw    *string    `json:"client_name_raw"`
	ClientEmailRaw   *string    `json:"client_email_raw"  validate:"omitempty,email"`
	ScheduledStartAt *time.Time `json:"scheduled_start_at"`
	DeadlineAt       *time.Time `json:"deadline_at"`
}

type UpdateProjectRequest struct {
	Name             string     `json:"name"              validate:"required,min=1,max=200"`
	Description      *string    `json:"description"`
	Status           string     `json:"status"            validate:"required"`
	ClientPersonID   *string    `json:"client_person_id"  validate:"omitempty,uuid"`
	ClientNameRaw    *string    `json:"client_name_raw"`
	ClientEmailRaw   *string    `json:"client_email_raw"  validate:"omitempty,email"`
	ScheduledStartAt *time.Time `json:"scheduled_start_at"`
	DeadlineAt       *time.Time `json:"deadline_at"`
}

type CloseProjectRequest struct {
	Reason   string     `json:"reason"    validate:"required,oneof=won_completed lost_to_competitor client_declined we_withdrew client_abandoned budget_cut cancelled_internal"`
	ClosedAt *time.Time `json:"closed_at" validate:"required"`
}

type UpsertProjectFinancialPlanRequest struct {
	ContractAmount      decimal.Decimal                       `json:"contract_amount"       validate:"required"`
	RetentionAmount     decimal.Decimal                       `json:"retention_amount"`
	PlannedDeliveryDate *time.Time                            `json:"planned_delivery_date"`
	PaymentNote         *string                               `json:"payment_note"`
	Installments        []UpsertProjectInstallmentRequestItem `json:"installments"`
}

type UpsertProjectInstallmentRequestItem struct {
	SortOrder           int             `json:"sort_order"`
	Title               string          `json:"title"                  validate:"required,min=1,max=200"`
	Amount              decimal.Decimal `json:"amount"                 validate:"required"`
	PlannedDeliveryDate *time.Time      `json:"planned_delivery_date"`
	PlannedReceiveDate  *time.Time      `json:"planned_receive_date"`
	Note                *string         `json:"note"`
}

// ----- Member -----

type AddMemberRequest struct {
	PersonID string `json:"person_id" validate:"required,uuid"`
}

type TransferLeadRequest struct {
	MemberID string `json:"member_id" validate:"required,uuid"`
}

// ----- Transaction -----

type CreateTransactionRequest struct {
	CategoryID   string          `json:"category_id"   validate:"required,uuid"`
	Amount       decimal.Decimal `json:"amount"        validate:"required"`
	CurrencyCode string          `json:"currency_code"`
	Note         *string         `json:"note"`
	OccurredAt   time.Time       `json:"occurred_at"   validate:"required"`
}

type UpdateTransactionRequest struct {
	CategoryID   string          `json:"category_id"   validate:"required,uuid"`
	Amount       decimal.Decimal `json:"amount"        validate:"required"`
	CurrencyCode string          `json:"currency_code"`
	Note         *string         `json:"note"`
	OccurredAt   time.Time       `json:"occurred_at"   validate:"required"`
}

// parseUUIDPtr parses a *string UUID into *uuid.UUID. Empty / nil input returns nil.
func parseUUIDPtr(s *string) (*uuid.UUID, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	u, err := uuid.Parse(*s)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
