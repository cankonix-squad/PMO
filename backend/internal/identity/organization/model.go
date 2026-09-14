package organization

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrgUnitLevel maps to the 5-level government hierarchy.
// 1=Kementerian, 2=Ditjen, 3=Direktorat, 4=Subdit, 5=Unit
type OrgUnitLevel int

const (
	LevelKementerian OrgUnitLevel = 1
	LevelDitjen      OrgUnitLevel = 2
	LevelDirektorat  OrgUnitLevel = 3
	LevelSubdit      OrgUnitLevel = 4
	LevelUnit        OrgUnitLevel = 5
)

// Organization is the top-level tenant entity.
type Organization struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code      string         `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Name      string         `gorm:"size:300;not null" json:"name"`
	ShortName string         `gorm:"size:100" json:"short_name,omitempty"`
	LogoURL   string         `gorm:"type:text" json:"logo_url,omitempty"`
	Address   string         `gorm:"type:text" json:"address,omitempty"`
	Website   string         `gorm:"size:300" json:"website,omitempty"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	OrgUnits []OrgUnit `gorm:"foreignKey:OrganizationID" json:"org_units,omitempty"`
}

// OrgUnit represents a structural unit within an organization.
// Self-referencing via ParentID to support 5-level hierarchy.
type OrgUnit struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	ParentID       *uuid.UUID     `gorm:"type:uuid" json:"parent_id,omitempty"`
	Code           string         `gorm:"size:100;not null" json:"code"`
	Name           string         `gorm:"size:300;not null" json:"name"`
	Level          OrgUnitLevel   `gorm:"not null" json:"level"`
	HeadUserID     *uuid.UUID     `gorm:"type:uuid" json:"head_user_id,omitempty"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	SortOrder      int            `gorm:"default:0" json:"sort_order"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Children []OrgUnit `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// --- DTOs ---

type CreateOrganizationRequest struct {
	Code      string `json:"code" binding:"required,max=50"`
	Name      string `json:"name" binding:"required,max=300"`
	ShortName string `json:"short_name" binding:"max=100"`
	Address   string `json:"address"`
	Website   string `json:"website" binding:"max=300"`
}

type UpdateOrganizationRequest struct {
	Name      string `json:"name" binding:"max=300"`
	ShortName string `json:"short_name" binding:"max=100"`
	LogoURL   string `json:"logo_url"`
	Address   string `json:"address"`
	Website   string `json:"website" binding:"max=300"`
	IsActive  *bool  `json:"is_active"`
}

type CreateOrgUnitRequest struct {
	ParentID  *uuid.UUID   `json:"parent_id"`
	Code      string       `json:"code" binding:"required,max=100"`
	Name      string       `json:"name" binding:"required,max=300"`
	Level     OrgUnitLevel `json:"level" binding:"required,min=1,max=5"`
	SortOrder int          `json:"sort_order"`
}

type UpdateOrgUnitRequest struct {
	ParentID   *uuid.UUID   `json:"parent_id"`
	Code       string       `json:"code" binding:"max=100"`
	Name       string       `json:"name" binding:"max=300"`
	Level      OrgUnitLevel `json:"level" binding:"min=1,max=5"`
	HeadUserID *uuid.UUID   `json:"head_user_id"`
	IsActive   *bool        `json:"is_active"`
	SortOrder  int          `json:"sort_order"`
}
