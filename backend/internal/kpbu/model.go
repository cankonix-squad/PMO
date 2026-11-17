// Package kpbu implements KPBU (Kerjasama Pemerintah dengan Badan Usaha /
// public-private partnership) monitoring — phase K1: project registry,
// lifecycle stage FSM, stage targets, and parties.
// See docs/planning/KPBU-GRAND-DESIGN.md.
package kpbu

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/shared/types"
	"gorm.io/gorm"
)

// Lifecycle stages.
const (
	StageIdentified     = "IDENTIFIED"
	StagePlanning       = "PLANNING"
	StagePreparationOBC = "PREPARATION_OBC"
	StagePreparationFBC = "PREPARATION_FBC"
	StageTransaction    = "TRANSACTION"
	StageContractSigned = "CONTRACT_SIGNED"
	StageFinancialClose = "FINANCIAL_CLOSE"
	StageConstruction   = "CONSTRUCTION"
	StageOperation      = "OPERATION"
	StageHandover       = "HANDOVER"
	StageClosed         = "CLOSED"
	StageOnHold         = "ON_HOLD"
	StageCancelled      = "CANCELLED"
	StageTerminated     = "TERMINATED"
)

// MainStages is the ordered happy path. Stage milestones are tracked for these.
var MainStages = []string{
	StageIdentified, StagePlanning, StagePreparationOBC, StagePreparationFBC,
	StageTransaction, StageContractSigned, StageFinancialClose, StageConstruction,
	StageOperation, StageHandover,
}

var (
	validSectors       = map[string]bool{"SDA": true, "JALAN": true, "JALAN_TOL": true, "AIR_MINUM": true, "SANITASI": true, "PERUMAHAN": true, "LAINNYA": true}
	validInitiatives   = map[string]bool{"SOLICITED": true, "UNSOLICITED": true}
	validReturnSchemes = map[string]bool{"USER_CHARGE": true, "AVAILABILITY_PAYMENT": true, "HYBRID": true}
	validPartyRoles    = map[string]bool{"PJPK": true, "BUP": true, "SPONSOR": true, "LENDER": true, "GUARANTOR": true, "CONSULTANT": true, "INDEPENDENT_ENGINEER": true, "OTHER": true}
)

var (
	ErrNotFound          = errors.New("kpbu: not found")
	ErrCodeTaken         = errors.New("kpbu: code already in use")
	ErrInvalidTransition = errors.New("kpbu: invalid stage transition")
	ErrInvalidStage      = errors.New("kpbu: invalid stage")
)

// ValidationError is a client-side input error (HTTP 400).
type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }

func invalid(msg string) error { return &ValidationError{Msg: msg} }

// Project is a KPBU project.
type Project struct {
	ID                   uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID       uuid.UUID       `gorm:"type:uuid;not null" json:"organization_id"`
	ProjectID            *uuid.UUID      `gorm:"type:uuid" json:"project_id"`
	Code                 string          `json:"code"`
	Name                 string          `json:"name"`
	Description          string          `json:"description"`
	Sector               string          `json:"sector"`
	Stage                string          `json:"stage"`
	PreviousStage        *string         `json:"previous_stage"`
	Initiative           string          `json:"initiative"`
	ReturnScheme         string          `json:"return_scheme"`
	PJPKName             string          `gorm:"column:pjpk_name" json:"pjpk_name"`
	PJPKUnit             string          `gorm:"column:pjpk_unit" json:"pjpk_unit"`
	InvestmentValue      float64         `json:"investment_value"`
	Currency             string          `json:"currency"`
	ConcessionYears      *int            `json:"concession_years"`
	ConcessionStart      *types.FlexTime `json:"concession_start"`
	ConcessionEnd        *types.FlexTime `json:"concession_end"`
	FinancialCloseTarget *types.FlexTime `json:"financial_close_target"`
	FinancialCloseActual *types.FlexTime `json:"financial_close_actual"`
	CODTarget            *types.FlexTime `gorm:"column:cod_target" json:"cod_target"`
	CODActual            *types.FlexTime `gorm:"column:cod_actual" json:"cod_actual"`
	ProcurementMethod    *string         `json:"procurement_method"`
	ProcurementReason    string          `json:"procurement_reason"`
	AgreementSignedDate  *types.FlexTime `json:"agreement_signed_date"`
	OrgUnitID            *uuid.UUID      `gorm:"type:uuid" json:"org_unit_id"`
	RegionID             *uuid.UUID      `gorm:"type:uuid" json:"region_id"`
	RiverBasinID         *uuid.UUID      `gorm:"type:uuid" json:"river_basin_id"`
	Province             string          `json:"province"`
	Latitude             *float64        `json:"latitude"`
	Longitude            *float64        `json:"longitude"`
	CreatedBy            *uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
	DeletedAt            gorm.DeletedAt  `gorm:"index" json:"-"`

	// Computed, not stored.
	OrgUnitName     string          `gorm:"-" json:"org_unit_name,omitempty"`
	StageDelayDays  int             `gorm:"-" json:"stage_delay_days"`
	StageIsDelayed  bool            `gorm:"-" json:"stage_is_delayed"`
	StagePlannedEnd *types.FlexTime `gorm:"-" json:"stage_planned_end"`

	// Financial close deadline = agreement signed + 12 months (Perpres 38/2015 Ps.42).
	FinancialCloseDeadline *types.FlexTime `gorm:"-" json:"financial_close_deadline"`
	FinancialCloseDaysLeft *int            `gorm:"-" json:"financial_close_days_left"`
	// NONE | ON_TRACK | WARNING | OVERDUE | ACHIEVED
	FinancialCloseStatus string `gorm:"-" json:"financial_close_status"`
}

