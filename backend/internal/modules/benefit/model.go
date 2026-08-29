package benefit

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Indicator struct {
	ID                uuid.UUID      `json:"id"`
	OrganizationID    uuid.UUID      `json:"organization_id"`
	ProjectID         *uuid.UUID     `json:"project_id,omitempty"`
	Name              string         `json:"name"`
	Unit              string         `json:"unit"`
	AggregationMethod string         `json:"aggregation_method"`
	OwnerID           *uuid.UUID     `json:"owner_id,omitempty"`
	Source            string         `json:"source,omitempty"`
	Description       string         `json:"description,omitempty"`
	CreatedBy         uuid.UUID      `json:"created_by"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Indicator) TableName() string { return "benefit_indicators" }

type Measurement struct {
	ID               uuid.UUID      `json:"id"`
	OrganizationID   uuid.UUID      `json:"organization_id"`
	IndicatorID      uuid.UUID      `json:"indicator_id"`
	PeriodYear       int            `json:"period_year"`
	PeriodMonth      int            `json:"period_month"`
	Baseline         float64        `json:"baseline"`
	Target           float64        `json:"target"`
	Actual           float64        `json:"actual"`
	Source           string         `json:"source,omitempty"`
	ValidationStatus string         `json:"validation_status"`
	CreatedBy        uuid.UUID      `json:"created_by"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Measurement) TableName() string { return "benefit_measurements" }

type CreateIndicatorRequest struct {
	ProjectID         *uuid.UUID `json:"project_id"`
	Name              string     `json:"name" binding:"required,max=200"`
	Unit              string     `json:"unit" binding:"required,max=50"`
	AggregationMethod string     `json:"aggregation_method" binding:"required,oneof=SUM AVERAGE LATEST"`
	OwnerID           *uuid.UUID `json:"owner_id"`
	Source            string     `json:"source"`
	Description       string     `json:"description"`
}
type CreateMeasurementRequest struct {
	PeriodYear       int     `json:"period_year" binding:"required,min=2000,max=2100"`
	PeriodMonth      int     `json:"period_month" binding:"required,min=1,max=12"`
	Baseline         float64 `json:"baseline"`
	Target           float64 `json:"target"`
	Actual           float64 `json:"actual"`
	Source           string  `json:"source"`
	ValidationStatus string  `json:"validation_status" binding:"omitempty,oneof=DRAFT SUBMITTED VALID REJECTED STALE"`
}

type UpdateIndicatorRequest struct {
	ProjectID         *uuid.UUID `json:"project_id"`
	Name              string     `json:"name" binding:"omitempty,max=200"`
	Unit              string     `json:"unit" binding:"omitempty,max=50"`
	AggregationMethod string     `json:"aggregation_method" binding:"omitempty,oneof=SUM AVERAGE LATEST"`
	OwnerID           *uuid.UUID `json:"owner_id"`
	Source            string     `json:"source"`
	Description       string     `json:"description"`
}

type UpdateMeasurementRequest struct {
	PeriodYear       *int     `json:"period_year" binding:"omitempty,min=2000,max=2100"`
	PeriodMonth      *int     `json:"period_month" binding:"omitempty,min=1,max=12"`
	Baseline         *float64 `json:"baseline"`
	Target           *float64 `json:"target"`
	Actual           *float64 `json:"actual"`
	Source           string   `json:"source"`
	ValidationStatus string   `json:"validation_status" binding:"omitempty,oneof=DRAFT SUBMITTED VALID REJECTED STALE"`
}

type AggregateResult struct {
	Indicator Indicator `json:"indicator"`
	Count     int       `json:"count"`
	Value     *float64  `json:"value"`
}

type SummaryItem struct {
	Unit              string  `json:"unit"`
	AggregationMethod string  `json:"aggregation_method"`
	Count             int     `json:"count"`
	Value             float64 `json:"value"`
}
