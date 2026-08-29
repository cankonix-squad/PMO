package project

import (
	"time"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/shared/types"
	"gorm.io/gorm"
)

// Project is the core project entity.
type Project struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID       `gorm:"type:uuid;not null" json:"organization_id"`
	OrgUnitID      *uuid.UUID      `gorm:"type:uuid" json:"org_unit_id,omitempty"`
	ProgramID      *uuid.UUID      `gorm:"type:uuid" json:"program_id,omitempty"`
	SectorID       *uuid.UUID      `gorm:"type:uuid" json:"sector_id,omitempty"`
	RegionID       *uuid.UUID      `gorm:"type:uuid" json:"region_id,omitempty"`
	RiverBasinID   *uuid.UUID      `gorm:"type:uuid" json:"river_basin_id,omitempty"`
	Code           string          `gorm:"size:100;not null" json:"code"`
	Name           string          `gorm:"size:500;not null" json:"name"`
	Description    string          `gorm:"type:text" json:"description,omitempty"`
	Objectives     string          `gorm:"type:text" json:"objectives,omitempty"`
	Status         string          `gorm:"size:50;default:'DRAFT'" json:"status"`
	Priority       string          `gorm:"size:50;default:'MEDIUM'" json:"priority"`
	Category       string          `gorm:"size:100" json:"category,omitempty"`
	StartDate      *types.FlexTime `json:"start_date,omitempty"`
	EndDate        *types.FlexTime `json:"end_date,omitempty"`
	ActualEndDate  *types.FlexTime `json:"actual_end_date,omitempty"`
	BudgetTotal    float64         `gorm:"type:decimal(20,2);default:0" json:"budget_total"`
	Currency       string          `gorm:"size:10;default:'IDR'" json:"currency"`
	ProgressPct    float64         `gorm:"type:decimal(5,2);default:0" json:"progress_pct"`
	ManagerID      *uuid.UUID      `gorm:"type:uuid" json:"manager_id,omitempty"`
	CreatedBy      uuid.UUID       `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`

	// Computed via batch lookup after query; not a DB column
	OrgUnitName string `gorm:"-" json:"org_unit_name,omitempty"`

	// Preloaded associations
	Milestones []Milestone  `gorm:"foreignKey:ProjectID" json:"milestones,omitempty"`
	Team       []TeamMember `gorm:"foreignKey:ProjectID" json:"team,omitempty"`
}

// TeamMember links a user to a project with a specific role.
type TeamMember struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectID uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	Role      string         `gorm:"size:100;default:'MEMBER'" json:"role"` // PROJECT_MANAGER | PROJECT_OFFICER | MEMBER | VIEWER
	JoinedAt  time.Time      `json:"joined_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TeamMember) TableName() string {
	return "project_teams"
}

// Milestone represents a key checkpoint within a project.
type Milestone struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID       `gorm:"type:uuid;not null" json:"organization_id"`
	ProjectID      uuid.UUID       `gorm:"type:uuid;not null;index" json:"project_id"`
	Title          string          `gorm:"size:500;not null" json:"title"`
	Description    string          `gorm:"type:text" json:"description,omitempty"`
	DueDate        *types.FlexTime `json:"due_date,omitempty"`
	Status         string          `gorm:"size:50;default:'PENDING'" json:"status"`
	ProgressPct    float64         `gorm:"type:decimal(5,2);default:0" json:"progress_pct"`
	CreatedBy      uuid.UUID       `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
}

// Task represents a work item in the WBS (Work Breakdown Structure).
// Self-referencing via ParentID to support up to 3 levels.
type Task struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID       `gorm:"type:uuid;not null" json:"organization_id"`
	ProjectID      uuid.UUID       `gorm:"type:uuid;not null;index" json:"project_id"`
	MilestoneID    *uuid.UUID      `gorm:"type:uuid" json:"milestone_id,omitempty"`
	ParentID       *uuid.UUID      `gorm:"type:uuid" json:"parent_id,omitempty"`
	WBSCode        string          `gorm:"size:50" json:"wbs_code,omitempty"`
	Title          string          `gorm:"size:500;not null" json:"title"`
	Description    string          `gorm:"type:text" json:"description,omitempty"`
	Status         string          `gorm:"size:50;default:'TODO'" json:"status"`
	Priority       string          `gorm:"size:50;default:'MEDIUM'" json:"priority"`
	Type           string          `gorm:"size:50;default:'TASK'" json:"type"` // TASK | BUG | FEATURE | RESEARCH
	StartDate      *types.FlexTime `json:"start_date,omitempty"`
	DueDate        *types.FlexTime `json:"due_date,omitempty"`
	EstHours       float64         `gorm:"type:decimal(8,2);default:0" json:"est_hours"`
	ActualHours    float64         `gorm:"type:decimal(8,2);default:0" json:"actual_hours"`
	ProgressPct    float64         `gorm:"type:decimal(5,2);default:0" json:"progress_pct"`
	CreatedBy      uuid.UUID       `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`

	Subtasks    []Task           `gorm:"foreignKey:ParentID" json:"subtasks,omitempty"`
	Assignments []TaskAssignment `gorm:"foreignKey:TaskID" json:"assignments,omitempty"`
}

