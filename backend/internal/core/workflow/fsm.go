package workflow

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrTransitionNotAllowed is returned when a status transition is not permitted.
var ErrTransitionNotAllowed = errors.New("transition not allowed from current status")

// EntityType identifies which entity the FSM applies to.
type EntityType string

const (
	EntityProject          EntityType = "project"
	EntityTask             EntityType = "task"
	EntityMilestone        EntityType = "milestone"
	EntityIssue            EntityType = "issue"
	EntityRisk             EntityType = "risk"
	EntityCorrectiveAction EntityType = "corrective_action"
)

// Transition defines a single valid state change.
type Transition struct {
	From         string
	To           string
	RequiresRole string // optional: role code required to perform this transition
}

// FSM is a Finite State Machine that enforces valid status transitions per entity type.
type FSM struct {
	transitions map[EntityType][]Transition
}

// New creates an FSM pre-loaded with all CANKORA business transitions.
func New() *FSM {
	fsm := &FSM{
		transitions: make(map[EntityType][]Transition),
	}

	// Project: DRAFT → PLANNING → ACTIVE → ON_HOLD → COMPLETED | CANCELLED
	fsm.transitions[EntityProject] = []Transition{
		{From: "DRAFT", To: "PLANNING"},
		{From: "PLANNING", To: "ACTIVE"},
		{From: "PLANNING", To: "CANCELLED"},
		{From: "ACTIVE", To: "ON_HOLD"},
		{From: "ON_HOLD", To: "ACTIVE"},
		{From: "ACTIVE", To: "COMPLETED"},
		{From: "COMPLETED", To: "ACTIVE"}, // reopen
		{From: "DRAFT", To: "CANCELLED"},
		{From: "ACTIVE", To: "CANCELLED"},
		{From: "ON_HOLD", To: "CANCELLED"},
	}

	// Task: TODO → IN_PROGRESS → IN_REVIEW → DONE | BLOCKED
	fsm.transitions[EntityTask] = []Transition{
		{From: "TODO", To: "IN_PROGRESS"},
		{From: "IN_PROGRESS", To: "IN_REVIEW"},
		{From: "IN_PROGRESS", To: "BLOCKED"},
		{From: "IN_REVIEW", To: "DONE"},
		{From: "IN_REVIEW", To: "IN_PROGRESS"}, // revision
		{From: "BLOCKED", To: "IN_PROGRESS"},
		{From: "DONE", To: "IN_PROGRESS"}, // reopen
	}

	// Milestone: PENDING → IN_PROGRESS → COMPLETED | DELAYED
	fsm.transitions[EntityMilestone] = []Transition{
		{From: "PENDING", To: "IN_PROGRESS"},
		{From: "IN_PROGRESS", To: "COMPLETED"},
		{From: "IN_PROGRESS", To: "DELAYED"},
		{From: "DELAYED", To: "IN_PROGRESS"},
	}

	// Issue: OPEN → IN_PROGRESS → RESOLVED → CLOSED | REOPENED
	fsm.transitions[EntityIssue] = []Transition{
		{From: "OPEN", To: "IN_PROGRESS"},
		{From: "IN_PROGRESS", To: "RESOLVED"},
		{From: "RESOLVED", To: "CLOSED"},
		{From: "RESOLVED", To: "OPEN"}, // reopen
		{From: "CLOSED", To: "OPEN"},
	}

	// Risk: IDENTIFIED → ASSESSED → MITIGATED | ACCEPTED | ESCALATED | CLOSED
	fsm.transitions[EntityRisk] = []Transition{
		{From: "IDENTIFIED", To: "ASSESSED"},
		{From: "ASSESSED", To: "MITIGATED"},
		{From: "ASSESSED", To: "ACCEPTED"},
		{From: "ASSESSED", To: "ESCALATED"},
		{From: "MITIGATED", To: "CLOSED"},
		{From: "ACCEPTED", To: "CLOSED"},
		{From: "ESCALATED", To: "MITIGATED"},
	}

	// CorrectiveAction: DRAFT → SUBMITTED → IN_PROGRESS → COMPLETED | REJECTED
	//                   REJECTED → DRAFT (revise & resubmit)
	fsm.transitions[EntityCorrectiveAction] = []Transition{
		{From: "DRAFT", To: "SUBMITTED"},
		{From: "SUBMITTED", To: "IN_PROGRESS"},
		{From: "SUBMITTED", To: "REJECTED"},
		{From: "IN_PROGRESS", To: "COMPLETED"},
		{From: "IN_PROGRESS", To: "REJECTED"},
		{From: "REJECTED", To: "DRAFT"}, // revise & resubmit
	}

	return fsm
}

// CanTransition returns true if moving from → to is permitted for the given entity type.
func (f *FSM) CanTransition(entity EntityType, from, to string) bool {
	transitions, ok := f.transitions[entity]
	if !ok {
		return false
	}
	for _, t := range transitions {
		if t.From == from && t.To == to {
			return true
		}
	}
	return false
}

// Transition validates and returns the matching Transition, or ErrTransitionNotAllowed.
func (f *FSM) Transition(entity EntityType, from, to string) (*Transition, error) {
	transitions, ok := f.transitions[entity]
	if !ok {
		return nil, ErrTransitionNotAllowed
	}
	for _, t := range transitions {
		if t.From == from && t.To == to {
			tt := t
			return &tt, nil
		}
	}
	return nil, ErrTransitionNotAllowed
}

// AllowedTransitions returns the list of valid "to" states from the current state.
func (f *FSM) AllowedTransitions(entity EntityType, from string) []string {
	transitions, ok := f.transitions[entity]
	if !ok {
		return nil
	}
	var allowed []string
	for _, t := range transitions {
		if t.From == from {
			allowed = append(allowed, t.To)
		}
	}
	return allowed
}

// --- Approval workflow model ---

// ApprovalRequest tracks a pending approval for a status transition.
type ApprovalRequest struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EntityType  string     `gorm:"size:100;not null" json:"entity_type"`
	EntityID    uuid.UUID  `gorm:"type:uuid;not null" json:"entity_id"`
	FromStatus  string     `gorm:"size:100;not null" json:"from_status"`
	ToStatus    string     `gorm:"size:100;not null" json:"to_status"`
	RequestedBy uuid.UUID  `gorm:"type:uuid;not null" json:"requested_by"`
	ReviewedBy  *uuid.UUID `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	Status      string     `gorm:"size:50;default:'PENDING'" json:"status"` // PENDING | APPROVED | REJECTED
	Comment     string     `gorm:"type:text" json:"comment,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
