package governance

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

// SourceType enumerates where the submitted data came from.
const (
	SourceManual     = "MANUAL"
	SourceCSVImport  = "CSV_IMPORT"
	SourcePrimavera  = "PRIMAVERA"
	SourceGovernment = "GOVERNMENT"
	SourceBIM        = "BIM"
	SourceAPI        = "API"
)

// DatasetType enumerates the business data domains a submission can cover.
const (
	DatasetProjectProgress = "PROJECT_PROGRESS"
	DatasetBudget          = "BUDGET"
	DatasetRisk            = "RISK"
	DatasetIssue           = "ISSUE"
	DatasetBenefit         = "BENEFIT"
	DatasetLocation        = "LOCATION"
	DatasetContract        = "CONTRACT"
	DatasetDocument        = "DOCUMENT"
	DatasetOther           = "OTHER"
)

// SubmissionStatus is the lifecycle state of a data submission.
const (
	StatusDraft     = "DRAFT"
	StatusSubmitted = "SUBMITTED"
	StatusInReview  = "IN_REVIEW"
	StatusApproved  = "APPROVED"
	StatusRejected  = "REJECTED"
	StatusLocked    = "LOCKED"
	StatusCancelled = "CANCELLED"
)

// ItemAction is the planned write action for a submission item.
const (
	ItemActionCreate       = "CREATE"
	ItemActionUpdate       = "UPDATE"
	ItemActionDelete       = "DELETE"
	ItemActionUpsert       = "UPSERT"
	ItemActionValidateOnly = "VALIDATE_ONLY"
)

// ItemValidationStatus is the per-item validation result.
const (
	ItemValidationPending = "PENDING"
	ItemValidationValid   = "VALID"
	ItemValidationInvalid = "INVALID"
)

// LockStatus is the lifecycle state of a data lock period.
const (
	LockOpen   = "OPEN"
	LockLocked = "LOCKED"
)

// AllowedSubmissionStatuses is the full set of valid submission statuses.
var AllowedSubmissionStatuses = map[string]bool{
	StatusDraft: true, StatusSubmitted: true, StatusInReview: true,
	StatusApproved: true, StatusRejected: true, StatusLocked: true,
	StatusCancelled: true,
}

// AllowedDatasetTypes is the set of valid dataset types.
var AllowedDatasetTypes = map[string]bool{
	DatasetProjectProgress: true, DatasetBudget: true, DatasetRisk: true,
	DatasetIssue: true, DatasetBenefit: true, DatasetLocation: true,
	DatasetContract: true, DatasetDocument: true, DatasetOther: true,
}

// AllowedSourceTypes is the set of valid source types.
var AllowedSourceTypes = map[string]bool{
	SourceManual: true, SourceCSVImport: true, SourcePrimavera: true,
	SourceGovernment: true, SourceBIM: true, SourceAPI: true,
}

// ---------------------------------------------------------------------------
// SubmissionItem — a single entity inside a submission
// ---------------------------------------------------------------------------