// TaskAssignment links a user to a task.
type TaskAssignment struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TaskID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"task_id"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	IsLead     bool           `gorm:"default:false" json:"is_lead"`
	AssignedAt time.Time      `json:"assigned_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// Issue tracks a problem or blocker within a project.
type Issue struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID       `gorm:"type:uuid;not null" json:"organization_id"`
	ProjectID      uuid.UUID       `gorm:"type:uuid;not null;index" json:"project_id"`
	TaskID         *uuid.UUID      `gorm:"type:uuid" json:"task_id,omitempty"`
	Title          string          `gorm:"size:500;not null" json:"title"`
	Description    string          `gorm:"type:text" json:"description,omitempty"`
	Status         string          `gorm:"size:50;default:'OPEN'" json:"status"`
	Severity       string          `gorm:"size:50;default:'MEDIUM'" json:"severity"` // LOW | MEDIUM | HIGH | CRITICAL
	Escalation     string          `gorm:"size:50;default:'NONE'" json:"escalation"` // NONE | PROJECT_MANAGER | PROGRAM_MANAGER | EXECUTIVE
	ReportedBy     uuid.UUID       `gorm:"type:uuid;not null" json:"reported_by"`
	AssignedTo     *uuid.UUID      `gorm:"type:uuid" json:"assigned_to,omitempty"`
	DueDate        *types.FlexTime `json:"due_date,omitempty"`
	Resolution     string          `gorm:"type:text" json:"resolution,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
}

