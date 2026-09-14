package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Profile is a read-only view model for user listings.
// It does NOT include sensitive fields like password_hash.
type Profile struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	OrgUnitID      *uuid.UUID `json:"org_unit_id,omitempty"`
	OrgUnitName    string     `json:"org_unit_name,omitempty"`
	EmployeeID     *string    `json:"employee_id,omitempty"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	FullName       string     `json:"full_name"`
	Email          string     `json:"email"`
	Phone          string     `json:"phone,omitempty"`
	JobTitle       string     `json:"job_title,omitempty"`
	AvatarURL      string     `json:"avatar_url,omitempty"`
	IsActive       bool       `json:"is_active"`
	MustChangePwd  bool       `json:"must_change_pwd"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
	Roles          []RoleRef  `json:"roles,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// RoleRef is a lightweight role reference embedded in a user profile.
type RoleRef struct {
	ID   uuid.UUID `json:"id"`
	Code string    `json:"code"`
	Name string    `json:"name"`
}

// CreateUserRequest is the admin payload to provision a new user.
type CreateUserRequest struct {
	OrganizationID uuid.UUID   `json:"organization_id"`
	OrgUnitID      *uuid.UUID  `json:"org_unit_id"`
	EmployeeID     *string     `json:"employee_id"`
	FirstName      string      `json:"first_name" binding:"required,max=100"`
	LastName       string      `json:"last_name" binding:"max=100"`
	Email          string      `json:"email" binding:"required,email"`
	Password       string      `json:"password" binding:"required,min=8"`
	Phone          string      `json:"phone" binding:"max=50"`
	JobTitle       string      `json:"job_title" binding:"max=200"`
	IsActive       *bool       `json:"is_active"`
	RoleIDs        []uuid.UUID `json:"role_ids"`
}

// UpdateUserRequest is the admin payload to update an existing user.
type UpdateUserRequest struct {
	OrgUnitID  *uuid.UUID  `json:"org_unit_id"`
	EmployeeID *string     `json:"employee_id"`
	FirstName  string      `json:"first_name" binding:"max=100"`
	LastName   string      `json:"last_name" binding:"max=100"`
	Phone      string      `json:"phone" binding:"max=50"`
	JobTitle   string      `json:"job_title" binding:"max=200"`
	AvatarURL  string      `json:"avatar_url"`
	IsActive   *bool       `json:"is_active"`
	RoleIDs    []uuid.UUID `json:"role_ids"`
}

// UpdateProfileRequest is the self-service payload for updating own profile.
type UpdateProfileRequest struct {
	FirstName string `json:"first_name" binding:"max=100"`
	LastName  string `json:"last_name" binding:"max=100"`
	Phone     string `json:"phone" binding:"max=50"`
	JobTitle  string `json:"job_title" binding:"max=200"`
	AvatarURL string `json:"avatar_url"`
}

// UserListFilter defines query filters for listing users.
type UserListFilter struct {
	OrganizationID uuid.UUID
	OrgUnitID      *uuid.UUID
	Search         string
	IsActive       *bool
	Page           int
	PageSize       int
}

// userGorm is an internal GORM model alias used only by the repository.
// We reuse auth.User directly to avoid duplication.
type userGorm struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null"`
	OrgUnitID      *uuid.UUID `gorm:"type:uuid"`
	EmployeeID     *string    `gorm:"size:100"`
	FirstName      string     `gorm:"size:100"`
	LastName       string     `gorm:"size:100"`
	Email          string     `gorm:"size:200;uniqueIndex"`
	PasswordHash   string     `gorm:"size:255"`
	Phone          string     `gorm:"size:50"`
	JobTitle       string     `gorm:"size:200"`
	AvatarURL      string     `gorm:"type:text"`
	IsActive       bool       `gorm:"default:true"`
	MustChangePwd  bool       `gorm:"default:true"`
	LastLoginAt    *time.Time
	LoginFailed    int `gorm:"default:0"`
	LockedUntil    *time.Time
	CreatedBy      *uuid.UUID `gorm:"type:uuid"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (userGorm) TableName() string { return "users" }
