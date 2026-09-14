package dataquality

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	StatusDraft     = "DRAFT"
	StatusSubmitted = "SUBMITTED"
	StatusValid     = "VALID"
	StatusRejected  = "REJECTED"
	StatusStale     = "STALE"
)

type Submission struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID  uuid.UUID       `gorm:"type:uuid;not null;index" json:"organization_id"`
	ProjectID       uuid.UUID       `gorm:"type:uuid;not null;index" json:"project_id"`
	SnapshotID      uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex" json:"snapshot_id"`
	Source          string          `gorm:"size:100" json:"source,omitempty"`
	SourceReference string          `gorm:"size:255" json:"source_reference,omitempty"`
	PeriodYear      int             `json:"period_year"`
	PeriodMonth     int             `json:"period_month"`
	Status          string          `gorm:"size:20;not null" json:"status"`
	CompletenessPct float64         `gorm:"type:decimal(5,2);not null" json:"completeness_pct"`
	FreshnessAt     *time.Time      `json:"freshness_at,omitempty"`
	FreshnessDays   *int            `json:"freshness_days,omitempty"`
	SLADueAt        *time.Time      `json:"sla_due_at,omitempty"`
	SubmittedBy     *uuid.UUID      `gorm:"type:uuid" json:"submitted_by,omitempty"`
	SubmittedAt     *time.Time      `json:"submitted_at,omitempty"`
	ValidatorID     *uuid.UUID      `gorm:"type:uuid" json:"validator_id,omitempty"`
	ValidatedAt     *time.Time      `json:"validated_at,omitempty"`
	RejectionReason string          `gorm:"type:text" json:"rejection_reason,omitempty"`
	Lineage         json.RawMessage `gorm:"type:jsonb;not null" json:"lineage"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (Submission) TableName() string { return "data_submissions" }

type CreateSubmissionRequest struct {
	SnapshotID      string                 `json:"snapshot_id" binding:"required"`
	CompletenessPct float64                `json:"completeness_pct" binding:"min=0,max=100"`
	FreshnessAt     *time.Time             `json:"freshness_at"`
	SLAHours        int                    `json:"sla_hours"`
	SourceReference string                 `json:"source_reference"`
	Lineage         map[string]interface{} `json:"lineage"`
}

type TransitionRequest struct {
	Status          string `json:"status" binding:"required,oneof=VALID REJECTED STALE"`
	RejectionReason string `json:"rejection_reason"`
}
