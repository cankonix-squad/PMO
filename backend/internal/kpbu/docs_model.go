package kpbu

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/shared/types"
	"gorm.io/gorm"
)

// Checklist categories (rows A–E of the document matrix).
const (
	CatPrakarsa        = "PRAKARSA"
	CatPrakualifikasi  = "PRAKUALIFIKASI"
	CatPenawaran       = "PENAWARAN"
	CatPenandatanganan = "PENANDATANGANAN"
	CatPembiayaan      = "PEMBIAYAAN_PELAKSANAAN"
)

// Which initiative a requirement applies to.
const (
	AppliesAll         = "ALL"
	AppliesSolicited   = "SOLICITED"
	AppliesUnsolicited = "UNSOLICITED"
)

// Document review statuses. MISSING is only used in checklist output.
const (
	DocMissing   = "MISSING"
	DocSubmitted = "SUBMITTED"
	DocApproved  = "APPROVED"
	DocRejected  = "REJECTED"
)

// FinancialCloseMonths is the deadline after the KPBU agreement is signed
// (Perpres 38/2015 Ps.42).
const FinancialCloseMonths = 12

// FinancialCloseWarnDays starts the "approaching deadline" warning.
const FinancialCloseWarnDays = 90

// PerformanceGuaranteeMaxPct is the cap of the performance guarantee relative
// to investment value (Perpres 38/2015).
const PerformanceGuaranteeMaxPct = 5.0

var (
	validCategories = map[string]bool{CatPrakarsa: true, CatPrakualifikasi: true, CatPenawaran: true, CatPenandatanganan: true, CatPembiayaan: true}
	validAppliesTo  = map[string]bool{AppliesAll: true, AppliesSolicited: true, AppliesUnsolicited: true}
	validSources    = map[string]bool{SourceVerified: true, SourceSecondary: true, SourceUnverified: true}
	validProcMethod = map[string]bool{"TENDER": true, "PENUNJUKAN_LANGSUNG": true}
	validGuarTypes  = map[string]bool{"BID": true, "PERFORMANCE": true, "GOVERNMENT": true, "OTHER": true}
	validGuarStatus = map[string]bool{"ACTIVE": true, "RELEASED": true, "CLAIMED": true, "EXPIRED": true}
)

// GateError lists mandatory checklist items without an approved document.
type GateError struct {
	Stage   string
	Missing []string
}

func (e *GateError) Error() string {
	return "required documents not approved for stage " + e.Stage + ": " + strings.Join(e.Missing, ", ")
}

// DocumentRequirement is one checklist row (editable master per organization).
type DocumentRequirement struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID      uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	Code                string         `json:"code"`
	Category            string         `json:"category"`
	Name                string         `json:"name"`
	Description         string         `json:"description"`
	AppliesTo           string         `json:"applies_to"`
	RequiredBeforeStage *string        `json:"required_before_stage"`
	IsMandatory         bool           `json:"is_mandatory"`
	LegalBasis          string         `json:"legal_basis"`
	SourceStatus        string         `json:"source_status"`
	SortOrder           int            `json:"sort_order"`
	IsActive            bool           `json:"is_active"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DocumentRequirement) TableName() string { return "kpbu_document_requirements" }

// Document is an uploaded business-entity document awaiting or past review.
type Document struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID       `gorm:"type:uuid;not null" json:"organization_id"`
	KPBUProjectID  uuid.UUID       `gorm:"column:kpbu_project_id;type:uuid;not null" json:"kpbu_project_id"`
	RequirementID  *uuid.UUID      `gorm:"type:uuid" json:"requirement_id"`
	Name           string          `json:"name"`
	FileURL        string          `json:"-"`
	MimeType       string          `json:"mime_type"`
	FileSize       int64           `json:"file_size"`
	DocumentNumber string          `json:"document_number"`
	DocumentDate   *types.FlexTime `json:"document_date"`
	Notes          string          `json:"notes"`
	Status         string          `json:"status"`
	UploadedBy     uuid.UUID       `gorm:"type:uuid" json:"uploaded_by"`
	UploadedAt     time.Time       `json:"uploaded_at"`
	ReviewedBy     *uuid.UUID      `gorm:"type:uuid" json:"reviewed_by"`
	ReviewedAt     *time.Time      `json:"reviewed_at"`
	ReviewNotes    string          `json:"review_notes"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (Document) TableName() string { return "kpbu_documents" }

