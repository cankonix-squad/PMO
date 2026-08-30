package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents the users table.
type User struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	OrgUnitID      *uuid.UUID     `gorm:"type:uuid" json:"org_unit_id"`
	EmployeeID     *string        `gorm:"size:100;uniqueIndex" json:"employee_id,omitempty"`
	FirstName      string         `gorm:"size:100;not null" json:"first_name"`
	LastName       string         `gorm:"size:100" json:"last_name"`
	Email          string         `gorm:"size:200;uniqueIndex;not null" json:"email"`
	PasswordHash   string         `gorm:"size:255;not null" json:"-"`
	Phone          string         `gorm:"size:50" json:"phone,omitempty"`
	JobTitle       string         `gorm:"size:200" json:"job_title,omitempty"`
	AvatarURL      string         `gorm:"type:text" json:"avatar_url,omitempty"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	MustChangePwd  bool           `gorm:"default:true" json:"must_change_pwd"`
	LastLoginAt    *time.Time     `json:"last_login_at,omitempty"`
	LoginFailed    int            `gorm:"default:0" json:"-"`
	LockedUntil    *time.Time     `json:"-"`
	CreatedBy      *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserSession tracks active JWT sessions for invalidation on logout.
type UserSession struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	JTI        string    `gorm:"size:255;uniqueIndex;not null" json:"jti"`
	RefreshJTI string    `gorm:"size:255;uniqueIndex;not null" json:"refresh_jti"`
	UserAgent  string    `gorm:"type:text" json:"user_agent,omitempty"`
	IPAddress  string    `gorm:"size:50" json:"ip_address,omitempty"`
	IsRevoked  bool      `gorm:"default:false" json:"is_revoked"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// PasswordResetToken stores one-time password reset tokens.
type PasswordResetToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash string     `gorm:"size:255;not null" json:"token_hash"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// --- Request / Response DTOs ---

// LoginRequest is the payload for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// TokenPair holds the access and refresh tokens returned on login.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserInfo  `json:"user"`
}

// UserInfo is the safe user payload embedded in TokenPair.
type UserInfo struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	OrgUnitID      *uuid.UUID `json:"org_unit_id,omitempty"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	Email          string     `json:"email"`
	JobTitle       string     `json:"job_title,omitempty"`
	AvatarURL      string     `json:"avatar_url,omitempty"`
	MustChangePwd  bool       `json:"must_change_pwd"`
	Roles          []string   `json:"roles"`
}

// RefreshRequest is the payload for POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest is the payload for POST /auth/change-password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// ForgotPasswordRequest is the payload for POST /auth/forgot-password.
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest is the payload for POST /auth/reset-password.
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// Claims represents JWT payload.
type Claims struct {
	UserID         uuid.UUID  `json:"sub"`
	Email          string     `json:"email"`
	OrganizationID uuid.UUID  `json:"org_id"`
	OrgUnitID      *uuid.UUID `json:"org_unit_id,omitempty"`
	JTI            string     `json:"jti"`
	IsRefresh      bool       `json:"is_refresh,omitempty"`
}

// ContextKey is used as a typed key for context values to avoid collisions.
type ContextKey string

const (
	ContextKeyUser   ContextKey = "auth_user"
	ContextKeyClaims ContextKey = "auth_claims"
)
