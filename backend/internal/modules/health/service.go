package health

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/audit"
	"gorm.io/gorm"
	"math"
	"time"
)

var ErrNotFound = errors.New("health resource not found")

type component struct {
	Score     float64 `json:"score"`
	Weight    float64 `json:"weight"`
	Available bool    `json:"available"`
	Reason    string  `json:"reason"`
}
type Service struct {
	db    *gorm.DB
	audit *audit.Writer
}

func NewService(db *gorm.DB, audit *audit.Writer) *Service { return &Service{db: db, audit: audit} }

func (s *Service) ListFormulas(orgID uuid.UUID) ([]Formula, error) {
	var v []Formula
	err := s.db.Where("organization_id = ?", orgID).Order("version DESC").Find(&v).Error
	return v, err
}
func (s *Service) CreateFormula(orgID, actor uuid.UUID, req CreateFormulaRequest) (*Formula, error) {
	if req.Version <= 0 {
		return nil, errors.New("version must be positive")
	}
	if len(req.Weights) != 8 {
		return nil, errors.New("weights must define all 8 dimensions")
	}
	if req.MissingDataRule == "" {
		req.MissingDataRule = "PENALIZE"
	}
	wb, _ := json.Marshal(req.Weights)
	tb, _ := json.Marshal(req.Thresholds)
	f := &Formula{ID: uuid.New(), OrganizationID: orgID, Version: req.Version, Status: "DRAFT", Weights: wb, Thresholds: tb, MissingDataRule: req.MissingDataRule, CreatedBy: &actor}
	if err := s.db.Create(f).Error; err != nil {
		return nil, err
	}
	s.auditRecord(orgID, actor, f.ID, "health_formula.created")
	return f, nil
}
func (s *Service) TransitionFormula(orgID, actor, id uuid.UUID, status string) (*Formula, error) {
	var f Formula
	if err := s.db.Where("organization_id = ? AND id = ?", orgID, id).First(&f).Error; err != nil {
		return nil, ErrNotFound
	}
	if status == "APPROVED" {
		var count int64
		s.db.Model(&Formula{}).Where("organization_id = ? AND status = 'APPROVED'", orgID).Count(&count)
		if count > 0 {
			return nil, errors.New("an approved formula already exists")
		}
		now := time.Now()
		f.Status = status
		f.ApprovedAt = &now
		f.ApprovedBy = &actor
	} else if status == "RETIRED" && f.Status != "APPROVED" {
		return nil, errors.New("only APPROVED formula can be retired")
	} else if status != "RETIRED" {
		return nil, errors.New("unsupported formula transition")
	}
	if err := s.db.Save(&f).Error; err != nil {
		return nil, err
	}
	s.auditRecord(orgID, actor, id, "health_formula."+status)
	return &f, nil
}
func (s *Service) Calculate(orgID, actor, projectID uuid.UUID, req CalculateRequest) (*Snapshot, error) {
	var f Formula
	if req.FormulaID != "" {
		id, err := uuid.Parse(req.FormulaID)
		if err != nil {
			return nil, errors.New("invalid formula_id")
		}
		if err = s.db.Where("organization_id = ? AND id = ?", orgID, id).First(&f).Error; err != nil {
			return nil, ErrNotFound
		}
	} else if err := s.db.Where("organization_id = ? AND status = 'APPROVED'", orgID).Order("version DESC").First(&f).Error; err != nil {
		return nil, errors.New("no approved health formula")
	}
	var projectCount int64
	if err := s.db.Table("projects").Where("organization_id = ? AND id = ? AND deleted_at IS NULL", orgID, projectID).Count(&projectCount).Error; err != nil || projectCount == 0 {
		return nil, ErrNotFound
	}
	var weights map[string]float64
	var thresholds map[string]float64
	_ = json.Unmarshal(f.Weights, &weights)
	_ = json.Unmarshal(f.Thresholds, &thresholds)
	components := map[string]component{}
	for _, d := range Dimensions {
		score, available, reason := s.dimension(orgID, projectID, d)
		components[d] = component{Score: score, Weight: weights[d], Available: available, Reason: reason}
	}
	totalWeight := 0.0
	total := 0.0
	missing := 0
	for _, d := range Dimensions {
		c := components[d]
		if c.Weight <= 0 {
			continue
		}
		totalWeight += c.Weight
		if !c.Available {
			missing++
			if f.MissingDataRule == "EXCLUDE" {
				totalWeight -= c.Weight
				continue
			}
			c.Score = thresholds["missing_score"]
		}
		total += c.Score * c.Weight
	}
	score := 0.0
	if totalWeight > 0 {
		score = total / totalWeight
	}
	score = math.Round(score*100) / 100
	green := thresholds["green"]
	yellow := thresholds["yellow"]
	red := thresholds["red"]
	if green == 0 {
		green = 80
	}
	if yellow == 0 {
		yellow = 60
	}
	if red == 0 {
		red = 40
	}
	class := "CRITICAL"
	if score >= green {
		class = "GREEN"
	} else if score >= yellow {
		class = "YELLOW"
	} else if score >= red {
		class = "RED"
	}
	raw, _ := json.Marshal(components)
	now := time.Now()
	snap := &Snapshot{ID: uuid.New(), OrganizationID: orgID, ProjectID: projectID, FormulaID: f.ID, PeriodYear: req.PeriodYear, PeriodMonth: req.PeriodMonth, Score: score, HealthClass: class, Components: raw, Explanation: fmt.Sprintf("Formula v%d; %d/%d dimensions available.", f.Version, 8-missing, 8), CalculatedAt: now}
	if err := s.db.Create(snap).Error; err != nil {
		return nil, err
	}
	s.auditRecord(orgID, actor, snap.ID, "health_snapshot.calculated")
	return snap, nil
}
func (s *Service) dimension(orgID, projectID uuid.UUID, d string) (float64, bool, string) {
	switch d {
	case "risk":
		var n int64
		s.db.Table("risks").Where("organization_id = ? AND project_id = ? AND deleted_at IS NULL AND status NOT IN ('CLOSED','ACCEPTED')", orgID, projectID).Count(&n)
		return math.Max(0, 100-float64(n)*20), true, fmt.Sprintf("%d open risks", n)
	case "issue":
		var n int64
		s.db.Table("issues").Where("organization_id = ? AND project_id = ? AND deleted_at IS NULL AND status NOT IN ('CLOSED','RESOLVED')", orgID, projectID).Count(&n)
		return math.Max(0, 100-float64(n)*20), true, fmt.Sprintf("%d open issues", n)
	case "contract":
		var n int64
		s.db.Table("contracts").Where("organization_id = ? AND project_id = ? AND deleted_at IS NULL", orgID, projectID).Count(&n)
		return func() float64 {
			if n > 0 {
				return 100
			}
			return 50
		}(), true, fmt.Sprintf("%d contracts", n)
	case "physical":
		var p float64
		err := s.db.Table("project_snapshots").Where("organization_id = ? AND project_id = ? AND status = 'VALID'", orgID, projectID).Order("period_year DESC, period_month DESC").Select("COALESCE(physical_actual,0)").Limit(1).Scan(&p).Error
		return p, err == nil && p > 0, "validated physical snapshot"
	default:
		return 0, false, "no validated source available"
	}
}
func (s *Service) ListSnapshots(orgID, projectID uuid.UUID) ([]Snapshot, error) {
	var v []Snapshot
	err := s.db.Where("organization_id = ? AND project_id = ?", orgID, projectID).Order("period_year DESC, period_month DESC, calculated_at DESC").Find(&v).Error
	return v, err
}
func (s *Service) auditRecord(orgID, actor, id uuid.UUID, action string) {
	if s.audit != nil {
		s.audit.Record(audit.WriteRequest{OrganizationID: orgID, ActorID: &actor, Action: action, EntityType: "health", EntityID: id.String()})
	}
}