// Risk tracks a potential threat to the project.
type Risk struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID       `gorm:"type:uuid;not null" json:"organization_id"`
	ProjectID      uuid.UUID       `gorm:"type:uuid;not null;index" json:"project_id"`
	Title          string          `gorm:"size:500;not null" json:"title"`
	Description    string          `gorm:"type:text" json:"description,omitempty"`
	Status         string          `gorm:"size:50;default:'IDENTIFIED'" json:"status"` // IDENTIFIED | ASSESSED | MITIGATED | ACCEPTED | ESCALATED | CLOSED
	Probability    int             `gorm:"not null;default:3" json:"probability"`      // 1-5
	Impact         int             `gorm:"not null;default:3" json:"impact"`           // 1-5
	RiskScore      int             `gorm:"not null;default:9" json:"risk_score"`       // probability × impact
	Severity       string          `gorm:"size:20;default:'MEDIUM'" json:"severity"`   // LOW | MEDIUM | HIGH | CRITICAL
	Mitigation     string          `gorm:"type:text" json:"mitigation,omitempty"`
	OwnedBy        *uuid.UUID      `gorm:"type:uuid" json:"owned_by,omitempty"`
	DueDate        *types.FlexTime `json:"due_date,omitempty"`
	CreatedBy      uuid.UUID       `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
}

// ProjectBudget tracks financial line items for a project.
type ProjectBudget struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	Category    string         `gorm:"size:200;not null" json:"category"`
	Description string         `gorm:"type:text" json:"description,omitempty"`
	Planned     float64        `gorm:"type:decimal(20,2);default:0" json:"planned"`
	Actual      float64        `gorm:"type:decimal(20,2);default:0" json:"actual"`
	Currency    string         `gorm:"size:10;default:'IDR'" json:"currency"`
	CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// VendorType classifies a party as a provider (VENDOR) or a supervisory
// consultant (CONSULTANT).
type VendorType string

const (
	VendorTypeVendor     VendorType = "VENDOR"
	VendorTypeConsultant VendorType = "CONSULTANT"
)

// Vendor is a tenant-scoped master record for a business party: either a
// construction provider (penyedia) or a supervisory consultant (konsultan
// supervisi).
type Vendor struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name           string         `gorm:"size:500;not null" json:"name"`
	Type           string         `gorm:"size:50;not null;default:'VENDOR'" json:"type"` // VENDOR | CONSULTANT
	LegalName      string         `gorm:"size:500" json:"legal_name,omitempty"`
	TaxID          string         `gorm:"size:100" json:"tax_id,omitempty"`
	ContactPerson  string         `gorm:"size:200" json:"contact_person,omitempty"`
	Email          string         `gorm:"size:200" json:"email,omitempty"`
	Phone          string         `gorm:"size:50" json:"phone,omitempty"`
	Address        string         `gorm:"type:text" json:"address,omitempty"`
	IsActive       bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedBy      uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Vendor) TableName() string {
	return "vendors"
}

// ContractStatus classifies a project contract lifecycle.
type ContractStatus string

const (
	ContractStatusDraft      ContractStatus = "DRAFT"
	ContractStatusActive     ContractStatus = "ACTIVE"
	ContractStatusAmended    ContractStatus = "AMENDED"
	ContractStatusCompleted  ContractStatus = "COMPLETED"
	ContractStatusTerminated ContractStatus = "TERMINATED"
)

// Contract records a project's contract, linked to a project and an
// organization. vendor_id is required (penyedia); consultant_id is optional
// (konsultan supervisi).
type Contract struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID       `gorm:"type:uuid;not null;index" json:"organization_id"`
	ProjectID      uuid.UUID       `gorm:"type:uuid;not null;index" json:"project_id"`
	ContractNumber string          `gorm:"size:200;not null" json:"contract_number"`
	Title          string          `gorm:"size:500;not null" json:"title"`
	VendorID       uuid.UUID       `gorm:"type:uuid;not null" json:"vendor_id"`
	ConsultantID   *uuid.UUID      `gorm:"type:uuid" json:"consultant_id,omitempty"`
	ContractValue  float64         `gorm:"type:decimal(20,2);not null;default:0" json:"contract_value"`
	Currency       string          `gorm:"size:10;not null;default:'IDR'" json:"currency"`
	SignedDate     *types.FlexTime `json:"signed_date,omitempty"`
	StartDate      *types.FlexTime `json:"start_date,omitempty"`
	EndDate        *types.FlexTime `json:"end_date,omitempty"`
	Status         string          `gorm:"size:50;not null;default:'DRAFT'" json:"status"` // DRAFT | ACTIVE | AMENDED | COMPLETED | TERMINATED
	ScopeOfWork    string          `gorm:"type:text" json:"scope_of_work,omitempty"`
	CreatedBy      uuid.UUID       `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`

	// Preloaded associations (vendor display names for the response).
	Vendor     *Vendor `gorm:"foreignKey:VendorID" json:"vendor,omitempty"`
	Consultant *Vendor `gorm:"foreignKey:ConsultantID" json:"consultant,omitempty"`
}

func (Contract) TableName() string {
	return "contracts"
}

// ProjectDocument links an uploaded file to a project. FileURL stores the
// RELATIVE storage key (e.g. documents/{orgID}/{projectID}/{documentID}/file)
// — never an absolute local filesystem path. The file itself lives on local
// storage (P1-005) and is planned to move to MinIO/S3-compatible object
// storage later.
type ProjectDocument struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	Name       string         `gorm:"size:500;not null" json:"name"`
	Category   string         `gorm:"size:100" json:"category,omitempty"`
	Version    string         `gorm:"size:100" json:"version,omitempty"`
	FileURL    string         `gorm:"type:text;not null" json:"file_url"` // relative storage key
	FileSize   int64          `json:"file_size"`
	MimeType   string         `gorm:"size:200" json:"mime_type,omitempty"`
	UploadedBy uuid.UUID      `gorm:"type:uuid;not null" json:"uploaded_by"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProjectDocument) TableName() string {
	return "project_documents"
}

// ProgressHistory records a snapshot of project progress over time.
type ProgressHistory struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectID   uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	ProgressPct float64   `gorm:"type:decimal(5,2)" json:"progress_pct"`
	Notes       string    `gorm:"type:text" json:"notes,omitempty"`
	RecordedBy  uuid.UUID `gorm:"type:uuid;not null" json:"recorded_by"`
	RecordedAt  time.Time `json:"recorded_at"`
}

