package project

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PeriodicReport is the official periodic progress & financial input for dashboard trend.
// Classification: OPERATIONAL (laporan periodik operasional — belum official-governed).
type PeriodicReport struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID      uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	ProjectID           uuid.UUID      `gorm:"type:uuid;not null" json:"project_id"`
	PeriodYear          int            `gorm:"not null" json:"period_year"`
	PeriodMonth         int            `gorm:"not null" json:"period_month"`
	PhysicalProgressPct float64        `gorm:"type:numeric(5,2);not null;default:0" json:"physical_progress_pct"`
	FinancialPlanned    float64        `gorm:"type:numeric(20,2);not null;default:0" json:"financial_planned"`
	FinancialActual     float64        `gorm:"type:numeric(20,2);not null;default:0" json:"financial_actual"`
	FinancialPct        float64        `gorm:"type:numeric(8,4);not null;default:0" json:"financial_pct"` // backend-computed
	Notes               string         `gorm:"type:text" json:"notes,omitempty"`
	ReportedBy          *uuid.UUID     `gorm:"type:uuid" json:"reported_by,omitempty"`
	ReportedAt          time.Time      `gorm:"not null;default:now()" json:"reported_at"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName sets the GORM table name.
func (PeriodicReport) TableName() string { return "project_periodic_reports" }

// computeFinancialPct computes financial_pct safely (0 if planned is 0).
func computeFinancialPct(planned, actual float64) float64 {
	if planned <= 0 {
		return 0
	}
	return (actual / planned) * 100
}

// PeriodicReportListFilter holds query parameters for listing periodic reports.
type PeriodicReportListFilter struct {
	OrganizationID uuid.UUID
	ProjectID      uuid.UUID
	Year           int
	Month          int
	Page           int
	PageSize       int
}

// CreatePeriodicReportRequest is the input payload for creating a periodic report.
type CreatePeriodicReportRequest struct {
	PeriodYear          int     `json:"period_year" binding:"required,min=2000,max=2100"`
	PeriodMonth         int     `json:"period_month" binding:"required,min=1,max=12"`
	PhysicalProgressPct float64 `json:"physical_progress_pct" binding:"min=0,max=100"`
	FinancialPlanned    float64 `json:"financial_planned" binding:"min=0"`
	FinancialActual     float64 `json:"financial_actual" binding:"min=0"`
	Notes               string  `json:"notes"`
	ReportedAt          string  `json:"reported_at"` // optional ISO8601; defaults to now
}

// UpdatePeriodicReportRequest is the input payload for updating a periodic report.
type UpdatePeriodicReportRequest struct {
	PhysicalProgressPct *float64 `json:"physical_progress_pct"`
	FinancialPlanned    *float64 `json:"financial_planned"`
	FinancialActual     *float64 `json:"financial_actual"`
	Notes               *string  `json:"notes"`
	ReportedAt          *string  `json:"reported_at"`
}

// ErrPeriodicReportNotFound is returned when a report lookup yields no result.
var ErrPeriodicReportNotFound = errors.New("periodic report not found")

// ErrPeriodicReportDuplicate is returned when a report for the same org+project+year+month already exists.
var ErrPeriodicReportDuplicate = errors.New("periodic report already exists for this period")

// periodicReportRepository defines data access for periodic reports.
// Implemented directly on gorm.DB for simplicity (consistent with other modules like dashboard/auditlog).

// ListPeriodicReports returns paginated periodic reports for a project.
func listPeriodicReports(ctx context.Context, db *gorm.DB, filter PeriodicReportListFilter) ([]PeriodicReport, int64, error) {
	q := db.WithContext(ctx).
		Where("organization_id = ? AND project_id = ? AND deleted_at IS NULL", filter.OrganizationID, filter.ProjectID)
	if filter.Year > 0 {
		q = q.Where("period_year = ?", filter.Year)
	}
	if filter.Month > 0 {
		q = q.Where("period_month = ?", filter.Month)
	}

	var total int64
	if err := q.Model(&PeriodicReport{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var reports []PeriodicReport
	if err := q.
		Order("period_year DESC, period_month DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&reports).Error; err != nil {
		return nil, 0, err
	}
	return reports, total, nil
}

// getPeriodicReport fetches a single report by ID with tenant+project guard.
func getPeriodicReport(ctx context.Context, db *gorm.DB, orgID, projectID, reportID uuid.UUID) (*PeriodicReport, error) {
	var r PeriodicReport
	err := db.WithContext(ctx).
		Where("id = ? AND organization_id = ? AND project_id = ? AND deleted_at IS NULL",
			reportID, orgID, projectID).
		First(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPeriodicReportNotFound
	}
	return &r, err
}

// createPeriodicReport inserts a new periodic report row.
func createPeriodicReport(ctx context.Context, db *gorm.DB, r *PeriodicReport) error {
	return db.WithContext(ctx).Create(r).Error
}

// updatePeriodicReport saves changes to a periodic report row.
func updatePeriodicReport(ctx context.Context, db *gorm.DB, r *PeriodicReport) error {
	return db.WithContext(ctx).Save(r).Error
}

// softDeletePeriodicReport sets deleted_at on the periodic report.
func softDeletePeriodicReport(ctx context.Context, db *gorm.DB, orgID, projectID, reportID uuid.UUID) error {
	res := db.WithContext(ctx).
		Where("id = ? AND organization_id = ? AND project_id = ? AND deleted_at IS NULL",
			reportID, orgID, projectID).
		Delete(&PeriodicReport{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrPeriodicReportNotFound
	}
	return nil
}

// latestProjectProgressFromPeriodic returns the latest physical_progress_pct for a project from its periodic reports.
func latestProjectProgressFromPeriodic(ctx context.Context, db *gorm.DB, orgID, projectID uuid.UUID) (float64, bool, error) {
	type row struct{ PhysicalProgressPct float64 }
	var r row
	err := db.WithContext(ctx).
		Table("project_periodic_reports").
		Select("physical_progress_pct").
		Where("organization_id = ? AND project_id = ? AND deleted_at IS NULL", orgID, projectID).
		Order("period_year DESC, period_month DESC").
		Limit(1).
		Scan(&r).Error
	if err != nil {
		return 0, false, err
	}
	if r.PhysicalProgressPct == 0 {
		// Distinguish "no row" vs "row with zero"
		var cnt int64
		db.WithContext(ctx).
			Table("project_periodic_reports").
			Where("organization_id = ? AND project_id = ? AND deleted_at IS NULL", orgID, projectID).
			Count(&cnt)
		if cnt == 0 {
			return 0, false, nil
		}
	}
	return r.PhysicalProgressPct, true, nil
}
