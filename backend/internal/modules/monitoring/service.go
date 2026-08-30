package monitoring

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Models
// ---------------------------------------------------------------------------

// Baseline adalah rencana awal yang dibekukan sebagai acuan pengukuran.
type Baseline struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	ProjectID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	Version        int            `gorm:"default:1" json:"version"`
	Label          string         `gorm:"size:200" json:"label,omitempty"`
	ApprovedAt     *time.Time     `json:"approved_at,omitempty"`
	ApprovedBy     *uuid.UUID     `gorm:"type:uuid" json:"approved_by,omitempty"`
	PhysicalTarget float64        `gorm:"type:decimal(5,2);default:0" json:"physical_target"`
	BudgetTotal    float64        `gorm:"type:decimal(20,2);default:0" json:"budget_total"`
	Currency       string         `gorm:"size:10;default:'IDR'" json:"currency"`
	PlannedStart   time.Time      `gorm:"type:date" json:"planned_start"`
	PlannedEnd     time.Time      `gorm:"type:date" json:"planned_end"`
	Source         string         `gorm:"size:100" json:"source,omitempty"`
	Notes          string         `gorm:"type:text" json:"notes,omitempty"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	CreatedBy      *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Baseline) TableName() string { return "project_baselines" }

// Snapshot adalah rekaman realisasi pada satu cut-off period.
type Snapshot struct {
	ID                    uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID        uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	ProjectID             uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	BaselineID            *uuid.UUID     `gorm:"type:uuid" json:"baseline_id,omitempty"`
	PeriodYear            int            `gorm:"not null" json:"period_year"`
	PeriodMonth           int            `gorm:"not null" json:"period_month"`
	PeriodLabel           string         `gorm:"size:50" json:"period_label,omitempty"`
	PhysicalActual        float64        `gorm:"type:decimal(5,2);default:0" json:"physical_actual"`
	PhysicalTarget        float64        `gorm:"type:decimal(5,2);default:0" json:"physical_target"`
	PhysicalVariance      float64        `gorm:"type:decimal(6,2);->" json:"physical_variance"` // GENERATED
	FinancialActual       float64        `gorm:"type:decimal(20,2);default:0" json:"financial_actual"`
	FinancialTarget       float64        `gorm:"type:decimal(20,2);default:0" json:"financial_target"`
	FinancialVariance     float64        `gorm:"type:decimal(20,2);->" json:"financial_variance"` // GENERATED
	Currency              string         `gorm:"size:10;default:'IDR'" json:"currency"`
	ScheduleActualStart   *time.Time     `gorm:"type:date" json:"schedule_actual_start,omitempty"`
	ScheduleActualEnd     *time.Time     `gorm:"type:date" json:"schedule_actual_end,omitempty"`
	ScheduleDeviationDays *int           `json:"schedule_deviation_days,omitempty"`
	Status                string         `gorm:"size:20;default:'DRAFT'" json:"status"`
	SubmittedAt           *time.Time     `json:"submitted_at,omitempty"`
	SubmittedBy           *uuid.UUID     `gorm:"type:uuid" json:"submitted_by,omitempty"`
	ValidatedAt           *time.Time     `json:"validated_at,omitempty"`
	ValidatedBy           *uuid.UUID     `gorm:"type:uuid" json:"validated_by,omitempty"`
	RejectionReason       string         `gorm:"type:text" json:"rejection_reason,omitempty"`
	Source                string         `gorm:"size:100" json:"source,omitempty"`
	Notes                 string         `gorm:"type:text" json:"notes,omitempty"`
	CreatedBy             *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Snapshot) TableName() string { return "project_snapshots" }

// ---------------------------------------------------------------------------
// DTOs — Baseline
// ---------------------------------------------------------------------------

type CreateBaselineRequest struct {
	Version        int     `json:"version"`
	Label          string  `json:"label"`
	PhysicalTarget float64 `json:"physical_target" binding:"min=0,max=100"`
	BudgetTotal    float64 `json:"budget_total" binding:"min=0"`
	Currency       string  `json:"currency"`
	PlannedStart   string  `json:"planned_start" binding:"required"`
	PlannedEnd     string  `json:"planned_end" binding:"required"`
	Source         string  `json:"source"`
	Notes          string  `json:"notes"`
}

type UpdateBaselineRequest struct {
	Label          string  `json:"label"`
	PhysicalTarget float64 `json:"physical_target"`
	BudgetTotal    float64 `json:"budget_total"`
	PlannedStart   string  `json:"planned_start"`
	PlannedEnd     string  `json:"planned_end"`
	Source         string  `json:"source"`
	Notes          string  `json:"notes"`
	IsActive       *bool   `json:"is_active"`
}

// ---------------------------------------------------------------------------
// DTOs — Snapshot
// ---------------------------------------------------------------------------

type CreateSnapshotRequest struct {
	BaselineID            *string `json:"baseline_id"`
	PeriodYear            int     `json:"period_year" binding:"required,min=2000,max=2100"`
	PeriodMonth           int     `json:"period_month" binding:"required,min=1,max=12"`
	PhysicalActual        float64 `json:"physical_actual" binding:"min=0,max=100"`
	PhysicalTarget        float64 `json:"physical_target" binding:"min=0,max=100"`
	FinancialActual       float64 `json:"financial_actual" binding:"min=0"`
	FinancialTarget       float64 `json:"financial_target" binding:"min=0"`
	Currency              string  `json:"currency"`
	ScheduleDeviationDays *int    `json:"schedule_deviation_days"`
	Source                string  `json:"source"`
	Notes                 string  `json:"notes"`
}

type UpdateSnapshotRequest struct {
	PhysicalActual        float64 `json:"physical_actual"`
	PhysicalTarget        float64 `json:"physical_target"`
	FinancialActual       float64 `json:"financial_actual"`
	FinancialTarget       float64 `json:"financial_target"`
	ScheduleDeviationDays *int    `json:"schedule_deviation_days"`
	Source                string  `json:"source"`
	Notes                 string  `json:"notes"`
}

type SubmitSnapshotRequest struct {
	Status          string `json:"status" binding:"required,oneof=SUBMITTED VALID REJECTED STALE"`
	RejectionReason string `json:"rejection_reason"`
}

// ---------------------------------------------------------------------------
// Repository — Baseline
// ---------------------------------------------------------------------------

type BaselineRepository interface {
	List(orgID, projectID uuid.UUID) ([]Baseline, error)
	Get(orgID, projectID, id uuid.UUID) (*Baseline, error)
	Create(b *Baseline) error
	Update(b *Baseline) error
	Delete(orgID, projectID, id uuid.UUID) error
	DeactivateAll(projectID uuid.UUID) error
}

type baselineRepo struct{ db *gorm.DB }

func newBaselineRepo(db *gorm.DB) BaselineRepository { return &baselineRepo{db: db} }

func (r *baselineRepo) List(orgID, projectID uuid.UUID) ([]Baseline, error) {
	var list []Baseline
	err := r.db.Where("organization_id = ? AND project_id = ? AND deleted_at IS NULL", orgID, projectID).
		Order("version DESC").Find(&list).Error
	return list, err
}

func (r *baselineRepo) Get(orgID, projectID, id uuid.UUID) (*Baseline, error) {
	var b Baseline
	err := r.db.Where("organization_id = ? AND project_id = ? AND id = ? AND deleted_at IS NULL", orgID, projectID, id).
		First(&b).Error
	return &b, err
}

func (r *baselineRepo) Create(b *Baseline) error { return r.db.Create(b).Error }
func (r *baselineRepo) Update(b *Baseline) error { return r.db.Save(b).Error }

func (r *baselineRepo) Delete(orgID, projectID, id uuid.UUID) error {
	return r.db.Where("organization_id = ? AND project_id = ? AND id = ?", orgID, projectID, id).
		Delete(&Baseline{}).Error
}

func (r *baselineRepo) DeactivateAll(projectID uuid.UUID) error {
	return r.db.Model(&Baseline{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		Update("is_active", false).Error
}

// ---------------------------------------------------------------------------
// Repository — Snapshot
// ---------------------------------------------------------------------------

type SnapshotRepository interface {
	List(orgID, projectID uuid.UUID, status string) ([]Snapshot, error)
	Get(orgID, projectID, id uuid.UUID) (*Snapshot, error)
	Create(s *Snapshot) error
	Update(s *Snapshot) error
	Delete(orgID, projectID, id uuid.UUID) error
}

type snapshotRepo struct{ db *gorm.DB }

func newSnapshotRepo(db *gorm.DB) SnapshotRepository { return &snapshotRepo{db: db} }

func (r *snapshotRepo) List(orgID, projectID uuid.UUID, status string) ([]Snapshot, error) {
	var list []Snapshot
	q := r.db.Where("organization_id = ? AND project_id = ? AND deleted_at IS NULL", orgID, projectID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("period_year DESC, period_month DESC").Find(&list).Error
	return list, err
}

func (r *snapshotRepo) Get(orgID, projectID, id uuid.UUID) (*Snapshot, error) {
	var s Snapshot
	err := r.db.Where("organization_id = ? AND project_id = ? AND id = ? AND deleted_at IS NULL", orgID, projectID, id).
		First(&s).Error
	return &s, err
}

func (r *snapshotRepo) Create(s *Snapshot) error { return r.db.Create(s).Error }
func (r *snapshotRepo) Update(s *Snapshot) error { return r.db.Save(s).Error }

func (r *snapshotRepo) Delete(orgID, projectID, id uuid.UUID) error {
	return r.db.Where("organization_id = ? AND project_id = ? AND id = ?", orgID, projectID, id).
		Delete(&Snapshot{}).Error
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type Service struct {
	baseRepo BaselineRepository
	snapRepo SnapshotRepository
	log      *zap.Logger
}

func NewService(db *gorm.DB, log *zap.Logger) *Service {
	return &Service{
		baseRepo: newBaselineRepo(db),
		snapRepo: newSnapshotRepo(db),
		log:      log,
	}
}

// ── Baseline ──────────────────────────────────────────────────────────────

func (s *Service) ListBaselines(orgID, projectID uuid.UUID) ([]Baseline, error) {
	return s.baseRepo.List(orgID, projectID)
}

func (s *Service) GetBaseline(orgID, projectID, id uuid.UUID) (*Baseline, error) {
	return s.baseRepo.Get(orgID, projectID, id)
}

func (s *Service) CreateBaseline(orgID, projectID, createdBy uuid.UUID, req CreateBaselineRequest) (*Baseline, error) {
	start, err := time.Parse("2006-01-02", req.PlannedStart)
	if err != nil {
		return nil, fmt.Errorf("invalid planned_start: %w", err)
	}
	end, err := time.Parse("2006-01-02", req.PlannedEnd)
	if err != nil {
		return nil, fmt.Errorf("invalid planned_end: %w", err)
	}
	if end.Before(start) {
		return nil, errors.New("planned_end must be after planned_start")
	}

	cur := req.Currency
	if cur == "" {
		cur = "IDR"
	}
	ver := req.Version
	if ver == 0 {
		ver = 1
	}

	b := &Baseline{
		OrganizationID: orgID,
		ProjectID:      projectID,
		Version:        ver,
		Label:          req.Label,
		PhysicalTarget: req.PhysicalTarget,
		BudgetTotal:    req.BudgetTotal,
		Currency:       cur,
		PlannedStart:   start,
		PlannedEnd:     end,
		Source:         req.Source,
		Notes:          req.Notes,
		IsActive:       true,
		CreatedBy:      &createdBy,
	}
	if err := s.baseRepo.Create(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Service) UpdateBaseline(orgID, projectID, id uuid.UUID, req UpdateBaselineRequest) (*Baseline, error) {
	b, err := s.baseRepo.Get(orgID, projectID, id)
	if err != nil {
		return nil, err
	}
	if req.Label != "" {
		b.Label = req.Label
	}
	if req.PhysicalTarget != 0 {
		b.PhysicalTarget = req.PhysicalTarget
	}
	if req.BudgetTotal != 0 {
		b.BudgetTotal = req.BudgetTotal
	}
	if req.PlannedStart != "" {
		t, err := time.Parse("2006-01-02", req.PlannedStart)
		if err != nil {
			return nil, fmt.Errorf("invalid planned_start: %w", err)
		}
		b.PlannedStart = t
	}
	if req.PlannedEnd != "" {
		t, err := time.Parse("2006-01-02", req.PlannedEnd)
		if err != nil {
			return nil, fmt.Errorf("invalid planned_end: %w", err)
		}
		b.PlannedEnd = t
	}
	if req.Source != "" {
		b.Source = req.Source
	}
	b.Notes = req.Notes
	if req.IsActive != nil {
		if *req.IsActive {
			// Deactivate others before activating this one
			_ = s.baseRepo.DeactivateAll(projectID)
		}
		b.IsActive = *req.IsActive
	}
	if err := s.baseRepo.Update(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Service) DeleteBaseline(orgID, projectID, id uuid.UUID) error {
	return s.baseRepo.Delete(orgID, projectID, id)
}

// ── Snapshot ──────────────────────────────────────────────────────────────

func (s *Service) ListSnapshots(orgID, projectID uuid.UUID, status string) ([]Snapshot, error) {
	// Only return non-draft/non-rejected for dashboard cut-off consistency
	return s.snapRepo.List(orgID, projectID, status)
}

func (s *Service) GetSnapshot(orgID, projectID, id uuid.UUID) (*Snapshot, error) {
	return s.snapRepo.Get(orgID, projectID, id)
}

func (s *Service) CreateSnapshot(orgID, projectID, createdBy uuid.UUID, req CreateSnapshotRequest) (*Snapshot, error) {
	cur := req.Currency
	if cur == "" {
		cur = "IDR"
	}

	label := fmt.Sprintf("%s %d", monthName(req.PeriodMonth), req.PeriodYear)

	snap := &Snapshot{
		OrganizationID:        orgID,
		ProjectID:             projectID,
		PeriodYear:            req.PeriodYear,
		PeriodMonth:           req.PeriodMonth,
		PeriodLabel:           label,
		PhysicalActual:        req.PhysicalActual,
		PhysicalTarget:        req.PhysicalTarget,
		FinancialActual:       req.FinancialActual,
		FinancialTarget:       req.FinancialTarget,
		Currency:              cur,
		ScheduleDeviationDays: req.ScheduleDeviationDays,
		Source:                req.Source,
		Notes:                 req.Notes,
		Status:                "DRAFT",
		CreatedBy:             &createdBy,
	}

	if req.BaselineID != nil && *req.BaselineID != "" {
		bid, err := uuid.Parse(*req.BaselineID)
		if err != nil {
			return nil, fmt.Errorf("invalid baseline_id: %w", err)
		}
		snap.BaselineID = &bid
	}

	if err := s.snapRepo.Create(snap); err != nil {
		return nil, err
	}
	return snap, nil
}

func (s *Service) UpdateSnapshot(orgID, projectID, id uuid.UUID, req UpdateSnapshotRequest) (*Snapshot, error) {
	snap, err := s.snapRepo.Get(orgID, projectID, id)
	if err != nil {
		return nil, err
	}
	if snap.Status != "DRAFT" {
		return nil, errors.New("only DRAFT snapshots can be edited")
	}
	if req.PhysicalActual != 0 {
		snap.PhysicalActual = req.PhysicalActual
	}
	if req.PhysicalTarget != 0 {
		snap.PhysicalTarget = req.PhysicalTarget
	}
	if req.FinancialActual != 0 {
		snap.FinancialActual = req.FinancialActual
	}
	if req.FinancialTarget != 0 {
		snap.FinancialTarget = req.FinancialTarget
	}
	snap.ScheduleDeviationDays = req.ScheduleDeviationDays
	if req.Source != "" {
		snap.Source = req.Source
	}
	snap.Notes = req.Notes
	if err := s.snapRepo.Update(snap); err != nil {
		return nil, err
	}
	return snap, nil
}

func (s *Service) TransitionSnapshot(orgID, projectID, id, actorID uuid.UUID, req SubmitSnapshotRequest) (*Snapshot, error) {
	snap, err := s.snapRepo.Get(orgID, projectID, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	switch req.Status {
	case "SUBMITTED":
		if snap.Status != "DRAFT" {
			return nil, errors.New("only DRAFT can be submitted")
		}
		snap.Status = "SUBMITTED"
		snap.SubmittedAt = &now
		snap.SubmittedBy = &actorID
	case "VALID":
		if snap.Status != "SUBMITTED" {
			return nil, errors.New("only SUBMITTED can be validated")
		}
		snap.Status = "VALID"
		snap.ValidatedAt = &now
		snap.ValidatedBy = &actorID
	case "REJECTED":
		if snap.Status != "SUBMITTED" {
			return nil, errors.New("only SUBMITTED can be rejected")
		}
		snap.Status = "REJECTED"
		snap.ValidatedAt = &now
		snap.ValidatedBy = &actorID
		snap.RejectionReason = req.RejectionReason
	case "STALE":
		snap.Status = "STALE"
	default:
		return nil, fmt.Errorf("unsupported transition: %s", req.Status)
	}

	if err := s.snapRepo.Update(snap); err != nil {
		return nil, err
	}
	return snap, nil
}

func (s *Service) DeleteSnapshot(orgID, projectID, id uuid.UUID) error {
	snap, err := s.snapRepo.Get(orgID, projectID, id)
	if err != nil {
		return err
	}
	if snap.Status != "DRAFT" {
		return errors.New("only DRAFT snapshots can be deleted")
	}
	return s.snapRepo.Delete(orgID, projectID, id)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func monthName(m int) string {
	months := []string{"", "Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	if m < 1 || m > 12 {
		return fmt.Sprintf("M%02d", m)
	}
	return months[m]
}
