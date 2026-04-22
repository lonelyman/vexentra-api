package projecthdl

import (
	"time"

	"vexentra-api/internal/modules/project"
	"vexentra-api/internal/modules/txcategory"

	"github.com/shopspring/decimal"
)

type ProjectResponse struct {
	ID          string  `json:"id"`
	ProjectCode string  `json:"project_code"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`

	Status        string  `json:"status"`
	ClosureReason *string `json:"closure_reason,omitempty"`

	ClientPersonID *string `json:"client_person_id,omitempty"`
	ClientNameRaw  *string `json:"client_name_raw,omitempty"`
	ClientEmailRaw *string `json:"client_email_raw,omitempty"`

	ScheduledStartAt *time.Time `json:"scheduled_start_at,omitempty"`
	DeadlineAt       *time.Time `json:"deadline_at,omitempty"`
	ActivatedAt      *time.Time `json:"activated_at,omitempty"`
	ClosedAt         *time.Time `json:"closed_at,omitempty"`

	CreatedByUserID string    `json:"created_by_user_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ProjectStatusResponse struct {
	Status         string `json:"status"`
	LabelTH        string `json:"label_th"`
	Phase          string `json:"phase"`
	SortOrder      int    `json:"sort_order"`
	IsTerminal     bool   `json:"is_terminal"`
	RequiresClient bool   `json:"requires_client"`
	IsActive       bool   `json:"is_active"`
}

func NewProjectStatusResponse(s project.ProjectStatusMeta) ProjectStatusResponse {
	return ProjectStatusResponse{
		Status:         string(s.Status),
		LabelTH:        s.LabelTH,
		Phase:          string(s.Phase),
		SortOrder:      s.SortOrder,
		IsTerminal:     s.IsTerminal,
		RequiresClient: s.RequiresClient,
		IsActive:       s.IsActive,
	}
}

func NewProjectResponse(p *project.Project) ProjectResponse {
	r := ProjectResponse{
		ID:               p.ID.String(),
		ProjectCode:      p.ProjectCode,
		Name:             p.Name,
		Description:      p.Description,
		Status:           string(p.Status),
		ClientNameRaw:    p.ClientNameRaw,
		ClientEmailRaw:   p.ClientEmailRaw,
		ScheduledStartAt: p.ScheduledStartAt,
		DeadlineAt:       p.DeadlineAt,
		ActivatedAt:      p.ActivatedAt,
		ClosedAt:         p.ClosedAt,
		CreatedByUserID:  p.CreatedByUserID.String(),
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
	if p.ClosureReason != nil {
		s := string(*p.ClosureReason)
		r.ClosureReason = &s
	}
	if p.ClientPersonID != nil {
		s := p.ClientPersonID.String()
		r.ClientPersonID = &s
	}
	return r
}

type MemberResponse struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	PersonID      string    `json:"person_id"`
	IsLead        bool      `json:"is_lead"`
	AddedByUserID string    `json:"added_by_user_id"`
	JoinedAt      time.Time `json:"joined_at"`
}

func NewMemberResponse(m *project.ProjectMember) MemberResponse {
	return MemberResponse{
		ID:            m.ID.String(),
		ProjectID:     m.ProjectID.String(),
		PersonID:      m.PersonID.String(),
		IsLead:        m.IsLead,
		AddedByUserID: m.AddedByUserID.String(),
		JoinedAt:      m.CreatedAt,
	}
}

type TransactionResponse struct {
	ID              string          `json:"id"`
	ProjectID       string          `json:"project_id"`
	CategoryID      string          `json:"category_id"`
	Amount          decimal.Decimal `json:"amount"`
	CurrencyCode    string          `json:"currency_code"`
	Note            *string         `json:"note,omitempty"`
	OccurredAt      time.Time       `json:"occurred_at"`
	CreatedByUserID string          `json:"created_by_user_id"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func NewTransactionResponse(t *project.ProjectTransaction) TransactionResponse {
	return TransactionResponse{
		ID:              t.ID.String(),
		ProjectID:       t.ProjectID.String(),
		CategoryID:      t.CategoryID.String(),
		Amount:          t.Amount,
		CurrencyCode:    t.CurrencyCode,
		Note:            t.Note,
		OccurredAt:      t.OccurredAt,
		CreatedByUserID: t.CreatedByUserID.String(),
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}

type ProjectTotalsResponse struct {
	Income  decimal.Decimal `json:"income"`
	Expense decimal.Decimal `json:"expense"`
	Net     decimal.Decimal `json:"net"`
}

func NewProjectTotalsResponse(t *project.ProjectTotals) ProjectTotalsResponse {
	return ProjectTotalsResponse{
		Income:  t.Income,
		Expense: t.Expense,
		Net:     t.Net,
	}
}

type ProjectFinancialPlanResponse struct {
	ProjectID            string                              `json:"project_id"`
	ContractAmount       decimal.Decimal                     `json:"contract_amount"`
	RetentionAmount      decimal.Decimal                     `json:"retention_amount"`
	PlannedDeliveryDate  *time.Time                          `json:"planned_delivery_date,omitempty"`
	PaymentNote          *string                             `json:"payment_note,omitempty"`
	Installments         []ProjectPaymentInstallmentResponse `json:"installments"`
	InstallmentsTotal    decimal.Decimal                     `json:"installments_total"`
	NetReceivable        decimal.Decimal                     `json:"net_receivable"`
	UnallocatedRemaining decimal.Decimal                     `json:"unallocated_remaining"`
	CreatedAt            time.Time                           `json:"created_at"`
	UpdatedAt            time.Time                           `json:"updated_at"`
}

type ProjectPaymentInstallmentResponse struct {
	ID                  string          `json:"id"`
	ProjectID           string          `json:"project_id"`
	SortOrder           int             `json:"sort_order"`
	Title               string          `json:"title"`
	Amount              decimal.Decimal `json:"amount"`
	PlannedDeliveryDate *time.Time      `json:"planned_delivery_date,omitempty"`
	PlannedReceiveDate  *time.Time      `json:"planned_receive_date,omitempty"`
	Note                *string         `json:"note,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

func NewProjectFinancialPlanResponse(p *project.ProjectFinancialPlan) ProjectFinancialPlanResponse {
	items := make([]ProjectPaymentInstallmentResponse, len(p.Installments))
	for i := range p.Installments {
		items[i] = NewProjectPaymentInstallmentResponse(p.Installments[i])
	}
	return ProjectFinancialPlanResponse{
		ProjectID:            p.ProjectID.String(),
		ContractAmount:       p.ContractAmount,
		RetentionAmount:      p.RetentionAmount,
		PlannedDeliveryDate:  p.PlannedDeliveryDate,
		PaymentNote:          p.PaymentNote,
		Installments:         items,
		InstallmentsTotal:    p.InstallmentsTotal,
		NetReceivable:        p.NetReceivable,
		UnallocatedRemaining: p.UnallocatedRemaining,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
	}
}

func NewProjectPaymentInstallmentResponse(i *project.ProjectPaymentInstallment) ProjectPaymentInstallmentResponse {
	return ProjectPaymentInstallmentResponse{
		ID:                  i.ID.String(),
		ProjectID:           i.ProjectID.String(),
		SortOrder:           i.SortOrder,
		Title:               i.Title,
		Amount:              i.Amount,
		PlannedDeliveryDate: i.PlannedDeliveryDate,
		PlannedReceiveDate:  i.PlannedReceiveDate,
		Note:                i.Note,
		CreatedAt:           i.CreatedAt,
		UpdatedAt:           i.UpdatedAt,
	}
}

// ----- TransactionCategory (lives in txcategory module but exposed via /tx-categories) -----

type TransactionCategoryResponse struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	IconKey   *string   `json:"icon_key,omitempty"`
	IsSystem  bool      `json:"is_system"`
	IsActive  bool      `json:"is_active"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewTransactionCategoryResponse(c *txcategory.TransactionCategory) TransactionCategoryResponse {
	return TransactionCategoryResponse{
		ID:        c.ID.String(),
		Code:      c.Code,
		Name:      c.Name,
		Type:      string(c.Type),
		IconKey:   c.IconKey,
		IsSystem:  c.IsSystem,
		IsActive:  c.IsActive,
		SortOrder: c.SortOrder,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
