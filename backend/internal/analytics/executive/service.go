package executive

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct{ db *gorm.DB }

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// GetDashboard returns the full Level 1 executive dashboard.
func (s *Service) GetDashboard(ctx context.Context, orgID uuid.UUID, f Filter) (*ExecutiveDashboard, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	summary, err := s.getNationalSummary(ctx, orgID)
	if err != nil {
		return nil, err
	}

	critical, err := s.getCriticalProjects(ctx, orgID)
	if err != nil {
		return nil, err
	}

	escalations, err := s.getOpenEscalations(ctx, orgID)
	if err != nil {
		return nil, err
	}

	decisions, err := s.getPendingDecisions(ctx, orgID)
	if err != nil {
		return nil, err
	}

	programs, err := s.getProgramSummaries(ctx, orgID)
	if err != nil {
		return nil, err
	}

	benefits, err := s.getBenefitSummary(ctx, orgID)
	if err != nil {
		return nil, err
	}

	return &ExecutiveDashboard{
		Summary:          *summary,
		CriticalProjects: critical,
		Escalations:      escalations,
		PendingDecisions: decisions,
		Programs:         programs,
		Benefits:         *benefits,
		AsOf:             now,
	}, nil
}

// getNationalSummary aggregates national KPIs from projects and related tables.
func (s *Service) getNationalSummary(ctx context.Context, orgID uuid.UUID) (*NationalSummary, error) {
	type projectStats struct {
		TotalProjects  int64
		ActiveProjects int64
		DraftProjects  int64
		TotalBudget    float64
		AvgProgress    float64
	}
	var ps projectStats
	if err := s.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) AS total_projects,
			COUNT(*) FILTER (WHERE status IN ('ACTIVE','ON_HOLD')) AS active_projects,
			COUNT(*) FILTER (WHERE status = 'DRAFT') AS draft_projects,
			COALESCE(SUM(budget_total), 0) AS total_budget,
			COALESCE(AVG(progress_pct), 0) AS avg_progress
		FROM projects
		WHERE organization_id = ? AND deleted_at IS NULL
	`, orgID).Scan(&ps).Error; err != nil {
		return nil, err
	}

	// health distribution — latest snapshot per project
	type healthStats struct {
		HealthGreen    int64
		HealthYellow   int64
		HealthRed      int64
		HealthCritical int64
	}
	var hs healthStats
	if err := s.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) FILTER (WHERE h.health_class = 'GREEN')    AS health_green,
			COUNT(*) FILTER (WHERE h.health_class = 'YELLOW')   AS health_yellow,
			COUNT(*) FILTER (WHERE h.health_class = 'RED')      AS health_red,
			COUNT(*) FILTER (WHERE h.health_class = 'CRITICAL') AS health_critical
		FROM (
			SELECT DISTINCT ON (hs2.project_id) hs2.health_class
			FROM health_snapshots hs2
			JOIN projects p ON p.id = hs2.project_id
			WHERE hs2.organization_id = ? AND p.deleted_at IS NULL
			ORDER BY hs2.project_id, hs2.calculated_at DESC
		) h
	`, orgID).Scan(&hs).Error; err != nil {
		return nil, err
	}

	type riskStats struct {
		OpenRisks int64
		HighRisks int64
	}
	var rs riskStats
	if err := s.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) FILTER (WHERE r.status NOT IN ('CLOSED','RESOLVED')) AS open_risks,
			COUNT(*) FILTER (WHERE r.status NOT IN ('CLOSED','RESOLVED') AND r.severity IN ('HIGH','CRITICAL')) AS high_risks
		FROM risks r
		WHERE r.organization_id = ? AND r.deleted_at IS NULL
	`, orgID).Scan(&rs).Error; err != nil {
		return nil, err
	}

	type issueStats struct {
		OpenIssues     int64
		CriticalIssues int64
	}
	var is issueStats
	if err := s.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) FILTER (WHERE i.status NOT IN ('CLOSED','RESOLVED')) AS open_issues,
			COUNT(*) FILTER (WHERE i.status NOT IN ('CLOSED','RESOLVED') AND i.severity IN ('HIGH','CRITICAL')) AS critical_issues
		FROM issues i
		WHERE i.organization_id = ? AND i.deleted_at IS NULL
	`, orgID).Scan(&is).Error; err != nil {
		return nil, err
	}

	type cmdStats struct {
		OpenEscalations  int64
		PendingDecisions int64
		OverdueDecisions int64
	}
	var cs cmdStats
	if err := s.db.WithContext(ctx).Raw(`
		SELECT
			(SELECT COUNT(*) FROM command_escalations WHERE organization_id = ? AND status IN ('OPEN','ACKNOWLEDGED') AND deleted_at IS NULL) AS open_escalations,
			(SELECT COUNT(*) FROM executive_decisions   WHERE organization_id = ? AND status = 'OPEN' AND deleted_at IS NULL) AS pending_decisions,
			(SELECT COUNT(*) FROM executive_decisions   WHERE organization_id = ? AND status = 'OPEN' AND due_date < CURRENT_DATE AND deleted_at IS NULL) AS overdue_decisions
	`, orgID, orgID, orgID).Scan(&cs).Error; err != nil {
		return nil, err
	}

	type benefitCount struct{ Total int64 }
	var bc benefitCount
	s.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) AS total FROM benefit_indicators WHERE organization_id = ? AND deleted_at IS NULL`,
		orgID,
	).Scan(&bc)

	scored := hs.HealthGreen + hs.HealthYellow + hs.HealthRed + hs.HealthCritical
	unscored := ps.TotalProjects - scored
	if unscored < 0 {
		unscored = 0
	}

	return &NationalSummary{
		TotalProjects:       ps.TotalProjects,
		ActiveProjects:      ps.ActiveProjects,
		DraftProjects:       ps.DraftProjects,
		TotalBudget:         ps.TotalBudget,
		BudgetRealized:      0,
		BudgetUsagePct:      0,
		AvgPhysicalProgress: ps.AvgProgress,
		HealthGreen:         hs.HealthGreen,
		HealthYellow:        hs.HealthYellow,
		HealthRed:           hs.HealthRed,
		HealthCritical:      hs.HealthCritical,
		HealthUnscored:      unscored,
		OpenRisks:           rs.OpenRisks,
		HighRisks:           rs.HighRisks,
		OpenIssues:          is.OpenIssues,
		CriticalIssues:      is.CriticalIssues,
		OverdueActions:      0,
		OpenEscalations:     cs.OpenEscalations,
		PendingDecisions:    cs.PendingDecisions,
		OverdueDecisions:    cs.OverdueDecisions,
		BenefitIndicators:   bc.Total,
		AsOf:                time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// getCriticalProjects returns projects with CRITICAL/RED health or high priority score.
func (s *Service) getCriticalProjects(ctx context.Context, orgID uuid.UUID) ([]CriticalProject, error) {
	type row struct {
		ProjectID      uuid.UUID
		ProjectCode    string
		ProjectName    string
		Status         string
		HealthClass    string
		PhysicalActual float64
		BudgetTotal    float64
		OpenRisks      int64
		OpenIssues     int64
		PriorityScore  float64
		PriorityClass  string
		ProgramName    string
		SectorName     string
	}
	var rows []row
	if err := s.db.WithContext(ctx).Raw(`
		SELECT
			p.id AS project_id,
			p.code AS project_code,
			p.name AS project_name,
			p.status,
			COALESCE(hs.health_class, 'UNSCORED') AS health_class,
			COALESCE(p.progress_pct, 0) AS physical_actual,
			COALESCE(p.budget_total, 0) AS budget_total,
			(SELECT COUNT(*) FROM risks   r WHERE r.project_id = p.id AND r.deleted_at IS NULL AND r.status NOT IN ('CLOSED','RESOLVED')) AS open_risks,
			(SELECT COUNT(*) FROM issues  i WHERE i.project_id = p.id AND i.deleted_at IS NULL AND i.status NOT IN ('CLOSED','RESOLVED')) AS open_issues,
			COALESCE(pps.total_score, 0) AS priority_score,
			COALESCE(pps.score_category, '') AS priority_class,
			COALESCE(prg.name, '') AS program_name,
			COALESCE(sec.name, '') AS sector_name
		FROM projects p
		LEFT JOIN LATERAL (
			SELECT health_class
			FROM health_snapshots
			WHERE project_id = p.id
			ORDER BY calculated_at DESC
			LIMIT 1
		) hs ON true
		LEFT JOIN LATERAL (
			SELECT total_score, score_category
			FROM project_priority_scores
			WHERE project_id = p.id
			ORDER BY calculated_at DESC
			LIMIT 1
		) pps ON true
		LEFT JOIN programs prg ON prg.id = p.program_id AND prg.deleted_at IS NULL
		LEFT JOIN sectors sec ON sec.id = p.sector_id AND sec.deleted_at IS NULL
		WHERE p.organization_id = ? AND p.deleted_at IS NULL
		  AND (
			COALESCE(hs.health_class, 'UNSCORED') IN ('CRITICAL', 'RED')
			OR COALESCE(pps.score_category, '') IN ('CRITICAL', 'HIGH')
		  )
		ORDER BY
			CASE COALESCE(hs.health_class, 'UNSCORED')
				WHEN 'CRITICAL' THEN 1
				WHEN 'RED' THEN 2
				ELSE 3
			END,
			COALESCE(pps.total_score, 0) DESC
		LIMIT 20
	`, orgID).Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]CriticalProject, 0, len(rows))
	for _, r := range rows {
		cp := CriticalProject{
			ProjectID:      r.ProjectID,
			ProjectCode:    r.ProjectCode,
			ProjectName:    r.ProjectName,
			Status:         r.Status,
			HealthClass:    r.HealthClass,
			PhysicalActual: r.PhysicalActual,
			BudgetTotal:    r.BudgetTotal,
			OpenRisks:      r.OpenRisks,
			OpenIssues:     r.OpenIssues,
			PriorityScore:  r.PriorityScore,
			PriorityClass:  r.PriorityClass,
			ProgramName:    r.ProgramName,
			SectorName:     r.SectorName,
		}
		result = append(result, cp)
	}
	return result, nil
}

// getOpenEscalations returns OPEN escalations for the org (FR-CTL1-02).
func (s *Service) getOpenEscalations(ctx context.Context, orgID uuid.UUID) ([]EscalationItem, error) {
	type row struct {
		ID          uuid.UUID
		ProjectID   *uuid.UUID
		ProjectName string
		Level       string
		SourceType  string
		Reason      string
		Status      string
		CreatedAt   time.Time
	}
	var rows []row
	if err := s.db.WithContext(ctx).Raw(`
		SELECT
			e.id,
			e.project_id,
			COALESCE(p.name, '') AS project_name,
			e.level,
			e.source_type,
			e.reason,
			e.status,
			e.created_at
		FROM command_escalations e
		LEFT JOIN projects p ON p.id = e.project_id AND p.deleted_at IS NULL
		WHERE e.organization_id = ? AND e.deleted_at IS NULL
		  AND e.status IN ('OPEN', 'ACKNOWLEDGED')
		ORDER BY e.created_at DESC
		LIMIT 50
	`, orgID).Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]EscalationItem, 0, len(rows))
	for _, r := range rows {
		item := EscalationItem{
			ID:          r.ID,
			ProjectID:   r.ProjectID,
			ProjectName: r.ProjectName,
			Level:       r.Level,
			SourceType:  r.SourceType,
			Reason:      r.Reason,
			Status:      r.Status,
			CreatedAt:   r.CreatedAt.UTC().Format(time.RFC3339),
		}
		result = append(result, item)
	}
	return result, nil
}

// getPendingDecisions returns OPEN executive decisions (FR-CTL1-02 — antrean keputusan).
func (s *Service) getPendingDecisions(ctx context.Context, orgID uuid.UUID) ([]DecisionItem, error) {
	type row struct {
		ID           uuid.UUID
		ProjectID    *uuid.UUID
		ProjectName  string
		Subject      string
		DecisionText string
		Status       string
		DueDate      *time.Time
		CreatedAt    time.Time
	}
	var rows []row
	if err := s.db.WithContext(ctx).Raw(`
		SELECT
			d.id,
			d.project_id,
			COALESCE(p.name, '') AS project_name,
			d.subject,
			d.decision_text,
			d.status,
			d.due_date,
			d.created_at
		FROM executive_decisions d
		LEFT JOIN projects p ON p.id = d.project_id AND p.deleted_at IS NULL
		WHERE d.organization_id = ? AND d.deleted_at IS NULL
		  AND d.status = 'OPEN'
		ORDER BY d.due_date ASC NULLS LAST, d.created_at DESC
		LIMIT 50
	`, orgID).Scan(&rows).Error; err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	result := make([]DecisionItem, 0, len(rows))
	for _, r := range rows {
		var dueDateStr *string
		isOverdue := false
		if r.DueDate != nil {
			s := r.DueDate.Format("2006-01-02")
			dueDateStr = &s
			isOverdue = r.DueDate.Before(now)
		}
		item := DecisionItem{
			ID:           r.ID,
			ProjectID:    r.ProjectID,
			ProjectName:  r.ProjectName,
			Subject:      r.Subject,
			DecisionText: r.DecisionText,
			Status:       r.Status,
			DueDate:      dueDateStr,
			IsOverdue:    isOverdue,
			CreatedAt:    r.CreatedAt.UTC().Format(time.RFC3339),
		}
		result = append(result, item)
	}
	return result, nil
}

// getProgramSummaries returns per-program KPI for comparison table (FR-CTL1-03).
func (s *Service) getProgramSummaries(ctx context.Context, orgID uuid.UUID) ([]ProgramKPISummary, error) {
	type row struct {
		ProgramID      uuid.UUID
		ProgramCode    string
		ProgramName    string
		TotalProjects  int64
		ActiveProjects int64
		TotalBudget    float64
		AvgProgress    float64
		OpenRisks      int64
		OpenIssues     int64
	}
	var rows []row
	if err := s.db.WithContext(ctx).Raw(`
		SELECT
			prg.id AS program_id,
			prg.code AS program_code,
			prg.name AS program_name,
			COUNT(p.id) AS total_projects,
			COUNT(p.id) FILTER (WHERE p.status IN ('ACTIVE','ON_HOLD')) AS active_projects,
			COALESCE(SUM(p.budget_total), 0) AS total_budget,
			COALESCE(AVG(p.progress_pct), 0) AS avg_progress,
			(
				SELECT COUNT(*) FROM risks r
				JOIN projects pp ON pp.id = r.project_id
				WHERE pp.program_id = prg.id AND pp.organization_id = ? AND pp.deleted_at IS NULL
				  AND r.deleted_at IS NULL AND r.status NOT IN ('CLOSED','RESOLVED')
			) AS open_risks,
			(
				SELECT COUNT(*) FROM issues i
				JOIN projects pp ON pp.id = i.project_id
				WHERE pp.program_id = prg.id AND pp.organization_id = ? AND pp.deleted_at IS NULL
				  AND i.deleted_at IS NULL AND i.status NOT IN ('CLOSED','RESOLVED')
			) AS open_issues
		FROM programs prg
		LEFT JOIN projects p ON p.program_id = prg.id AND p.organization_id = ? AND p.deleted_at IS NULL
		WHERE prg.organization_id = ? AND prg.deleted_at IS NULL AND prg.is_active = true
		GROUP BY prg.id, prg.code, prg.name, prg.sort_order
		ORDER BY prg.sort_order, prg.name
	`, orgID, orgID, orgID, orgID).Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]ProgramKPISummary, 0, len(rows))
	for _, r := range rows {
		result = append(result, ProgramKPISummary{
			ProgramID:      r.ProgramID,
			ProgramCode:    r.ProgramCode,
			ProgramName:    r.ProgramName,
			TotalProjects:  r.TotalProjects,
			ActiveProjects: r.ActiveProjects,
			TotalBudget:    r.TotalBudget,
			AvgProgress:    r.AvgProgress,
			OpenRisks:      r.OpenRisks,
			OpenIssues:     r.OpenIssues,
		})
	}
	return result, nil
}

// getBenefitSummary returns national benefit indicator summary (FR-CTL1-04).
func (s *Service) getBenefitSummary(ctx context.Context, orgID uuid.UUID) (*BenefitSummary, error) {
	type indicatorRow struct {
		ID                uuid.UUID
		Name              string
		Unit              string
		TargetValue       float64
		AggregationMethod string
		ActualValue       float64
	}
	var rows []indicatorRow
	if err := s.db.WithContext(ctx).Raw(`
		SELECT
			bi.id,
			bi.name,
			bi.unit,
			COALESCE((SELECT AVG(bm.target) FROM benefit_measurements bm WHERE bm.indicator_id = bi.id AND bm.deleted_at IS NULL), 0) AS target_value,
			bi.aggregation_method,
			COALESCE(
				CASE bi.aggregation_method
					WHEN 'SUM'     THEN (SELECT SUM(bm.actual) FROM benefit_measurements bm WHERE bm.indicator_id = bi.id AND bm.deleted_at IS NULL)
					WHEN 'AVERAGE' THEN (SELECT AVG(bm.actual) FROM benefit_measurements bm WHERE bm.indicator_id = bi.id AND bm.deleted_at IS NULL)
					WHEN 'LATEST'  THEN (SELECT bm.actual FROM benefit_measurements bm WHERE bm.indicator_id = bi.id AND bm.deleted_at IS NULL ORDER BY bm.period_year DESC, bm.period_month DESC LIMIT 1)
					ELSE                (SELECT SUM(bm.actual) FROM benefit_measurements bm WHERE bm.indicator_id = bi.id AND bm.deleted_at IS NULL)
				END,
			0) AS actual_value
		FROM benefit_indicators bi
		WHERE bi.organization_id = ? AND bi.deleted_at IS NULL
		ORDER BY bi.name
	`, orgID).Scan(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]BenefitItem, 0, len(rows))
	var onTrack, behind int64
	for _, r := range rows {
		pct := 0.0
		if r.TargetValue > 0 {
			pct = (r.ActualValue / r.TargetValue) * 100
		}
		if pct >= 80 {
			onTrack++
		} else {
			behind++
		}
		items = append(items, BenefitItem{
			ID:                r.ID,
			Name:              r.Name,
			Unit:              r.Unit,
			TargetValue:       r.TargetValue,
			ActualValue:       r.ActualValue,
			AchievementPct:    pct,
			AggregationMethod: r.AggregationMethod,
		})
	}

	return &BenefitSummary{
		TotalIndicators: int64(len(rows)),
		OnTrackCount:    onTrack,
		BehindCount:     behind,
		Indicators:      items,
	}, nil
}
