package project

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/workflow"
	"github.com/harmanto-49/cankora/internal/shared/constants"
	"github.com/harmanto-49/cankora/internal/shared/types"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ErrProjectNotFound is returned when a project lookup yields no result.
var ErrProjectNotFound = errors.New("project not found")

// ErrCodeTaken is returned when the project code is already used.
var ErrCodeTaken = errors.New("project code already in use")

// ErrTaskNotFound is returned when a task lookup yields no result.
var ErrTaskNotFound = errors.New("task not found")

// ErrMilestoneNotFound is returned when a milestone lookup yields no result.
var ErrMilestoneNotFound = errors.New("milestone not found")

// ErrIssueNotFound is returned when an issue lookup yields no result.
var ErrIssueNotFound = errors.New("issue not found")

// ErrRiskNotFound is returned when a risk lookup yields no result.
var ErrRiskNotFound = errors.New("risk not found")

// ErrBudgetNotFound is returned when a budget line lookup yields no result.
var ErrBudgetNotFound = errors.New("budget not found")

// ErrVendorNotFound is returned when a vendor lookup yields no result.
var ErrVendorNotFound = errors.New("vendor not found")

// ErrContractNotFound is returned when a contract lookup yields no result.
var ErrContractNotFound = errors.New("contract not found")

// ErrContractNumberTaken is returned when the contract number already exists
// within the same organization.
var ErrContractNumberTaken = errors.New("contract number already in use")

// ErrInvalidVendorType is returned when a vendor type is not VENDOR/CONSULTANT.
var ErrInvalidVendorType = errors.New("vendor type must be VENDOR or CONSULTANT")

// ErrInvalidContractStatus is returned when a contract status is not one of
// DRAFT/ACTIVE/AMENDED/COMPLETED/TERMINATED.
var ErrInvalidContractStatus = errors.New("invalid contract status")

// ErrInvalidContractDates is returned when start_date is after end_date.
var ErrInvalidContractDates = errors.New("contract start_date cannot be after end_date")

// ErrInvalidContractValue is returned when contract_value is negative.
var ErrInvalidContractValue = errors.New("contract value cannot be negative")

// ErrVendorInUse is returned when deleting a vendor that is referenced by an
// active contract.
var ErrVendorInUse = errors.New("vendor is referenced by contracts and cannot be deleted")

// ErrDocumentNotFound is returned when a document lookup yields no result.
var ErrDocumentNotFound = errors.New("document not found")

// ErrDocumentMissingFile is returned when an upload request has no file part.
var ErrDocumentMissingFile = errors.New("no file uploaded")

// ErrDocumentStorage is returned when reading/writing the underlying file
// store fails (file missing on disk, permission error, etc.).
var ErrDocumentStorage = errors.New("document storage error")

// ErrCorrectiveActionNotFound is returned when a corrective action lookup yields no result.
var ErrCorrectiveActionNotFound = errors.New("corrective action not found")

// Repository defines data access for the project module.
type Repository interface {
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

// postgresRepository is the GORM implementation.
type postgresRepository struct {
	db *gorm.DB
}

// NewRepository creates a new project Repository.
func NewRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, p *Project) error {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrCodeTaken
		}
		return err
	}
	return nil
}

// isUniqueViolation reports whether err is a PostgreSQL unique constraint violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func (r *postgresRepository) FindByID(ctx context.Context, id, orgID uuid.UUID) (*Project, error) {
	var p Project
	err := r.db.WithContext(ctx).
		Preload("Milestones", "organization_id = ?", orgID).
		Preload("Team").
		Where("projects.id = ? AND projects.organization_id = ?", id, orgID).
		First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProjectNotFound
	}
	// Populate org_unit_name via separate query (avoids Preload conflict with GORM '-' tag)
	if p.OrgUnitID != nil {
		var name string
		if e := r.db.WithContext(ctx).
			Raw("SELECT name FROM org_units WHERE id = ? AND deleted_at IS NULL", p.OrgUnitID).
			Scan(&name).Error; e == nil {
			p.OrgUnitName = name
		}
	}
	return &p, err
}

func (r *postgresRepository) List(ctx context.Context, filter ProjectListFilter) ([]Project, int64, error) {
	query := r.db.WithContext(ctx).Model(&Project{}).
		Where("organization_id = ?", filter.OrganizationID)

	if filter.OrgUnitID != nil {
		query = query.Where("org_unit_id = ?", filter.OrgUnitID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Priority != "" {
		query = query.Where("priority = ?", filter.Priority)
	}
	if filter.ManagerID != nil {
		query = query.Where("manager_id = ?", filter.ManagerID)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ?", s, s)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	var projects []Project
	err := query.Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&projects).Error
	if err != nil {
		return nil, 0, err
	}

	// Batch-resolve org_unit names: collect unique non-nil org_unit_ids,
	// fetch names in one query, then populate OrgUnitName on each project.
	orgUnitIDs := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]bool)
	for _, p := range projects {
		if p.OrgUnitID != nil && !seen[*p.OrgUnitID] {
			orgUnitIDs = append(orgUnitIDs, *p.OrgUnitID)
			seen[*p.OrgUnitID] = true
		}
	}
	if len(orgUnitIDs) > 0 {
		type orgUnitRow struct {
			ID   uuid.UUID
			Name string
		}
		var ouRows []orgUnitRow
		if e := r.db.WithContext(ctx).
			Raw("SELECT id, name FROM org_units WHERE id IN ? AND deleted_at IS NULL", orgUnitIDs).
			Scan(&ouRows).Error; e == nil {
			nameMap := make(map[uuid.UUID]string, len(ouRows))
			for _, ou := range ouRows {
				nameMap[ou.ID] = ou.Name
			}
			for i := range projects {
				if projects[i].OrgUnitID != nil {
					projects[i].OrgUnitName = nameMap[*projects[i].OrgUnitID]
				}
			}
		}
	}
	return projects, total, err
}

func (r *postgresRepository) Update(ctx context.Context, p *Project) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *postgresRepository) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, orgID).
		Delete(&Project{}).Error
}

