package kpbu

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	notification "github.com/harmanto-49/cankora/internal/notifications/delivery"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// K3 — service KPIs. A KPI has a target and a direction; a VERIFIED
// measurement that misses the target suggests a deduction (deduction_pct of
// that period's scheduled availability payment). The suggestion is never
// applied automatically: the PJPK enters the deduction on the payment row.

const (
	KPIHigherBetter = "HIGHER_BETTER"
	KPILowerBetter  = "LOWER_BETTER"
	MeasureInternal = "INTERNAL"
	MeasurePortal   = "PORTAL"
)

var ErrKPICodeTaken = invalid("a KPI with this code already exists in the project")
var ErrMeasureExists = invalid("a measurement for this KPI and period already exists")
var ErrMeasureDecided = invalid("measurement has already been reviewed")

type KPI struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	KPBUProjectID  uuid.UUID      `gorm:"column:kpbu_project_id;type:uuid;not null" json:"kpbu_project_id"`
	Code           string         `json:"code"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Unit           string         `json:"unit"`
	TargetValue    float64        `json:"target_value"`
	Direction      string         `json:"direction"`
	DeductionPct   float64        `json:"deduction_pct"`
	IsActive       bool           `json:"is_active"`
	SortOrder      int            `json:"sort_order"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (KPI) TableName() string { return "kpbu_kpis" }

type KPIRequest struct {
	Code         string  `json:"code" binding:"required,max=30"`
	Name         string  `json:"name" binding:"required,max=300"`
	Description  string  `json:"description"`
	Unit         string  `json:"unit" binding:"max=50"`
	TargetValue  float64 `json:"target_value"`
	Direction    string  `json:"direction"`
	DeductionPct float64 `json:"deduction_pct" binding:"min=0,max=100"`
	IsActive     *bool   `json:"is_active"`
	SortOrder    int     `json:"sort_order"`
}

