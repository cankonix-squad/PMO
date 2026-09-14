package notification

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Notification status values.
const (
	StatusPending = "PENDING"
	StatusSent    = "SENT"
	StatusFailed  = "FAILED"
	StatusRead    = "READ"
)

// Notification channel values.
const (
	ChannelInApp = "IN_APP"
	ChannelEmail = "EMAIL"
)

// Notification priority values.
const (
	PriorityLow    = "LOW"
	PriorityNormal = "NORMAL"
	PriorityHigh   = "HIGH"
	PriorityUrgent = "URGENT"
)

// ErrNotFound is returned when a notification row cannot be found (or belongs to another tenant).
var ErrNotFound = errors.New("notification not found")

// Notification is a persistent notification record stored in the DB.
// It serves as both the in-app inbox and the email outbox.
type Notification struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID  uuid.UUID      `gorm:"type:uuid;not null;index"                       json:"organization_id"`
	RecipientUserID *uuid.UUID     `gorm:"type:uuid;index"                                json:"recipient_user_id,omitempty"`
	Channel         string         `gorm:"size:20;not null;default:IN_APP"                json:"channel"`
	Status          string         `gorm:"size:20;not null;default:PENDING"               json:"status"`
	Priority        string         `gorm:"size:20;not null;default:NORMAL"                json:"priority"`
	Subject         string         `gorm:"size:500;not null"                              json:"subject"`
	Body            string         `gorm:"type:text;not null"                             json:"body"`
	SourceType      string         `gorm:"size:100"                                       json:"source_type,omitempty"`
	SourceID        string         `gorm:"size:255"                                       json:"source_id,omitempty"`
	ErrorMessage    string         `gorm:"type:text"                                      json:"error_message,omitempty"`
	SentAt          *time.Time     `gorm:"index"                                          json:"sent_at,omitempty"`
	ReadAt          *time.Time     `                                                      json:"read_at,omitempty"`
	CreatedAt       time.Time      `gorm:"index"                                          json:"created_at"`
	UpdatedAt       time.Time      `                                                      json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index"                                          json:"-"`
}

// TableName tells GORM which table to use.
func (Notification) TableName() string { return "notifications" }

// ListFilter holds all query filters for listing notifications.
type ListFilter struct {
	OrganizationID  uuid.UUID
	RecipientUserID *uuid.UUID // nil = all users in org (admin view)
	Status          string
	Channel         string
	Priority        string
	SourceType      string
	UnreadOnly      bool
	Page            int
	PageSize        int
}

// EnqueueRequest is the input for creating a notification record.
type EnqueueRequest struct {
	OrganizationID  uuid.UUID
	RecipientUserID *uuid.UUID
	Channel         string
	Priority        string
	Subject         string
	Body            string
	SourceType      string
	SourceID        string
}

// Summary holds aggregate stats for the notification inbox header.
type Summary struct {
	Total   int64 `json:"total"`
	Unread  int64 `json:"unread"`
	Pending int64 `json:"pending"`
	Failed  int64 `json:"failed"`
}