func (ProgressHistory) TableName() string {
	return "project_progress_history"
}

// --- DTOs ---

type CreateProjectRequest struct {
	OrgUnitID   *uuid.UUID      `json:"org_unit_id"`
	Code        string          `json:"code" binding:"required,max=100"`
	Name        string          `json:"name" binding:"required,max=500"`
	Description string          `json:"description"`
	Objectives  string          `json:"objectives"`
	Priority    string          `json:"priority" binding:"max=50"`
	Category    string          `json:"category" binding:"max=100"`
	StartDate   *types.FlexTime `json:"start_date"`
	EndDate     *types.FlexTime `json:"end_date"`
	BudgetTotal float64         `json:"budget_total"`
	Currency    string          `json:"currency" binding:"max=10"`
	ManagerID   *uuid.UUID      `json:"manager_id"`
}

type UpdateProjectRequest struct {
	OrgUnitID   *uuid.UUID      `json:"org_unit_id"`
	Name        string          `json:"name" binding:"max=500"`
	Description string          `json:"description"`
	Objectives  string          `json:"objectives"`
	Status      string          `json:"status" binding:"max=50"`
	Priority    string          `json:"priority" binding:"max=50"`
	Category    string          `json:"category" binding:"max=100"`
	StartDate   *types.FlexTime `json:"start_date"`
	EndDate     *types.FlexTime `json:"end_date"`
	BudgetTotal *float64        `json:"budget_total"`
	Currency    string          `json:"currency" binding:"max=10"`
	ManagerID   *uuid.UUID      `json:"manager_id"`
	ProgressPct *float64        `json:"progress_pct"`
}

type TransitionRequest struct {
	ToStatus string `json:"to_status" binding:"required"`
	Comment  string `json:"comment"`
}

type ProjectListFilter struct {
	OrganizationID uuid.UUID
	OrgUnitID      *uuid.UUID
	Status         string
	Priority       string
	ManagerID      *uuid.UUID
	Search         string
	Page           int
	PageSize       int
}

// TaskListFilter carries query params for listing tasks.
type TaskListFilter struct {
	OrganizationID uuid.UUID
	ProjectID      uuid.UUID
	MilestoneID    *uuid.UUID
	Status         string
	AssignedTo     *uuid.UUID
	Search         string
	Page           int
	PageSize       int
}

// CreateTaskRequest is the DTO for task creation.
type CreateTaskRequest struct {
	MilestoneID *uuid.UUID      `json:"milestone_id"`
	ParentID    *uuid.UUID      `json:"parent_id"`
	WBSCode     string          `json:"wbs_code" binding:"max=50"`
	Title       string          `json:"title" binding:"required,max=500"`
	Description string          `json:"description"`
	Priority    string          `json:"priority" binding:"max=50"`
	Type        string          `json:"type" binding:"max=50"`
	StartDate   *types.FlexTime `json:"start_date"`
	DueDate     *types.FlexTime `json:"due_date"`
	EstHours    float64         `json:"est_hours"`
}

// UpdateTaskRequest is the DTO for task updates.
type UpdateTaskRequest struct {
	MilestoneID *uuid.UUID      `json:"milestone_id"`
	WBSCode     string          `json:"wbs_code" binding:"max=50"`
	Title       string          `json:"title" binding:"max=500"`
	Description string          `json:"description"`
	Status      string          `json:"status" binding:"max=50"`
	Priority    string          `json:"priority" binding:"max=50"`
	Type        string          `json:"type" binding:"max=50"`
	StartDate   *types.FlexTime `json:"start_date"`
	DueDate     *types.FlexTime `json:"due_date"`
	EstHours    float64         `json:"est_hours"`
	ActualHours *float64        `json:"actual_hours"`
	ProgressPct *float64        `json:"progress_pct"`
}

// CreateMilestoneRequest is the DTO for milestone creation.
type CreateMilestoneRequest struct {
	Title       string          `json:"title" binding:"required,max=500"`
	Description string          `json:"description"`
	DueDate     *types.FlexTime `json:"due_date"`
}

// UpdateMilestoneRequest is the DTO for milestone updates.
type UpdateMilestoneRequest struct {
	Title       string          `json:"title" binding:"max=500"`
	Description string          `json:"description"`
	Status      string          `json:"status" binding:"max=50"`
	ProgressPct *float64        `json:"progress_pct"`
	DueDate     *types.FlexTime `json:"due_date"`
}

