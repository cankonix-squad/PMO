package projectcontrol

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

var ErrNotFound = errors.New("project control not found")

type Service struct{ db *gorm.DB }

func NewService(db *gorm.DB) *Service { return &Service{db: db} }
func (s *Service) Get(ctx context.Context, orgID, projectID uuid.UUID, year, month int) (*Control, error) {
	var p ProjectSummary
	if err := s.db.WithContext(ctx).Table("projects").Where("organization_id=? AND id=? AND deleted_at IS NULL", orgID, projectID).First(&p).Error; err != nil {
		return nil, ErrNotFound
	}
	if year == 0 {
		now := time.Now()
		year, month = now.Year(), int(now.Month())
	}
	out := &Control{Project: p, AsOf: time.Now().UTC().Format(time.RFC3339), Issues: []Item{}, Risks: []Item{}, Actions: []Item{}}
	var c struct {
		Count       int64
		TotalValue  float64
		ActiveCount int64
		Currency    string
	}
	s.db.WithContext(ctx).Table("contracts").Select("count(*) count, coalesce(sum(contract_value),0) total_value, count(*) filter (where status='ACTIVE') active_count, coalesce(max(currency),'IDR') currency").Where("organization_id=? AND project_id=? AND deleted_at IS NULL", orgID, projectID).Scan(&c)
	out.Contract = ContractSummary{Count: c.Count, TotalValue: c.TotalValue, ActiveCount: c.ActiveCount, Currency: c.Currency}
	var snap SnapshotSummary
	if err := s.db.WithContext(ctx).Table("project_snapshots").Where("organization_id=? AND project_id=? AND period_year=? AND period_month=? AND status='VALID' AND deleted_at IS NULL", orgID, projectID, year, month).First(&snap).Error; err == nil {
		out.Snapshot = &snap
	}
	var hs HealthSummary
	if err := s.db.WithContext(ctx).Table("health_snapshots").Select("score, health_class class, formula_id, explanation").Where("organization_id=? AND project_id=? AND period_year=? AND period_month=?", orgID, projectID, year, month).Order("calculated_at DESC").First(&hs).Error; err == nil {
		out.Health = &hs
	}
	s.db.WithContext(ctx).Table("field_inspections").Select("count(*) inspections, count(*) filter (where verification_status='VERIFIED') verified_inspections").Where("organization_id=? AND project_id=? AND deleted_at IS NULL", orgID, projectID).Scan(&out.Evidence)
	s.db.WithContext(ctx).Table("field_evidence").Where("organization_id=? AND project_id=? AND deleted_at IS NULL", orgID, projectID).Count(&out.Evidence.EvidenceFiles)
	var issues []Item
	s.db.WithContext(ctx).Table("issues").Select("id,title,status,severity,coalesce(due_date::text,'') due_date").Where("organization_id=? AND project_id=? AND deleted_at IS NULL AND status NOT IN ('CLOSED','RESOLVED')", orgID, projectID).Order("due_date ASC NULLS LAST").Limit(20).Scan(&issues)
	out.Issues = issues
	var risks []Item
	s.db.WithContext(ctx).Table("risks").Select("id,title,status,severity,coalesce(due_date::text,'') due_date").Where("organization_id=? AND project_id=? AND deleted_at IS NULL AND status NOT IN ('CLOSED','ACCEPTED')", orgID, projectID).Order("risk_score DESC").Limit(20).Scan(&risks)
	out.Risks = risks
	var actions []Item
	s.db.WithContext(ctx).Table("corrective_actions").Select("id,title,status,coalesce(target_date::text,'') due_date").Where("organization_id=? AND project_id=? AND deleted_at IS NULL AND status NOT IN ('COMPLETED','REJECTED')", orgID, projectID).Order("target_date ASC NULLS LAST").Limit(20).Scan(&actions)
	out.Actions = actions
	return out, nil
}
