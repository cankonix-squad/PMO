package audit

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound is returned when an audit log entry cannot be found.
var ErrNotFound = errors.New("audit log not found")

// Log is an IMMUTABLE audit record. No soft delete, no UpdatedAt.
// Once written it must never be modified or deleted.
type Log struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index" json:"organization_id"`
	ActorID        *uuid.UUID `gorm:"type:uuid;index" json:"actor_id,omitempty"`
	ActorEmail     string     `gorm:"size:200" json:"actor_email,omitempty"`
	Action         string     `gorm:"size:100;not null" json:"action"`
	EntityType     string     `gorm:"size:100;not null;index" json:"entity_type"`
	EntityID       string     `gorm:"size:255;index" json:"entity_id,omitempty"`
	EntityLabel    string     `gorm:"size:500" json:"entity_label,omitempty"`
	OldValues      *string    `gorm:"type:jsonb" json:"old_values,omitempty"`
	NewValues      *string    `gorm:"type:jsonb" json:"new_values,omitempty"`
	IPAddress      string     `gorm:"size:50" json:"ip_address,omitempty"`
	UserAgent      string     `gorm:"type:text" json:"user_agent,omitempty"`
	RequestID      string     `gorm:"size:255" json:"request_id,omitempty"`
	CreatedAt      time.Time  `gorm:"index" json:"created_at"`
}

// TableName ensures GORM uses "audit_logs".
func (Log) TableName() string {
	return "audit_logs"
}

// WriteRequest is the input for recording a single audit event.
type WriteRequest struct {
	OrganizationID uuid.UUID
	ActorID        *uuid.UUID
	ActorEmail     string
	Action         string
	EntityType     string
	EntityID       string
	EntityLabel    string
	OldValues      interface{}
	NewValues      interface{}
	IPAddress      string
	UserAgent      string
	RequestID      string
}

// ListFilter defines query filters for fetching audit logs.
type ListFilter struct {
	OrganizationID uuid.UUID
	EntityType     string
	Action         string
	ActorID        *uuid.UUID
	EntityID       string
	Search         string // ILIKE match on actor_email, action, entity_type, entity_label
	From           *time.Time
	To             *time.Time
	Page           int
	PageSize       int
}

// Summary holds aggregate statistics for the audit log viewer header.
type Summary struct {
	TotalEvents  int64         `json:"total_events"`
	UniqueActors int64         `json:"unique_actors"`
	TopActions   []ActionCount `json:"top_actions"`
	TopEntities  []EntityCount `json:"top_entities"`
}

// ActionCount is an action + its occurrence count.
type ActionCount struct {
	Action string `json:"action"`
	Count  int64  `json:"count"`
}

// EntityCount is an entity_type + its occurrence count.
type EntityCount struct {
	EntityType string `json:"entity_type"`
	Count      int64  `json:"count"`
}
