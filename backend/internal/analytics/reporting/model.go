package reporting

import (
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Enums
// ---------------------------------------------------------------------------

type ExportStatus string

const (
	ExportStatusPending    ExportStatus = "PENDING"
	ExportStatusProcessing ExportStatus = "PROCESSING"
	ExportStatusCompleted  ExportStatus = "COMPLETED"
	ExportStatusFailed     ExportStatus = "FAILED"
)

type ExportFormat string

const (
	ExportFormatPDF  ExportFormat = "PDF"
	ExportFormatXLSX ExportFormat = "XLSX"
	ExportFormatCSV  ExportFormat = "CSV"
)

// ---------------------------------------------------------------------------
// DB models
// ---------------------------------------------------------------------------

// ReportDefinition maps to report_definitions table.
type ReportDefinition struct {
	ID                uuid.UUID `json:"id"                 gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrganizationID    uuid.UUID `json:"organization_id"    gorm:"type:uuid;not null"`
	Name              string    `json:"name"               gorm:"not null"`
	Description       string    `json:"description"`
	Category          string    `json:"category"           gorm:"default:GENERAL"`
	DatasetKey        string    `json:"dataset_key"        gorm:"not null"`
	VisualizationType string    `json:"visualization_type" gorm:"default:TABLE"`
	Available         bool      `json:"available"          gorm:"default:true"`
	RequiresPowerBI   bool      `json:"requires_powerbi"   gorm:"default:false"`
	EmbedConfigured   bool      `json:"embed_configured"   gorm:"default:false"`
	SortOrder         int       `json:"sort_order"         gorm:"default:0"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (ReportDefinition) TableName() string { return "report_definitions" }

// ReportExportRequest maps to report_export_requests table.
type ReportExportRequest struct {
	ID             uuid.UUID    `json:"id"              gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrganizationID uuid.UUID    `json:"organization_id" gorm:"type:uuid;not null"`
	ReportID       *uuid.UUID   `json:"report_id,omitempty" gorm:"type:uuid"`
	DatasetKey     string       `json:"dataset_key"     gorm:"not null"`
	Format         ExportFormat `json:"format"          gorm:"type:export_format_type;not null;default:XLSX"`
	Status         ExportStatus `json:"status"          gorm:"type:export_status;not null;default:PENDING"`
	Parameters     []byte       `json:"parameters"      gorm:"type:jsonb;default:'{}'"`
	// Legacy field — kept for backward compat; new code uses storage_key + download endpoint
	FileURL      *string `json:"file_url,omitempty"`
	ErrorMessage *string `json:"error_message,omitempty"`
	// File metadata (populated by UAT-002 export generation)
	FileName    *string    `json:"file_name,omitempty"    gorm:"column:file_name"`
	StorageKey  *string    `json:"storage_key,omitempty"  gorm:"column:storage_key"`
	MimeType    *string    `json:"mime_type,omitempty"    gorm:"column:mime_type"`
	FileSize    *int64     `json:"file_size,omitempty"    gorm:"column:file_size"`
	GeneratedAt *time.Time `json:"generated_at,omitempty" gorm:"column:generated_at"`
	RequestedBy uuid.UUID  `json:"requested_by"    gorm:"type:uuid;not null"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (ReportExportRequest) TableName() string { return "report_export_requests" }

// ---------------------------------------------------------------------------
// Dataset response types
// ---------------------------------------------------------------------------

// DatasetFilter — common query params for all dataset endpoints.
type DatasetFilter struct {
	PeriodStart *time.Time `form:"period_start"`
	PeriodEnd   *time.Time `form:"period_end"`
	ProgramID   *string    `form:"program_id"`
	Status      *string    `form:"status"`
	Province    *string    `form:"province"`
}

// ExecutiveSummaryRow — one row in executive-summary dataset.
type ExecutiveSummaryRow struct {
	TotalProjects     int     `json:"total_projects"`
	ActiveProjects    int     `json:"active_projects"`
	CompletedProjects int     `json:"completed_projects"`
	OnHoldProjects    int     `json:"on_hold_projects"`
	AvgProgressPct    float64 `json:"avg_progress_pct"`
	TotalBudgetPlan   float64 `json:"total_budget_plan"`
	TotalBudgetActual float64 `json:"total_budget_actual"`
	BudgetUsagePct    float64 `json:"budget_usage_pct"`
	TotalRisks        int     `json:"total_risks"`
	OpenRisks         int     `json:"open_risks"`
	HighRisks         int     `json:"high_risks"`
	TotalIssues       int     `json:"total_issues"`
	OpenIssues        int     `json:"open_issues"`
	GreenHealth       int     `json:"green_health"`
	YellowHealth      int     `json:"yellow_health"`
	RedHealth         int     `json:"red_health"`
	CriticalHealth    int     `json:"critical_health"`
}

// ProjectPerformanceRow — one project row in project-performance dataset.
type ProjectPerformanceRow struct {
	ProjectID        uuid.UUID  `json:"project_id"`
	ProjectCode      string     `json:"project_code"`
	ProjectName      string     `json:"project_name"`
	Status           string     `json:"status"`
	ProgressPct      float64    `json:"progress_pct"`
	StartDate        *time.Time `json:"start_date,omitempty"`
	EndDate          *time.Time `json:"end_date,omitempty"`
	BudgetPlan       float64    `json:"budget_plan"`
	BudgetActual     float64    `json:"budget_actual"`
	BudgetUsagePct   float64    `json:"budget_usage_pct"`
	HealthClass      *string    `json:"health_class,omitempty"`
	Province         *string    `json:"province,omitempty"`
	PriorityScore    *float64   `json:"priority_score,omitempty"`
	PriorityCategory *string    `json:"priority_category,omitempty"`
}

// RiskIssueRow — aggregated risk/issue row per project.
type RiskIssueRow struct {
	ProjectID      uuid.UUID `json:"project_id"`
	ProjectCode    string    `json:"project_code"`
	ProjectName    string    `json:"project_name"`
	TotalRisks     int       `json:"total_risks"`
	OpenRisks      int       `json:"open_risks"`
	HighRisks      int       `json:"high_risks"`
	CriticalRisks  int       `json:"critical_risks"`
	TotalIssues    int       `json:"total_issues"`
	OpenIssues     int       `json:"open_issues"`
	HighIssues     int       `json:"high_issues"`
	CriticalIssues int       `json:"critical_issues"`
}

// BudgetRow — budget summary per project.
type BudgetRow struct {
	ProjectID    uuid.UUID `json:"project_id"`
	ProjectCode  string    `json:"project_code"`
	ProjectName  string    `json:"project_name"`
	Status       string    `json:"status"`
	BudgetPlan   float64   `json:"budget_plan"`
	BudgetActual float64   `json:"budget_actual"`
	Variance     float64   `json:"variance"`
	UsagePct     float64   `json:"usage_pct"`
}

// BenefitRow — benefit indicators per project.
type BenefitRow struct {
	ProjectID         uuid.UUID `json:"project_id"`
	ProjectCode       string    `json:"project_code"`
	ProjectName       string    `json:"project_name"`
	IndicatorID       uuid.UUID `json:"indicator_id"`
	IndicatorName     string    `json:"indicator_name"`
	Unit              string    `json:"unit"`
	Target            float64   `json:"target"`
	Actual            float64   `json:"actual"`
	AchievementPct    float64   `json:"achievement_pct"`
	AggregationMethod string    `json:"aggregation_method"`
}

// PriorityRow — priority score per project.
type PriorityRow struct {
	ProjectID    uuid.UUID `json:"project_id"`
	ProjectCode  string    `json:"project_code"`
	ProjectName  string    `json:"project_name"`
	TotalScore   float64   `json:"total_score"`
	Category     string    `json:"score_category"`
	CalculatedAt time.Time `json:"calculated_at"`
}

// ---------------------------------------------------------------------------
// Power BI config response
// ---------------------------------------------------------------------------

// PowerBIConfig — sanitised config returned to frontend.
// Actual credentials/tokens are NEVER exposed.
type PowerBIConfig struct {
	Configured  bool   `json:"configured"`
	WorkspaceID string `json:"workspace_id,omitempty"` // safe to expose
	ReportID    string `json:"report_id,omitempty"`    // safe to expose
	TenantID    string `json:"tenant_id,omitempty"`    // safe to expose
	EmbedURL    string `json:"embed_url,omitempty"`    // safe to expose
	// TokenEndpoint is provided only to indicate auth flow, not the actual token
	AuthMethod string `json:"auth_method,omitempty"` // "service_principal" | "master_user"
}

// ---------------------------------------------------------------------------
// Export request DTO
// ---------------------------------------------------------------------------

type CreateExportRequestInput struct {
	DatasetKey string                 `json:"dataset_key" binding:"required"`
	ReportID   *uuid.UUID             `json:"report_id,omitempty"`
	Format     ExportFormat           `json:"format"      binding:"required,oneof=PDF XLSX CSV"`
	Parameters map[string]interface{} `json:"parameters"`
}
