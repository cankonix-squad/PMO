package constants

// App
const (
	AppName    = "PMO"
	AppVersion = "0.1.0"
	APIVersion = "v1"
	APIPrefix  = "/api/v1"
)

// Pagination defaults
const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Auth
const (
	BcryptCost          = 12
	MaxLoginAttempts    = 5
	LockoutDurationMins = 30
)

// JWT
const (
	JWTIssuer = "cankora"
)

// Scope types
const (
	ScopeAll             = "ALL"
	ScopeOrgUnit         = "ORG_UNIT"
	ScopeAssignedProject = "ASSIGNED_PROJECT"
	ScopeMemberProject   = "MEMBER_PROJECT"
)

// Default role codes (seed)
const (
	RoleSuperAdmin      = "SUPER_ADMIN"
	RoleAdmin           = "ADMIN"
	RolePMO             = "PMO"
	RoleProjectManager  = "PROJECT_MANAGER"
	RoleProjectOfficer  = "PROJECT_OFFICER"
	RoleExecutiveViewer = "EXECUTIVE_VIEWER"
	RoleAuditor         = "AUDITOR"
)

// Permission resources
const (
	ResourceUser             = "user"
	ResourceOrganization     = "organization"
	ResourceOrgUnit          = "org_unit"
	ResourceRole             = "role"
	ResourcePermission       = "permission"
	ResourceGroup            = "group"
	ResourceAuditLog         = "audit_log"
	ResourceProject          = "project"
	ResourceProjectTeam      = "project_team"
	ResourceMilestone        = "milestone"
	ResourceTask             = "task"
	ResourceIssue            = "issue"
	ResourceRisk             = "risk"
	ResourceBudget           = "budget"
	ResourceVendor           = "vendor"
	ResourceContract         = "contract"
	ResourceDocument         = "document"
	ResourceCorrectiveAction = "corrective_action"
	ResourceMeeting          = "meeting"
	ResourceReport           = "report"
	ResourceWorkflow         = "workflow"
	ResourceApprovalRequest  = "approval_request"

	// P1-010: Portfolio & Spatial masters
	ResourceProgram    = "program"
	ResourceSector     = "sector"
	ResourceRegion     = "region"
	ResourceRiverBasin = "river_basin"

	// P2-006: Program / Sector dashboard analytics
	ResourceProgramDashboard = "program_dashboard"

	// P2-007: Level 1 Executive dashboard
	ResourceExecutiveDashboard = "executive_dashboard"

	// P2-008: GIS Map
	ResourceGISMap = "gis_map"

	// P2-001: CSV/Excel Import
	ResourceImport = "import"

	// P2-011: Primavera P6 Integration
	ResourcePrimaveraSync = "primavera_sync"

	// P2-002: Government Connector Foundation
	ResourceGovernmentConnector = "government_connector"

	// P3-001: BIM/Digital Twin Integration
	ResourceBIMIntegration = "bim_integration"

	// PROJECT-FORM-002: Project Category master
	ResourceProjectCategory = "project_category"

	// P3-003: Data Governance — official validation & approval workflow
	ResourceDataGovernance = "data_governance"

	// UAT-004: Notification Delivery Foundation
	ResourceNotification  = "notification"
	ResourceNotifications = ResourceNotification

	// PMO-DASH-002: Periodic progress & financial report
	ResourcePeriodicReport = "periodic_report"

	// P1-011: Monitoring — baseline & snapshot
	ResourceBaseline          = "baseline"
	ResourceSnapshot          = "snapshot"
	ResourceDataSubmission    = "data_submission"
	ResourceValidationQueue   = "validation_queue"
	ResourceFieldInspection   = "field_inspection"
	ResourceFieldEvidence     = "field_evidence"
	ResourceHealthFormula     = "health_formula"
	ResourceHealthSnapshot    = "health_snapshot"
	ResourceCommandCenter     = "command_center"
	ResourceBenefit           = "benefit"
	ResourcePriority          = "priority"
	ResourceProjects          = ResourceProject
	ResourceTasks             = ResourceTask
	ResourceMilestones        = ResourceMilestone
	ResourceIssues            = ResourceIssue
	ResourceRisks             = ResourceRisk
	ResourceBudgets           = ResourceBudget
	ResourceVendors           = ResourceVendor
	ResourceContracts         = ResourceContract
	ResourceDocuments         = ResourceDocument
	ResourceCorrectiveActions = ResourceCorrectiveAction
	ResourceTeam              = ResourceProjectTeam
	ResourceUsers             = ResourceUser
	ResourceRoles             = ResourceRole
	ResourceOrganizations     = ResourceOrganization
	ResourceAuditLogs         = ResourceAuditLog
	ResourceReports           = ResourceReport
)

