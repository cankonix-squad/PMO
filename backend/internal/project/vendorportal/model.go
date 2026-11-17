// Package vendorportal implements the contractor portal for APBN projects
// (VENDOR-PORTAL-001): contractor accounts linked to a vendor, contractor-side
// progress reports and documents scoped to the vendor's own contracts, and the
// Balai review queue that verifies or rejects those submissions.
package vendorportal

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Submission statuses.
const (
	StatusSubmitted = "SUBMITTED"
	StatusVerified  = "VERIFIED"
	StatusRejected  = "REJECTED"
)

// Document categories a contractor can upload.
var validDocCategories = map[string]bool{
	"PROGRESS_REPORT": true, // laporan kemajuan pekerjaan
	"MC":              true, // monthly certificate / berita acara kemajuan
	"BAST":            true, // berita acara serah terima
	"PHOTO":           true, // dokumentasi foto lapangan
	"SHOP_DRAWING":    true,
	"AS_BUILT":        true,
	"OTHER":           true,
}

var (
	ErrNotVendor      = errors.New("vendorportal: user is not a contractor account")
	ErrNotFound       = errors.New("vendorportal: not found")
	ErrDuplicate      = errors.New("vendorportal: a live report already exists for this period")
	ErrAlreadyDecided = errors.New("vendorportal: submission already reviewed")
	ErrEmailTaken     = errors.New("vendorportal: email already in use")
	ErrNotContractPPK = errors.New("vendorportal: only the contract PPK may review this submission")
)

// ValidationError is a client input error (HTTP 400).
type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }

func invalid(msg string) error { return &ValidationError{Msg: msg} }

// ProgressReport is a monthly physical progress report from a contractor.
type ProgressReport struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID      uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	ProjectID           uuid.UUID      `gorm:"type:uuid;not null" json:"project_id"`
	ContractID          uuid.UUID      `gorm:"type:uuid;not null" json:"contract_id"`
	VendorID            uuid.UUID      `gorm:"type:uuid;not null" json:"vendor_id"`
	PeriodYear          int            `json:"period_year"`
	PeriodMonth         int            `json:"period_month"`
	PhysicalProgressPct float64        `json:"physical_progress_pct"`
	PlannedProgressPct  *float64       `json:"planned_progress_pct"`
	Notes               string         `json:"notes"`
	Status              string         `json:"status"`
	SubmittedBy         uuid.UUID      `gorm:"type:uuid" json:"submitted_by"`
	SubmittedAt         time.Time      `json:"submitted_at"`
	ReviewedBy          *uuid.UUID     `gorm:"type:uuid" json:"reviewed_by"`
	ReviewedAt          *time.Time     `json:"reviewed_at"`
	ReviewNotes         string         `json:"review_notes"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`

	Context `gorm:"-"`
}

func (ProgressReport) TableName() string { return "vendor_progress_reports" }

// Document is a file uploaded by a contractor.
type Document struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	ProjectID      uuid.UUID      `gorm:"type:uuid;not null" json:"project_id"`
	ContractID     uuid.UUID      `gorm:"type:uuid;not null" json:"contract_id"`
	VendorID       uuid.UUID      `gorm:"type:uuid;not null" json:"vendor_id"`
	Name           string         `json:"name"`
	Category       string         `json:"category"`
	PeriodYear     *int           `json:"period_year"`
	PeriodMonth    *int           `json:"period_month"`
	FileURL        string         `gorm:"column:file_url" json:"-"`
	MimeType       string         `json:"mime_type"`
	FileSize       int64          `json:"file_size"`
	Notes          string         `json:"notes"`
	Status         string         `json:"status"`
	SubmittedBy    uuid.UUID      `gorm:"type:uuid" json:"submitted_by"`
	SubmittedAt    time.Time      `json:"submitted_at"`
	ReviewedBy     *uuid.UUID     `gorm:"type:uuid" json:"reviewed_by"`
	ReviewedAt     *time.Time     `json:"reviewed_at"`
	ReviewNotes    string         `json:"review_notes"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Context `gorm:"-"`
}

func (Document) TableName() string { return "vendor_documents" }

// Context carries display names for review screens. Not stored.
type Context struct {
	ContractNumber string     `gorm:"-" json:"contract_number,omitempty"`
	ProjectCode    string     `gorm:"-" json:"project_code,omitempty"`
	ProjectName    string     `gorm:"-" json:"project_name,omitempty"`
	VendorName     string     `gorm:"-" json:"vendor_name,omitempty"`
	PPKUserID      *uuid.UUID `gorm:"-" json:"ppk_user_id,omitempty"`
	PPKName        string     `gorm:"-" json:"ppk_name,omitempty"`
}

// ContractSummary is a contract as seen by its contractor.
type ContractSummary struct {
	ID                  uuid.UUID  `json:"id"`
	ContractNumber      string     `json:"contract_number"`
	Title               string     `json:"title"`
	Status              string     `json:"status"`
	ContractValue       float64    `json:"contract_value"`
	Currency            string     `json:"currency"`
	StartDate           *time.Time `json:"start_date"`
	EndDate             *time.Time `json:"end_date"`
	ProjectID           uuid.UUID  `json:"project_id"`
	ProjectCode         string     `json:"project_code"`
	ProjectName         string     `json:"project_name"`
	LatestVerifiedPct   *float64   `json:"latest_verified_progress_pct"`
	PendingReports      int64      `json:"pending_reports"`
	PendingDocuments    int64      `json:"pending_documents"`
	RejectedSubmissions int64      `json:"rejected_submissions"`
}

// CreateReportRequest is a contractor progress report.
type CreateReportRequest struct {
	PeriodYear          int      `json:"period_year" binding:"required,min=2000,max=2100"`
	PeriodMonth         int      `json:"period_month" binding:"required,min=1,max=12"`
	PhysicalProgressPct *float64 `json:"physical_progress_pct" binding:"required,min=0,max=100"`
	PlannedProgressPct  *float64 `json:"planned_progress_pct" binding:"omitempty,min=0,max=100"`
	Notes               string   `json:"notes"`
}

// ReviewRequest verifies or rejects a submission.
type ReviewRequest struct {
	Decision string `json:"decision" binding:"required"`
	Notes    string `json:"notes"`
}

// SubmissionFilter filters the Balai review queue.
type SubmissionFilter struct {
	Status    string
	ProjectID *uuid.UUID
}

// Submissions is the Balai review queue.
type Submissions struct {
	ProgressReports []ProgressReport `json:"progress_reports"`
	Documents       []Document       `json:"documents"`
}

// VendorAccount is a contractor login as listed for admins.
type VendorAccount struct {
	ID            uuid.UUID  `json:"id"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	Email         string     `json:"email"`
	Phone         string     `json:"phone"`
	IsActive      bool       `json:"is_active"`
	MustChangePwd bool       `json:"must_change_pwd"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

// CreateAccountRequest invites a contractor user.
type CreateAccountRequest struct {
	FirstName string `json:"first_name" binding:"required,max=100"`
	LastName  string `json:"last_name" binding:"max=100"`
	Email     string `json:"email" binding:"required,email,max=200"`
	Phone     string `json:"phone" binding:"max=50"`
}

// CreatedAccount is returned once, with the temporary password.
type CreatedAccount struct {
	Account           VendorAccount `json:"account"`
	TemporaryPassword string        `json:"temporary_password"`
}