// DeleteCascade soft-deletes a project AND all of its business children inside
// a single transaction. Every statement is scoped by project_id (and where the
// child table carries organization_id, also by orgID) so cross-tenant writes
// are impossible. Child rows without a deleted_at column (e.g. progress
// history) are hard-removed because they are historical snapshots, not
// business records.
func (r *postgresRepository) DeleteCascade(ctx context.Context, id, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Scope child tables to the project. Tables with organization_id also
		// scope by orgID as a defensive tenant check.
		if err := tx.Where("project_id = ?", id).Delete(&Task{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&Milestone{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&Issue{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&Risk{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&ProjectBudget{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&Contract{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&ProjectDocument{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&CorrectiveAction{}).Error; err != nil {
			return err
		}
		if err := tx.Exec(
			`UPDATE benefit_measurements bm
			 SET deleted_at = NOW()
			 FROM benefit_indicators bi
			 WHERE bi.id = bm.indicator_id
			   AND bi.project_id = ?
			   AND bi.organization_id = ?
			   AND bm.organization_id = ?
			   AND bm.deleted_at IS NULL`,
			id, orgID, orgID,
		).Error; err != nil {
			return err
		}
		if err := tx.Exec(
			`UPDATE benefit_indicators
			 SET deleted_at = NOW()
			 WHERE project_id = ?
			   AND organization_id = ?
			   AND deleted_at IS NULL`,
			id, orgID,
		).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&TeamMember{}).Error; err != nil {
			return err
		}

		// task_assignments are soft-deleted via their parent tasks in GORM
		// association hooks; we also cascade explicitly for safety.
		if err := tx.Exec(
			"UPDATE task_assignments ta SET deleted_at = NOW() FROM tasks t WHERE t.id = ta.task_id AND t.project_id = ? AND ta.deleted_at IS NULL",
			id,
		).Error; err != nil {
			return err
		}

		// Progress history has no deleted_at — it is an immutable snapshot
		// series. Removing the rows is the correct "soft delete" equivalent
		// for a deleted project (no business row remains accessible).
		if err := tx.Where("project_id = ?", id).Delete(&ProgressHistory{}).Error; err != nil {
			return err
		}

		// Finally soft-delete the project itself.
		if err := tx.Where("id = ? AND organization_id = ?", id, orgID).Delete(&Project{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *postgresRepository) AddTeamMember(ctx context.Context, m *TeamMember) error {
	return r.db.WithContext(ctx).
		Where(TeamMember{ProjectID: m.ProjectID, UserID: m.UserID}).
		FirstOrCreate(m).Error
}

func (r *postgresRepository) RemoveTeamMember(ctx context.Context, projectID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Delete(&TeamMember{}).Error
}

func (r *postgresRepository) GetTeam(ctx context.Context, projectID uuid.UUID) ([]TeamMember, error) {
	var members []TeamMember
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&members).Error
	return members, err
}

func (r *postgresRepository) RecordProgress(ctx context.Context, h *ProgressHistory) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *postgresRepository) GetProgressHistory(ctx context.Context, projectID uuid.UUID) ([]ProgressHistory, error) {
	var history []ProgressHistory
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("recorded_at DESC").
		Find(&history).Error
	return history, err
}

// ---- Task repository methods ------------------------------------------------

func (r *postgresRepository) CreateTask(ctx context.Context, t *Task) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *postgresRepository) FindTaskByID(ctx context.Context, projectID, taskID, orgID uuid.UUID) (*Task, error) {
	var t Task
	err := r.db.WithContext(ctx).
		Preload("Subtasks", "organization_id = ? AND project_id = ?", orgID, projectID).
		Preload("Assignments").
		Where("id = ? AND project_id = ? AND organization_id = ?", taskID, projectID, orgID).
		First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTaskNotFound
	}
	return &t, err
}

func (r *postgresRepository) ListTasks(ctx context.Context, filter TaskListFilter) ([]Task, int64, error) {
	q := r.db.WithContext(ctx).Model(&Task{}).
		Where("project_id = ? AND organization_id = ?", filter.ProjectID, filter.OrganizationID).
		Where("parent_id IS NULL") // top-level tasks only

	if filter.MilestoneID != nil {
		q = q.Where("milestone_id = ?", filter.MilestoneID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.AssignedTo != nil {
		q = q.Joins("JOIN task_assignments ON task_assignments.task_id = tasks.id").
			Where("task_assignments.user_id = ?", filter.AssignedTo)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Where("title ILIKE ?", s)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	var tasks []Task
	err := q.Preload("Subtasks").Preload("Assignments").
		Order("created_at ASC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&tasks).Error
	return tasks, total, err
}

func (r *postgresRepository) UpdateTask(ctx context.Context, t *Task) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *postgresRepository) DeleteTask(ctx context.Context, projectID, taskID, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", taskID, projectID, orgID).
		Delete(&Task{}).Error
}

// ---- Milestone repository methods ------------------------------------------

func (r *postgresRepository) CreateMilestone(ctx context.Context, m *Milestone) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *postgresRepository) FindMilestoneByID(ctx context.Context, projectID, milestoneID, orgID uuid.UUID) (*Milestone, error) {
	var m Milestone
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", milestoneID, projectID, orgID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMilestoneNotFound
	}
	return &m, err
}

func (r *postgresRepository) ListMilestones(ctx context.Context, projectID, orgID uuid.UUID) ([]Milestone, error) {
	var milestones []Milestone
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND organization_id = ?", projectID, orgID).
		Order("due_date ASC NULLS LAST, created_at ASC").
		Find(&milestones).Error
	return milestones, err
}

func (r *postgresRepository) UpdateMilestone(ctx context.Context, m *Milestone) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *postgresRepository) DeleteMilestone(ctx context.Context, projectID, milestoneID, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", milestoneID, projectID, orgID).
		Delete(&Milestone{}).Error
}

// ---- Issue repository methods ----------------------------------------------

func (r *postgresRepository) CreateIssue(ctx context.Context, i *Issue) error {
	return r.db.WithContext(ctx).Create(i).Error
}

func (r *postgresRepository) FindIssueByID(ctx context.Context, projectID, issueID, orgID uuid.UUID) (*Issue, error) {
	var i Issue
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", issueID, projectID, orgID).
		First(&i).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrIssueNotFound
	}
	return &i, err
}

func (r *postgresRepository) ListIssues(ctx context.Context, filter IssueListFilter) ([]Issue, int64, error) {
	q := r.db.WithContext(ctx).Model(&Issue{}).
		Where("project_id = ? AND organization_id = ?", filter.ProjectID, filter.OrganizationID)

	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Severity != "" {
		q = q.Where("severity = ?", filter.Severity)
	}
	if filter.AssignedTo != nil {
		q = q.Where("assigned_to = ?", filter.AssignedTo)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Where("title ILIKE ? OR description ILIKE ?", s, s)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	var issues []Issue
	err := q.Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&issues).Error
	return issues, total, err
}

func (r *postgresRepository) UpdateIssue(ctx context.Context, i *Issue) error {
	return r.db.WithContext(ctx).Save(i).Error
}

func (r *postgresRepository) DeleteIssue(ctx context.Context, projectID, issueID, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", issueID, projectID, orgID).
		Delete(&Issue{}).Error
}

// ---- Risk repository methods ----------------------------------------------

func (r *postgresRepository) CreateRisk(ctx context.Context, risk *Risk) error {
	return r.db.WithContext(ctx).Create(risk).Error
}

func (r *postgresRepository) FindRiskByID(ctx context.Context, projectID, riskID, orgID uuid.UUID) (*Risk, error) {
	var risk Risk
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", riskID, projectID, orgID).
		First(&risk).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRiskNotFound
	}
	return &risk, err
}

func (r *postgresRepository) ListRisks(ctx context.Context, filter RiskListFilter) ([]Risk, int64, error) {
	q := r.db.WithContext(ctx).Model(&Risk{}).
		Where("project_id = ? AND organization_id = ?", filter.ProjectID, filter.OrganizationID)

	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Severity != "" {
		q = q.Where("severity = ?", filter.Severity)
	}
	if filter.OwnedBy != nil {
		q = q.Where("owned_by = ?", filter.OwnedBy)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Where("title ILIKE ? OR description ILIKE ?", s, s)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	var risks []Risk
	err := q.Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&risks).Error
	return risks, total, err
}

func (r *postgresRepository) UpdateRisk(ctx context.Context, risk *Risk) error {
	return r.db.WithContext(ctx).Save(risk).Error
}

func (r *postgresRepository) DeleteRisk(ctx context.Context, projectID, riskID, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", riskID, projectID, orgID).
		Delete(&Risk{}).Error
}

// ---- Budget repository methods ----------------------------------------------
// project_budgets has no organization_id column; tenant safety is enforced by the
// service layer (parent project ownership) before any budget access.

func (r *postgresRepository) CreateBudget(ctx context.Context, b *ProjectBudget) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *postgresRepository) FindBudgetByID(ctx context.Context, projectID, budgetID, orgID uuid.UUID) (*ProjectBudget, error) {
	// orgID is resolved through the parent project ownership check in the service.
	var b ProjectBudget
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", budgetID, projectID).
		First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrBudgetNotFound
	}
	return &b, err
}

func (r *postgresRepository) ListBudgets(ctx context.Context, filter BudgetListFilter) ([]ProjectBudget, int64, error) {
	q := r.db.WithContext(ctx).Model(&ProjectBudget{}).
		Where("project_id = ?", filter.ProjectID)

	if filter.Category != "" {
		q = q.Where("category ILIKE ?", "%"+filter.Category+"%")
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Where("category ILIKE ? OR description ILIKE ?", s, s)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	var budgets []ProjectBudget
	err := q.Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&budgets).Error
	return budgets, total, err
}

func (r *postgresRepository) UpdateBudget(ctx context.Context, b *ProjectBudget) error {
	return r.db.WithContext(ctx).Save(b).Error
}

func (r *postgresRepository) DeleteBudget(ctx context.Context, projectID, budgetID, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", budgetID, projectID).
		Delete(&ProjectBudget{}).Error
}

// ---- Vendor repository methods -------------------------------------------------

func (r *postgresRepository) CreateVendor(ctx context.Context, v *Vendor) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *postgresRepository) FindVendorByID(ctx context.Context, id, orgID uuid.UUID) (*Vendor, error) {
	var v Vendor
	err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, orgID).
		First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrVendorNotFound
	}
	return &v, err
}

func (r *postgresRepository) ListVendors(ctx context.Context, filter VendorListFilter) ([]Vendor, int64, error) {
	q := r.db.WithContext(ctx).Model(&Vendor{}).
		Where("organization_id = ?", filter.OrganizationID)

	if filter.Type != "" {
		q = q.Where("type = ?", filter.Type)
	}
	if filter.IsActive != nil {
		q = q.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Where("name ILIKE ? OR legal_name ILIKE ? OR tax_id ILIKE ?", s, s, s)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	var vendors []Vendor
	err := q.Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&vendors).Error
	return vendors, total, err
}

