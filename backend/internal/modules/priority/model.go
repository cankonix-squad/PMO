package priority

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

// ComponentKeys defines all scoring dimensions for priority.
var ComponentKeys = []string{
	"health_score",
	"risk_score",
	"issue_severity",
	"budget_usage",
	"schedule_variance",
	"corrective_action_overdue",
	"benefit_indicator",
}

// Formula statuses
const (
	FormulaStatusDraft    = "DRAFT"
	FormulaStatusActive   = "ACTIVE"
	FormulaStatusArchived = "ARCHIVED"
)

// Score categories
const (
	CategoryLow      = "LOW"
	CategoryMedium   = "MEDIUM"
	CategoryHigh     = "HIGH"
	CategoryCritical = "CRITICAL"
)

// MissingDataRule options
const (
	MissingPenalize = "PENALIZE"
	MissingExclude  = "EXCLUDE"
	MissingNeutral  = "NEUTRAL"
)

// ------------------------------------------------------------------
// DB models
// ------------------------------------------------------------------

// Formula represents a versioned priority scoring formula.
type Formula struct {
	ID                 uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID     uuid.UUID       `gorm:"type:uuid;not null"                              json:"organization_id"`
	Name               string          `gorm:"size:200;not null"                               json:"name"`
	Description        string          `gorm:"type:text"                                       json:"description,omitempty"`
	Version            int             `gorm:"not null;default:1"                              json:"version"`
	Status             string          `gorm:"size:20;not null;default:'DRAFT'"                json:"status"`
	MissingDataRule    string          `gorm:"size:20;not null;default:'PENALIZE'"             json:"missing_data_rule"`
	CategoryThresholds json.RawMessage `gorm:"type:jsonb;not null;default:'{}'"                json:"category_thresholds"`
	ActivatedBy        *uuid.UUID      `gorm:"type:uuid"                                       json:"activated_by,omitempty"`
	ActivatedAt        *time.Time      `                                                       json:"activated_at,omitempty"`
	CreatedBy          *uuid.UUID      `gorm:"type:uuid"                                       json:"created_by,omitempty"`
	CreatedAt          time.Time       `                                                       json:"created_at"`
	UpdatedAt          time.Time       `                                                       json:"updated_at"`
	DeletedAt          *time.Time      `gorm:"index"                                           json:"-"`

	Components []FormulaComponent `gorm:"foreignKey:FormulaID" json:"components,omitempty"`
}

func (Formula) TableName() string { return "priority_formulas" }

// FormulaComponent holds weight for one scoring dimension.
type FormulaComponent struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FormulaID      uuid.UUID `gorm:"type:uuid;not null"                              json:"formula_id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null"                              json:"organization_id"`
	ComponentKey   string    `gorm:"size:100;not null"                               json:"component_key"`
	Label          string    `gorm:"size:200;not null"                               json:"label"`
	Weight         float64   `gorm:"type:numeric(5,4);not null;default:0"            json:"weight"`
	SortOrder      int       `gorm:"not null;default:0"                              json:"sort_order"`
	CreatedAt      time.Time `                                                       json:"created_at"`
	UpdatedAt      time.Time `                                                       json:"updated_at"`
}

func (FormulaComponent) TableName() string { return "priority_formula_components" }

// Score is a priority score snapshot for one project.
type Score struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID    uuid.UUID  `gorm:"type:uuid;not null"                              json:"organization_id"`
	ProjectID         uuid.UUID  `gorm:"type:uuid;not null"                              json:"project_id"`
	FormulaID         uuid.UUID  `gorm:"type:uuid;not null"                              json:"formula_id"`
	FormulaVersion    int        `gorm:"not null"                                        json:"formula_version"`
	TotalScore        float64    `gorm:"type:numeric(6,3);not null;default:0"            json:"total_score"`
	ScoreCategory     string     `gorm:"size:20;not null;default:'LOW'"                  json:"score_category"`
	RankInOrg         *int       `                                                       json:"rank_in_org,omitempty"`
	MissingComponents int        `gorm:"not null;default:0"                              json:"missing_components"`
	CalculatedAt      time.Time  `                                                       json:"calculated_at"`
	CalculatedBy      *uuid.UUID `gorm:"type:uuid"                                       json:"calculated_by,omitempty"`
	CreatedAt         time.Time  `                                                       json:"created_at"`

	Components []ScoreComponent `gorm:"foreignKey:ScoreID" json:"components,omitempty"`
}

