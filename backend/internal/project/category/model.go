package projectcategory

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ProjectCategory is a tenant-scoped master category for projects.
type ProjectCategory struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"          json:"id"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index"      json:"organization_id"`
	Code           string     `gorm:"type:varchar(100);not null"    json:"code"`
	Name           string     `gorm:"type:varchar(300);not null"    json:"name"`
	Description    *string    `gorm:"type:text"                     json:"description,omitempty"`
	IsActive       bool       `gorm:"not null;default:true"         json:"is_active"`
	SortOrder      int        `gorm:"not null;default:0"            json:"sort_order"`
	CreatedBy      *uuid.UUID `gorm:"type:uuid"                     json:"created_by,omitempty"`
	CreatedAt      time.Time  `gorm:"not null;autoCreateTime"       json:"created_at"`
	UpdatedAt      time.Time  `gorm:"not null;autoUpdateTime"       json:"updated_at"`
	DeletedAt      *time.Time `gorm:"index"                         json:"deleted_at,omitempty"`
}

func (ProjectCategory) TableName() string { return "project_categories" }

// ---- Request / Response DTOs ------------------------------------------------

type CreateRequest struct {
	Code        string  `json:"code"        binding:"required,max=100"`
	Name        string  `json:"name"        binding:"required,max=300"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
	SortOrder   *int    `json:"sort_order"`
}

type UpdateRequest struct {
	Code        *string `json:"code"        binding:"omitempty,max=100"`
	Name        *string `json:"name"        binding:"omitempty,max=300"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
	SortOrder   *int    `json:"sort_order"`
}

// ---- Sentinel errors --------------------------------------------------------

var (
	ErrNotFound  = errors.New("project category not found")
	ErrCodeTaken = errors.New("project category code already in use")
)
