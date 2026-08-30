package government

import (
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

// ConnectorKey identifies a government data source connector.
const (
	ConnectorProjectRegistry   = "government_project_registry"
	ConnectorBudgetReference   = "government_budget_reference"
	ConnectorLocationReference = "government_location_reference"
	ConnectorVendorReference   = "government_vendor_reference"
)

// DatasetType enumerates the kinds of data each connector can provide.
const (
	DatasetProjects         = "projects"
	DatasetBudgetAllocation = "budget_allocations"
	DatasetLocations        = "locations"
	DatasetVendors          = "vendors"
)

// SyncMode controls how the ingestor handles incoming records.
const (
	ModeSample = "SAMPLE"  // fetch a small preview sample, never write
	ModeDryRun = "DRY_RUN" // validate + full preview, never write
	ModeCommit = "COMMIT"  // validate + upsert lineage mappings
)

// SyncStatus is the lifecycle state of a government sync run.
const (
	StatusPending   = "PENDING"
	StatusRunning   = "RUNNING"
	StatusSucceeded = "SUCCEEDED"
	StatusFailed    = "FAILED"
	StatusCancelled = "CANCELLED"
)

// RecordStatus is the outcome of processing a single ingested record.
const (
	RecordAccepted = "ACCEPTED"
	RecordRejected = "REJECTED"
	RecordSkipped  = "SKIPPED"
)

// RecordAction is the write action taken for an accepted record.
const (
	ActionCreate   = "CREATE"
	ActionUpdate   = "UPDATE"
	ActionSkip     = "SKIP"
	ActionConflict = "CONFLICT"
)

// MatchStatus describes whether an external mapping has been resolved
// to a known internal CANKORA entity.
const (
	MatchStatusPendingMatch = "PENDING_MATCH" // external record not yet linked to an internal entity
	MatchStatusMatched      = "MATCHED"       // external record resolved to a real internal entity
	MatchStatusRejected     = "REJECTED"      // mapping explicitly rejected — no internal entity expected
)

// MatchReason is a confidence/reason code returned by the candidate matcher.
const (
	MatchReasonExactCode     = "EXACT_CODE"     // external payload code/ref matched internal entity code exactly
	MatchReasonExactName     = "EXACT_NAME"     // normalised name match (case-insensitive, whitespace trimmed)
	MatchReasonExactNPWP     = "EXACT_NPWP"     // vendor NPWP matched exactly
	MatchReasonPartialName   = "PARTIAL_NAME"   // one side is a substring of the other
	MatchReasonLowConfidence = "LOW_CONFIDENCE" // weak similarity only; requires human review
)

// MatchConfidence is a numeric score 0–100 corresponding to MatchReason.
const (
	ConfidenceExactCode     = 100
	ConfidenceExactName     = 90
	ConfidenceExactNPWP     = 100
	ConfidencePartialName   = 60
	ConfidenceLowConfidence = 30
)

// ConnectorState describes the operational state of a connector.
const (
	ConnectorStateNotConfigured = "NOT_CONFIGURED"
	ConnectorStateSandboxSample = "SANDBOX_SAMPLE"
	ConnectorStateActive        = "ACTIVE"
)

// ---------------------------------------------------------------------------
// SyncRun — top-level record per government data ingestion attempt
// ---------------------------------------------------------------------------

// SyncRun represents one government connector ingestion attempt.
type SyncRun struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	OrganizationID  uuid.UUID  `gorm:"type:uuid;not null"   json:"organization_id"`
	ConnectorKey    string     `gorm:"size:100;not null"    json:"connector_key"`
	DatasetType     string     `gorm:"size:100;not null"    json:"dataset_type"`
	Mode            string     `gorm:"size:20;not null"     json:"mode"`
	Status          string     `gorm:"size:20;not null"     json:"status"`
	StartedBy       uuid.UUID  `gorm:"type:uuid;not null"   json:"started_by"`
	TotalRecords    int        `gorm:"not null;default:0"   json:"total_records"`
	AcceptedRecords int        `gorm:"not null;default:0"   json:"accepted_records"`
	RejectedRecords int        `gorm:"not null;default:0"   json:"rejected_records"`
	ErrorSummary    []byte     `gorm:"type:jsonb;not null"  json:"error_summary"`
	SourceHash      string     `gorm:"size:128;not null"    json:"source_hash"`
	IdempotencyKey  string     `gorm:"size:256;not null"    json:"idempotency_key"`
	StartedAt       *time.Time `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (SyncRun) TableName() string { return "government_sync_runs" }

// ---------------------------------------------------------------------------
// SyncRecord — record-level ingestion log per run
// ---------------------------------------------------------------------------

// SyncRecord logs the outcome of processing one government source record.
type SyncRecord struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SyncRunID        uuid.UUID `gorm:"type:uuid;not null"   json:"sync_run_id"`
	OrganizationID   uuid.UUID `gorm:"type:uuid;not null"   json:"organization_id"`
	ExternalID       string    `gorm:"size:500;not null"    json:"external_id"`
	DatasetType      string    `gorm:"size:100;not null"    json:"dataset_type"`
	Status           string    `gorm:"size:20;not null"     json:"status"`
	Action           string    `gorm:"size:20;not null"     json:"action"`
	ValidationErrors []byte    `gorm:"type:jsonb;not null"  json:"validation_errors"`
	RawPayload       []byte    `gorm:"type:jsonb;not null"  json:"raw_payload"`
	CreatedAt        time.Time `json:"created_at"`
}

func (SyncRecord) TableName() string { return "government_sync_records" }

// ---------------------------------------------------------------------------
// ExternalMapping — external_id → internal CANKORA entity lineage
// ---------------------------------------------------------------------------

// ExternalMapping tracks the link between a government external ID and an
// internal CANKORA entity. Provides idempotency for repeated ingestion.
type ExternalMapping struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OrganizationID     uuid.UUID `gorm:"type:uuid;not null"   json:"organization_id"`
	ConnectorKey       string    `gorm:"size:100;not null"    json:"connector_key"`
	DatasetType        string    `gorm:"size:100;not null"    json:"dataset_type"`
	ExternalID         string    `gorm:"size:500;not null"    json:"external_id"`
	InternalEntityType string    `gorm:"size:100;not null"    json:"internal_entity_type"`
	// InternalEntityID is NULL when the external record has not yet been matched
	// to a known CANKORA entity (MatchStatus = PENDING_MATCH).
	// It is populated only after a real resolution step sets MatchStatus = MATCHED.
	// Never write a random uuid.New() placeholder here.
	InternalEntityID *uuid.UUID `gorm:"type:uuid"            json:"internal_entity_id"`
	// MatchStatus is PENDING_MATCH until the external record is resolved to an
	// internal entity; then MATCHED or REJECTED.
	// Never fabricate a fake UUID to skip this.
	MatchStatus string `gorm:"size:30;not null"     json:"match_status"`
	// MatchConfidence is a 0–100 score set by the candidate matcher (migration 000029).
	MatchConfidence *int `gorm:"column:match_confidence"  json:"match_confidence,omitempty"`
	// MatchReason is a human-readable reason code (EXACT_CODE, EXACT_NAME, etc.).
	MatchReason *string `gorm:"size:50"              json:"match_reason,omitempty"`
	// MatchedBy / MatchedAt record who confirmed the match and when.
	MatchedBy *uuid.UUID `gorm:"type:uuid"            json:"matched_by,omitempty"`
	MatchedAt *time.Time `json:"matched_at,omitempty"`
	// RejectedBy / RejectedAt / RejectReason record rejection details.
	RejectedBy        *uuid.UUID `gorm:"type:uuid"            json:"rejected_by,omitempty"`
	RejectedAt        *time.Time `json:"rejected_at,omitempty"`
	RejectReason      *string    `gorm:"type:text"            json:"reject_reason,omitempty"`
	SourcePayloadHash string     `gorm:"size:128;not null"    json:"source_payload_hash"`
	LastSeenAt        time.Time  `json:"last_seen_at"`
	SyncRunID         *uuid.UUID `gorm:"type:uuid"            json:"sync_run_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (ExternalMapping) TableName() string { return "government_external_mappings" }

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

// ConnectorDefinition describes a registered government connector.
type ConnectorDefinition struct {
	Key          string   `json:"key"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	DatasetTypes []string `json:"dataset_types"`
	State        string   `json:"state"`
	// BaseURL is the configured endpoint; empty string if not set.
	// Never expose credentials/tokens in this field.
	BaseURL string `json:"base_url,omitempty"`
}

// ConnectorConfig holds non-secret configuration metadata for a connector.
// Never include API keys, passwords, or tokens.
type ConnectorConfig struct {
	ConnectorKey string `json:"connector_key"`
	Enabled      bool   `json:"enabled"`
	BaseURL      string `json:"base_url"`
	State        string `json:"state"`
	// SandboxMode is true when the system falls back to mock/sample data.
	SandboxMode bool `json:"sandbox_mode"`
}

// CreateRunRequest is the payload for POST /integrations/government/runs.
type CreateRunRequest struct {
	ConnectorKey   string `json:"connector_key" binding:"required"`
	DatasetType    string `json:"dataset_type"  binding:"required"`
	Mode           string `json:"mode"          binding:"required"`
	IdempotencyKey string `json:"idempotency_key"`
}

// ListRunsFilter holds query parameters for listing sync runs.
type ListRunsFilter struct {
	ConnectorKey string
	DatasetType  string
	Status       string
	Page         int
	PageSize     int
}

// ListRecordsFilter holds query parameters for listing sync records.
type ListRecordsFilter struct {
	Status   string
	Action   string
	Page     int
	PageSize int
}

// ListMappingsFilter holds query parameters for listing external mappings.
type ListMappingsFilter struct {
	ConnectorKey       string
	DatasetType        string
	InternalEntityType string
	MatchStatus        string
	Page               int
	PageSize           int
}

// ---------------------------------------------------------------------------
// Resolution DTOs (P3-002)
// ---------------------------------------------------------------------------

// ResolutionCandidate is an internal CANKORA entity suggested as a match target.
type ResolutionCandidate struct {
	EntityID   string `json:"entity_id"`
	EntityType string `json:"entity_type"`
	Name       string `json:"name"`
	Code       string `json:"code,omitempty"`
	Confidence int    `json:"confidence"`
	Reason     string `json:"reason"`
}

// MatchMappingRequest is the payload for POST /mappings/:id/match.
type MatchMappingRequest struct {
	InternalEntityID   string `json:"internal_entity_id"  binding:"required"`
	InternalEntityType string `json:"internal_entity_type" binding:"required"`
	MatchReason        string `json:"match_reason"`
	MatchConfidence    *int   `json:"match_confidence"`
}

// UnmatchMappingRequest is the payload for POST /mappings/:id/unmatch.
type UnmatchMappingRequest struct {
	// No required fields — all metadata is cleared server-side.
}

// RejectMappingRequest is the payload for POST /mappings/:id/reject.
type RejectMappingRequest struct {
	RejectReason string `json:"reject_reason"`
}

// ListPendingMappingsFilter holds query parameters for the pending queue.
type ListPendingMappingsFilter struct {
	ConnectorKey string
	DatasetType  string
	Page         int
	PageSize     int
}
