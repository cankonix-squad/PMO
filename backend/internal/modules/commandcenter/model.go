package commandcenter

import "github.com/google/uuid"

type Item struct {
	ID          string     `json:"id"`
	Kind        string     `json:"kind"`
	Title       string     `json:"title"`
	Message     string     `json:"message"`
	Severity    string     `json:"severity"`
	Status      string     `json:"status"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty"`
	ProjectName string     `json:"project_name,omitempty"`
	PICUserID   *uuid.UUID `json:"pic_user_id,omitempty"`
	DueAt       string     `json:"due_at,omitempty"`
	SourceType  string     `json:"source_type,omitempty"`
	SourceID    string     `json:"source_id,omitempty"`
	AgingDays   int        `json:"aging_days,omitempty"`
}

type Summary struct {
	AsOf        string `json:"as_of"`
	Alerts      []Item `json:"alerts"`
	Actions     []Item `json:"actions"`
	Validations []Item `json:"validations"`
	Watchlist   []Item `json:"watchlist"`
	Escalations []Item `json:"escalations"`
	Decisions   []Item `json:"decisions"`
}
