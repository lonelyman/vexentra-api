package project

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProjectStatus is the lifecycle state of a Project.
// Mirrors the PostgreSQL enum `project_status` — keep values in sync with migration 20260422000003.
type ProjectStatus string

const (
	ProjectStatusDraft   ProjectStatus = "draft"
	ProjectStatusPlanned ProjectStatus = "planned"
	ProjectStatusBidding ProjectStatus = "bidding"
	ProjectStatusActive  ProjectStatus = "active"
	ProjectStatusOnHold  ProjectStatus = "on_hold"
	ProjectStatusClosed  ProjectStatus = "closed"
)

type ProjectStatusPhase string

const (
	ProjectStatusPhaseBacklog  ProjectStatusPhase = "backlog"
	ProjectStatusPhasePreSales ProjectStatusPhase = "pre_sales"
	ProjectStatusPhaseDelivery ProjectStatusPhase = "delivery"
	ProjectStatusPhaseTerminal ProjectStatusPhase = "terminal"
)

type ProjectStatusMeta struct {
	Status         ProjectStatus
	LabelTH        string
	Phase          ProjectStatusPhase
	SortOrder      int
	IsTerminal     bool
	RequiresClient bool
	IsActive       bool
}

// ProjectClosureReason records why a project reached the terminal `closed` state.
// Required when Status == ProjectStatusClosed; must be nil otherwise (DB CHECK enforces bijection).
type ProjectClosureReason string

const (
	ClosureReasonWonCompleted      ProjectClosureReason = "won_completed"
	ClosureReasonLostToCompetitor  ProjectClosureReason = "lost_to_competitor"
	ClosureReasonClientDeclined    ProjectClosureReason = "client_declined"
	ClosureReasonWeWithdrew        ProjectClosureReason = "we_withdrew"
	ClosureReasonClientAbandoned   ProjectClosureReason = "client_abandoned"
	ClosureReasonBudgetCut         ProjectClosureReason = "budget_cut"
	ClosureReasonCancelledInternal ProjectClosureReason = "cancelled_internal"
)

// Project is the central aggregate of the Project Management module.
// Client handling: once promoted from raw snapshot to a linked Person,
// ClientPersonID is set and ClientNameRaw/ClientEmailRaw remain as the historical record.
type Project struct {
	ID          uuid.UUID
	ProjectCode string // format PREFIX-YYYY-NNNN (DB CHECK: ^[A-Z]+-[0-9]{4}-[0-9]{4}$)

	Name        string
	Description *string

	Status        ProjectStatus
	ClosureReason *ProjectClosureReason

	ClientPersonID *uuid.UUID
	ClientNameRaw  *string
	ClientEmailRaw *string

	ScheduledStartAt *time.Time
	DeadlineAt       *time.Time
	ActivatedAt      *time.Time
	ClosedAt         *time.Time

	CreatedByUserID uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type ProjectPaymentInstallment struct {
	ID                  uuid.UUID
	ProjectID           uuid.UUID
	SortOrder           int
	Title               string
	Amount              decimal.Decimal
	PlannedDeliveryDate *time.Time
	PlannedReceiveDate  *time.Time
	Note                *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type ProjectFinancialPlan struct {
	ProjectID            uuid.UUID
	ContractAmount       decimal.Decimal
	RetentionAmount      decimal.Decimal
	PlannedDeliveryDate  *time.Time
	PaymentNote          *string
	Installments         []*ProjectPaymentInstallment
	InstallmentsTotal    decimal.Decimal
	NetReceivable        decimal.Decimal
	UnallocatedRemaining decimal.Decimal
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