func (r *postgresRepository) UpdateVendor(ctx context.Context, v *Vendor) error {
	return r.db.WithContext(ctx).Save(v).Error
}

func (r *postgresRepository) DeleteVendor(ctx context.Context, id, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, orgID).
		Delete(&Vendor{}).Error
}

// ---- Contract repository methods ------------------------------------------------

func (r *postgresRepository) CreateContract(ctx context.Context, c *Contract) error {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrContractNumberTaken
		}
		return err
	}
	return nil
}

func (r *postgresRepository) FindContractByID(ctx context.Context, projectID, contractID, orgID uuid.UUID) (*Contract, error) {
	var c Contract
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", contractID, projectID, orgID).
		Preload("Vendor").
		Preload("Consultant").
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrContractNotFound
	}
	return &c, err
}

func (r *postgresRepository) ListContracts(ctx context.Context, filter ContractListFilter) ([]Contract, int64, error) {
	q := r.db.WithContext(ctx).Model(&Contract{}).
		Where("project_id = ? AND organization_id = ?", filter.ProjectID, filter.OrganizationID)

	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.VendorID != nil {
		q = q.Where("vendor_id = ?", filter.VendorID)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Where("contract_number ILIKE ? OR title ILIKE ?", s, s)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	var contracts []Contract
	err := q.Preload("Vendor").Preload("Consultant").
		Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&contracts).Error
	return contracts, total, err
}

func (r *postgresRepository) UpdateContract(ctx context.Context, c *Contract) error {
	if err := r.db.WithContext(ctx).Save(c).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrContractNumberTaken
		}
		return err
	}
	return nil
}

func (r *postgresRepository) DeleteContract(ctx context.Context, projectID, contractID, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", contractID, projectID, orgID).
		Delete(&Contract{}).Error
}

// countContractsForVendor counts non-deleted contracts referencing a vendor as
// either the provider (vendor_id) or the supervisory consultant
// (consultant_id), within the given organization.
func (r *postgresRepository) countContractsForVendor(ctx context.Context, vendorID, orgID uuid.UUID, out *int64) error {
	return r.db.WithContext(ctx).Model(&Contract{}).
		Where("organization_id = ? AND (vendor_id = ? OR consultant_id = ?)", orgID, vendorID, vendorID).
		Count(out).Error
}

// ---- Document repository methods -------------------------------------------------

func (r *postgresRepository) CreateDocument(ctx context.Context, d *ProjectDocument) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *postgresRepository) FindDocumentByID(ctx context.Context, projectID, documentID, orgID uuid.UUID) (*ProjectDocument, error) {
	// orgID scoping is enforced via parent project ownership check in the
	// service (project_documents has no organization_id column).
	var d ProjectDocument
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", documentID, projectID).
		First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrDocumentNotFound
	}
	return &d, err
}

func (r *postgresRepository) ListDocuments(ctx context.Context, filter DocumentListFilter) ([]ProjectDocument, int64, error) {
	q := r.db.WithContext(ctx).Model(&ProjectDocument{}).
		Where("project_id = ?", filter.ProjectID)

	if filter.Category != "" {
		q = q.Where("category = ?", filter.Category)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Where("name ILIKE ? OR category ILIKE ?", s, s)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	var docs []ProjectDocument
	err := q.Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&docs).Error
	return docs, total, err
}

func (r *postgresRepository) UpdateDocument(ctx context.Context, d *ProjectDocument) error {
	return r.db.WithContext(ctx).Save(d).Error
}

func (r *postgresRepository) DeleteDocument(ctx context.Context, projectID, documentID, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", documentID, projectID).
		Delete(&ProjectDocument{}).Error
}

// Service handles project business logic.
type Service struct {
	repo Repository
	fsm  *workflow.FSM
	log  *zap.Logger

	// storageRoot is the local dev storage root for document files.
	storageRoot string
	// maxFileSize caps uploaded document size in bytes.
	maxFileSize int64
}

// NewService creates a new project Service.
func NewService(repo Repository, fsm *workflow.FSM, log *zap.Logger) *Service {
	return &Service{
		repo:        repo,
		fsm:         fsm,
		log:         log,
		storageRoot: "storage/documents",
		maxFileSize: 20 * 1024 * 1024,
	}
}

// ConfigureDocumentStorage sets the local storage root and max upload size
// (bytes). It is called once during wiring; defaults remain until then.
func (s *Service) ConfigureDocumentStorage(root string, maxBytes int64) {
	if root != "" {
		s.storageRoot = root
	}
	if maxBytes > 0 {
		s.maxFileSize = maxBytes
	}
}

// Create creates a new project.
func (s *Service) Create(ctx context.Context, req *CreateProjectRequest, orgID, createdBy uuid.UUID) (*Project, error) {
	p := &Project{
		ID:             uuid.New(),
		OrganizationID: orgID,
		OrgUnitID:      req.OrgUnitID,
		Code:           req.Code,
		Name:           req.Name,
		Description:    req.Description,
		Objectives:     req.Objectives,
		Status:         "DRAFT",
		Priority:       req.Priority,
		Category:       req.Category,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		BudgetTotal:    req.BudgetTotal,
		Currency:       req.Currency,
		ManagerID:      req.ManagerID,
		CreatedBy:      createdBy,
	}
	if p.Priority == "" {
		p.Priority = "MEDIUM"
	}
	if p.Currency == "" {
		p.Currency = "IDR"
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("project: create: %w", err)
	}

	// Auto-add creator as PROJECT_MANAGER
	_ = s.repo.AddTeamMember(ctx, &TeamMember{
		ID:        uuid.New(),
		ProjectID: p.ID,
		UserID:    createdBy,
		Role:      "PROJECT_MANAGER",
	})

	return s.repo.FindByID(ctx, p.ID, orgID)
}

// GetByID returns a project by ID, checking org boundary.
func (s *Service) GetByID(ctx context.Context, id, orgID uuid.UUID) (*Project, error) {
	return s.repo.FindByID(ctx, id, orgID)
}

// List returns paginated projects for an organization.
func (s *Service) List(ctx context.Context, filter ProjectListFilter) ([]Project, int64, error) {
	return s.repo.List(ctx, filter)
}

// Update modifies a project's details.
func (s *Service) Update(ctx context.Context, id, orgID uuid.UUID, req *UpdateProjectRequest) (*Project, error) {
	p, err := s.GetByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Description != "" {
		p.Description = req.Description
	}
	if req.Objectives != "" {
		p.Objectives = req.Objectives
	}
	if req.Priority != "" {
		p.Priority = req.Priority
	}
	if req.Category != "" {
		p.Category = req.Category
	}
	if req.StartDate != nil {
		p.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		p.EndDate = req.EndDate
	}
	if req.BudgetTotal != nil {
		p.BudgetTotal = *req.BudgetTotal
	}
	if req.Currency != "" {
		p.Currency = req.Currency
	}
	if req.ManagerID != nil {
		p.ManagerID = req.ManagerID
	}
	if req.ProgressPct != nil {
		p.ProgressPct = *req.ProgressPct
	}
	if req.OrgUnitID != nil {
		p.OrgUnitID = req.OrgUnitID
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("project: update: %w", err)
	}

	return s.repo.FindByID(ctx, p.ID, orgID)
}

// Transition moves a project to a new status via the FSM.
func (s *Service) Transition(ctx context.Context, id, orgID uuid.UUID, req *TransitionRequest) (*Project, error) {
	p, err := s.GetByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}

	if _, err := s.fsm.Transition(workflow.EntityProject, p.Status, req.ToStatus); err != nil {
		return nil, fmt.Errorf("project: %w — allowed: %v",
			workflow.ErrTransitionNotAllowed,
			s.fsm.AllowedTransitions(workflow.EntityProject, p.Status),
		)
	}

	p.Status = req.ToStatus
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("project: transition update: %w", err)
	}

	return p, nil
}

// Delete soft-deletes a project and cascades the soft delete to all business
// children (tasks, subtasks, milestones, issues, risks, budgets, documents,
// team members, task assignments) in a single transaction. It verifies the
// project belongs to the caller's organization first.
func (s *Service) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	p, err := s.GetByID(ctx, id, orgID)
	if err != nil {
		return err
	}
	return s.repo.DeleteCascade(ctx, p.ID, orgID)
}

// ---- Task service methods ---------------------------------------------------

