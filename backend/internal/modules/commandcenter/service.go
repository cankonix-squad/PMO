package commandcenter

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Service struct{ db *gorm.DB }

func NewService(db *gorm.DB) *Service { return &Service{db: db} }
func (s *Service) Summary(ctx context.Context, orgID uuid.UUID) (*Summary, error) {
	now := time.Now()
	result := &Summary{AsOf: now.UTC().Format(time.RFC3339), Alerts: []Item{}, Actions: []Item{}, Validations: []Item{}, Watchlist: []Item{}, Escalations: []Item{}, Decisions: []Item{}}
	var risks []struct {
		ID          uuid.UUID
		Title       string
		ProjectID   uuid.UUID
		ProjectName string
		Severity    string
		DueDate     *time.Time
	}
	if err := s.db.WithContext(ctx).Table("risks r").Select("r.id,r.title,r.project_id,p.name project_name,r.severity,r.due_date").Joins("JOIN projects p ON p.id=r.project_id AND p.organization_id=? AND p.deleted_at IS NULL", orgID).Where("r.organization_id=? AND r.deleted_at IS NULL AND r.status NOT IN ('CLOSED','ACCEPTED')", orgID).Order("r.risk_score DESC").Limit(10).Scan(&risks).Error; err != nil {
		return nil, err
	}
	for _, row := range risks {
		item := Item{ID: row.ID.String(), Kind: "RISK", Title: row.Title, Message: "Open risk requires monitoring", Severity: row.Severity, Status: "OPEN", ProjectID: &row.ProjectID, ProjectName: row.ProjectName, SourceType: "risk", SourceID: row.ID.String()}
		if row.DueDate != nil {
			item.DueAt = row.DueDate.UTC().Format(time.RFC3339)
			item.AgingDays = maxDays(now, *row.DueDate)
		}
		result.Alerts = append(result.Alerts, item)
		result.Watchlist = append(result.Watchlist, item)
	}
	var actions []struct {
		ID          uuid.UUID
		Title       string
		ProjectID   uuid.UUID
		ProjectName string
		Status      string
		PICUserID   *uuid.UUID
		TargetDate  *time.Time
	}
	if err := s.db.WithContext(ctx).Table("corrective_actions ca").Select("ca.id,ca.title,ca.project_id,p.name project_name,ca.status,ca.pic_user_id,ca.target_date").Joins("JOIN projects p ON p.id=ca.project_id AND p.organization_id=? AND p.deleted_at IS NULL", orgID).Where("ca.organization_id=? AND ca.deleted_at IS NULL AND ca.status NOT IN ('COMPLETED','REJECTED')", orgID).Order("ca.target_date ASC NULLS LAST").Limit(20).Scan(&actions).Error; err != nil {
		return nil, err
	}
	for _, row := range actions {
		item := Item{ID: row.ID.String(), Kind: "CORRECTIVE_ACTION", Title: row.Title, Message: "Corrective action needs follow-up", Severity: "HIGH", Status: row.Status, ProjectID: &row.ProjectID, ProjectName: row.ProjectName, PICUserID: row.PICUserID, SourceType: "corrective_action", SourceID: row.ID.String()}
		if row.TargetDate != nil {
			item.DueAt = row.TargetDate.UTC().Format(time.RFC3339)
			item.AgingDays = maxDays(now, *row.TargetDate)
		}
		result.Actions = append(result.Actions, item)
	}
	var validations []struct {
		ID          uuid.UUID
		ProjectID   uuid.UUID
		ProjectName string
		Status      string
		SubmittedAt *time.Time
		SLADueAt    *time.Time
	}
	if err := s.db.WithContext(ctx).Table("data_submissions ds").Select("ds.id,ds.project_id,p.name project_name,ds.status,ds.submitted_at,ds.sla_due_at").Joins("JOIN projects p ON p.id=ds.project_id AND p.organization_id=? AND p.deleted_at IS NULL", orgID).Where("ds.organization_id=? AND ds.deleted_at IS NULL AND ds.status IN ('SUBMITTED','STALE','REJECTED')", orgID).Order("ds.sla_due_at ASC NULLS LAST").Limit(20).Scan(&validations).Error; err != nil {
		return nil, err
	}
	for _, row := range validations {
		item := Item{ID: row.ID.String(), Kind: "VALIDATION", Title: "Snapshot validation", Message: "Submission awaits data quality decision", Severity: "MEDIUM", Status: row.Status, ProjectID: &row.ProjectID, ProjectName: row.ProjectName, SourceType: "data_submission", SourceID: row.ID.String()}
		if row.SLADueAt != nil {
			item.DueAt = row.SLADueAt.UTC().Format(time.RFC3339)
			item.AgingDays = maxDays(now, *row.SLADueAt)
		}
		result.Validations = append(result.Validations, item)
	}
	var escalations []Escalation
	if err := s.db.WithContext(ctx).Where("organization_id=? AND status IN ('OPEN','ACKNOWLEDGED') AND deleted_at IS NULL", orgID).Order("created_at DESC").Limit(20).Find(&escalations).Error; err != nil {
		return nil, err
	}
	for _, row := range escalations {
		result.Escalations = append(result.Escalations, Item{ID: row.ID.String(), Kind: "ESCALATION", Title: row.Level, Message: row.Reason, Severity: "HIGH", Status: row.Status, ProjectID: row.ProjectID, SourceType: row.SourceType, SourceID: row.SourceID.String()})
	}
	var decisions []Decision
	if err := s.db.WithContext(ctx).Where("organization_id=? AND status IN ('OPEN','IN_PROGRESS') AND deleted_at IS NULL", orgID).Order("due_date ASC NULLS LAST, created_at DESC").Limit(20).Find(&decisions).Error; err != nil {
		return nil, err
	}
	for _, row := range decisions {
		item := Item{ID: row.ID.String(), Kind: "DECISION", Title: row.Subject, Message: row.DecisionText, Severity: "MEDIUM", Status: row.Status, ProjectID: row.ProjectID, SourceType: "executive_decision", SourceID: row.ID.String()}
		if row.DueDate != nil {
			item.DueAt = row.DueDate.UTC().Format(time.RFC3339)
			item.AgingDays = maxDays(now, row.DueDate.Time)
		}
		result.Decisions = append(result.Decisions, item)
	}
	return result, nil
}
func maxDays(now, due time.Time) int {
	days := int(now.Sub(due).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}
