package report

import (
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Enums
// ---------------------------------------------------------------------------

type PeriodType string

const (
	PeriodWeekly    PeriodType = "WEEKLY"
	PeriodMonthly   PeriodType = "MONTHLY"
	PeriodQuarterly PeriodType = "QUARTERLY"
)

type ReportStatus string

const (
	ReportStatusDraft     ReportStatus = "DRAFT"
	ReportStatusPublished ReportStatus = "PUBLISHED"
	ReportStatusArchived  ReportStatus = "ARCHIVED"
)

// ---------------------------------------------------------------------------
// SnapshotMetrics — JSON blob stored in report_snapshots.metrics
// ---------------------------------------------------------------------------

type SnapshotMetrics struct {
	// Projects
	TotalProjects  int `json:"total_projects"`
	ActiveProjects int `json:"active_projects"`
	DoneProjects   int `json:"done_projects"`

	// Tasks
	TotalTasks     int     `json:"total_tasks"`
	DoneTasks      int     `json:"done_tasks"`
	OverdueTasks   int     `json:"overdue_tasks"`
	AvgProgressPct float64 `json:"avg_progress_pct"`

	// Milestones
	TotalMilestones   int `json:"total_milestones"`
	DoneMilestones    int `json:"done_milestones"`
	OverdueMilestones int `json:"overdue_milestones"`

	// Risks
	TotalRisks int `json:"total_risks"`
	OpenRisks  int `json:"open_risks"`
	HighRisks  int `json:"high_risks"`

	// Issues
	TotalIssues int `json:"total_issues"`
	OpenIssues  int `json:"open_issues"`

	// Budget (org-wide totals in IDR)
	TotalPlannedBudget float64 `json:"total_planned_budget"`
	TotalActualBudget  float64 `json:"total_actual_budget"`
	BudgetUsagePct     float64 `json:"budget_usage_pct"`

	// Corrective Actions
	TotalCorrectiveActions int `json:"total_corrective_actions"`
	OpenCorrectiveActions  int `json:"open_corrective_actions"`
}

// ---------------------------------------------------------------------------
// ReportSnapshot — domain model (maps to report_snapshots table)
// ---------------------------------------------------------------------------

type ReportSnapshot struct {
	ID               uuid.UUID       `json:"id"              gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrganizationID   uuid.UUID       `json:"organization_id" gorm:"type:uuid;not null"`
	PeriodType       PeriodType      `json:"period_type"     gorm:"type:report_period;not null"`
	PeriodLabel      string          `json:"period_label"    gorm:"not null"`
	PeriodStart      time.Time       `json:"period_start"    gorm:"type:date;not null"`
	PeriodEnd        time.Time       `json:"period_end"      gorm:"type:date;not null"`
	ProjectID        *uuid.UUID      `json:"project_id,omitempty" gorm:"type:uuid"`
	Metrics          SnapshotMetrics `json:"metrics"      gorm:"type:jsonb;serializer:json"`
	ExecutiveSummary *string         `json:"executive_summary,omitempty"`
	Status           ReportStatus    `json:"status"          gorm:"type:report_status;not null;default:'DRAFT'"`
	ExportFormat     *string         `json:"export_format,omitempty"`
	ExportURL        *string         `json:"export_url,omitempty"`
	CreatedBy        uuid.UUID       `json:"created_by"      gorm:"type:uuid;not null"`
	PublishedAt      *time.Time      `json:"published_at,omitempty"`
	PublishedBy      *uuid.UUID      `json:"published_by,omitempty" gorm:"type:uuid"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        *time.Time      `json:"deleted_at,omitempty"  gorm:"index"`

	// Preloaded relations
	CreatedByUser *ReportUser    `json:"created_by_user,omitempty" gorm:"foreignKey:CreatedBy;references:ID"`
	Project       *ReportProject `json:"project,omitempty"       gorm:"foreignKey:ProjectID;references:ID"`
}

func (ReportSnapshot) TableName() string { return "report_snapshots" }

// Lightweight user info for embedding in report response.
type ReportUser struct {
	ID   uuid.UUID `json:"id"   gorm:"primaryKey"`
	Name string    `json:"name"`
}

func (ReportUser) TableName() string { return "users" }

// Lightweight project info for embedding.
type ReportProject struct {
	ID   uuid.UUID `json:"id"   gorm:"primaryKey"`
	Name string    `json:"name"`
}

func (ReportProject) TableName() string { return "projects" }

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type ListReportFilter struct {
	PeriodType PeriodType   `form:"period_type"`
	Status     ReportStatus `form:"status"`
	ProjectID  string       `form:"project_id"`
	Page       int          `form:"page"`
	PageSize   int          `form:"page_size"`
}

type CreateReportRequest struct {
	PeriodType       PeriodType `json:"period_type"        binding:"required,oneof=WEEKLY MONTHLY QUARTERLY"`
	PeriodLabel      string     `json:"period_label"       binding:"required"`
	PeriodStart      string     `json:"period_start"       binding:"required"` // YYYY-MM-DD
	PeriodEnd        string     `json:"period_end"         binding:"required"` // YYYY-MM-DD
	ProjectID        *string    `json:"project_id"`
	ExecutiveSummary *string    `json:"executive_summary"`
}

type UpdateReportRequest struct {
	ExecutiveSummary *string `json:"executive_summary"`
}

type TransitionReportRequest struct {
	ToStatus ReportStatus `json:"to_status" binding:"required,oneof=PUBLISHED ARCHIVED DRAFT"`
}

// GenerateReportRequest triggers metric computation + snapshot creation.
type GenerateReportRequest struct {
	PeriodType  PeriodType `json:"period_type"  binding:"required,oneof=WEEKLY MONTHLY QUARTERLY"`
	PeriodLabel string     `json:"period_label" binding:"required"`
	PeriodStart string     `json:"period_start" binding:"required"`
	PeriodEnd   string     `json:"period_end"   binding:"required"`
	ProjectID   *string    `json:"project_id"`
}