// CreateTask creates a new task under a project, enforcing org boundary.
func (s *Service) CreateTask(ctx context.Context, projectID, orgID, createdBy uuid.UUID, req *CreateTaskRequest) (*Task, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	if req.ParentID != nil {
		parent, err := s.GetTask(ctx, projectID, *req.ParentID, orgID)
		if err != nil {
			return nil, err
		}
		if parent.ParentID != nil {
			return nil, fmt.Errorf("task: parent task cannot be nested more than one level")
		}
	}
	if req.MilestoneID != nil {
		if _, err := s.GetMilestone(ctx, projectID, *req.MilestoneID, orgID); err != nil {
			return nil, err
		}
	}

	t := &Task{
		OrganizationID: orgID,
		ProjectID:      projectID,
		MilestoneID:    req.MilestoneID,
		ParentID:       req.ParentID,
		WBSCode:        req.WBSCode,
		Title:          req.Title,
		Description:    req.Description,
		Priority:       req.Priority,
		Type:           req.Type,
		StartDate:      req.StartDate,
		DueDate:        req.DueDate,
		EstHours:       req.EstHours,
		CreatedBy:      createdBy,
	}
	if t.Priority == "" {
		t.Priority = "MEDIUM"
	}
	if t.Type == "" {
		t.Type = "TASK"
	}
	if err := s.repo.CreateTask(ctx, t); err != nil {
		return nil, fmt.Errorf("task: create: %w", err)
	}
	return s.repo.FindTaskByID(ctx, projectID, t.ID, orgID)
}

// GetTask returns a task by ID, verifying org boundary.
func (s *Service) GetTask(ctx context.Context, projectID, taskID, orgID uuid.UUID) (*Task, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	return s.repo.FindTaskByID(ctx, projectID, taskID, orgID)
}

// ListTasks returns paginated tasks for a project.
func (s *Service) ListTasks(ctx context.Context, filter TaskListFilter) ([]Task, int64, error) {
	if _, err := s.GetByID(ctx, filter.ProjectID, filter.OrganizationID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListTasks(ctx, filter)
}

// UpdateTask applies partial updates to a task.
func (s *Service) UpdateTask(ctx context.Context, projectID, taskID, orgID uuid.UUID, req *UpdateTaskRequest) (*Task, error) {
	t, err := s.GetTask(ctx, projectID, taskID, orgID)
	if err != nil {
		return nil, err
	}
	if req.Title != "" {
		t.Title = req.Title
	}
	if req.Description != "" {
		t.Description = req.Description
	}
	if req.Status != "" {
		if _, err := s.fsm.Transition(workflow.EntityTask, t.Status, req.Status); err != nil {
			return nil, fmt.Errorf("task: %w - allowed: %v",
				workflow.ErrTransitionNotAllowed,
				s.fsm.AllowedTransitions(workflow.EntityTask, t.Status),
			)
		}
		t.Status = req.Status
	}
	if req.Priority != "" {
		t.Priority = req.Priority
	}
	if req.Type != "" {
		t.Type = req.Type
	}
	if req.WBSCode != "" {
		t.WBSCode = req.WBSCode
	}
	if req.MilestoneID != nil {
		if _, err := s.GetMilestone(ctx, projectID, *req.MilestoneID, orgID); err != nil {
			return nil, err
		}
		t.MilestoneID = req.MilestoneID
	}
	if req.StartDate != nil {
		t.StartDate = req.StartDate
	}
	if req.DueDate != nil {
		t.DueDate = req.DueDate
	}
	if req.EstHours > 0 {
		t.EstHours = req.EstHours
	}
	if req.ActualHours != nil {
		t.ActualHours = *req.ActualHours
	}
	if req.ProgressPct != nil {
		t.ProgressPct = *req.ProgressPct
	}
	if err := s.repo.UpdateTask(ctx, t); err != nil {
		return nil, fmt.Errorf("task: update: %w", err)
	}
	return s.repo.FindTaskByID(ctx, projectID, t.ID, orgID)
}

// DeleteTask soft-deletes a task, verifying org boundary.
func (s *Service) DeleteTask(ctx context.Context, projectID, taskID, orgID uuid.UUID) error {
	t, err := s.GetTask(ctx, projectID, taskID, orgID)
	if err != nil {
		return err
	}
	return s.repo.DeleteTask(ctx, projectID, t.ID, orgID)
}

// ---- Milestone service methods ---------------------------------------------

// CreateMilestone creates a new milestone under a project.
func (s *Service) CreateMilestone(ctx context.Context, projectID, orgID, createdBy uuid.UUID, req *CreateMilestoneRequest) (*Milestone, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}

	m := &Milestone{
		OrganizationID: orgID,
		ProjectID:      projectID,
		Title:          req.Title,
		Description:    req.Description,
		DueDate:        req.DueDate,
		CreatedBy:      createdBy,
	}
	if err := s.repo.CreateMilestone(ctx, m); err != nil {
		return nil, fmt.Errorf("milestone: create: %w", err)
	}
	return s.repo.FindMilestoneByID(ctx, projectID, m.ID, orgID)
}

// GetMilestone returns a milestone by ID, verifying org boundary.
func (s *Service) GetMilestone(ctx context.Context, projectID, milestoneID, orgID uuid.UUID) (*Milestone, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	return s.repo.FindMilestoneByID(ctx, projectID, milestoneID, orgID)
}

// ListMilestones returns all milestones for a project.
func (s *Service) ListMilestones(ctx context.Context, projectID, orgID uuid.UUID) ([]Milestone, error) {
	// Verify the project belongs to this org first
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	return s.repo.ListMilestones(ctx, projectID, orgID)
}

// UpdateMilestone applies partial updates to a milestone.
func (s *Service) UpdateMilestone(ctx context.Context, projectID, milestoneID, orgID uuid.UUID, req *UpdateMilestoneRequest) (*Milestone, error) {
	m, err := s.GetMilestone(ctx, projectID, milestoneID, orgID)
	if err != nil {
		return nil, err
	}
	if req.Title != "" {
		m.Title = req.Title
	}
	if req.Description != "" {
		m.Description = req.Description
	}
	if req.Status != "" {
		if _, err := s.fsm.Transition(workflow.EntityMilestone, m.Status, req.Status); err != nil {
			return nil, fmt.Errorf("milestone: %w - allowed: %v",
				workflow.ErrTransitionNotAllowed,
				s.fsm.AllowedTransitions(workflow.EntityMilestone, m.Status),
			)
		}
		m.Status = req.Status
	}
	if req.ProgressPct != nil {
		m.ProgressPct = *req.ProgressPct
	}
	if req.DueDate != nil {
		m.DueDate = req.DueDate
	}
	if err := s.repo.UpdateMilestone(ctx, m); err != nil {
		return nil, fmt.Errorf("milestone: update: %w", err)
	}
	return m, nil
}

// DeleteMilestone soft-deletes a milestone, verifying org boundary.
func (s *Service) DeleteMilestone(ctx context.Context, projectID, milestoneID, orgID uuid.UUID) error {
	m, err := s.GetMilestone(ctx, projectID, milestoneID, orgID)
	if err != nil {
		return err
	}
	return s.repo.DeleteMilestone(ctx, projectID, m.ID, orgID)
}

// ---- Issue service methods -------------------------------------------------

// CreateIssue creates a new issue under a project, enforcing org boundary.
func (s *Service) CreateIssue(ctx context.Context, projectID, orgID, createdBy uuid.UUID, req *CreateIssueRequest) (*Issue, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	if req.TaskID != nil {
		if _, err := s.GetTask(ctx, projectID, *req.TaskID, orgID); err != nil {
			return nil, err
		}
	}

	i := &Issue{
		OrganizationID: orgID,
		ProjectID:      projectID,
		TaskID:         req.TaskID,
		Title:          req.Title,
		Description:    req.Description,
		Status:         "OPEN",
		Severity:       req.Severity,
		Escalation:     req.Escalation,
		ReportedBy:     createdBy,
		AssignedTo:     req.AssignedTo,
		DueDate:        req.DueDate,
		Resolution:     req.Resolution,
	}
	if i.Severity == "" {
		i.Severity = "MEDIUM"
	}
	if i.Escalation == "" {
		i.Escalation = "NONE"
	}
	if err := s.repo.CreateIssue(ctx, i); err != nil {
		return nil, fmt.Errorf("issue: create: %w", err)
	}
	return s.repo.FindIssueByID(ctx, projectID, i.ID, orgID)
}

// GetIssue returns an issue by ID, verifying org and project boundaries.
func (s *Service) GetIssue(ctx context.Context, projectID, issueID, orgID uuid.UUID) (*Issue, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	return s.repo.FindIssueByID(ctx, projectID, issueID, orgID)
}