func (Score) TableName() string { return "project_priority_scores" }

// ScoreComponent holds explainability data for one dimension of a Score.
type ScoreComponent struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ScoreID         uuid.UUID `gorm:"type:uuid;not null"                              json:"score_id"`
	OrganizationID  uuid.UUID `gorm:"type:uuid;not null"                              json:"organization_id"`
	ProjectID       uuid.UUID `gorm:"type:uuid;not null"                              json:"project_id"`
	ComponentKey    string    `gorm:"size:100;not null"                               json:"component_key"`
	Label           string    `gorm:"size:200;not null"                               json:"label"`
	RawValue        *float64  `gorm:"type:numeric(10,4)"                              json:"raw_value"`
	RawUnit         string    `gorm:"size:50"                                         json:"raw_unit,omitempty"`
	NormalizedScore *float64  `gorm:"type:numeric(6,3)"                               json:"normalized_score"`
	Weight          float64   `gorm:"type:numeric(5,4);not null;default:0"            json:"weight"`
	WeightedScore   *float64  `gorm:"type:numeric(6,3)"                               json:"weighted_score"`
	Available       bool      `gorm:"not null;default:true"                           json:"available"`
	Note            string    `gorm:"type:text"                                       json:"note,omitempty"`
	CreatedAt       time.Time `                                                       json:"created_at"`
}

func (ScoreComponent) TableName() string { return "project_priority_score_components" }

// ------------------------------------------------------------------
// Request / Response DTOs
// ------------------------------------------------------------------

type ComponentWeightInput struct {
	ComponentKey string  `json:"component_key" binding:"required"`
	Label        string  `json:"label"         binding:"required,max=200"`
	Weight       float64 `json:"weight"        binding:"required,min=0,max=1"`
	SortOrder    int     `json:"sort_order"`
}

type CategoryThreshold struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type CreateFormulaRequest struct {
	Name               string                       `json:"name"                binding:"required,max=200"`
	Description        string                       `json:"description"`
	MissingDataRule    string                       `json:"missing_data_rule"   binding:"omitempty,oneof=PENALIZE EXCLUDE NEUTRAL"`
	CategoryThresholds map[string]CategoryThreshold `json:"category_thresholds"`
	Components         []ComponentWeightInput       `json:"components"          binding:"required,min=1"`
}

type UpdateFormulaRequest struct {
	Name               string                       `json:"name"                binding:"omitempty,max=200"`
	Description        string                       `json:"description"`
	MissingDataRule    string                       `json:"missing_data_rule"   binding:"omitempty,oneof=PENALIZE EXCLUDE NEUTRAL"`
	CategoryThresholds map[string]CategoryThreshold `json:"category_thresholds"`
	Components         []ComponentWeightInput       `json:"components"`
}

type CalculateRequest struct {
	FormulaID string `json:"formula_id"` // optional; defaults to ACTIVE formula
}

type BatchCalculateRequest struct {
	FormulaID  string      `json:"formula_id"`  // optional
	ProjectIDs []uuid.UUID `json:"project_ids"` // optional; empty = all active projects
}

// ProjectScoreSummary is used in the ranking list endpoint.
type ProjectScoreSummary struct {
	ScoreID           uuid.UUID `json:"score_id"`
	ProjectID         uuid.UUID `json:"project_id"`
	ProjectName       string    `json:"project_name"`
	ProjectStatus     string    `json:"project_status"`
	FormulaID         uuid.UUID `json:"formula_id"`
	FormulaName       string    `json:"formula_name"`
	FormulaVersion    int       `json:"formula_version"`
	TotalScore        float64   `json:"total_score"`
	ScoreCategory     string    `json:"score_category"`
	RankInOrg         *int      `json:"rank_in_org"`
	MissingComponents int       `json:"missing_components"`
	CalculatedAt      time.Time `json:"calculated_at"`
}

// RankingResponse wraps the ranked list with summary counts.
type RankingResponse struct {
	Counts   map[string]int        `json:"counts"`
	Projects []ProjectScoreSummary `json:"projects"`
}

// BatchCalculateResponse summarises a batch calculation run.
type BatchCalculateResponse struct {
	Calculated     int    `json:"calculated"`
	Skipped        int    `json:"skipped"`
	FormulaID      string `json:"formula_id"`
	FormulaVersion int    `json:"formula_version"`
}