// Guarantee is a bid, performance, or government guarantee.
type Guarantee struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID  uuid.UUID       `gorm:"type:uuid;not null" json:"organization_id"`
	KPBUProjectID   uuid.UUID       `gorm:"column:kpbu_project_id;type:uuid;not null" json:"kpbu_project_id"`
	Type            string          `json:"type"`
	Issuer          string          `json:"issuer"`
	GuaranteeNumber string          `json:"guarantee_number"`
	Amount          float64         `json:"amount"`
	Currency        string          `json:"currency"`
	IssuedDate      *types.FlexTime `json:"issued_date"`
	ExpiryDate      *types.FlexTime `json:"expiry_date"`
	Status          string          `json:"status"`
	Notes           string          `json:"notes"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"-"`

	// Computed.
	PctOfInvestment *float64 `gorm:"-" json:"pct_of_investment"`
	ExceedsLimit    bool     `gorm:"-" json:"exceeds_limit"`
	DaysToExpiry    *int     `gorm:"-" json:"days_to_expiry"`
}

func (Guarantee) TableName() string { return "kpbu_guarantees" }

// RequirementRequest creates or updates a checklist row.
type RequirementRequest struct {
	Code                string  `json:"code" binding:"required,max=30"`
	Category            string  `json:"category" binding:"required"`
	Name                string  `json:"name" binding:"required,max=300"`
	Description         string  `json:"description"`
	AppliesTo           string  `json:"applies_to"`
	RequiredBeforeStage *string `json:"required_before_stage"`
	IsMandatory         bool    `json:"is_mandatory"`
	LegalBasis          string  `json:"legal_basis" binding:"max=500"`
	SourceStatus        string  `json:"source_status"`
	SortOrder           int     `json:"sort_order"`
	IsActive            *bool   `json:"is_active"`
}

// DocumentUpload carries a multipart upload.
type DocumentUpload struct {
	Filename       string
	Data           []byte
	RequirementID  *uuid.UUID
	Name           string
	DocumentNumber string
	DocumentDate   *types.FlexTime
	Notes          string
}

// ReviewRequest approves or rejects a document. Rejection requires notes.
type ReviewRequest struct {
	Decision string `json:"decision" binding:"required"`
	Notes    string `json:"notes"`
}

// GuaranteeRequest creates or updates a guarantee.
type GuaranteeRequest struct {
	Type            string          `json:"type" binding:"required"`
	Issuer          string          `json:"issuer" binding:"max=300"`
	GuaranteeNumber string          `json:"guarantee_number" binding:"max=200"`
	Amount          float64         `json:"amount" binding:"min=0"`
	Currency        string          `json:"currency"`
	IssuedDate      *types.FlexTime `json:"issued_date"`
	ExpiryDate      *types.FlexTime `json:"expiry_date"`
	Status          string          `json:"status"`
	Notes           string          `json:"notes"`
}

// ChecklistItem is one requirement with its documents for a project.
type ChecklistItem struct {
	Requirement DocumentRequirement `json:"requirement"`
	Status      string              `json:"status"`
	Documents   []Document          `json:"documents"`
}

// ChecklistSummary counts mandatory items.
type ChecklistSummary struct {
	MandatoryTotal    int      `json:"mandatory_total"`
	MandatoryApproved int      `json:"mandatory_approved"`
	PendingReview     int      `json:"pending_review"`
	BlockingNextStage []string `json:"blocking_next_stage"`
	NextStage         string   `json:"next_stage"`
}

// Checklist is the project's document checklist.
type Checklist struct {
	Items      []ChecklistItem  `json:"items"`
	Other      []Document       `json:"other_documents"`
	Summary    ChecklistSummary `json:"summary"`
	Initiative string           `json:"initiative"`
}