// SubmissionItem represents one entity (project, budget, risk, ...) being
// submitted for official validation.
type SubmissionItem struct {
	ID               uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SubmissionID     uuid.UUID       `gorm:"type:uuid;not null;index" json:"submission_id"`
	EntityType       string          `gorm:"size:100;not null" json:"entity_type"`
	EntityID         *uuid.UUID      `gorm:"type:uuid" json:"entity_id,omitempty"`
	Action           string          `gorm:"size:20;not null;default:'CREATE'" json:"action"`
	PayloadBefore    json.RawMessage `gorm:"type:jsonb" json:"payload_before,omitempty"`
	PayloadAfter     json.RawMessage `gorm:"type:jsonb;not null" json:"payload_after"`
	ValidationStatus string          `gorm:"size:20;not null;default:'PENDING'" json:"validation_status"`
	ValidationErrors json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"validation_errors"`
	CreatedAt        time.Time       `json:"created_at"`
}

func (SubmissionItem) TableName() string { return "data_submission_items" }

// ---------------------------------------------------------------------------
// LockPeriod — period-level protection of approved data
// ---------------------------------------------------------------------------

// LockPeriod represents a closed accounting/reporting period for a dataset.
type LockPeriod struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index" json:"organization_id"`
	DatasetType    string         `gorm:"size:50;not null" json:"dataset_type"`
	PeriodYear     int            `gorm:"not null" json:"period_year"`
	PeriodMonth    *int           `json:"period_month,omitempty"` // NULL = full-year lock
	Status         string         `gorm:"size:20;not null;default:'OPEN'" json:"status"`
	LockedBy       *uuid.UUID     `gorm:"type:uuid" json:"locked_by,omitempty"`
	LockedAt       *time.Time     `json:"locked_at,omitempty"`
	LockReason     string         `gorm:"type:text" json:"lock_reason,omitempty"`
	CreatedBy      uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (LockPeriod) TableName() string { return "data_lock_periods" }

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

// CreateSubmissionRequest is the payload for POST /governance/submissions.
type CreateSubmissionRequest struct {
	DatasetType      string              `json:"dataset_type" binding:"required"`
	SourceType       string              `json:"source_type" binding:"required"`
	SourceEntityType string              `json:"source_entity_type"`
	SourceEntityID   string              `json:"source_entity_id"`
	PeriodYear       int                 `json:"period_year" binding:"required"`
	PeriodMonth      *int                `json:"period_month"`
	Items            []CreateItemRequest `json:"items"`
	SourceReference  string              `json:"source_reference"`
}

// CreateItemRequest is a single line item in a create-submission payload.
type CreateItemRequest struct {
	EntityType    string                 `json:"entity_type" binding:"required"`
	EntityID      string                 `json:"entity_id"`
	Action        string                 `json:"action" binding:"required"`
	PayloadAfter  map[string]interface{} `json:"payload_after"`
	PayloadBefore map[string]interface{} `json:"payload_before"`
}

// ReviewRequest is the payload for POST /governance/submissions/:id/review.
type ReviewRequest struct {
	ReviewNotes string `json:"review_notes"`
}

// RejectRequest is the payload for POST /governance/submissions/:id/reject.
type RejectRequest struct {
	RejectionReason string `json:"rejection_reason" binding:"required"`
}

// LockRequest is the payload for POST /governance/submissions/:id/lock.
type LockRequest struct {
	LockReason string `json:"lock_reason"`
}

// CancelRequest is the payload for POST /governance/submissions/:id/cancel.
type CancelRequest struct {
	CancelReason string `json:"cancel_reason"`
}

// CreateLockPeriodRequest is the payload for POST /governance/lock-periods.
type CreateLockPeriodRequest struct {
	DatasetType string `json:"dataset_type" binding:"required"`
	PeriodYear  int    `json:"period_year" binding:"required"`
	PeriodMonth *int   `json:"period_month"`
	LockReason  string `json:"lock_reason"`
	// LockNow true → immediately LOCKED; false/omitted → OPEN
	LockNow bool `json:"lock_now"`
}

// ListSubmissionsFilter holds query parameters for listing submissions.
type ListSubmissionsFilter struct {
	Status      string
	DatasetType string
	SourceType  string
	PeriodYear  int
	PeriodMonth int
	Page        int
	PageSize    int
}

// ListLockPeriodsFilter holds query parameters for listing lock periods.
type ListLockPeriodsFilter struct {
	DatasetType string
	Status      string
	PeriodYear  int
	Page        int
	PageSize    int
}

// SubmissionDetail is the response shape for GET /governance/submissions/:id.
// It embeds the submission plus its items.
type SubmissionDetail struct {
	*Submission
	Items []SubmissionItem `json:"items"`
}

// Submission is the extended data_submissions model. The base table already
// exists (created by the dataquality module); this struct carries the extra
// official-approval fields added by migration 000030/000031.
//
// Official governance submissions are NOT tied to a monitoring snapshot by
// default: ProjectID and SnapshotID are nullable so both dataquality (snapshot
// validation queue) and governance (standalone official submission) can share
// the table without FK/unique constraint collisions.
type Submission struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"organization_id"`
	ProjectID        *uuid.UUID `gorm:"type:uuid;index" json:"project_id,omitempty"`
	SnapshotID       *uuid.UUID `gorm:"type:uuid;index" json:"snapshot_id,omitempty"`
	Source           string     `gorm:"size:100" json:"source,omitempty"`
	SourceReference  string     `gorm:"size:255" json:"source_reference,omitempty"`
	SourceType       string     `gorm:"size:50;not null;default:'MANUAL'" json:"source_type"`
	DatasetType      string     `gorm:"size:50;not null;default:'OTHER'" json:"dataset_type"`
	SourceEntityType string     `gorm:"size:100" json:"source_entity_type,omitempty"`
	SourceEntityID   *uuid.UUID `gorm:"type:uuid" json:"source_entity_id,omitempty"`
	PeriodYear       int        `json:"period_year"`
	PeriodMonth      *int       `json:"period_month,omitempty"`
	Status           string     `gorm:"size:20;not null" json:"status"`
	CompletenessPct  float64    `gorm:"type:decimal(5,2);not null" json:"completeness_pct"`
	FreshnessAt      *time.Time `json:"freshness_at,omitempty"`
	FreshnessDays    *int       `json:"freshness_days,omitempty"`
	SLADueAt         *time.Time `json:"sla_due_at,omitempty"`
	// CreatedBy records who created the DRAFT submission. It is set at creation
	// time (audit event is also recorded). It must NOT be confused with
	// SubmittedBy/SubmittedAt which are only populated at Submit time.
	CreatedBy       *uuid.UUID      `gorm:"type:uuid" json:"created_by,omitempty"`
	SubmittedBy     *uuid.UUID      `gorm:"type:uuid" json:"submitted_by,omitempty"`
	SubmittedAt     *time.Time      `json:"submitted_at,omitempty"`
	ValidatorID     *uuid.UUID      `gorm:"type:uuid" json:"validator_id,omitempty"`
	ValidatedAt     *time.Time      `json:"validated_at,omitempty"`
	RejectionReason string          `gorm:"type:text" json:"rejection_reason,omitempty"`
	ReviewNotes     string          `gorm:"type:text" json:"review_notes,omitempty"`
	ReviewedBy      *uuid.UUID      `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time      `json:"reviewed_at,omitempty"`
	ApprovedBy      *uuid.UUID      `gorm:"type:uuid" json:"approved_by,omitempty"`
	ApprovedAt      *time.Time      `json:"approved_at,omitempty"`
	LockedBy        *uuid.UUID      `gorm:"type:uuid" json:"locked_by,omitempty"`
	LockedAt        *time.Time      `json:"locked_at,omitempty"`
	Lineage         json.RawMessage `gorm:"type:jsonb;not null" json:"lineage"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (Submission) TableName() string { return "data_submissions" }