type KPIMeasurement struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	KPBUProjectID  uuid.UUID      `gorm:"column:kpbu_project_id;type:uuid;not null" json:"kpbu_project_id"`
	KPIID          uuid.UUID      `gorm:"column:kpi_id;type:uuid;not null" json:"kpi_id"`
	PeriodYear     int            `json:"period_year"`
	PeriodMonth    int            `json:"period_month"`
	ActualValue    float64        `json:"actual_value"`
	Notes          string         `json:"notes"`
	Source         string         `json:"source"`
	Status         string         `json:"status"`
	SubmittedBy    *uuid.UUID     `gorm:"type:uuid" json:"submitted_by"`
	SubmittedAt    time.Time      `json:"submitted_at"`
	ReviewedBy     *uuid.UUID     `gorm:"type:uuid" json:"reviewed_by"`
	ReviewedAt     *time.Time     `json:"reviewed_at"`
	ReviewNotes    string         `json:"review_notes"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Computed.
	KPICode string `gorm:"-" json:"kpi_code,omitempty"`
	KPIName string `gorm:"-" json:"kpi_name,omitempty"`
	Met     *bool  `gorm:"-" json:"met"`
}

func (KPIMeasurement) TableName() string { return "kpbu_kpi_measurements" }

type MeasurementRequest struct {
	KPIID       uuid.UUID `json:"kpi_id" binding:"required"`
	PeriodYear  int       `json:"period_year" binding:"required,min=2000,max=2100"`
	PeriodMonth int       `json:"period_month" binding:"required,min=1,max=12"`
	ActualValue *float64  `json:"actual_value" binding:"required"`
	Notes       string    `json:"notes"`
}

type ScorecardRow struct {
	KPI         KPI             `json:"kpi"`
	Measurement *KPIMeasurement `json:"measurement"`
	Status      string          `json:"status"` // MET | MISSED | PENDING | NO_DATA
}

type Scorecard struct {
	PeriodYear            int            `json:"period_year"`
	PeriodMonth           int            `json:"period_month"`
	Rows                  []ScorecardRow `json:"rows"`
	Met                   int            `json:"met"`
	Missed                int            `json:"missed"`
	Pending               int            `json:"pending"`
	NoData                int            `json:"no_data"`
	SuggestedDeductionPct float64        `json:"suggested_deduction_pct"`
	ScheduledPayment      *float64       `json:"scheduled_payment"`
	SuggestedDeduction    *float64       `json:"suggested_deduction"`
	RecordedDeduction     *float64       `json:"recorded_deduction"`
}

// kpiMet reports whether actual meets the target in the KPI's direction.
func kpiMet(k KPI, actual float64) bool {
	if k.Direction == KPILowerBetter {
		return actual <= k.TargetValue
	}
	return actual >= k.TargetValue
}

func normalizeKPI(r *KPIRequest) error {
	r.Code = strings.ToUpper(strings.TrimSpace(r.Code))
	r.Name = strings.TrimSpace(r.Name)
	r.Direction = upperOr(r.Direction, KPIHigherBetter)
	if r.Code == "" || r.Name == "" {
		return invalid("code and name are required")
	}
	if r.Direction != KPIHigherBetter && r.Direction != KPILowerBetter {
		return invalid("direction must be HIGHER_BETTER or LOWER_BETTER")
	}
	if r.DeductionPct < 0 || r.DeductionPct > 100 {
		return invalid("deduction_pct must be between 0 and 100")
	}
	return nil
}

// ── KPI definitions ─────────────────────────────────────────────────────────

func (s *Service) ListKPIs(orgID, projectID uuid.UUID, activeOnly bool) ([]KPI, error) {
	if _, err := s.findProject(orgID, projectID); err != nil {
		return nil, err
	}
	q := s.db.Where("organization_id = ? AND kpbu_project_id = ?", orgID, projectID)
	if activeOnly {
		q = q.Where("is_active")
	}
	var list []KPI
	return list, q.Order("sort_order ASC, code ASC").Find(&list).Error
}

func (s *Service) findKPI(orgID, projectID, id uuid.UUID) (*KPI, error) {
	var k KPI
	err := s.db.Where("id = ? AND organization_id = ? AND kpbu_project_id = ?", id, orgID, projectID).First(&k).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &k, err
}

func (s *Service) SaveKPI(orgID, projectID uuid.UUID, id *uuid.UUID, r KPIRequest) (*KPI, error) {
	if err := normalizeKPI(&r); err != nil {
		return nil, err
	}
	if _, err := s.findProject(orgID, projectID); err != nil {
		return nil, err
	}
	k := &KPI{OrganizationID: orgID, KPBUProjectID: projectID, IsActive: true}
	if id != nil {
		var err error
		if k, err = s.findKPI(orgID, projectID, *id); err != nil {
			return nil, err
		}
	}
	k.Code, k.Name, k.Description, k.Unit = r.Code, r.Name, strings.TrimSpace(r.Description), strings.TrimSpace(r.Unit)
	k.TargetValue, k.Direction, k.DeductionPct, k.SortOrder = r.TargetValue, r.Direction, r.DeductionPct, r.SortOrder
	if r.IsActive != nil {
		k.IsActive = *r.IsActive
	}
	if err := s.db.Save(k).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, ErrKPICodeTaken
		}
		return nil, err
	}
	return k, nil
}

// DeleteKPI soft-deletes a KPI and its measurements.
func (s *Service) DeleteKPI(orgID, projectID, id uuid.UUID) (*KPI, error) {
	k, err := s.findKPI(orgID, projectID, id)
	if err != nil {
		return nil, err
	}
	return k, s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("kpi_id = ? AND organization_id = ?", k.ID, orgID).Delete(&KPIMeasurement{}).Error; err != nil {
			return err
		}
		return tx.Delete(k).Error
	})
}

// ── Measurements ────────────────────────────────────────────────────────────

func (s *Service) enrichMeasurements(orgID, projectID uuid.UUID, list []KPIMeasurement) error {
	var kpis []KPI
	if err := s.db.Unscoped().Where("organization_id = ? AND kpbu_project_id = ?", orgID, projectID).Find(&kpis).Error; err != nil {
		return err
	}
	byID := map[uuid.UUID]KPI{}
	for _, k := range kpis {
		byID[k.ID] = k
	}
	for i := range list {
		m := &list[i]
		if k, ok := byID[m.KPIID]; ok {
			m.KPICode, m.KPIName = k.Code, k.Name
			met := kpiMet(k, m.ActualValue)
			m.Met = &met
		}
	}
	return nil
}

// ListMeasurements lists measurements, optionally filtered by status/period.
func (s *Service) ListMeasurements(orgID, projectID uuid.UUID, status string, year, month int) ([]KPIMeasurement, error) {
	if _, err := s.findProject(orgID, projectID); err != nil {
		return nil, err
	}
	q := s.db.Where("organization_id = ? AND kpbu_project_id = ?", orgID, projectID)
	if status != "" {
		q = q.Where("status = ?", strings.ToUpper(status))
	}
	if year > 0 {
		q = q.Where("period_year = ?", year)
	}
	if month > 0 {
		q = q.Where("period_month = ?", month)
	}
	var list []KPIMeasurement
	if err := q.Order("period_year DESC, period_month DESC, submitted_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, s.enrichMeasurements(orgID, projectID, list)
}

// CreateMeasurement records a KPI value. Internal (PJPK) entries are verified
// at once by the entering user; portal (BUP) entries wait for review.
func (s *Service) CreateMeasurement(orgID, userID, projectID uuid.UUID, source string, r MeasurementRequest) (*KPIMeasurement, error) {
	k, err := s.findKPI(orgID, projectID, r.KPIID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, invalid("kpi_id not found in this project")
		}
		return nil, err
	}
	if !k.IsActive {
		return nil, invalid("KPI is inactive")
	}
	now := s.now()
	if r.PeriodYear > now.Year() || (r.PeriodYear == now.Year() && r.PeriodMonth > int(now.Month())) {
		return nil, invalid("period cannot be in the future")
	}
	uid := userID
	m := KPIMeasurement{
		OrganizationID: orgID, KPBUProjectID: projectID, KPIID: k.ID,
		PeriodYear: r.PeriodYear, PeriodMonth: r.PeriodMonth, ActualValue: *r.ActualValue,
		Notes: strings.TrimSpace(r.Notes), Source: source, Status: DocSubmitted,
		SubmittedBy: &uid, SubmittedAt: now,
	}
	if source == MeasureInternal {
		m.Status, m.ReviewedBy, m.ReviewedAt = "VERIFIED", &uid, &now
	}
	if err := s.db.Create(&m).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, ErrMeasureExists
		}
		return nil, err
	}
	met := kpiMet(*k, m.ActualValue)
	m.KPICode, m.KPIName, m.Met = k.Code, k.Name, &met
	return &m, nil
}

// ReviewMeasurement verifies or rejects a SUBMITTED measurement.
func (s *Service) ReviewMeasurement(orgID, reviewerID, projectID, id uuid.UUID, req ReviewRequest) (*KPIMeasurement, error) {
	decision := strings.ToUpper(strings.TrimSpace(req.Decision))
	if decision == "APPROVED" {
		decision = "VERIFIED"
	}
	notes := strings.TrimSpace(req.Notes)
	if decision != "VERIFIED" && decision != "REJECTED" {
		return nil, invalid("decision must be VERIFIED or REJECTED")
	}
	if decision == "REJECTED" && notes == "" {
		return nil, invalid("notes are required when rejecting")
	}
	var m KPIMeasurement
	err := s.db.Where("id = ? AND organization_id = ? AND kpbu_project_id = ?", id, orgID, projectID).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	now := s.now()
	res := s.db.Model(&KPIMeasurement{}).Where("id = ? AND status = ?", m.ID, DocSubmitted).
		Updates(map[string]interface{}{"status": decision, "review_notes": notes, "reviewed_by": reviewerID, "reviewed_at": now, "updated_at": now})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrMeasureDecided
	}
	m.Status, m.ReviewNotes, m.ReviewedBy, m.ReviewedAt = decision, notes, &reviewerID, &now
	one := []KPIMeasurement{m}
	if err := s.enrichMeasurements(orgID, projectID, one); err != nil {
		return nil, err
	}
	return &one[0], nil
}

// ── Scorecard ───────────────────────────────────────────────────────────────

// buildScorecard is pure: rows per active KPI with the period's live
// measurement; only VERIFIED misses add to the suggested deduction (cap 100%).
func buildScorecard(year, month int, kpis []KPI, ms []KPIMeasurement) *Scorecard {
	sc := &Scorecard{PeriodYear: year, PeriodMonth: month, Rows: []ScorecardRow{}}
	byKPI := map[uuid.UUID]*KPIMeasurement{}
	for i := range ms {
		m := &ms[i]
		if m.Status == "REJECTED" {
			continue
		}
		byKPI[m.KPIID] = m
	}
	for _, k := range kpis {
		row := ScorecardRow{KPI: k, Measurement: byKPI[k.ID], Status: "NO_DATA"}
		switch {
		case row.Measurement == nil:
			sc.NoData++
		case row.Measurement.Status != "VERIFIED":
			row.Status = "PENDING"
			sc.Pending++
		case kpiMet(k, row.Measurement.ActualValue):
			row.Status = "MET"
			sc.Met++
		default:
			row.Status = "MISSED"
			sc.Missed++
			sc.SuggestedDeductionPct += k.DeductionPct
		}
		sc.Rows = append(sc.Rows, row)
	}
	sc.SuggestedDeductionPct = math.Min(100, round2(sc.SuggestedDeductionPct))
	return sc
}

func (s *Service) KPIScorecard(orgID, projectID uuid.UUID, year, month int) (*Scorecard, error) {
	if month < 1 || month > 12 || year < 2000 || year > 2100 {
		return nil, invalid("valid year and month are required")
	}
	kpis, err := s.ListKPIs(orgID, projectID, true)
	if err != nil {
		return nil, err
	}
	ms, err := s.ListMeasurements(orgID, projectID, "", year, month)
	if err != nil {
		return nil, err
	}
	sc := buildScorecard(year, month, kpis, ms)
	var ap AvailabilityPayment
	err = s.db.Where("organization_id = ? AND kpbu_project_id = ? AND period_year = ? AND period_month = ?", orgID, projectID, year, month).
		Take(&ap).Error
	if err == nil {
		sched, rec := ap.ScheduledAmount, ap.DeductionAmount
		sugg := round2(sched * sc.SuggestedDeductionPct / 100)
		sc.ScheduledPayment, sc.SuggestedDeduction, sc.RecordedDeduction = &sched, &sugg, &rec
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return sc, nil
}

// ── Notifications (optional) ────────────────────────────────────────────────

// Notifier is the subset of the notification service used by KPBU.
type Notifier interface {
	Enqueue(req notification.EnqueueRequest) error
}

func (s *Service) SetNotifier(n Notifier) { s.notifier = n }

func (s *Service) notify(orgID uuid.UUID, to []uuid.UUID, priority, subject, body, sourceType string, sourceID uuid.UUID) {
	if s.notifier == nil {
		return
	}
	seen := map[uuid.UUID]bool{}
	for _, uid := range to {
		if uid == uuid.Nil || seen[uid] {
			continue
		}
		seen[uid] = true
		u := uid
		if err := s.notifier.Enqueue(notification.EnqueueRequest{
			OrganizationID: orgID, RecipientUserID: &u, Priority: priority,
			Subject: subject, Body: body, SourceType: sourceType, SourceID: sourceID.String(),
		}); err != nil && s.log != nil {
			s.log.Warn("kpbu notification failed", zap.Error(err))
		}
	}
}

// internalReviewers are active SUPER_ADMIN/ADMIN/PMO users of the organization.
func (s *Service) internalReviewers(orgID uuid.UUID) []uuid.UUID {
	var ids []uuid.UUID
	s.db.Table("users u").Distinct("u.id").
		Joins("JOIN user_roles ur ON ur.user_id = u.id").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Where("u.organization_id = ? AND u.is_active = TRUE AND u.deleted_at IS NULL AND r.code IN ?", orgID, []string{"SUPER_ADMIN", "ADMIN", "PMO"}).
		Pluck("u.id", &ids)
	return ids
}

func monthYear(year, month int) string {
	months := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	if month < 1 || month > 12 {
		return fmt.Sprintf("%d", year)
	}
	return fmt.Sprintf("%s %d", months[month-1], year)
}