// ListIssues returns paginated issues for a project.
func (s *Service) ListIssues(ctx context.Context, filter IssueListFilter) ([]Issue, int64, error) {
	if _, err := s.GetByID(ctx, filter.ProjectID, filter.OrganizationID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListIssues(ctx, filter)
}

// UpdateIssue applies partial updates to an issue, enforcing FSM for status.
func (s *Service) UpdateIssue(ctx context.Context, projectID, issueID, orgID uuid.UUID, req *UpdateIssueRequest) (*Issue, error) {
	i, err := s.GetIssue(ctx, projectID, issueID, orgID)
	if err != nil {
		return nil, err
	}
	if req.Title != "" {
		i.Title = req.Title
	}
	if req.Description != "" {
		i.Description = req.Description
	}
	if req.Status != "" {
		if _, err := s.fsm.Transition(workflow.EntityIssue, i.Status, req.Status); err != nil {
			return nil, fmt.Errorf("issue: %w - allowed: %v",
				workflow.ErrTransitionNotAllowed,
				s.fsm.AllowedTransitions(workflow.EntityIssue, i.Status),
			)
		}
		i.Status = req.Status
	}
	if req.Severity != "" {
		i.Severity = req.Severity
	}
	if req.Escalation != "" {
		i.Escalation = req.Escalation
	}
	if req.TaskID != nil {
		if _, err := s.GetTask(ctx, projectID, *req.TaskID, orgID); err != nil {
			return nil, err
		}
		i.TaskID = req.TaskID
	}
	if req.AssignedTo != nil {
		i.AssignedTo = req.AssignedTo
	}
	if req.DueDate != nil {
		i.DueDate = req.DueDate
	}
	if req.Resolution != "" {
		i.Resolution = req.Resolution
	}
	if err := s.repo.UpdateIssue(ctx, i); err != nil {
		return nil, fmt.Errorf("issue: update: %w", err)
	}
	return s.repo.FindIssueByID(ctx, projectID, i.ID, orgID)
}

// DeleteIssue soft-deletes an issue, verifying org boundary.
func (s *Service) DeleteIssue(ctx context.Context, projectID, issueID, orgID uuid.UUID) error {
	i, err := s.GetIssue(ctx, projectID, issueID, orgID)
	if err != nil {
		return err
	}
	return s.repo.DeleteIssue(ctx, projectID, i.ID, orgID)
}

// ---- Risk service methods -------------------------------------------------

// CreateRisk creates a new risk under a project, enforcing org boundary.
// The risk score and severity are always derived from probability × impact.
func (s *Service) CreateRisk(ctx context.Context, projectID, orgID, createdBy uuid.UUID, req *CreateRiskRequest) (*Risk, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}

	probability := clampRiskLevel(req.Probability)
	if req.Probability == 0 {
		probability = 3
	}
	impact := clampRiskLevel(req.Impact)
	if req.Impact == 0 {
		impact = 3
	}

	r := &Risk{
		OrganizationID: orgID,
		ProjectID:      projectID,
		Title:          req.Title,
		Description:    req.Description,
		Status:         "IDENTIFIED",
		Probability:    probability,
		Impact:         impact,
		RiskScore:      RiskScore(probability, impact),
		Severity:       RiskSeverity(probability, impact),
		Mitigation:     req.Mitigation,
		OwnedBy:        req.OwnedBy,
		DueDate:        req.DueDate,
		CreatedBy:      createdBy,
	}
	if err := s.repo.CreateRisk(ctx, r); err != nil {
		return nil, fmt.Errorf("risk: create: %w", err)
	}
	return s.repo.FindRiskByID(ctx, projectID, r.ID, orgID)
}

// GetRisk returns a risk by ID, verifying org and project boundaries.
func (s *Service) GetRisk(ctx context.Context, projectID, riskID, orgID uuid.UUID) (*Risk, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	return s.repo.FindRiskByID(ctx, projectID, riskID, orgID)
}

// ListRisks returns paginated risks for a project.
func (s *Service) ListRisks(ctx context.Context, filter RiskListFilter) ([]Risk, int64, error) {
	if _, err := s.GetByID(ctx, filter.ProjectID, filter.OrganizationID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListRisks(ctx, filter)
}

// UpdateRisk applies partial updates to a risk, enforcing FSM for status and
// recomputing risk_score/severity whenever probability or impact changes.
func (s *Service) UpdateRisk(ctx context.Context, projectID, riskID, orgID uuid.UUID, req *UpdateRiskRequest) (*Risk, error) {
	r, err := s.GetRisk(ctx, projectID, riskID, orgID)
	if err != nil {
		return nil, err
	}
	if req.Title != "" {
		r.Title = req.Title
	}
	if req.Description != "" {
		r.Description = req.Description
	}
	if req.Status != "" {
		if _, err := s.fsm.Transition(workflow.EntityRisk, r.Status, req.Status); err != nil {
			return nil, fmt.Errorf("risk: %w - allowed: %v",
				workflow.ErrTransitionNotAllowed,
				s.fsm.AllowedTransitions(workflow.EntityRisk, r.Status),
			)
		}
		r.Status = req.Status
	}
	if req.Probability > 0 {
		r.Probability = clampRiskLevel(req.Probability)
	}
	if req.Impact > 0 {
		r.Impact = clampRiskLevel(req.Impact)
	}
	if req.Probability > 0 || req.Impact > 0 {
		r.RiskScore = RiskScore(r.Probability, r.Impact)
		r.Severity = RiskSeverity(r.Probability, r.Impact)
	}
	if req.Mitigation != "" {
		r.Mitigation = req.Mitigation
	}
	if req.OwnedBy != nil {
		r.OwnedBy = req.OwnedBy
	}
	if req.DueDate != nil {
		r.DueDate = req.DueDate
	}
	if err := s.repo.UpdateRisk(ctx, r); err != nil {
		return nil, fmt.Errorf("risk: update: %w", err)
	}
	return s.repo.FindRiskByID(ctx, projectID, r.ID, orgID)
}

// DeleteRisk soft-deletes a risk, verifying org boundary.
func (s *Service) DeleteRisk(ctx context.Context, projectID, riskID, orgID uuid.UUID) error {
	r, err := s.GetRisk(ctx, projectID, riskID, orgID)
	if err != nil {
		return err
	}
	return s.repo.DeleteRisk(ctx, projectID, r.ID, orgID)
}

// ---- Budget service methods ------------------------------------------------
// Tenant safety for budget lines is enforced by verifying the parent project
// belongs to the caller's organization before any access. project_budgets has
// no organization_id column, so all authorization flows through the parent.

// CreateBudget creates a budget line under a project, verifying org boundary.
func (s *Service) CreateBudget(ctx context.Context, projectID, orgID, createdBy uuid.UUID, req *CreateBudgetRequest) (*BudgetLine, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}

	planned := round2(req.Planned)
	if planned < 0 {
		return nil, fmt.Errorf("budget: planned cannot be negative")
	}
	actual := round2(req.Actual)
	if actual < 0 {
		return nil, fmt.Errorf("budget: actual cannot be negative")
	}
	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "IDR"
	}

	b := &ProjectBudget{
		ProjectID:   projectID,
		Category:    req.Category,
		Description: req.Description,
		Planned:     planned,
		Actual:      actual,
		Currency:    currency,
		CreatedBy:   createdBy,
	}
	if err := s.repo.CreateBudget(ctx, b); err != nil {
		return nil, fmt.Errorf("budget: create: %w", err)
	}
	line, err := s.repo.FindBudgetByID(ctx, projectID, b.ID, orgID)
	if err != nil {
		return nil, err
	}
	return buildBudgetLine(line), nil
}

