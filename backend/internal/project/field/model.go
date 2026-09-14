package field

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Inspection struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"organization_id"`
	ProjectID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	InspectedAt        time.Time      `json:"inspected_at"`
	Latitude           *float64       `json:"latitude,omitempty"`
	Longitude          *float64       `json:"longitude,omitempty"`
	InspectorID        uuid.UUID      `gorm:"type:uuid;not null" json:"inspector_id"`
	Notes              string         `gorm:"type:text" json:"notes,omitempty"`
	VerificationStatus string         `gorm:"size:20;not null" json:"verification_status"`
	VerifiedBy         *uuid.UUID     `gorm:"type:uuid" json:"verified_by,omitempty"`
	VerifiedAt         *time.Time     `json:"verified_at,omitempty"`
	Evidence           []Evidence     `json:"evidence,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Inspection) TableName() string { return "field_inspections" }

type Evidence struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"organization_id"`
	ProjectID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	InspectionID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"inspection_id"`
	FileName           string         `gorm:"size:500;not null" json:"file_name"`
	StorageKey         string         `gorm:"size:1000;not null" json:"-"`
	MimeType           string         `gorm:"size:255;not null" json:"mime_type"`
	FileSize           int64          `json:"file_size"`
	ChecksumSHA256     string         `gorm:"size:64;not null" json:"checksum_sha256"`
	CapturedAt         *time.Time     `json:"captured_at,omitempty"`
	Latitude           *float64       `json:"latitude,omitempty"`
	Longitude          *float64       `json:"longitude,omitempty"`
	VerificationStatus string         `gorm:"size:20;not null" json:"verification_status"`
	VerifiedBy         *uuid.UUID     `gorm:"type:uuid" json:"verified_by,omitempty"`
	VerifiedAt         *time.Time     `json:"verified_at,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Evidence) TableName() string { return "field_evidence" }

type CreateInspectionRequest struct {
	InspectedAt time.Time `form:"inspected_at" binding:"required"`
	Latitude    *float64  `form:"latitude"`
	Longitude   *float64  `form:"longitude"`
	Notes       string    `form:"notes"`
}

type VerifyRequest struct {
	Status string `json:"status" binding:"required,oneof=VERIFIED REJECTED"`
}