func (Project) TableName() string { return "kpbu_projects" }

// StageMilestone holds planned/actual dates of one lifecycle stage.
type StageMilestone struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID       `gorm:"type:uuid;not null" json:"organization_id"`
	KPBUProjectID  uuid.UUID       `gorm:"column:kpbu_project_id;type:uuid;not null" json:"kpbu_project_id"`
	Stage          string          `json:"stage"`
	PlannedStart   *types.FlexTime `json:"planned_start"`
	PlannedEnd     *types.FlexTime `json:"planned_end"`
	ActualStart    *types.FlexTime `json:"actual_start"`
	ActualEnd      *types.FlexTime `json:"actual_end"`
	Notes          string          `json:"notes"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`

	// Computed.
	IsDelayed bool `gorm:"-" json:"is_delayed"`
	DelayDays int  `gorm:"-" json:"delay_days"`
}

func (StageMilestone) TableName() string { return "kpbu_stage_milestones" }

// Party is a stakeholder in a KPBU project.
type Party struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	KPBUProjectID  uuid.UUID      `gorm:"column:kpbu_project_id;type:uuid;not null" json:"kpbu_project_id"`
	Role           string         `json:"role"`
	Name           string         `json:"name"`
	VendorID       *uuid.UUID     `gorm:"type:uuid" json:"vendor_id"`
	SharePct       *float64       `json:"share_pct"`
	ContactPerson  string         `json:"contact_person"`
	Email          string         `json:"email"`
	Phone          string         `json:"phone"`
	Notes          string         `json:"notes"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Party) TableName() string { return "kpbu_parties" }

// ProjectRequest is the body for create (POST) and full update (PUT).
// Stage is not editable here; use the transition endpoint.
type ProjectRequest struct {
	Code                 string          `json:"code" binding:"required,max=50"`
	Name                 string          `json:"name" binding:"required,max=300"`
	Description          string          `json:"description"`
	Sector               string          `json:"sector" binding:"required"`
	Initiative           string          `json:"initiative"`
	ReturnScheme         string          `json:"return_scheme"`
	PJPKName             string          `json:"pjpk_name" binding:"max=200"`
	PJPKUnit             string          `json:"pjpk_unit" binding:"max=200"`
	InvestmentValue      float64         `json:"investment_value" binding:"min=0"`
	Currency             string          `json:"currency"`
	ConcessionYears      *int            `json:"concession_years" binding:"omitempty,min=1,max=99"`
	ConcessionStart      *types.FlexTime `json:"concession_start"`
	ConcessionEnd        *types.FlexTime `json:"concession_end"`
	FinancialCloseTarget *types.FlexTime `json:"financial_close_target"`
	CODTarget            *types.FlexTime `json:"cod_target"`
	ProcurementMethod    *string         `json:"procurement_method"`
	ProcurementReason    string          `json:"procurement_reason"`
	AgreementSignedDate  *types.FlexTime `json:"agreement_signed_date"`
	ProjectID            *uuid.UUID      `json:"project_id"`
	OrgUnitID            *uuid.UUID      `json:"org_unit_id"`
	Province             string          `json:"province" binding:"max=100"`
	Latitude             *float64        `json:"latitude" binding:"omitempty,min=-90,max=90"`
	Longitude            *float64        `json:"longitude" binding:"omitempty,min=-180,max=180"`
}

// TransitionRequest moves a project to another stage.
type TransitionRequest struct {
	ToStage string          `json:"to_stage" binding:"required"`
	Date    *types.FlexTime `json:"date"`
	Notes   string          `json:"notes"`
	// OverrideDocumentCheck skips the document gate; OverrideReason is required
	// and recorded in the audit log.
	OverrideDocumentCheck bool   `json:"override_document_check"`
	OverrideReason        string `json:"override_reason"`
}

// StageMilestoneRequest upserts planned/actual dates for one stage.
type StageMilestoneRequest struct {
	Stage        string          `json:"stage" binding:"required"`
	PlannedStart *types.FlexTime `json:"planned_start"`
	PlannedEnd   *types.FlexTime `json:"planned_end"`
	ActualStart  *types.FlexTime `json:"actual_start"`
	ActualEnd    *types.FlexTime `json:"actual_end"`
	Notes        string          `json:"notes"`
}

// PartyRequest is the body for party create/update.
type PartyRequest struct {
	Role          string     `json:"role" binding:"required"`
	Name          string     `json:"name" binding:"required,max=300"`
	VendorID      *uuid.UUID `json:"vendor_id"`
	SharePct      *float64   `json:"share_pct" binding:"omitempty,min=0,max=100"`
	ContactPerson string     `json:"contact_person" binding:"max=200"`
	Email         string     `json:"email" binding:"omitempty,email,max=200"`
	Phone         string     `json:"phone" binding:"max=50"`
	Notes         string     `json:"notes"`
}

// ListFilter filters the project list.
type ListFilter struct {
	Stage  string
	Sector string
	Search string
}

// Detail is the full project view.
type Detail struct {
	Project         *Project         `json:"project"`
	StageMilestones []StageMilestone `json:"stage_milestones"`
	Parties         []Party          `json:"parties"`
	AllowedNext     []string         `json:"allowed_next_stages"`
}
