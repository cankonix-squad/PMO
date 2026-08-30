package report

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------------

var ErrReportNotFound = errors.New("report snapshot not found")

// ---------------------------------------------------------------------------
// Repository
// ---------------------------------------------------------------------------

type Repository interface {
	Create(snapshot *ReportSnapshot) error
	FindByID(orgID, id uuid.UUID) (*ReportSnapshot, error)
	List(orgID uuid.UUID, filter ListReportFilter) ([]ReportSnapshot, int64, error)
	Update(snapshot *ReportSnapshot) error
	Delete(orgID, id uuid.UUID) error
}

// ---------------------------------------------------------------------------
// postgresRepository
// ---------------------------------------------------------------------------

type postgresRepository struct {
	db *gorm.DB
}

func newPostgresRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(s *ReportSnapshot) error {
	return r.db.Create(s).Error
}

func (r *postgresRepository) FindByID(orgID, id uuid.UUID) (*ReportSnapshot, error) {
	var s ReportSnapshot
	err := r.db.
		Preload("CreatedByUser").
		Preload("Project").
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, orgID).
		First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrReportNotFound
	}
	return &s, err
}

func (r *postgresRepository) List(orgID uuid.UUID, f ListReportFilter) ([]ReportSnapshot, int64, error) {
	q := r.db.Model(&ReportSnapshot{}).
		Preload("CreatedByUser").
		Preload("Project").
		Where("organization_id = ? AND deleted_at IS NULL", orgID)

	if f.PeriodType != "" {
		q = q.Where("period_type = ?", f.PeriodType)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.ProjectID != "" {
		q = q.Where("project_id = ?", f.ProjectID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := f.Page
	if page < 1 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var snapshots []ReportSnapshot
	err := q.Order("period_start DESC").Offset(offset).Limit(pageSize).Find(&snapshots).Error
	return snapshots, total, err
}

func (r *postgresRepository) Update(s *ReportSnapshot) error {
	return r.db.Save(s).Error
}

func (r *postgresRepository) Delete(orgID, id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&ReportSnapshot{}).
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, orgID).
		Update("deleted_at", now).Error
}

// ---------------------------------------------------------------------------
// Metric computation helpers (reads live DB data for the given period)
// ---------------------------------------------------------------------------

func computeMetrics(db *gorm.DB, orgID uuid.UUID, projectID *uuid.UUID, start, end time.Time) (SnapshotMetrics, error) {
	var m SnapshotMetrics
	now := time.Now()

	// count helper: scan into int64 then assign
	count := func(q *gorm.DB) int {
		var n int64
		q.Count(&n)
		return int(n)
	}

	// ---- Projects ----
	projBase := db.Table("projects").Where("organization_id = ? AND deleted_at IS NULL", orgID)
	if projectID != nil {
		projBase = projBase.Where("id = ?", projectID)
	}
	m.TotalProjects = count(projBase)
	m.ActiveProjects = count(db.Table("projects").Where("organization_id = ? AND deleted_at IS NULL AND status IN ('ACTIVE','IN_PROGRESS')", orgID))
	m.DoneProjects = count(db.Table("projects").Where("organization_id = ? AND deleted_at IS NULL AND status = 'CLOSED'", orgID))

	// ---- Tasks ----
	taskBase := db.Table("tasks").Joins("JOIN projects p ON p.id = tasks.project_id").Where("p.organization_id = ? AND tasks.deleted_at IS NULL", orgID)
	if projectID != nil {
		taskBase = taskBase.Where("tasks.project_id = ?", projectID)
	}
	m.TotalTasks = count(taskBase)
	m.DoneTasks = count(db.Table("tasks").Joins("JOIN projects p ON p.id = tasks.project_id").Where("p.organization_id = ? AND tasks.deleted_at IS NULL AND tasks.status = 'DONE'", orgID))
	m.OverdueTasks = count(db.Table("tasks").Joins("JOIN projects p ON p.id = tasks.project_id").Where("p.organization_id = ? AND tasks.deleted_at IS NULL AND tasks.due_date < ? AND tasks.status != 'DONE'", orgID, now))

	var avgProgress *float64
	db.Table("tasks").Joins("JOIN projects p ON p.id = tasks.project_id").Where("p.organization_id = ? AND tasks.deleted_at IS NULL", orgID).Select("AVG(tasks.progress_pct)").Scan(&avgProgress)
	if avgProgress != nil {
		m.AvgProgressPct = roundTwo(*avgProgress)
	}

	// ---- Milestones ----
	m.TotalMilestones = count(db.Table("milestones").Joins("JOIN projects p ON p.id = milestones.project_id").Where("p.organization_id = ? AND milestones.deleted_at IS NULL", orgID))
	m.DoneMilestones = count(db.Table("milestones").Joins("JOIN projects p ON p.id = milestones.project_id").Where("p.organization_id = ? AND milestones.deleted_at IS NULL AND milestones.status = 'COMPLETED'", orgID))
	m.OverdueMilestones = count(db.Table("milestones").Joins("JOIN projects p ON p.id = milestones.project_id").Where("p.organization_id = ? AND milestones.deleted_at IS NULL AND milestones.due_date < ? AND milestones.status != 'COMPLETED'", orgID, now))

	// ---- Risks ----
	m.TotalRisks = count(db.Table("risks").Joins("JOIN projects p ON p.id = risks.project_id").Where("p.organization_id = ? AND risks.deleted_at IS NULL", orgID))
	m.OpenRisks = count(db.Table("risks").Joins("JOIN projects p ON p.id = risks.project_id").Where("p.organization_id = ? AND risks.deleted_at IS NULL AND risks.status NOT IN ('CLOSED','MITIGATED')", orgID))
	m.HighRisks = count(db.Table("risks").Joins("JOIN projects p ON p.id = risks.project_id").Where("p.organization_id = ? AND risks.deleted_at IS NULL AND risks.severity IN ('HIGH','CRITICAL')", orgID))

	// ---- Issues ----
	m.TotalIssues = count(db.Table("issues").Joins("JOIN projects p ON p.id = issues.project_id").Where("p.organization_id = ? AND issues.deleted_at IS NULL", orgID))
	m.OpenIssues = count(db.Table("issues").Joins("JOIN projects p ON p.id = issues.project_id").Where("p.organization_id = ? AND issues.deleted_at IS NULL AND issues.status NOT IN ('CLOSED','RESOLVED')", orgID))

	// ---- Budget ----
	type budgetRow struct {
		TotalPlanned float64
		TotalActual  float64
	}
	var br budgetRow
	db.Table("project_budgets pb").Joins("JOIN projects p ON p.id = pb.project_id").Where("p.organization_id = ? AND pb.deleted_at IS NULL", orgID).Select("COALESCE(SUM(pb.planned_amount),0) AS total_planned, COALESCE(SUM(pb.actual_amount),0) AS total_actual").Scan(&br)
	m.TotalPlannedBudget = roundTwo(br.TotalPlanned)
	m.TotalActualBudget = roundTwo(br.TotalActual)
	if br.TotalPlanned > 0 {
		m.BudgetUsagePct = roundTwo(br.TotalActual / br.TotalPlanned * 100)
	}

	// ---- Corrective Actions ----
	m.TotalCorrectiveActions = count(db.Table("corrective_actions ca").Joins("JOIN projects p ON p.id = ca.project_id").Where("p.organization_id = ? AND ca.deleted_at IS NULL", orgID))
	m.OpenCorrectiveActions = count(db.Table("corrective_actions ca").Joins("JOIN projects p ON p.id = ca.project_id").Where("p.organization_id = ? AND ca.deleted_at IS NULL AND ca.status NOT IN ('COMPLETED','REJECTED')", orgID))

	_ = start
	_ = end
	return m, nil
}

func roundTwo(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type Service struct {
	repo Repository
	db   *gorm.DB
	log  *zap.Logger
}

func NewService(db *gorm.DB, log *zap.Logger) *Service {
	return &Service{
		repo: newPostgresRepository(db),
		db:   db,
		log:  log,
	}
}

func (s *Service) GenerateReport(claims *auth.Claims, req GenerateReportRequest) (*ReportSnapshot, error) {
	start, err := time.Parse("2006-01-02", req.PeriodStart)
	if err != nil {
		return nil, fmt.Errorf("invalid period_start: %w", err)
	}
	end, err := time.Parse("2006-01-02", req.PeriodEnd)
	if err != nil {
		return nil, fmt.Errorf("invalid period_end: %w", err)
	}
	if end.Before(start) {
		return nil, fmt.Errorf("period_end must be after period_start")
	}

	orgID := claims.OrganizationID
	createdBy := claims.UserID

	var projectID *uuid.UUID
	if req.ProjectID != nil && *req.ProjectID != "" {
		pid, err := uuid.Parse(*req.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("invalid project_id: %w", err)
		}
		projectID = &pid
	}

	metrics, err := computeMetrics(s.db, orgID, projectID, start, end)
	if err != nil {
		return nil, fmt.Errorf("compute metrics: %w", err)
	}

	snapshot := &ReportSnapshot{
		OrganizationID: orgID,
		PeriodType:     req.PeriodType,
		PeriodLabel:    req.PeriodLabel,
		PeriodStart:    start,
		PeriodEnd:      end,
		ProjectID:      projectID,
		Metrics:        metrics,
		Status:         ReportStatusDraft,
		CreatedBy:      createdBy,
	}

	if err := s.repo.Create(snapshot); err != nil {
		return nil, fmt.Errorf("create snapshot: %w", err)
	}

	// reload with preloads
	return s.repo.FindByID(orgID, snapshot.ID)
}

func (s *Service) CreateReport(claims *auth.Claims, req CreateReportRequest) (*ReportSnapshot, error) {
	start, err := time.Parse("2006-01-02", req.PeriodStart)
	if err != nil {
		return nil, fmt.Errorf("invalid period_start: %w", err)
	}
	end, err := time.Parse("2006-01-02", req.PeriodEnd)
	if err != nil {
		return nil, fmt.Errorf("invalid period_end: %w", err)
	}

	orgID := claims.OrganizationID
	createdBy := claims.UserID

	var projectID *uuid.UUID
	if req.ProjectID != nil && *req.ProjectID != "" {
		pid, err := uuid.Parse(*req.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("invalid project_id: %w", err)
		}
		projectID = &pid
	}

	snapshot := &ReportSnapshot{
		OrganizationID:   orgID,
		PeriodType:       req.PeriodType,
		PeriodLabel:      req.PeriodLabel,
		PeriodStart:      start,
		PeriodEnd:        end,
		ProjectID:        projectID,
		ExecutiveSummary: req.ExecutiveSummary,
		Status:           ReportStatusDraft,
		CreatedBy:        createdBy,
	}

	if err := s.repo.Create(snapshot); err != nil {
		return nil, fmt.Errorf("create report: %w", err)
	}
	return s.repo.FindByID(orgID, snapshot.ID)
}

func (s *Service) GetReport(claims *auth.Claims, id uuid.UUID) (*ReportSnapshot, error) {
	return s.repo.FindByID(claims.OrganizationID, id)
}

func (s *Service) ListReports(claims *auth.Claims, filter ListReportFilter) ([]ReportSnapshot, int64, error) {
	return s.repo.List(claims.OrganizationID, filter)
}

func (s *Service) UpdateReport(claims *auth.Claims, id uuid.UUID, req UpdateReportRequest) (*ReportSnapshot, error) {
	orgID := claims.OrganizationID
	snap, err := s.repo.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	snap.ExecutiveSummary = req.ExecutiveSummary
	snap.UpdatedAt = time.Now()
	if err := s.repo.Update(snap); err != nil {
		return nil, err
	}
	return s.repo.FindByID(orgID, snap.ID)
}

func (s *Service) TransitionReport(claims *auth.Claims, id uuid.UUID, req TransitionReportRequest) (*ReportSnapshot, error) {
	orgID := claims.OrganizationID
	snap, err := s.repo.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	snap.Status = req.ToStatus
	snap.UpdatedAt = now

	if req.ToStatus == ReportStatusPublished {
		snap.PublishedAt = &now
		userID := claims.UserID
		snap.PublishedBy = &userID
	}

	if err := s.repo.Update(snap); err != nil {
		return nil, err
	}
	return s.repo.FindByID(orgID, snap.ID)
}

func (s *Service) DeleteReport(claims *auth.Claims, id uuid.UUID) error {
	orgID := claims.OrganizationID
	if _, err := s.repo.FindByID(orgID, id); err != nil {
		return err
	}
	return s.repo.Delete(orgID, id)
}
