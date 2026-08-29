package health

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

var Dimensions = []string{"schedule", "physical", "financial", "contract", "risk", "issue", "quality", "procurement"}

type Formula struct {
	ID              uuid.UUID       `json:"id"`
	OrganizationID  uuid.UUID       `json:"organization_id"`
	Version         int             `json:"version"`
	Status          string          `json:"status"`
	Weights         json.RawMessage `json:"weights"`
	Thresholds      json.RawMessage `json:"thresholds"`
	MissingDataRule string          `json:"missing_data_rule"`
	EffectiveFrom   *time.Time      `json:"effective_from,omitempty"`
	EffectiveTo     *time.Time      `json:"effective_to,omitempty"`
	ApprovedBy      *uuid.UUID      `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time      `json:"approved_at,omitempty"`
	CreatedBy       *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func (Formula) TableName() string { return "health_formulas" }

type Snapshot struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	ProjectID      uuid.UUID       `json:"project_id"`
	FormulaID      uuid.UUID       `json:"formula_id"`
	PeriodYear     int             `json:"period_year"`
	PeriodMonth    int             `json:"period_month"`
	Score          float64         `json:"score"`
	HealthClass    string          `json:"health_class"`
	Components     json.RawMessage `json:"components"`
	Explanation    string          `json:"explanation"`
	CalculatedAt   time.Time       `json:"calculated_at"`
	CreatedAt      time.Time       `json:"created_at"`
}

func (Snapshot) TableName() string { return "health_snapshots" }

type CreateFormulaRequest struct {
	Version         int                `json:"version"`
	Weights         map[string]float64 `json:"weights"`
	Thresholds      map[string]float64 `json:"thresholds"`
	MissingDataRule string             `json:"missing_data_rule"`
}
type FormulaTransitionRequest struct {
	Status string `json:"status" binding:"required,oneof=APPROVED RETIRED"`
}
type CalculateRequest struct {
	FormulaID   string `json:"formula_id"`
	PeriodYear  int    `json:"period_year" binding:"required,min=2000,max=2100"`
	PeriodMonth int    `json:"period_month" binding:"required,min=1,max=12"`
}