// Vendor types
const (
	VendorTypeVendor     = "VENDOR"
	VendorTypeConsultant = "CONSULTANT"
)

// Contract statuses
const (
	ContractStatusDraft      = "DRAFT"
	ContractStatusActive     = "ACTIVE"
	ContractStatusAmended    = "AMENDED"
	ContractStatusCompleted  = "COMPLETED"
	ContractStatusTerminated = "TERMINATED"
)

// Permission actions
const (
	ActionView    = "view"
	ActionCreate  = "create"
	ActionUpdate  = "update"
	ActionDelete  = "delete"
	ActionApprove = "approve"
	ActionExport  = "export"
	ActionAssign  = "assign"
)

// Project statuses
const (
	ProjectStatusDraft     = "DRAFT"
	ProjectStatusSubmitted = "SUBMITTED"
	ProjectStatusReviewed  = "REVIEWED"
	ProjectStatusApproved  = "APPROVED"
	ProjectStatusActive    = "ACTIVE"
	ProjectStatusOnHold    = "ON_HOLD"
	ProjectStatusCompleted = "COMPLETED"
	ProjectStatusClosed    = "CLOSED"
)

// Project health
const (
	HealthOnTrack = "ON_TRACK"
	HealthAtRisk  = "AT_RISK"
	HealthDelayed = "DELAYED"
)

// Priority levels
const (
	PriorityLow      = "LOW"
	PriorityMedium   = "MEDIUM"
	PriorityHigh     = "HIGH"
	PriorityCritical = "CRITICAL"
)

// Task statuses
const (
	TaskStatusBacklog     = "BACKLOG"
	TaskStatusAnalysis    = "ANALYSIS"
	TaskStatusDevelopment = "DEVELOPMENT"
	TaskStatusQA          = "QA"
	TaskStatusUAT         = "UAT"
	TaskStatusProduction  = "PRODUCTION"
	TaskStatusDone        = "DONE"
)

// Issue statuses
const (
	IssueStatusOpen       = "OPEN"
	IssueStatusInProgress = "IN_PROGRESS"
	IssueStatusResolved   = "RESOLVED"
	IssueStatusClosed     = "CLOSED"
)

// Risk statuses
const (
	RiskStatusIdentified = "IDENTIFIED"
	RiskStatusMonitored  = "MONITORED"
	RiskStatusMitigated  = "MITIGATED"
	RiskStatusClosed     = "CLOSED"
)

// Milestone statuses
const (
	MilestoneStatusUpcoming   = "UPCOMING"
	MilestoneStatusInProgress = "IN_PROGRESS"
	MilestoneStatusAchieved   = "ACHIEVED"
	MilestoneStatusMissed     = "MISSED"
)

// Approval statuses
const (
	ApprovalStatusPending   = "PENDING"
	ApprovalStatusApproved  = "APPROVED"
	ApprovalStatusRejected  = "REJECTED"
	ApprovalStatusCancelled = "CANCELLED"
)

// Audit actions
const (
	AuditActionCreate           = "CREATE"
	AuditActionUpdate           = "UPDATE"
	AuditActionDelete           = "DELETE"
	AuditActionApprove          = "APPROVE"
	AuditActionReject           = "REJECT"
	AuditActionLogin            = "LOGIN"
	AuditActionLogout           = "LOGOUT"
	AuditActionFailedLogin      = "FAILED_LOGIN"
	AuditActionExport           = "EXPORT"
	AuditActionRoleAssign       = "ROLE_ASSIGN"
	AuditActionRoleRevoke       = "ROLE_REVOKE"
	AuditActionPermissionChange = "PERMISSION_CHANGE"
	AuditActionPasswordChange   = "PASSWORD_CHANGE"
	AuditActionPasswordReset    = "PASSWORD_RESET"
)

// Project team roles
const (
	ProjectRolePM       = "PM"
	ProjectRoleOfficer  = "OFFICER"
	ProjectRoleQA       = "QA"
	ProjectRoleReviewer = "REVIEWER"
	ProjectRoleViewer   = "VIEWER"
)

// Org unit levels
const (
	OrgLevelKementerian = 1
	OrgLevelDitjen      = 2
	OrgLevelDirektorat  = 3
	OrgLevelSubdit      = 4
	OrgLevelUnit        = 5
)
