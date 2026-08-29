package bim

import (
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

// Discipline identifies the engineering discipline of a BIM model.
const (
	DisciplineArchitectural = "ARCHITECTURAL"
	DisciplineStructural    = "STRUCTURAL"
	DisciplineMEP           = "MEP"
	DisciplineCivil         = "CIVIL"
	DisciplineLandscape     = "LANDSCAPE"
	DisciplineOther         = "OTHER"
)

// Provider identifies the external BIM platform hosting the model.
const (
	ProviderAutodeskBIM360     = "autodesk_bim360"
	ProviderTrimbleConnect     = "trimble_connect"
	ProviderBentleyProjectWise = "bentley_projectwise"
	ProviderLocal              = "local"
	ProviderOther              = "other"
)

// ModelStatus is the lifecycle state of a BIM model registration.
const (
	ModelStatusDraft    = "DRAFT"
	ModelStatusActive   = "ACTIVE"
	ModelStatusArchived = "ARCHIVED"
)

// ModelRole describes how a BIM model is used within a project.
const (
	ModelRolePrimary   = "PRIMARY"
	ModelRoleReference = "REFERENCE"
	ModelRoleAsBuilt   = "ASBUILT"
	ModelRoleOther     = "OTHER"
)

// ---------------------------------------------------------------------------
// BIMModel — registered BIM model (metadata only, no binary storage)
// ---------------------------------------------------------------------------

// BIMModel represents the registration of an external BIM model in CANKORA.
// Binary files are never stored in database rows; only metadata and viewer URIs.
type BIMModel struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey"  json:"id"`
	OrganizationID  uuid.UUID  `gorm:"type:uuid;not null"    json:"organization_id"`
	Name            string     `gorm:"size:255;not null"     json:"name"`
	Description     string     `gorm:"type:text;not null"    json:"description"`
	Discipline      string     `gorm:"size:50;not null"      json:"discipline"`
	Provider        string     `gorm:"size:100;not null"     json:"provider"`
	ExternalModelID string     `gorm:"size:500;not null"     json:"external_model_id"`
	ViewerURL       string     `gorm:"type:text;not null"    json:"viewer_url"`
	Status          string     `gorm:"size:20;not null"      json:"status"`
	Metadata        []byte     `gorm:"type:jsonb;not null"   json:"metadata"`
	CreatedBy       uuid.UUID  `gorm:"type:uuid;not null"    json:"created_by"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (BIMModel) TableName() string { return "bim_models" }

// ---------------------------------------------------------------------------
// BIMModelVersion — immutable version record per BIM model
// ---------------------------------------------------------------------------

// BIMModelVersion is an immutable snapshot of a BIM model at a point in time.
// Published versions cannot be modified; corrections create a new version.
type BIMModelVersion struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	BIMModelID        uuid.UUID  `gorm:"type:uuid;not null"   json:"bim_model_id"`
	OrganizationID    uuid.UUID  `gorm:"type:uuid;not null"   json:"organization_id"`
	VersionLabel      string     `gorm:"size:100;not null"    json:"version_label"`
	ExternalVersionID string     `gorm:"size:500;not null"    json:"external_version_id"`
	ChangeSummary     string     `gorm:"type:text;not null"   json:"change_summary"`
	FileSizeBytes     int64      `gorm:"not null;default:0"   json:"file_size_bytes"`
	Checksum          string     `gorm:"size:128;not null"    json:"checksum"`
	PublishedAt       *time.Time `json:"published_at,omitempty"`
	CreatedBy         uuid.UUID  `gorm:"type:uuid;not null"   json:"created_by"`
	CreatedAt         time.Time  `json:"created_at"`
}

func (BIMModelVersion) TableName() string { return "bim_model_versions" }

// ---------------------------------------------------------------------------
// BIMProjectMapping — links a BIM model to a project
// ---------------------------------------------------------------------------

// BIMProjectMapping records the association between a BIM model and a CANKORA
// project, including the role the model plays within that project.
type BIMProjectMapping struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null"   json:"organization_id"`
	BIMModelID     uuid.UUID `gorm:"type:uuid;not null"   json:"bim_model_id"`
	ProjectID      uuid.UUID `gorm:"type:uuid;not null"   json:"project_id"`
	ModelRole      string    `gorm:"size:50;not null"     json:"model_role"`
	Notes          string    `gorm:"type:text;not null"   json:"notes"`
	LinkedBy       uuid.UUID `gorm:"type:uuid;not null"   json:"linked_by"`
	LinkedAt       time.Time `gorm:"not null"             json:"linked_at"`
}

func (BIMProjectMapping) TableName() string { return "bim_project_mappings" }

// ---------------------------------------------------------------------------
// Request / Response DTOs
// ---------------------------------------------------------------------------

// CreateBIMModelRequest is the payload for registering a new BIM model.
type CreateBIMModelRequest struct {
	Name            string                 `json:"name"              binding:"required,min=1,max=255"`
	Description     string                 `json:"description"`
	Discipline      string                 `json:"discipline"        binding:"required,oneof=ARCHITECTURAL STRUCTURAL MEP CIVIL LANDSCAPE OTHER"`
	Provider        string                 `json:"provider"          binding:"required,oneof=autodesk_bim360 trimble_connect bentley_projectwise local other"`
	ExternalModelID string                 `json:"external_model_id" binding:"required,min=1,max=500"`
	ViewerURL       string                 `json:"viewer_url"        binding:"omitempty,max=2048"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// UpdateBIMModelRequest is the payload for updating an existing BIM model.
type UpdateBIMModelRequest struct {
	Name        *string                 `json:"name"        binding:"omitempty,min=1,max=255"`
	Description *string                 `json:"description"`
	Discipline  *string                 `json:"discipline"  binding:"omitempty,oneof=ARCHITECTURAL STRUCTURAL MEP CIVIL LANDSCAPE OTHER"`
	ViewerURL   *string                 `json:"viewer_url"  binding:"omitempty,max=2048"`
	Status      *string                 `json:"status"      binding:"omitempty,oneof=DRAFT ACTIVE ARCHIVED"`
	Metadata    *map[string]interface{} `json:"metadata"`
}

// CreateVersionRequest is the payload for adding a new BIM model version.
type CreateVersionRequest struct {
	VersionLabel      string `json:"version_label"       binding:"required,min=1,max=100"`
	ExternalVersionID string `json:"external_version_id" binding:"omitempty,max=500"`
	ChangeSummary     string `json:"change_summary"`
	FileSizeBytes     int64  `json:"file_size_bytes"     binding:"omitempty,min=0"`
	Checksum          string `json:"checksum"            binding:"omitempty,max=128"`
}

// LinkProjectRequest is the payload for linking a BIM model to a project.
type LinkProjectRequest struct {
	ProjectID uuid.UUID `json:"project_id" binding:"required"`
	ModelRole string    `json:"model_role" binding:"required,oneof=PRIMARY REFERENCE ASBUILT OTHER"`
	Notes     string    `json:"notes"`
}

// BIMModelListResponse wraps a paginated list of BIM models.
type BIMModelListResponse struct {
	Items []BIMModel `json:"items"`
	Total int64      `json:"total"`
}

// BIMVersionListResponse wraps a list of versions for one model.
type BIMVersionListResponse struct {
	Items []BIMModelVersion `json:"items"`
}

// BIMMappingListResponse wraps a list of project mappings for one model.
type BIMMappingListResponse struct {
	Items []BIMProjectMapping `json:"items"`
}