// IssueListFilter carries query params for listing issues.
type IssueListFilter struct {
	OrganizationID uuid.UUID
	ProjectID      uuid.UUID
	Status         string
	Severity       string
	AssignedTo     *uuid.UUID
	Search         string
	Page           int
	PageSize       int
}

// CreateIssueRequest is the DTO for issue creation.
type CreateIssueRequest struct {
	TaskID      *uuid.UUID      `json:"task_id"`
	Title       string          `json:"title" binding:"required,max=500"`
	Description string          `json:"description"`
	Severity    string          `json:"severity" binding:"max=50"`
	Escalation  string          `json:"escalation" binding:"max=50"`
	AssignedTo  *uuid.UUID      `json:"assigned_to"`
	DueDate     *types.FlexTime `json:"due_date"`
	Resolution  string          `json:"resolution"`
}

// UpdateIssueRequest is the DTO for issue updates.
type UpdateIssueRequest struct {
	TaskID      *uuid.UUID      `json:"task_id"`
	Title       string          `json:"title" binding:"max=500"`
	Description string          `json:"description"`
	Status      string          `json:"status" binding:"max=50"`
	Severity    string          `json:"severity" binding:"max=50"`
	Escalation  string          `json:"escalation" binding:"max=50"`
	AssignedTo  *uuid.UUID      `json:"assigned_to"`
	DueDate     *types.FlexTime `json:"due_date"`
	Resolution  string          `json:"resolution"`
}

// RiskListFilter carries query params for listing risks.
type RiskListFilter struct {
	OrganizationID uuid.UUID
	ProjectID      uuid.UUID
	Status         string
	Severity       string
	OwnedBy        *uuid.UUID
	Search         string
	Page           int
	PageSize       int
}

// CreateRiskRequest is the DTO for risk creation.
type CreateRiskRequest struct {
	Title       string          `json:"title" binding:"required,max=500"`
	Description string          `json:"description"`
	Probability int             `json:"probability"` // 1-5
	Impact      int             `json:"impact"`      // 1-5
	Mitigation  string          `json:"mitigation"`
	OwnedBy     *uuid.UUID      `json:"owned_by"`
	DueDate     *types.FlexTime `json:"due_date"`
}

// UpdateRiskRequest is the DTO for risk updates.
type UpdateRiskRequest struct {
	Title       string          `json:"title" binding:"max=500"`
	Description string          `json:"description"`
	Status      string          `json:"status" binding:"max=50"`
	Probability int             `json:"probability"` // 1-5
	Impact      int             `json:"impact"`      // 1-5
	Mitigation  string          `json:"mitigation"`
	OwnedBy     *uuid.UUID      `json:"owned_by"`
	DueDate     *types.FlexTime `json:"due_date"`
}

// BudgetStatus classifies a budget line's usage against its planned amount.
// Normal < 80% · Watch >= 80% · Risk >= 90% · Overrun >= 100%.
type BudgetStatus string

const (
	BudgetStatusNormal  BudgetStatus = "NORMAL"
	BudgetStatusWatch   BudgetStatus = "WATCH"
	BudgetStatusRisk    BudgetStatus = "RISK"
	BudgetStatusOverrun BudgetStatus = "OVERRUN"
)

// BudgetListFilter carries query params for listing budget lines.
type BudgetListFilter struct {
	OrganizationID uuid.UUID
	ProjectID      uuid.UUID
	Category       string
	Search         string
	Page           int
	PageSize       int
}

// CreateBudgetRequest is the DTO for budget line creation.
type CreateBudgetRequest struct {
	Category    string  `json:"category" binding:"required,max=200"`
	Description string  `json:"description"`
	Planned     float64 `json:"planned"`
	Actual      float64 `json:"actual"`
	Currency    string  `json:"currency" binding:"max=10"`
}

// UpdateBudgetRequest is the DTO for budget line updates.
type UpdateBudgetRequest struct {
	Category    string   `json:"category" binding:"max=200"`
	Description string   `json:"description"`
	Planned     *float64 `json:"planned"`
	Actual      *float64 `json:"actual"`
	Currency    string   `json:"currency" binding:"max=10"`
}

