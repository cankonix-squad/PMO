package imports

import (
	"time"

	"github.com/google/uuid"
)

// DatasetType enumerates supported import dataset types.
const (
	DatasetProjectProgress     = "project_progress"
	DatasetProjectBudgets      = "project_budgets"
	DatasetRisks               = "risks"
	DatasetIssues              = "issues"
	DatasetBenefitMeasurements = "benefit_measurements"
)

// ImportStatus enumerates job lifecycle states.
const (
	StatusUploaded  = "UPLOADED"
	StatusValidated = "VALIDATED"
	StatusCommitted = "COMMITTED"
	StatusFailed    = "FAILED"
	StatusCancelled = "CANCELLED"
)

// RowAction is the planned action for a single import row.
const (
	ActionCreate = "CREATE"
	ActionUpdate = "UPDATE"
	ActionSkip   = "SKIP"
)

// MaxFileSizeBytes is the upper bound for uploaded files (10 MB).
const MaxFileSizeBytes = 10 * 1024 * 1024

// AllowedMIMETypes is the allowlist of accepted upload content types.
var AllowedMIMETypes = map[string]bool{
	"text/csv":                 true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	"application/octet-stream": true, // some browsers send this for csv
}

// Job is the top-level import job record.
type Job struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index"                      json:"organization_id"`
	DatasetType    string     `gorm:"size:50;not null"                              json:"dataset_type"`
	FileName       string     `gorm:"size:500;not null"                             json:"file_name"`
	FileSize       int64      `gorm:"not null;default:0"                            json:"file_size"`
	MIMEType       string     `gorm:"size:100;not null;default:'text/csv'"           json:"mime_type"`
	Status         string     `gorm:"size:20;not null;default:'UPLOADED'"           json:"status"`
	TotalRows      int        `gorm:"not null;default:0"                            json:"total_rows"`
	ValidRows      int        `gorm:"not null;default:0"                            json:"valid_rows"`
	InvalidRows    int        `gorm:"not null;default:0"                            json:"invalid_rows"`
	ErrorSummary   string     `gorm:"type:jsonb;not null;default:'[]'"              json:"error_summary"`
	UploadedBy     uuid.UUID  `gorm:"type:uuid;not null"                            json:"uploaded_by"`
	ValidatedAt    *time.Time `json:"validated_at,omitempty"`
	CommittedAt    *time.Time `json:"committed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (Job) TableName() string { return "import_jobs" }

// Row is a single parsed row result within a job.
type Row struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	JobID             uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"job_id"`
	RowNumber         int        `gorm:"not null"                                       json:"row_number"`
	RawPayload        string     `gorm:"type:jsonb;not null;default:'{}'"               json:"raw_payload"`
	NormalizedPayload string     `gorm:"type:jsonb;not null;default:'{}'"               json:"normalized_payload"`
	Valid             bool       `gorm:"not null;default:false"                         json:"valid"`
	Errors            string     `gorm:"type:jsonb;not null;default:'[]'"               json:"errors"`
	Action            string     `gorm:"size:20;not null;default:'SKIP'"                json:"action"`
	TargetEntityID    *uuid.UUID `gorm:"type:uuid"                                      json:"target_entity_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

func (Row) TableName() string { return "import_rows" }

// ColumnDef describes a single expected column in a template.
type ColumnDef struct {
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	Type        string `json:"type"` // string | number | date | enum
	Description string `json:"description"`
	Example     string `json:"example,omitempty"`
}

// TemplateDef describes a dataset type's expected schema.
type TemplateDef struct {
	DatasetType string      `json:"dataset_type"`
	DisplayName string      `json:"display_name"`
	Description string      `json:"description"`
	Columns     []ColumnDef `json:"columns"`
}

// --- API request/response types ---

// CreateJobRequest is used when uploading a file (multipart).
// DatasetType is passed as a form field.
type CreateJobRequest struct {
	DatasetType string `form:"dataset_type" binding:"required"`
}

// JobResponse is the API-facing view of a job.
type JobResponse struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	DatasetType    string     `json:"dataset_type"`
	FileName       string     `json:"file_name"`
	FileSize       int64      `json:"file_size"`
	MIMEType       string     `json:"mime_type"`
	Status         string     `json:"status"`
	TotalRows      int        `json:"total_rows"`
	ValidRows      int        `json:"valid_rows"`
	InvalidRows    int        `json:"invalid_rows"`
	ErrorSummary   []string   `json:"error_summary"`
	UploadedBy     uuid.UUID  `json:"uploaded_by"`
	ValidatedAt    *time.Time `json:"validated_at,omitempty"`
	CommittedAt    *time.Time `json:"committed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// RowResponse is the API-facing view of a single import row.
type RowResponse struct {
	ID                uuid.UUID  `json:"id"`
	JobID             uuid.UUID  `json:"job_id"`
	RowNumber         int        `json:"row_number"`
	RawPayload        any        `json:"raw_payload"`
	NormalizedPayload any        `json:"normalized_payload"`
	Valid             bool       `json:"valid"`
	Errors            []string   `json:"errors"`
	Action            string     `json:"action"`
	TargetEntityID    *uuid.UUID `json:"target_entity_id,omitempty"`
}

// ListJobsFilter are query params for the job list endpoint.
type ListJobsFilter struct {
	DatasetType string
	Status      string
	Page        int
	PageSize    int
}
