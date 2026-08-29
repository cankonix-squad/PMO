package rbac

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role represents a named set of permissions.
type Role struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	Code           string         `gorm:"size:100;not null" json:"code"`
	Name           string         `gorm:"size:200;not null" json:"name"`
	Description    string         `gorm:"type:text" json:"description,omitempty"`
	IsSystem       bool           `gorm:"default:false" json:"is_system"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

// Permission represents a single capability (resource + action).
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Resource    string    `gorm:"size:100;not null" json:"resource"`
	Action      string    `gorm:"size:100;not null" json:"action"`
	Description string    `gorm:"size:500" json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// RolePermission is the explicit join table for Role ↔ Permission.
type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"role_id"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// Group is an optional layer that aggregates users and inherits roles.
type Group struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	Name           string         `gorm:"size:200;not null" json:"name"`
	Description    string         `gorm:"type:text" json:"description,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// GroupRole binds a Role to a Group.
type GroupRole struct {
	GroupID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"group_id"`
	RoleID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

// UserGroup binds a User to a Group.
type UserGroup struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	GroupID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"group_id"`
	CreatedAt time.Time `json:"created_at"`
}

// UserRole binds a Role directly to a User (bypassing group).
type UserRole struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	RoleID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

// UserScope restricts a user's visibility/action to a specific org unit.
// ScopeType: "unit" | "direktorat" | "ditjen" | "organization"
type UserScope struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null" json:"organization_id"`
	ScopeType      string     `gorm:"size:50;not null" json:"scope_type"`
	OrgUnitID      *uuid.UUID `gorm:"type:uuid" json:"org_unit_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}