// BudgetLine is the computed response shape for a budget line. Derived fields
// (variance, usage_pct, status) are calculated by the backend from planned and
// actual so business numbers stay consistent regardless of client input.
type BudgetLine struct {
	ID          uuid.UUID    `json:"id"`
	ProjectID   uuid.UUID    `json:"project_id"`
	Category    string       `json:"category"`
	Description string       `json:"description,omitempty"`
	Planned     float64      `json:"planned"`
	Actual      float64      `json:"actual"`
	Currency    string       `json:"currency"`
	Variance    float64      `json:"variance"`  // planned - actual
	UsagePct    float64      `json:"usage_pct"` // actual / planned * 100 (0 when planned = 0)
	Status      BudgetStatus `json:"status"`    // NORMAL | WATCH | RISK | OVERRUN
	CreatedBy   uuid.UUID    `json:"created_by"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// VendorListFilter carries query params for listing vendors.
type VendorListFilter struct {
	OrganizationID uuid.UUID
	Type           string
	Search         string
	IsActive       *bool
	Page           int
	PageSize       int
}

// CreateVendorRequest is the DTO for vendor creation.
type CreateVendorRequest struct {
	Name          string `json:"name" binding:"required,max=500"`
	Type          string `json:"type" binding:"required,max=50"`
	LegalName     string `json:"legal_name" binding:"max=500"`
	TaxID         string `json:"tax_id" binding:"max=100"`
	ContactPerson string `json:"contact_person" binding:"max=200"`
	Email         string `json:"email" binding:"max=200"`
	Phone         string `json:"phone" binding:"max=50"`
	Address       string `json:"address"`
	IsActive      *bool  `json:"is_active"`
}

// UpdateVendorRequest is the DTO for vendor updates.
type UpdateVendorRequest struct {
	Name          string `json:"name" binding:"max=500"`
	Type          string `json:"type" binding:"max=50"`
	LegalName     string `json:"legal_name" binding:"max=500"`
	TaxID         string `json:"tax_id" binding:"max=100"`
	ContactPerson string `json:"contact_person" binding:"max=200"`
	Email         string `json:"email" binding:"max=200"`
	Phone         string `json:"phone" binding:"max=50"`
	Address       string `json:"address"`
	IsActive      *bool  `json:"is_active"`
}

// ContractListFilter carries query params for listing contracts.
type ContractListFilter struct {
	OrganizationID uuid.UUID
	ProjectID      uuid.UUID
	Status         string
	VendorID       *uuid.UUID
	Search         string
	Page           int
	PageSize       int
}

// CreateContractRequest is the DTO for contract creation.
type CreateContractRequest struct {
	ContractNumber string          `json:"contract_number" binding:"required,max=200"`
	Title          string          `json:"title" binding:"required,max=500"`
	VendorID       uuid.UUID       `json:"vendor_id" binding:"required"`
	ConsultantID   *uuid.UUID      `json:"consultant_id"`
	ContractValue  float64         `json:"contract_value"`
	Currency       string          `json:"currency" binding:"max=10"`
	SignedDate     *types.FlexTime `json:"signed_date"`
	StartDate      *types.FlexTime `json:"start_date"`
	EndDate        *types.FlexTime `json:"end_date"`
	Status         string          `json:"status" binding:"max=50"`
	ScopeOfWork    string          `json:"scope_of_work"`
}

// UpdateContractRequest is the DTO for contract updates.
type UpdateContractRequest struct {
	ContractNumber string          `json:"contract_number" binding:"max=200"`
	Title          string          `json:"title" binding:"max=500"`
	VendorID       *uuid.UUID      `json:"vendor_id"`
	ConsultantID   *uuid.UUID      `json:"consultant_id"`
	ContractValue  *float64        `json:"contract_value"`
	Currency       string          `json:"currency" binding:"max=10"`
	SignedDate     *types.FlexTime `json:"signed_date"`
	StartDate      *types.FlexTime `json:"start_date"`
	EndDate        *types.FlexTime `json:"end_date"`
	Status         string          `json:"status" binding:"max=50"`
	ScopeOfWork    string          `json:"scope_of_work"`
}

// DocumentListFilter carries query params for listing project documents.
type DocumentListFilter struct {
	OrganizationID uuid.UUID
	ProjectID      uuid.UUID
	Category       string
	Search         string
	Page           int
	PageSize       int
}

// UploadDocumentRequest carries multipart form fields plus file metadata for a
// document upload. The file itself is bound separately in the handler via
// c.FormFile("file").
type UploadDocumentRequest struct {
	Name     string `form:"name" binding:"max=500"`
	Category string `form:"category" binding:"max=100"`
	Version  string `form:"version" binding:"max=100"`
}

// UpdateDocumentRequest is the DTO for document metadata updates.
type UpdateDocumentRequest struct {
	Name     string `json:"name" binding:"max=500"`
	Category string `json:"category" binding:"max=100"`
	Version  string `json:"version" binding:"max=100"`
}

// ---------------------------------------------------------------------------
// Corrective Action (P1-006)
// ---------------------------------------------------------------------------

// CorrectiveAction records a deviation finding and its follow-up workflow.
// FSM: DRAFT → SUBMITTED → IN_PROGRESS → COMPLETED | REJECTED
//
//	REJECTED → DRAFT (revise & resubmit)
type CorrectiveAction struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null" json:"organization_id"`
	ProjectID      uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`

	// Deviation description
	Title          string `gorm:"size:500;not null" json:"title"`
	Deviation      string `gorm:"type:text;not null" json:"deviation"`
	RootCause      string `gorm:"type:text" json:"root_cause,omitempty"`
	Recommendation string `gorm:"type:text" json:"recommendation,omitempty"`

	// PIC & schedule
	PICUserID  *uuid.UUID      `gorm:"type:uuid" json:"pic_user_id,omitempty"`
	TargetDate *types.FlexTime `json:"target_date,omitempty"`

	// Source linkage (at most one non-null expected)
	SourceType    string     `gorm:"size:50" json:"source_type,omitempty"` // issue | risk | task
	SourceIssueID *uuid.UUID `gorm:"type:uuid" json:"source_issue_id,omitempty"`
	SourceRiskID  *uuid.UUID `gorm:"type:uuid" json:"source_risk_id,omitempty"`
	SourceTaskID  *uuid.UUID `gorm:"type:uuid" json:"source_task_id,omitempty"`

	// Workflow status
	Status string `gorm:"size:50;not null;default:'DRAFT'" json:"status"`
	// DRAFT | SUBMITTED | IN_PROGRESS | COMPLETED | REJECTED

	// Evidence note (free-text; full file linkage via project documents P1-005)
	EvidenceNote string `gorm:"type:text" json:"evidence_note,omitempty"`

	CreatedBy uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CorrectiveAction) TableName() string { return "corrective_actions" }

// CorrectiveActionListFilter carries query params for listing corrective actions.
type CorrectiveActionListFilter struct {
	OrganizationID uuid.UUID
	ProjectID      uuid.UUID
	Status         string
	SourceType     string
	Search         string
	Page           int
	PageSize       int
}

// CreateCorrectiveActionRequest is the DTO for corrective action creation.
type CreateCorrectiveActionRequest struct {
	Title          string          `json:"title" binding:"required,max=500"`
	Deviation      string          `json:"deviation" binding:"required"`
	RootCause      string          `json:"root_cause"`
	Recommendation string          `json:"recommendation"`
	PICUserID      *uuid.UUID      `json:"pic_user_id"`
	TargetDate     *types.FlexTime `json:"target_date"`
	SourceType     string          `json:"source_type" binding:"max=50"`
	SourceIssueID  *uuid.UUID      `json:"source_issue_id"`
	SourceRiskID   *uuid.UUID      `json:"source_risk_id"`
	SourceTaskID   *uuid.UUID      `json:"source_task_id"`
	EvidenceNote   string          `json:"evidence_note"`
}

// UpdateCorrectiveActionRequest is the DTO for corrective action updates.
type UpdateCorrectiveActionRequest struct {
	Title          string          `json:"title" binding:"max=500"`
	Deviation      string          `json:"deviation"`
	RootCause      string          `json:"root_cause"`
	Recommendation string          `json:"recommendation"`
	PICUserID      *uuid.UUID      `json:"pic_user_id"`
	TargetDate     *types.FlexTime `json:"target_date"`
	SourceType     string          `json:"source_type" binding:"max=50"`
	SourceIssueID  *uuid.UUID      `json:"source_issue_id"`
	SourceRiskID   *uuid.UUID      `json:"source_risk_id"`
	SourceTaskID   *uuid.UUID      `json:"source_task_id"`
	Status         string          `json:"status" binding:"max=50"`
	EvidenceNote   string          `json:"evidence_note"`
}
