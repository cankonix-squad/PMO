package project

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines data access for the project module.
type Repository interface {
	SyncProjectProgressFromPeriodic(ctx context.Context, orgID, projectID uuid.UUID) error
	Create(ctx context.Context, p *Project) error
	FindByID(ctx context.Context, id, orgID uuid.UUID) (*Project, error)
	List(ctx context.Context, filter ProjectListFilter) ([]Project, int64, error)
	Update(ctx context.Context, p *Project) error
	Delete(ctx context.Context, id, orgID uuid.UUID) error
	DeleteCascade(ctx context.Context, id, orgID uuid.UUID) error

	// Team
	AddTeamMember(ctx context.Context, m *TeamMember) error
	RemoveTeamMember(ctx context.Context, projectID, userID uuid.UUID) error
	GetTeam(ctx context.Context, projectID uuid.UUID) ([]TeamMember, error)

	// Progress
	RecordProgress(ctx context.Context, h *ProgressHistory) error
	GetProgressHistory(ctx context.Context, projectID uuid.UUID) ([]ProgressHistory, error)
	MasterDataExists(ctx context.Context, table string, id, orgID uuid.UUID) (bool, error)

	// Tasks
	CreateTask(ctx context.Context, t *Task) error
	FindTaskByID(ctx context.Context, projectID, taskID, orgID uuid.UUID) (*Task, error)
	ListTasks(ctx context.Context, filter TaskListFilter) ([]Task, int64, error)
	UpdateTask(ctx context.Context, t *Task) error
	DeleteTask(ctx context.Context, projectID, taskID, orgID uuid.UUID) error

	// Milestones
	CreateMilestone(ctx context.Context, m *Milestone) error
	FindMilestoneByID(ctx context.Context, projectID, milestoneID, orgID uuid.UUID) (*Milestone, error)
	ListMilestones(ctx context.Context, projectID, orgID uuid.UUID) ([]Milestone, error)
	UpdateMilestone(ctx context.Context, m *Milestone) error
	DeleteMilestone(ctx context.Context, projectID, milestoneID, orgID uuid.UUID) error

	// Issues
	CreateIssue(ctx context.Context, i *Issue) error
	FindIssueByID(ctx context.Context, projectID, issueID, orgID uuid.UUID) (*Issue, error)
	ListIssues(ctx context.Context, filter IssueListFilter) ([]Issue, int64, error)
	UpdateIssue(ctx context.Context, i *Issue) error
	DeleteIssue(ctx context.Context, projectID, issueID, orgID uuid.UUID) error

	// Risks
	CreateRisk(ctx context.Context, r *Risk) error
	FindRiskByID(ctx context.Context, projectID, riskID, orgID uuid.UUID) (*Risk, error)
	ListRisks(ctx context.Context, filter RiskListFilter) ([]Risk, int64, error)
	UpdateRisk(ctx context.Context, r *Risk) error
	DeleteRisk(ctx context.Context, projectID, riskID, orgID uuid.UUID) error

	// Budgets
	CreateBudget(ctx context.Context, b *ProjectBudget) error
	FindBudgetByID(ctx context.Context, projectID, budgetID, orgID uuid.UUID) (*ProjectBudget, error)
	ListBudgets(ctx context.Context, filter BudgetListFilter) ([]ProjectBudget, int64, error)
	UpdateBudget(ctx context.Context, b *ProjectBudget) error
	DeleteBudget(ctx context.Context, projectID, budgetID, orgID uuid.UUID) error

	// Vendors
	CreateVendor(ctx context.Context, v *Vendor) error
	FindVendorByID(ctx context.Context, id, orgID uuid.UUID) (*Vendor, error)
	ListVendors(ctx context.Context, filter VendorListFilter) ([]Vendor, int64, error)
	UpdateVendor(ctx context.Context, v *Vendor) error
	DeleteVendor(ctx context.Context, id, orgID uuid.UUID) error
	countContractsForVendor(ctx context.Context, vendorID, orgID uuid.UUID, out *int64) error

	// Contracts
	CreateContract(ctx context.Context, c *Contract) error
	FindContractByID(ctx context.Context, projectID, contractID, orgID uuid.UUID) (*Contract, error)
	ListContracts(ctx context.Context, filter ContractListFilter) ([]Contract, int64, error)
	UpdateContract(ctx context.Context, c *Contract) error
	DeleteContract(ctx context.Context, projectID, contractID, orgID uuid.UUID) error

	// Corrective Actions
	CreateCorrectiveAction(ctx context.Context, ca *CorrectiveAction) error
	FindCorrectiveActionByID(ctx context.Context, projectID, caID, orgID uuid.UUID) (*CorrectiveAction, error)
	ListCorrectiveActions(ctx context.Context, filter CorrectiveActionListFilter) ([]CorrectiveAction, int64, error)
	UpdateCorrectiveAction(ctx context.Context, ca *CorrectiveAction) error
	DeleteCorrectiveAction(ctx context.Context, projectID, caID, orgID uuid.UUID) error

	// Documents
	CreateDocument(ctx context.Context, d *ProjectDocument) error
	FindDocumentByID(ctx context.Context, projectID, documentID, orgID uuid.UUID) (*ProjectDocument, error)
	ListDocuments(ctx context.Context, filter DocumentListFilter) ([]ProjectDocument, int64, error)
	UpdateDocument(ctx context.Context, d *ProjectDocument) error
	DeleteDocument(ctx context.Context, projectID, documentID, orgID uuid.UUID) error
}