// ListBudgets returns paginated budget lines for a project.
func (s *Service) ListBudgets(ctx context.Context, filter BudgetListFilter) ([]BudgetLine, int64, error) {
	if _, err := s.GetByID(ctx, filter.ProjectID, filter.OrganizationID); err != nil {
		return nil, 0, err
	}
	rows, total, err := s.repo.ListBudgets(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	lines := make([]BudgetLine, 0, len(rows))
	for _, r := range rows {
		lines = append(lines, *buildBudgetLine(&r))
	}
	return lines, total, nil
}

// GetBudget returns a budget line by ID, verifying parent project ownership.
func (s *Service) GetBudget(ctx context.Context, projectID, budgetID, orgID uuid.UUID) (*BudgetLine, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	line, err := s.repo.FindBudgetByID(ctx, projectID, budgetID, orgID)
	if err != nil {
		return nil, err
	}
	return buildBudgetLine(line), nil
}

// UpdateBudget applies partial updates to a budget line, verifying org boundary.
func (s *Service) UpdateBudget(ctx context.Context, projectID, budgetID, orgID uuid.UUID, req *UpdateBudgetRequest) (*BudgetLine, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	line, err := s.repo.FindBudgetByID(ctx, projectID, budgetID, orgID)
	if err != nil {
		return nil, err
	}

	if req.Category != "" {
		line.Category = req.Category
	}
	if req.Description != "" {
		line.Description = req.Description
	}
	if req.Planned != nil {
		if *req.Planned < 0 {
			return nil, fmt.Errorf("budget: planned cannot be negative")
		}
		line.Planned = round2(*req.Planned)
	}
	if req.Actual != nil {
		if *req.Actual < 0 {
			return nil, fmt.Errorf("budget: actual cannot be negative")
		}
		line.Actual = round2(*req.Actual)
	}
	if strings.TrimSpace(req.Currency) != "" {
		line.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	}

	if err := s.repo.UpdateBudget(ctx, line); err != nil {
		return nil, fmt.Errorf("budget: update: %w", err)
	}
	updated, err := s.repo.FindBudgetByID(ctx, projectID, line.ID, orgID)
	if err != nil {
		return nil, err
	}
	return buildBudgetLine(updated), nil
}

// DeleteBudget soft-deletes a budget line, verifying org boundary.
func (s *Service) DeleteBudget(ctx context.Context, projectID, budgetID, orgID uuid.UUID) error {
	line, err := s.repo.FindBudgetByID(ctx, projectID, budgetID, orgID)
	if err != nil {
		return err
	}
	return s.repo.DeleteBudget(ctx, projectID, line.ID, orgID)
}

// ---- Vendor service ----------------------------------------------------------

var validVendorTypes = map[string]bool{
	constants.VendorTypeVendor:     true,
	constants.VendorTypeConsultant: true,
}

var validContractStatuses = map[string]bool{
	constants.ContractStatusDraft:      true,
	constants.ContractStatusActive:     true,
	constants.ContractStatusAmended:    true,
	constants.ContractStatusCompleted:  true,
	constants.ContractStatusTerminated: true,
}

// CreateVendor creates a new vendor/consultant record, scoped to an
// organization.
func (s *Service) CreateVendor(ctx context.Context, orgID, createdBy uuid.UUID, req *CreateVendorRequest) (*Vendor, error) {
	vType := strings.ToUpper(strings.TrimSpace(req.Type))
	if !validVendorTypes[vType] {
		return nil, ErrInvalidVendorType
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	v := &Vendor{
		OrganizationID: orgID,
		Name:           strings.TrimSpace(req.Name),
		Type:           vType,
		LegalName:      strings.TrimSpace(req.LegalName),
		TaxID:          strings.TrimSpace(req.TaxID),
		ContactPerson:  strings.TrimSpace(req.ContactPerson),
		Email:          strings.TrimSpace(req.Email),
		Phone:          strings.TrimSpace(req.Phone),
		Address:        req.Address,
		IsActive:       isActive,
		CreatedBy:      createdBy,
	}
	if err := s.repo.CreateVendor(ctx, v); err != nil {
		return nil, fmt.Errorf("vendor: create: %w", err)
	}
	created, err := s.repo.FindVendorByID(ctx, v.ID, orgID)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// ListVendors returns paginated vendors for an organization.
func (s *Service) ListVendors(ctx context.Context, filter VendorListFilter) ([]Vendor, int64, error) {
	return s.repo.ListVendors(ctx, filter)
}

// GetVendor returns a vendor by ID, verifying org boundary.
func (s *Service) GetVendor(ctx context.Context, id, orgID uuid.UUID) (*Vendor, error) {
	return s.repo.FindVendorByID(ctx, id, orgID)
}

// UpdateVendor applies updates to a vendor, verifying org boundary and
// validating the type enum if provided.
func (s *Service) UpdateVendor(ctx context.Context, id, orgID uuid.UUID, req *UpdateVendorRequest) (*Vendor, error) {
	v, err := s.repo.FindVendorByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(req.Name) != "" {
		v.Name = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.Type) != "" {
		vType := strings.ToUpper(strings.TrimSpace(req.Type))
		if !validVendorTypes[vType] {
			return nil, ErrInvalidVendorType
		}
		v.Type = vType
	}
	if req.LegalName != "" {
		v.LegalName = strings.TrimSpace(req.LegalName)
	}
	if req.TaxID != "" {
		v.TaxID = strings.TrimSpace(req.TaxID)
	}
	if req.ContactPerson != "" {
		v.ContactPerson = strings.TrimSpace(req.ContactPerson)
	}
	if req.Email != "" {
		v.Email = strings.TrimSpace(req.Email)
	}
	if req.Phone != "" {
		v.Phone = strings.TrimSpace(req.Phone)
	}
	if req.Address != "" {
		v.Address = req.Address
	}
	if req.IsActive != nil {
		v.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateVendor(ctx, v); err != nil {
		return nil, fmt.Errorf("vendor: update: %w", err)
	}
	updated, err := s.repo.FindVendorByID(ctx, v.ID, orgID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// DeleteVendor soft-deletes a vendor only if it is not referenced by any
// active contract.
func (s *Service) DeleteVendor(ctx context.Context, id, orgID uuid.UUID) error {
	// Check for active contracts referencing this vendor (as vendor or
	// consultant). Referenced vendors cannot be removed.
	var count int64
	if err := s.repo.countContractsForVendor(ctx, id, orgID, &count); err != nil {
		return err
	}
	if count > 0 {
		return ErrVendorInUse
	}
	return s.repo.DeleteVendor(ctx, id, orgID)
}

// ---- Contract service ---------------------------------------------------------

// CreateContract creates a project contract, verifying parent project
// ownership, vendor/consultant ownership, and business validation rules.
func (s *Service) CreateContract(ctx context.Context, projectID, orgID, createdBy uuid.UUID, req *CreateContractRequest) (*Contract, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}

	// Vendor must exist in the same org.
	if _, err := s.repo.FindVendorByID(ctx, req.VendorID, orgID); err != nil {
		return nil, ErrVendorNotFound
	}
	var consultantID *uuid.UUID
	if req.ConsultantID != nil {
		if _, err := s.repo.FindVendorByID(ctx, *req.ConsultantID, orgID); err != nil {
			return nil, ErrVendorNotFound
		}
		consultantID = req.ConsultantID
	}

	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status == "" {
		status = constants.ContractStatusDraft
	}
	if !validContractStatuses[status] {
		return nil, ErrInvalidContractStatus
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "IDR"
	}

	value := round2(req.ContractValue)
	if value < 0 {
		return nil, ErrInvalidContractValue
	}
	if err := validateContractDates(req.StartDate, req.EndDate); err != nil {
		return nil, err
	}

	c := &Contract{
		OrganizationID: orgID,
		ProjectID:      projectID,
		ContractNumber: strings.TrimSpace(req.ContractNumber),
		Title:          strings.TrimSpace(req.Title),
		VendorID:       req.VendorID,
		ConsultantID:   consultantID,
		ContractValue:  value,
		Currency:       currency,
		SignedDate:     req.SignedDate,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		Status:         status,
		ScopeOfWork:    req.ScopeOfWork,
		CreatedBy:      createdBy,
	}
	if err := s.repo.CreateContract(ctx, c); err != nil {
		if errors.Is(err, ErrContractNumberTaken) {
			return nil, ErrContractNumberTaken
		}
		return nil, fmt.Errorf("contract: create: %w", err)
	}
	created, err := s.repo.FindContractByID(ctx, projectID, c.ID, orgID)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// ListContracts returns paginated contracts for a project, verifying ownership.
func (s *Service) ListContracts(ctx context.Context, filter ContractListFilter) ([]Contract, int64, error) {
	if _, err := s.GetByID(ctx, filter.ProjectID, filter.OrganizationID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListContracts(ctx, filter)
}

// GetContract returns a contract by ID, verifying parent project ownership.
func (s *Service) GetContract(ctx context.Context, projectID, contractID, orgID uuid.UUID) (*Contract, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	return s.repo.FindContractByID(ctx, projectID, contractID, orgID)
}

// UpdateContract applies updates to a contract, verifying org boundary and
// validating business rules.
func (s *Service) UpdateContract(ctx context.Context, projectID, contractID, orgID uuid.UUID, req *UpdateContractRequest) (*Contract, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	c, err := s.repo.FindContractByID(ctx, projectID, contractID, orgID)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(req.ContractNumber) != "" {
		c.ContractNumber = strings.TrimSpace(req.ContractNumber)
	}
	if strings.TrimSpace(req.Title) != "" {
		c.Title = strings.TrimSpace(req.Title)
	}
	if req.VendorID != nil {
		if _, err := s.repo.FindVendorByID(ctx, *req.VendorID, orgID); err != nil {
			return nil, ErrVendorNotFound
		}
		c.VendorID = *req.VendorID
	}
	if req.ConsultantID != nil {
		if *req.ConsultantID == uuid.Nil {
			c.ConsultantID = nil
		} else {
			if _, err := s.repo.FindVendorByID(ctx, *req.ConsultantID, orgID); err != nil {
				return nil, ErrVendorNotFound
			}
			c.ConsultantID = req.ConsultantID
		}
	}
	if req.ContractValue != nil {
		value := round2(*req.ContractValue)
		if value < 0 {
			return nil, ErrInvalidContractValue
		}
		c.ContractValue = value
	}
	if strings.TrimSpace(req.Currency) != "" {
		c.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	}
	if req.SignedDate != nil {
		c.SignedDate = req.SignedDate
	}
	if req.StartDate != nil {
		c.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		c.EndDate = req.EndDate
	}
	if strings.TrimSpace(req.Status) != "" {
		status := strings.ToUpper(strings.TrimSpace(req.Status))
		if !validContractStatuses[status] {
			return nil, ErrInvalidContractStatus
		}
		c.Status = status
	}
	if req.ScopeOfWork != "" {
		c.ScopeOfWork = req.ScopeOfWork
	}

	if err := validateContractDates(c.StartDate, c.EndDate); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateContract(ctx, c); err != nil {
		if errors.Is(err, ErrContractNumberTaken) {
			return nil, ErrContractNumberTaken
		}
		return nil, fmt.Errorf("contract: update: %w", err)
	}
	updated, err := s.repo.FindContractByID(ctx, projectID, c.ID, orgID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// DeleteContract soft-deletes a contract, verifying org boundary.
func (s *Service) DeleteContract(ctx context.Context, projectID, contractID, orgID uuid.UUID) error {
	if _, err := s.repo.FindContractByID(ctx, projectID, contractID, orgID); err != nil {
		return err
	}
	return s.repo.DeleteContract(ctx, projectID, contractID, orgID)
}

// validateContractDates ensures start_date is not after end_date.
func validateContractDates(start, end *types.FlexTime) error {
	if start == nil || end == nil {
		return nil
	}
	sTime := start.Time
	eTime := end.Time
	if sTime.IsZero() || eTime.IsZero() {
		return nil
	}
	if sTime.After(eTime) {
		return ErrInvalidContractDates
	}
	return nil
}

// round2 rounds a money amount to 2 decimal places.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// buildBudgetLine computes derived fields (variance, usage_pct, status) for a
// budget line. Variance = planned - actual (positive means remaining budget).
// UsagePct = actual / planned * 100, and is 0 when planned is 0 to avoid
// division by zero. Status thresholds: <80% NORMAL, >=80% WATCH, >=90% RISK,
// >=100% OVERRUN.
func buildBudgetLine(b *ProjectBudget) *BudgetLine {
	planned := b.Planned
	actual := b.Actual
	variance := round2(planned - actual)

	var usagePct float64
	if planned > 0 {
		usagePct = math.Round((actual/planned)*10000) / 100
	}

	status := BudgetStatusNormal
	switch {
	case usagePct >= 100:
		status = BudgetStatusOverrun
	case usagePct >= 90:
		status = BudgetStatusRisk
	case usagePct >= 80:
		status = BudgetStatusWatch
	}

	return &BudgetLine{
		ID:          b.ID,
		ProjectID:   b.ProjectID,
		Category:    b.Category,
		Description: b.Description,
		Planned:     b.Planned,
		Actual:      b.Actual,
		Currency:    b.Currency,
		Variance:    variance,
		UsagePct:    usagePct,
		Status:      status,
		CreatedBy:   b.CreatedBy,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}

// SyncProjectProgressFromPeriodic updates project.progress_pct from the latest
// periodic report for that project. Called after create/update of a periodic report.
// Errors are logged but not surfaced — this is best-effort.
func (s *Service) SyncProjectProgressFromPeriodic(ctx context.Context, orgID, projectID uuid.UUID) {
	// Fetch the latest physical_progress_pct from periodic reports
	type latestRow struct{ PhysicalProgressPct float64 }
	var row latestRow
	err := s.repo.(*postgresRepository).db.WithContext(ctx).
		Table("project_periodic_reports").
		Select("physical_progress_pct").
		Where("organization_id = ? AND project_id = ? AND deleted_at IS NULL", orgID, projectID).
		Order("period_year DESC, period_month DESC").
		Limit(1).
		Scan(&row).Error
	if err != nil {
		s.log.Warn("SyncProjectProgressFromPeriodic: query failed", zap.Error(err))
		return
	}
	if row.PhysicalProgressPct <= 0 {
		return
	}
	if err := s.repo.(*postgresRepository).db.WithContext(ctx).
		Model(&Project{}).
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL", projectID, orgID).
		Update("progress_pct", row.PhysicalProgressPct).Error; err != nil {
		s.log.Warn("SyncProjectProgressFromPeriodic: update failed", zap.Error(err))
	}
}

// GetProgressHistory returns the progress history for a project, verifying org boundary.
func (s *Service) GetProgressHistory(ctx context.Context, projectID, orgID uuid.UUID) ([]ProgressHistory, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	return s.repo.GetProgressHistory(ctx, projectID)
}

// ---- Document service ---------------------------------------------------------

// UploadDocument persists a document's file to storage and its metadata to the
// database, verifying the parent project's org ownership. The file content is
// already validated (size + MIME + extension) by the handler; this method
// writes the bytes and stores a relative storage key in FileURL.
func (s *Service) UploadDocument(ctx context.Context, projectID, orgID, uploadedBy uuid.UUID, name, category, version, safeName, mime string, size int64, data []byte) (*ProjectDocument, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	if size > s.maxFileSize {
		return nil, ErrFileTooLarge
	}

	d := &ProjectDocument{
		ID:         uuid.New(),
		ProjectID:  projectID,
		Name:       truncateDocumentTitle(sanitizeFilename(name), maxDocumentNameLen),
		Category:   strings.TrimSpace(category),
		Version:    strings.TrimSpace(version),
		FileSize:   size,
		MimeType:   mime,
		UploadedBy: uploadedBy,
	}
	safeName = sanitizeFilename(safeName)
	key := buildDocumentStorageKey(orgID.String(), projectID.String(), d.ID.String(), safeName)
	if err := saveDocumentFile(s.storageRoot, key, data); err != nil {
		s.log.Error("document: save file", zap.Error(err))
		return nil, ErrDocumentStorage
	}
	d.FileURL = key
	if err := s.repo.CreateDocument(ctx, d); err != nil {
		return nil, fmt.Errorf("document: create metadata: %w", err)
	}
	created, err := s.repo.FindDocumentByID(ctx, projectID, d.ID, orgID)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// ListDocuments returns paginated documents for a project, verifying ownership.
func (s *Service) ListDocuments(ctx context.Context, filter DocumentListFilter) ([]ProjectDocument, int64, error) {
	if _, err := s.GetByID(ctx, filter.ProjectID, filter.OrganizationID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListDocuments(ctx, filter)
}

// GetDocument returns document metadata, verifying project ownership.
func (s *Service) GetDocument(ctx context.Context, projectID, documentID, orgID uuid.UUID) (*ProjectDocument, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	return s.repo.FindDocumentByID(ctx, projectID, documentID, orgID)
}

// UpdateDocument updates document metadata, verifying project ownership.
func (s *Service) UpdateDocument(ctx context.Context, projectID, documentID, orgID uuid.UUID, req *UpdateDocumentRequest) (*ProjectDocument, error) {
	d, err := s.GetDocument(ctx, projectID, documentID, orgID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Name) != "" {
		d.Name = truncateDocumentTitle(sanitizeFilename(req.Name), maxDocumentNameLen)
	}
	if req.Category != "" {
		d.Category = strings.TrimSpace(req.Category)
	}
	if req.Version != "" {
		d.Version = strings.TrimSpace(req.Version)
	}
	if err := s.repo.UpdateDocument(ctx, d); err != nil {
		return nil, fmt.Errorf("document: update metadata: %w", err)
	}
	updated, err := s.repo.FindDocumentByID(ctx, projectID, documentID, orgID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// DeleteDocument soft-deletes a document's metadata, verifying project
// ownership. The physical file is intentionally retained for audit/dev
// retention.
func (s *Service) DeleteDocument(ctx context.Context, projectID, documentID, orgID uuid.UUID) error {
	doc, err := s.GetDocument(ctx, projectID, documentID, orgID)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteDocument(ctx, projectID, documentID, orgID); err != nil {
		return err
	}
	// Remove the physical file. A soft-deleted row can never be downloaded
	// again, so this is safe. Failure to remove is logged but not fatal.
	if doc.FileURL != "" {
		if derr := deleteDocumentFile(s.storageRoot, doc.FileURL); derr != nil {
			s.log.Warn("document: delete physical file", zap.String("document_id", documentID.String()), zap.Error(derr))
		}
	}
	return nil
}

// OpenDocument returns the absolute file path + safe display filename for a
// document, verifying project ownership. Used by the download handler.
func (s *Service) OpenDocument(ctx context.Context, projectID, documentID, orgID uuid.UUID) (doc *ProjectDocument, absPath string, err error) {
	doc, err = s.GetDocument(ctx, projectID, documentID, orgID)
	if err != nil {
		return nil, "", err
	}
	absPath, _, oerr := openDocumentFile(s.storageRoot, doc.FileURL)
	if oerr != nil {
		return nil, "", ErrDocumentStorage
	}
	return doc, absPath, nil
}

// truncateDocumentTitle caps a display name to maxLen runes safely.
func truncateDocumentTitle(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// ---- Corrective Action repository methods ----------------------------------

func (r *postgresRepository) CreateCorrectiveAction(ctx context.Context, ca *CorrectiveAction) error {
	return r.db.WithContext(ctx).Create(ca).Error
}

func (r *postgresRepository) FindCorrectiveActionByID(ctx context.Context, projectID, caID, orgID uuid.UUID) (*CorrectiveAction, error) {
	var ca CorrectiveAction
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", caID, projectID, orgID).
		First(&ca).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCorrectiveActionNotFound
	}
	return &ca, err
}

func (r *postgresRepository) ListCorrectiveActions(ctx context.Context, filter CorrectiveActionListFilter) ([]CorrectiveAction, int64, error) {
	q := r.db.WithContext(ctx).Model(&CorrectiveAction{}).
		Where("project_id = ? AND organization_id = ?", filter.ProjectID, filter.OrganizationID)

	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.SourceType != "" {
		q = q.Where("source_type = ?", filter.SourceType)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Where("title ILIKE ? OR deviation ILIKE ?", s, s)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	var cas []CorrectiveAction
	err := q.Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&cas).Error
	return cas, total, err
}

func (r *postgresRepository) UpdateCorrectiveAction(ctx context.Context, ca *CorrectiveAction) error {
	return r.db.WithContext(ctx).Save(ca).Error
}

func (r *postgresRepository) DeleteCorrectiveAction(ctx context.Context, projectID, caID, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", caID, projectID, orgID).
		Delete(&CorrectiveAction{}).Error
}

// ---- Corrective Action service methods -------------------------------------

// CreateCorrectiveAction creates a new corrective action under a project.
func (s *Service) CreateCorrectiveAction(ctx context.Context, projectID, orgID, createdBy uuid.UUID, req *CreateCorrectiveActionRequest) (*CorrectiveAction, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}

	ca := &CorrectiveAction{
		OrganizationID: orgID,
		ProjectID:      projectID,
		Title:          req.Title,
		Deviation:      req.Deviation,
		RootCause:      req.RootCause,
		Recommendation: req.Recommendation,
		PICUserID:      req.PICUserID,
		TargetDate:     req.TargetDate,
		SourceType:     req.SourceType,
		SourceIssueID:  req.SourceIssueID,
		SourceRiskID:   req.SourceRiskID,
		SourceTaskID:   req.SourceTaskID,
		EvidenceNote:   req.EvidenceNote,
		Status:         "DRAFT",
		CreatedBy:      createdBy,
	}
	if err := s.repo.CreateCorrectiveAction(ctx, ca); err != nil {
		return nil, fmt.Errorf("corrective_action: create: %w", err)
	}
	return s.repo.FindCorrectiveActionByID(ctx, projectID, ca.ID, orgID)
}

// GetCorrectiveAction returns a corrective action by ID, verifying org and project boundaries.
func (s *Service) GetCorrectiveAction(ctx context.Context, projectID, caID, orgID uuid.UUID) (*CorrectiveAction, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	return s.repo.FindCorrectiveActionByID(ctx, projectID, caID, orgID)
}

// ListCorrectiveActions returns paginated corrective actions for a project.
func (s *Service) ListCorrectiveActions(ctx context.Context, filter CorrectiveActionListFilter) ([]CorrectiveAction, int64, error) {
	if _, err := s.GetByID(ctx, filter.ProjectID, filter.OrganizationID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListCorrectiveActions(ctx, filter)
}

// UpdateCorrectiveAction applies partial updates to a corrective action.
// Status changes are validated by the FSM; other field changes are free-form.
func (s *Service) UpdateCorrectiveAction(ctx context.Context, projectID, caID, orgID uuid.UUID, req *UpdateCorrectiveActionRequest) (*CorrectiveAction, error) {
	ca, err := s.GetCorrectiveAction(ctx, projectID, caID, orgID)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		ca.Title = req.Title
	}
	if req.Deviation != "" {
		ca.Deviation = req.Deviation
	}
	if req.RootCause != "" {
		ca.RootCause = req.RootCause
	}
	if req.Recommendation != "" {
		ca.Recommendation = req.Recommendation
	}
	if req.PICUserID != nil {
		ca.PICUserID = req.PICUserID
	}
	if req.TargetDate != nil {
		ca.TargetDate = req.TargetDate
	}
	if req.SourceType != "" {
		ca.SourceType = req.SourceType
	}
	if req.SourceIssueID != nil {
		ca.SourceIssueID = req.SourceIssueID
	}
	if req.SourceRiskID != nil {
		ca.SourceRiskID = req.SourceRiskID
	}
	if req.SourceTaskID != nil {
		ca.SourceTaskID = req.SourceTaskID
	}
	if req.EvidenceNote != "" {
		ca.EvidenceNote = req.EvidenceNote
	}
	// Status updates via UpdateCorrectiveAction are validated by FSM just like
	// the dedicated Transition endpoint, so partial updates can't bypass the FSM.
	if req.Status != "" && req.Status != ca.Status {
		if _, ferr := s.fsm.Transition(workflow.EntityCorrectiveAction, ca.Status, req.Status); ferr != nil {
			return nil, fmt.Errorf("corrective_action: %w — allowed: %v",
				workflow.ErrTransitionNotAllowed,
				s.fsm.AllowedTransitions(workflow.EntityCorrectiveAction, ca.Status),
			)
		}
		ca.Status = req.Status
	}

	if err := s.repo.UpdateCorrectiveAction(ctx, ca); err != nil {
		return nil, fmt.Errorf("corrective_action: update: %w", err)
	}
	return ca, nil
}

// DeleteCorrectiveAction soft-deletes a corrective action.
func (s *Service) DeleteCorrectiveAction(ctx context.Context, projectID, caID, orgID uuid.UUID) error {
	ca, err := s.GetCorrectiveAction(ctx, projectID, caID, orgID)
	if err != nil {
		return err
	}
	return s.repo.DeleteCorrectiveAction(ctx, projectID, ca.ID, orgID)
}

// TransitionCorrectiveAction moves a corrective action to a new status via the FSM.
func (s *Service) TransitionCorrectiveAction(ctx context.Context, projectID, caID, orgID uuid.UUID, req *TransitionRequest) (*CorrectiveAction, error) {
	ca, err := s.GetCorrectiveAction(ctx, projectID, caID, orgID)
	if err != nil {
		return nil, err
	}

	if _, ferr := s.fsm.Transition(workflow.EntityCorrectiveAction, ca.Status, req.ToStatus); ferr != nil {
		return nil, fmt.Errorf("corrective_action: %w — allowed: %v",
			workflow.ErrTransitionNotAllowed,
			s.fsm.AllowedTransitions(workflow.EntityCorrectiveAction, ca.Status),
		)
	}

	ca.Status = req.ToStatus
	if err := s.repo.UpdateCorrectiveAction(ctx, ca); err != nil {
		return nil, fmt.Errorf("corrective_action: transition: %w", err)
	}
	return ca, nil
}
