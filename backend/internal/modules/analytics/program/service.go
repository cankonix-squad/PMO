package program

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct{ db *gorm.DB }

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// ListPrograms returns KPI aggregation for all programs in the org.
func (s *Service) ListPrograms(ctx context.Context, orgID uuid.UUID, f Filter) (*ListResponse, error) {
	groups, err := s.listGroups(ctx, orgID, "program", f)
	if err != nil {
		return nil, err
	}
	return &ListResponse{Groups: groups, AsOf: time.Now().UTC().Format(time.RFC3339)}, nil
}

// ListSectors returns KPI aggregation for all sectors in the org.
func (s *Service) ListSectors(ctx context.Context, orgID uuid.UUID, f Filter) (*ListResponse, error) {
	groups, err := s.listGroups(ctx, orgID, "sector", f)
	if err != nil {
		return nil, err
	}
	return &ListResponse{Groups: groups, AsOf: time.Now().UTC().Format(time.RFC3339)}, nil
}

// GetProgram returns full dashboard for one program.
func (s *Service) GetProgram(ctx context.Context, orgID, programID uuid.UUID, f Filter) (*ProgramDashboard, error) {
	return s.getDashboard(ctx, orgID, programID, "program", f)
}

// GetSector returns full dashboard for one sector.
func (s *Service) GetSector(ctx context.Context, orgID, sectorID uuid.UUID, f Filter) (*ProgramDashboard, error) {
	return s.getDashboard(ctx, orgID, sectorID, "sector", f)
}

// ── Internal helpers ──────────────────────────────────────────────────────────

func (s *Service) listGroups(ctx context.Context, orgID uuid.UUID, groupType string, f Filter) ([]ProgramKPI, error) {
	year, month := resolveperiod(f)

	// Choose the master table based on group type
	masterTable := "programs"
	fkCol := "program_id"
	if groupType == "sector" {
		masterTable = "sectors"
		fkCol = "sector_id"
	}

	type groupRow struct {
		GroupID   uuid.UUID
		GroupCode string
		GroupName string
	}
	var groups []groupRow
	if err := s.db.WithContext(ctx).
		Table(masterTable).
		Select("id group_id, code group_code, name group_name").
		Where("organization_id = ? AND deleted_at IS NULL AND is_active = true", orgID).
		Order("sort_order, name").
		Scan(&groups).Error; err != nil {
		return nil, err
	}

	result := make([]ProgramKPI, 0, len(groups))
	for _, g := range groups {
		kpi, err := s.buildKPI(ctx, orgID, g.GroupID, g.GroupCode, g.GroupName, groupType, fkCol, year, month)
		if err != nil {
			continue
		}
		result = append(result, *kpi)
	}
	return result, nil
}

