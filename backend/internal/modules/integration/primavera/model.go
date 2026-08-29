package primavera

import (
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

// SyncDirection indicates data flow direction.
const (
	DirectionImport = "IMPORT"
	DirectionExport = "EXPORT"
)

// SyncFormat is the source file format.
const (
	FormatXER   = "XER"
	FormatPMXML = "PMXML"
	FormatJSON  = "JSON"
)

// SyncStatus represents the lifecycle state of a sync run.
const (
	StatusPending   = "PENDING"
	StatusRunning   = "RUNNING"
	StatusDone      = "DONE"
	StatusFailed    = "FAILED"
	StatusCancelled = "CANCELLED"
)

// ActivityAction is the outcome of processing a single P6 activity.
const (
	ActionCreate   = "CREATE"
	ActionUpdate   = "UPDATE"
	ActionSkip     = "SKIP"
	ActionConflict = "CONFLICT"
)

// MaxSyncFileSizeBytes is the upper bound for uploaded P6 export files (50 MB).
const MaxSyncFileSizeBytes = 50 * 1024 * 1024

// AllowedSyncMIMETypes is the allowlist of accepted upload content types.
var AllowedSyncMIMETypes = map[string]bool{
	"text/plain":               true, // XER files
	"application/xml":          true,
	"text/xml":                 true,
	"application/octet-stream": true, // some browsers/systems
}

// ---------------------------------------------------------------------------
// SyncRun — top-level record per sync attempt
// ---------------------------------------------------------------------------

// SyncRun represents one Primavera P6 sync attempt.
type SyncRun struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID     uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"organization_id"`
	ProjectID          *uuid.UUID `gorm:"type:uuid;index"                                json:"project_id,omitempty"`
	Direction          string     `gorm:"size:10;not null;default:'IMPORT'"              json:"direction"`
	SourceFileName     string     `gorm:"size:500;not null;default:''"                   json:"source_file_name"`
	SourceFileSize     int64      `gorm:"not null;default:0"                             json:"source_file_size"`
	SourceMIMEType     string     `gorm:"size:100;not null;default:'text/xml'"            json:"source_mime_type"`
	Format             string     `gorm:"size:20;not null;default:'XER'"                 json:"format"`
	Status             string     `gorm:"size:20;not null;default:'PENDING'"             json:"status"`
	TotalActivities    int        `gorm:"not null;default:0"                             json:"total_activities"`
	ImportedActivities int        `gorm:"not null;default:0"                             json:"imported_activities"`
	SkippedActivities  int        `gorm:"not null;default:0"                             json:"skipped_activities"`
	FailedActivities   int        `gorm:"not null;default:0"                             json:"failed_activities"`
	ConflictCount      int        `gorm:"not null;default:0"                             json:"conflict_count"`
	ErrorSummary       string     `gorm:"type:jsonb;not null;default:'[]'"               json:"error_summary"`
	ConflictReport     string     `gorm:"type:jsonb;not null;default:'[]'"               json:"conflict_report"`
	Lineage            string     `gorm:"type:jsonb;not null;default:'{}'"               json:"lineage"`
	StartedAt          *time.Time `json:"started_at,omitempty"`
	FinishedAt         *time.Time `json:"finished_at,omitempty"`
	TriggeredBy        uuid.UUID  `gorm:"type:uuid;not null"                             json:"triggered_by"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (SyncRun) TableName() string { return "primavera_sync_runs" }

// ---------------------------------------------------------------------------
// ActivityMapping — P6 activity_id → CANKORA entity link
// ---------------------------------------------------------------------------

// ActivityMapping links a Primavera P6 activity to a CANKORA entity.
type ActivityMapping struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID   uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"organization_id"`
	ProjectID        uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"project_id"`
	SyncRunID        uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"sync_run_id"`
	P6ActivityID     string     `gorm:"size:100;not null"                              json:"p6_activity_id"`
	P6WBSCode        string     `gorm:"size:200;not null;default:''"                   json:"p6_wbs_code"`
	P6ActivityName   string     `gorm:"size:500;not null;default:''"                   json:"p6_activity_name"`
	EntityType       string     `gorm:"size:50;not null"                               json:"entity_type"`
	EntityID         uuid.UUID  `gorm:"type:uuid;not null"                             json:"entity_id"`
	Action           string     `gorm:"size:20;not null;default:'CREATE'"              json:"action"`
	BaselinePhysical float64    `gorm:"type:numeric(5,2);not null;default:0"           json:"baseline_physical"`
	ActualPhysical   float64    `gorm:"type:numeric(5,2);not null;default:0"           json:"actual_physical"`
	PlannedStart     *time.Time `gorm:"type:date"                                      json:"planned_start,omitempty"`
	PlannedEnd       *time.Time `gorm:"type:date"                                      json:"planned_end,omitempty"`
	ActualStart      *time.Time `gorm:"type:date"                                      json:"actual_start,omitempty"`
	ActualEnd        *time.Time `gorm:"type:date"                                      json:"actual_end,omitempty"`
	RawPayload       string     `gorm:"type:jsonb;not null;default:'{}'"               json:"raw_payload"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (ActivityMapping) TableName() string { return "primavera_activity_mappings" }

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

// ListRunsFilter holds filter params for listing sync runs.
type ListRunsFilter struct {
	ProjectID string
	Status    string
	Format    string
	Page      int
	PageSize  int
}

// SyncRunSummary is the lightweight response for list endpoints.
type SyncRunSummary struct {
	ID                 uuid.UUID  `json:"id"`
	OrganizationID     uuid.UUID  `json:"organization_id"`
	ProjectID          *uuid.UUID `json:"project_id,omitempty"`
	Direction          string     `json:"direction"`
	SourceFileName     string     `json:"source_file_name"`
	Format             string     `json:"format"`
	Status             string     `json:"status"`
	TotalActivities    int        `json:"total_activities"`
	ImportedActivities int        `json:"imported_activities"`
	SkippedActivities  int        `json:"skipped_activities"`
	FailedActivities   int        `json:"failed_activities"`
	ConflictCount      int        `json:"conflict_count"`
	StartedAt          *time.Time `json:"started_at,omitempty"`
	FinishedAt         *time.Time `json:"finished_at,omitempty"`
	TriggeredBy        uuid.UUID  `json:"triggered_by"`
	CreatedAt          time.Time  `json:"created_at"`
}

// ConflictEntry describes a single field-level conflict.
type ConflictEntry struct {
	ActivityID string `json:"activity_id"`
	Field      string `json:"field"`
	Existing   string `json:"existing"`
	Incoming   string `json:"incoming"`
}

// SyncErrorEntry describes a single parse/import error.
type SyncErrorEntry struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	ActivityID string `json:"activity_id,omitempty"`
	Row        int    `json:"row,omitempty"`
}

// LineageMeta is stored as JSON in SyncRun.Lineage.
type LineageMeta struct {
	SourceProjectID string `json:"source_project_id,omitempty"`
	ExportedAt      string `json:"exported_at,omitempty"`
	P6Version       string `json:"p6_version,omitempty"`
	Operator        string `json:"operator,omitempty"`
}
