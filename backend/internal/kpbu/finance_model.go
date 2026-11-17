package kpbu

import (
	"time"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/shared/types"
	"gorm.io/gorm"
)

// K2 — KPBU finance: financing structure, government support, and
// availability payments. Values are operational input, not audited figures.

var (
	validFinancingSources = map[string]bool{
		"EQUITY": true, "SENIOR_DEBT": true, "SUBORDINATED_DEBT": true,
		"BOND": true, "SHAREHOLDER_LOAN": true, "OTHER": true,
	}
	debtSources           = map[string]bool{"SENIOR_DEBT": true, "SUBORDINATED_DEBT": true, "BOND": true, "SHAREHOLDER_LOAN": true}
	validFinancingStatus  = map[string]bool{"PLANNED": true, "COMMITTED": true, "DISBURSED": true, "CANCELLED": true}
	validSupportTypes     = map[string]bool{"VGF": true, "FISCAL_INCENTIVE": true, "LAND_ACQUISITION": true, "PERMIT": true, "GUARANTEE": true, "OTHER": true}
	validSupportStatus    = map[string]bool{"PROPOSED": true, "IN_PROCESS": true, "APPROVED": true, "REALIZED": true, "REJECTED": true}
	ErrDuplicatePeriod    = invalid("an availability payment for this period already exists")
	supportCountsAsActive = map[string]bool{"APPROVED": true, "REALIZED": true}
)

// Financing is one source of project funding (equity, loan, bond, ...).
type Financing struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID  uuid.UUID       `gorm:"type:uuid;not null" json:"organization_id"`
	KPBUProjectID   uuid.UUID       `gorm:"column:kpbu_project_id;type:uuid;not null" json:"kpbu_project_id"`
	Source          string          `json:"source"`
	Provider        string          `json:"provider"`
	AgreementNumber string          `json:"agreement_number"`
	AgreementDate   *types.FlexTime `json:"agreement_date"`
	Amount          float64         `json:"amount"`
	DisbursedAmount float64         `json:"disbursed_amount"`
	Currency        string          `json:"currency"`
	Status          string          `json:"status"`
	Notes           string          `json:"notes"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (Financing) TableName() string { return "kpbu_financings" }

type FinancingRequest struct {
	Source          string          `json:"source" binding:"required"`
	Provider        string          `json:"provider" binding:"max=300"`
	AgreementNumber string          `json:"agreement_number" binding:"max=200"`
	AgreementDate   *types.FlexTime `json:"agreement_date"`
	Amount          float64         `json:"amount" binding:"min=0"`
	DisbursedAmount float64         `json:"disbursed_amount" binding:"min=0"`
	Currency        string          `json:"currency"`
	Status          string          `json:"status"`
	Notes           string          `json:"notes"`
}

// GovernmentSupport is fiscal or non-fiscal support from the government
// (e.g. viability gap fund, land acquisition, permits, infrastructure guarantee).
type GovernmentSupport struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID       `gorm:"type:uuid;not null" json:"organization_id"`
	KPBUProjectID  uuid.UUID       `gorm:"column:kpbu_project_id;type:uuid;not null" json:"kpbu_project_id"`
	Type           string          `json:"type"`
	Provider       string          `json:"provider"`
	DecisionNumber string          `json:"decision_number"`
	DecisionDate   *types.FlexTime `json:"decision_date"`
	Amount         float64         `json:"amount"`
	Currency       string          `json:"currency"`
	Status         string          `json:"status"`
	Notes          string          `json:"notes"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (GovernmentSupport) TableName() string { return "kpbu_government_supports" }

type GovernmentSupportRequest struct {
	Type           string          `json:"type" binding:"required"`
	Provider       string          `json:"provider" binding:"max=300"`
	DecisionNumber string          `json:"decision_number" binding:"max=200"`
	DecisionDate   *types.FlexTime `json:"decision_date"`
	Amount         float64         `json:"amount" binding:"min=0"`
	Currency       string          `json:"currency"`
	Status         string          `json:"status"`
	Notes          string          `json:"notes"`
}