func (s *Service) getDashboard(ctx context.Context, orgID, groupID uuid.UUID, groupType string, f Filter) (*ProgramDashboard, error) {
	year, month := resolveperiod(f)

	masterTable := "programs"
	fkCol := "program_id"
	if groupType == "sector" {
		masterTable = "sectors"
		fkCol = "sector_id"
	}

	// Verify group exists and belongs to org
	type masterRow struct {
		ID   uuid.UUID
		Code string
		Name string
	}
	var master masterRow
	if err := s.db.WithContext(ctx).
		Table(masterTable).
		Select("id, code, name").
		Where("organization_id = ? AND id = ? AND deleted_at IS NULL", orgID, groupID).
		First(&master).Error; err != nil {
		return nil, gorm.ErrRecordNotFound
	}

	kpi, err := s.buildKPI(ctx, orgID, groupID, master.Code, master.Name, groupType, fkCol, year, month)
	if err != nil {
		return nil, err
	}

	projects, err := s.listProjects(ctx, orgID, groupID, fkCol, year, month)
	if err != nil {
		projects = []ProjectRow{}
	}

	topPhys := topPhysicalDeviation(projects, 5)
	topBudget := topBudgetDeviation(projects, 5)
	highRisk := highRiskProjects(projects, 5)

	return &ProgramDashboard{
		KPI:              *kpi,
		Projects:         projects,
		TopPhysicalDev:   topPhys,
		TopBudgetDev:     topBudget,
		HighRiskProjects: highRisk,
		AsOf:             time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *Service) buildKPI(ctx context.Context, orgID, groupID uuid.UUID, code, name, groupType, fkCol string, year, month int) (*ProgramKPI, error) {
	kpi := &ProgramKPI{
		GroupID:   groupID,
		GroupCode: code,
		GroupName: name,
		GroupType: groupType,
		Currency:  "IDR",
		AsOf:      time.Now().UTC().Format(time.RFC3339),
	}

	// Project counts + budget
	type projectAgg struct {
		Total          int64
		Active         int64
		TotalBudget    float64
		BudgetRealized float64
	}
	var pa projectAgg
	s.db.WithContext(ctx).Table("projects").
		Select(`
			COUNT(*) total,
			COUNT(*) FILTER (WHERE status IN ('ACTIVE','IN_PROGRESS','ON_TRACK')) active,
			COALESCE(SUM(budget_total), 0) total_budget,
			COALESCE(SUM(progress_pct * budget_total / 100.0), 0) budget_realized
		`).
		Where("organization_id = ? AND "+fkCol+" = ? AND deleted_at IS NULL", orgID, groupID).
		Scan(&pa)
	kpi.TotalProjects = pa.Total
	kpi.ActiveProjects = pa.Active
	kpi.TotalBudget = pa.TotalBudget
	kpi.BudgetRealized = pa.BudgetRealized
	if pa.TotalBudget > 0 {
		kpi.BudgetUsagePct = pa.BudgetRealized / pa.TotalBudget * 100
	}

	// Physical progress from latest VALID snapshots
	type snapAgg struct {
		AvgActual float64
		AvgTarget float64
	}
	var sa snapAgg
	s.db.WithContext(ctx).Table("project_snapshots ps").
		Select("COALESCE(AVG(ps.physical_actual), 0) avg_actual, COALESCE(AVG(ps.physical_target), 0) avg_target").
		Joins("JOIN projects p ON p.id = ps.project_id").
		Where("p.organization_id = ? AND p."+fkCol+" = ? AND p.deleted_at IS NULL AND ps.deleted_at IS NULL AND ps.status = 'VALID' AND ps.period_year = ? AND ps.period_month = ?",
			orgID, groupID, year, month).
		Scan(&sa)
	kpi.AvgPhysicalActual = sa.AvgActual
	kpi.AvgPhysicalTarget = sa.AvgTarget
	kpi.PhysicalVariance = sa.AvgActual - sa.AvgTarget

	// Health distribution from latest health snapshots
	type healthDist struct {
		Class string
		Count int64
	}
	var hdRows []healthDist
	s.db.WithContext(ctx).Raw(`
		SELECT hs.health_class class, COUNT(*) count
		FROM health_snapshots hs
		JOIN projects p ON p.id = hs.project_id
		WHERE p.organization_id = ? AND p.`+fkCol+` = ? AND p.deleted_at IS NULL
		  AND hs.id = (
		    SELECT id FROM health_snapshots h2
		    WHERE h2.project_id = hs.project_id
		    ORDER BY calculated_at DESC LIMIT 1
		  )
		GROUP BY hs.health_class
	`, orgID, groupID).Scan(&hdRows)

	scoredProjects := int64(0)
	for _, h := range hdRows {
		scoredProjects += h.Count
		switch h.Class {
		case "GREEN":
			kpi.HealthGreen = h.Count
		case "YELLOW":
			kpi.HealthYellow = h.Count
		case "RED":
			kpi.HealthRed = h.Count
		case "CRITICAL":
			kpi.HealthCritical = h.Count
		}
	}
	kpi.HealthUnscored = kpi.TotalProjects - scoredProjects

	// Risk & issue counts
	type riskAgg struct {
		Open int64
		High int64
	}
	var ra riskAgg
	s.db.WithContext(ctx).Table("risks r").
		Select("COUNT(*) open, COUNT(*) FILTER (WHERE r.severity IN ('HIGH','CRITICAL')) high").
		Joins("JOIN projects p ON p.id = r.project_id").
		Where("p.organization_id = ? AND p."+fkCol+" = ? AND p.deleted_at IS NULL AND r.deleted_at IS NULL AND r.status NOT IN ('CLOSED','ACCEPTED')",
			orgID, groupID).
		Scan(&ra)
	kpi.OpenRisks = ra.Open
	kpi.HighRisks = ra.High

	type issueAgg struct {
		Open     int64
		Critical int64
	}
	var ia issueAgg
	s.db.WithContext(ctx).Table("issues i").
		Select("COUNT(*) open, COUNT(*) FILTER (WHERE i.severity = 'CRITICAL') critical").
		Joins("JOIN projects p ON p.id = i.project_id").
		Where("p.organization_id = ? AND p."+fkCol+" = ? AND p.deleted_at IS NULL AND i.deleted_at IS NULL AND i.status NOT IN ('CLOSED','RESOLVED')",
			orgID, groupID).
		Scan(&ia)
	kpi.OpenIssues = ia.Open
	kpi.CriticalIssues = ia.Critical

	// Overdue corrective actions
	s.db.WithContext(ctx).Table("corrective_actions ca").
		Joins("JOIN projects p ON p.id = ca.project_id").
		Where("p.organization_id = ? AND p."+fkCol+" = ? AND p.deleted_at IS NULL AND ca.deleted_at IS NULL AND ca.status NOT IN ('COMPLETED','REJECTED') AND ca.target_date < NOW()",
			orgID, groupID).
		Count(&kpi.OverdueActions)

	// Priority score
	type prioAgg struct {
		AvgScore float64
		Critical int64
	}
	var pra prioAgg
	s.db.WithContext(ctx).Table("project_priority_scores pps").
		Select("COALESCE(AVG(pps.total_score), 0) avg_score, COUNT(*) FILTER (WHERE pps.score_category = 'CRITICAL') critical").
		Joins("JOIN projects p ON p.id = pps.project_id").
		Where("p.organization_id = ? AND p."+fkCol+" = ? AND p.deleted_at IS NULL AND pps.id = (SELECT id FROM project_priority_scores s2 WHERE s2.project_id = pps.project_id ORDER BY calculated_at DESC LIMIT 1)",
			orgID, groupID).
		Scan(&pra)
	kpi.AvgPriorityScore = pra.AvgScore
	kpi.CriticalPriority = pra.Critical

	// Benefit indicators
	s.db.WithContext(ctx).Table("benefit_indicators bi").
		Joins("JOIN projects p ON p.id = bi.project_id").
		Where("p.organization_id = ? AND p."+fkCol+" = ? AND p.deleted_at IS NULL AND bi.deleted_at IS NULL",
			orgID, groupID).
		Count(&kpi.BenefitIndicators)

	return kpi, nil
}

func (s *Service) listProjects(ctx context.Context, orgID, groupID uuid.UUID, fkCol string, year, month int) ([]ProjectRow, error) {
	type rawRow struct {
		ProjectID      uuid.UUID
		ProjectCode    string
		ProjectName    string
		Status         string
		PhysicalActual float64
		PhysicalTarget float64
		BudgetTotal    float64
		BudgetRealized float64
		HealthClass    string
		HealthScore    float64
		OpenRisks      int64
		OpenIssues     int64
		PriorityScore  float64
		PriorityClass  string
	}

	var rows []rawRow
	s.db.WithContext(ctx).Raw(`
		SELECT
			p.id project_id,
			p.code project_code,
			p.name project_name,
			p.status,
			COALESCE(ps.physical_actual, 0) physical_actual,
			COALESCE(ps.physical_target, 0) physical_target,
			COALESCE(p.budget_total, 0) budget_total,
			COALESCE(p.progress_pct * p.budget_total / 100.0, 0) budget_realized,
			COALESCE(hs.health_class, '') health_class,
			COALESCE(hs.score, 0) health_score,
			COALESCE(ri.open_risks, 0) open_risks,
			COALESCE(ii.open_issues, 0) open_issues,
			COALESCE(pps.total_score, 0) priority_score,
			COALESCE(pps.score_category, '') priority_class
		FROM projects p
		LEFT JOIN LATERAL (
			SELECT physical_actual, physical_target
			FROM project_snapshots
			WHERE project_id = p.id AND status = 'VALID'
			  AND period_year = ? AND period_month = ?
			  AND deleted_at IS NULL
			ORDER BY created_at DESC LIMIT 1
		) ps ON TRUE
		LEFT JOIN LATERAL (
			SELECT health_class, score
			FROM health_snapshots
			WHERE project_id = p.id
			ORDER BY calculated_at DESC LIMIT 1
		) hs ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(*) open_risks
			FROM risks
			WHERE project_id = p.id AND deleted_at IS NULL
			  AND status NOT IN ('CLOSED','ACCEPTED')
		) ri ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(*) open_issues
			FROM issues
			WHERE project_id = p.id AND deleted_at IS NULL
			  AND status NOT IN ('CLOSED','RESOLVED')
		) ii ON TRUE
		LEFT JOIN LATERAL (
			SELECT total_score, score_category
			FROM project_priority_scores
			WHERE project_id = p.id
			ORDER BY calculated_at DESC LIMIT 1
		) pps ON TRUE
		WHERE p.organization_id = ? AND p.`+fkCol+` = ? AND p.deleted_at IS NULL
		ORDER BY p.name
	`, year, month, orgID, groupID).Scan(&rows)

	out := make([]ProjectRow, 0, len(rows))
	for _, r := range rows {
		budgetUsage := float64(0)
		if r.BudgetTotal > 0 {
			budgetUsage = r.BudgetRealized / r.BudgetTotal * 100
		}
		out = append(out, ProjectRow{
			ProjectID:      r.ProjectID,
			ProjectCode:    r.ProjectCode,
			ProjectName:    r.ProjectName,
			Status:         r.Status,
			PhysicalActual: r.PhysicalActual,
			PhysicalTarget: r.PhysicalTarget,
			PhysicalVar:    r.PhysicalActual - r.PhysicalTarget,
			BudgetTotal:    r.BudgetTotal,
			BudgetUsagePct: budgetUsage,
			HealthClass:    r.HealthClass,
			HealthScore:    r.HealthScore,
			OpenRisks:      r.OpenRisks,
			OpenIssues:     r.OpenIssues,
			PriorityScore:  r.PriorityScore,
			PriorityClass:  r.PriorityClass,
		})
	}
	return out, nil
}

// ── Pure helpers ──────────────────────────────────────────────────────────────

func resolveperiod(f Filter) (int, int) {
	if f.PeriodYear > 0 {
		return f.PeriodYear, f.PeriodMonth
	}
	now := time.Now()
	return now.Year(), int(now.Month())
}

func topPhysicalDeviation(projects []ProjectRow, n int) []TopDeviation {
	// Sort by most negative physical variance (worst deviation first)
	sorted := make([]ProjectRow, len(projects))
	copy(sorted, projects)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].PhysicalVar < sorted[i].PhysicalVar {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	result := []TopDeviation{}
	for i, p := range sorted {
		if i >= n {
			break
		}
		if p.PhysicalVar >= 0 {
			break // no negative deviation
		}
		result = append(result, TopDeviation{
			ProjectID:   p.ProjectID,
			ProjectCode: p.ProjectCode,
			ProjectName: p.ProjectName,
			Value:       p.PhysicalVar,
			Label:       formatPct(p.PhysicalVar) + "% fisik",
		})
	}
	return result
}

func topBudgetDeviation(projects []ProjectRow, n int) []TopDeviation {
	sorted := make([]ProjectRow, len(projects))
	copy(sorted, projects)
	// Sort by highest budget usage (overspend risk first)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].BudgetUsagePct > sorted[i].BudgetUsagePct {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	result := []TopDeviation{}
	for i, p := range sorted {
		if i >= n {
			break
		}
		if p.BudgetUsagePct < 90 {
			break // only show near/over budget
		}
		result = append(result, TopDeviation{
			ProjectID:   p.ProjectID,
			ProjectCode: p.ProjectCode,
			ProjectName: p.ProjectName,
			Value:       p.BudgetUsagePct,
			Label:       formatPct(p.BudgetUsagePct) + "% anggaran terpakai",
		})
	}
	return result
}

func highRiskProjects(projects []ProjectRow, n int) []ProjectRow {
	sorted := make([]ProjectRow, len(projects))
	copy(sorted, projects)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].OpenRisks > sorted[i].OpenRisks {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	result := []ProjectRow{}
	for i, p := range sorted {
		if i >= n || p.OpenRisks == 0 {
			break
		}
		result = append(result, p)
	}
	return result
}

func formatPct(v float64) string {
	if v >= 0 {
		return "+" + formatFloat(v)
	}
	return formatFloat(v)
}

func formatFloat(v float64) string {
	// Simple sprintf-equivalent without importing fmt
	neg := v < 0
	if neg {
		v = -v
	}
	intPart := int64(v)
	fracPart := int64((v - float64(intPart)) * 10)
	s := itoa(intPart) + "." + itoa(fracPart)
	if neg {
		return "-" + s
	}
	return s
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