// AvailabilityPayment is one period of the PJPK's availability payment to the BUP.
type AvailabilityPayment struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID  uuid.UUID       `gorm:"type:uuid;not null" json:"organization_id"`
	KPBUProjectID   uuid.UUID       `gorm:"column:kpbu_project_id;type:uuid;not null" json:"kpbu_project_id"`
	PeriodYear      int             `json:"period_year"`
	PeriodMonth     int             `json:"period_month"`
	DueDate         *types.FlexTime `json:"due_date"`
	ScheduledAmount float64         `json:"scheduled_amount"`
	DeductionAmount float64         `json:"deduction_amount"`
	DeductionReason string          `json:"deduction_reason"`
	PaidAmount      float64         `json:"paid_amount"`
	PaidDate        *types.FlexTime `json:"paid_date"`
	InvoiceNumber   string          `json:"invoice_number"`
	Currency        string          `json:"currency"`
	Notes           string          `json:"notes"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"-"`

	// Computed.
	NetAmount   float64 `gorm:"-" json:"net_amount"`
	Outstanding float64 `gorm:"-" json:"outstanding"`
	Status      string  `gorm:"-" json:"status"` // SCHEDULED | PARTIAL | PAID | OVERDUE
	DaysOverdue int     `gorm:"-" json:"days_overdue"`
}

func (AvailabilityPayment) TableName() string { return "kpbu_availability_payments" }

type AvailabilityPaymentRequest struct {
	PeriodYear      int             `json:"period_year" binding:"required,min=2000,max=2100"`
	PeriodMonth     int             `json:"period_month" binding:"required,min=1,max=12"`
	DueDate         *types.FlexTime `json:"due_date"`
	ScheduledAmount float64         `json:"scheduled_amount" binding:"min=0"`
	DeductionAmount float64         `json:"deduction_amount" binding:"min=0"`
	DeductionReason string          `json:"deduction_reason"`
	PaidAmount      float64         `json:"paid_amount" binding:"min=0"`
	PaidDate        *types.FlexTime `json:"paid_date"`
	InvoiceNumber   string          `json:"invoice_number" binding:"max=200"`
	Currency        string          `json:"currency"`
	Notes           string          `json:"notes"`
}

// FinanceSummary aggregates K2 data for one project. Amounts in other
// currencies than the project currency are excluded and counted.
type FinanceSummary struct {
	Currency           string             `json:"currency"`
	InvestmentValue    float64            `json:"investment_value"`
	FinancingTotal     float64            `json:"financing_total"`
	FinancingBySource  map[string]float64 `json:"financing_by_source"`
	EquityTotal        float64            `json:"equity_total"`
	DebtTotal          float64            `json:"debt_total"`
	EquityPct          *float64           `json:"equity_pct"`
	DebtToEquity       *float64           `json:"debt_to_equity"`
	DisbursedTotal     float64            `json:"disbursed_total"`
	FundingCoveragePct *float64           `json:"funding_coverage_pct"` // financing_total / investment_value
	SupportApproved    float64            `json:"support_approved"`     // APPROVED + REALIZED
	SupportProposed    float64            `json:"support_proposed"`     // PROPOSED + IN_PROCESS
	SupportCount       int                `json:"support_count"`
	APScheduled        float64            `json:"ap_scheduled"`
	APDeductions       float64            `json:"ap_deductions"`
	APNet              float64            `json:"ap_net"`
	APPaid             float64            `json:"ap_paid"`
	APOutstandingDue   float64            `json:"ap_outstanding_due"` // outstanding on periods already due
	APOverdueCount     int                `json:"ap_overdue_count"`
	SkippedOtherCcy    int                `json:"skipped_other_currency"`
}
