package reporting

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// auditLog writes an audit event directly to the audit_logs table.
// Uses a minimal inline insert so reporting doesn't need to import the audit package.
func (s *Service) auditLog(action, entityType string, entityID, orgID, userID uuid.UUID, meta map[string]interface{}) {
	metaJSON, _ := json.Marshal(meta)
	_ = s.db.Exec(`
		INSERT INTO audit_logs (id, organization_id, actor_id, action, entity_type, entity_id, new_values, created_at)
		VALUES (gen_random_uuid(), ?, ?, ?, ?, ?, ?, now())`,
		orgID, userID, action, entityType, entityID.String(), metaJSON,
	).Error
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ---------------------------------------------------------------------------
// Catalog
// ---------------------------------------------------------------------------

// GetCatalog returns all report_definitions for the given org.
func (s *Service) GetCatalog(orgID uuid.UUID) ([]ReportDefinition, error) {
	var defs []ReportDefinition
	err := s.db.
		Where("organization_id = ?", orgID).
		Order("sort_order ASC, name ASC").
		Find(&defs).Error
	return defs, err
}

// ---------------------------------------------------------------------------
// Dataset: Executive Summary
// ---------------------------------------------------------------------------

func (s *Service) GetExecutiveSummary(orgID uuid.UUID, f DatasetFilter) (*ExecutiveSummaryRow, error) {
	row := &ExecutiveSummaryRow{}

	base := s.db.Table("projects p").
		Where("p.organization_id = ? AND p.deleted_at IS NULL", orgID)

	if f.Status != nil && *f.Status != "" {
		base = base.Where("p.status = ?", *f.Status)
	}
	if f.Province != nil && *f.Province != "" {
		base = base.Where("p.province = ?", *f.Province)
	}
	if f.PeriodStart != nil {
		base = base.Where("p.start_date >= ?", *f.PeriodStart)
	}
	if f.PeriodEnd != nil {
		base = base.Where("p.end_date <= ?", *f.PeriodEnd)
	}

	type projectCounts struct {
		Total     int
		Active    int
		Completed int
		OnHold    int
		AvgProg   float64
	}
	var pc projectCounts
	if err := base.Select(`
		COUNT(*) AS total,
		SUM(CASE WHEN p.status = 'ACTIVE' THEN 1 ELSE 0 END) AS active,
		SUM(CASE WHEN p.status = 'COMPLETED' THEN 1 ELSE 0 END) AS completed,
		SUM(CASE WHEN p.status = 'ON_HOLD' THEN 1 ELSE 0 END) AS on_hold,
		COALESCE(AVG(p.progress_pct), 0) AS avg_prog
	`).Scan(&pc).Error; err != nil {
		return nil, err
	}
	row.TotalProjects = pc.Total
	row.ActiveProjects = pc.Active
	row.CompletedProjects = pc.Completed
	row.OnHoldProjects = pc.OnHold
	row.AvgProgressPct = pc.AvgProg

	// Budget totals
	type budgetTotals struct {
		Plan   float64
		Actual float64
	}
	var bt budgetTotals
	if err := s.db.Table("project_budgets pb").
		Joins("JOIN projects p ON p.id = pb.project_id").
		Where("p.organization_id = ? AND p.deleted_at IS NULL AND pb.deleted_at IS NULL", orgID).
		Select("COALESCE(SUM(pb.planned),0) AS plan, COALESCE(SUM(pb.actual),0) AS actual").
		Scan(&bt).Error; err != nil {
		return nil, err
	}
	row.TotalBudgetPlan = bt.Plan
	row.TotalBudgetActual = bt.Actual
	if bt.Plan > 0 {
		row.BudgetUsagePct = bt.Actual / bt.Plan * 100
	}

	// Risk counts
	type riskCounts struct {
		Total int
		Open  int
		High  int
	}
	var rc riskCounts
	if err := s.db.Table("risks r").
		Joins("JOIN projects p ON p.id = r.project_id").
		Where("p.organization_id = ? AND p.deleted_at IS NULL AND r.deleted_at IS NULL", orgID).
		Select(`
			COUNT(*) AS total,
			SUM(CASE WHEN r.status NOT IN ('CLOSED','RESOLVED') THEN 1 ELSE 0 END) AS open,
			SUM(CASE WHEN r.severity IN ('HIGH','CRITICAL') AND r.status NOT IN ('CLOSED','RESOLVED') THEN 1 ELSE 0 END) AS high
		`).Scan(&rc).Error; err != nil {
		return nil, err
	}
	row.TotalRisks = rc.Total
	row.OpenRisks = rc.Open
	row.HighRisks = rc.High

	// Issue counts
	type issueCounts struct {
		Total int
		Open  int
	}
	var ic issueCounts
	if err := s.db.Table("issues i").
		Joins("JOIN projects p ON p.id = i.project_id").
		Where("p.organization_id = ? AND p.deleted_at IS NULL AND i.deleted_at IS NULL", orgID).
		Select(`
			COUNT(*) AS total,
			SUM(CASE WHEN i.status NOT IN ('CLOSED','RESOLVED') THEN 1 ELSE 0 END) AS open
		`).Scan(&ic).Error; err != nil {
		return nil, err
	}
	row.TotalIssues = ic.Total
	row.OpenIssues = ic.Open

	// Health class counts (latest snapshot per project — no deleted_at on health_snapshots)
	type healthCounts struct {
		Green    int
		Yellow   int
		Red      int
		Critical int
	}
	var hc healthCounts
	if err := s.db.Table("health_snapshots hs").
		Joins(`JOIN (
			SELECT project_id, MAX(calculated_at) AS latest
			FROM health_snapshots
			GROUP BY project_id
		) lhs ON lhs.project_id = hs.project_id AND lhs.latest = hs.calculated_at`).
		Joins("JOIN projects p ON p.id = hs.project_id").
		Where("p.organization_id = ? AND p.deleted_at IS NULL", orgID).
		Select(`
			SUM(CASE WHEN hs.health_class = 'GREEN' THEN 1 ELSE 0 END) AS green,
			SUM(CASE WHEN hs.health_class = 'YELLOW' THEN 1 ELSE 0 END) AS yellow,
			SUM(CASE WHEN hs.health_class = 'RED' THEN 1 ELSE 0 END) AS red,
			SUM(CASE WHEN hs.health_class = 'CRITICAL' THEN 1 ELSE 0 END) AS critical
		`).Scan(&hc).Error; err != nil {
		return nil, err
	}
	row.GreenHealth = hc.Green
	row.YellowHealth = hc.Yellow
	row.RedHealth = hc.Red
	row.CriticalHealth = hc.Critical

	return row, nil
}

// ---------------------------------------------------------------------------
// Dataset: Project Performance
// ---------------------------------------------------------------------------

func (s *Service) GetProjectPerformance(orgID uuid.UUID, f DatasetFilter) ([]ProjectPerformanceRow, error) {
	type rawRow struct {
		ProjectID        uuid.UUID  `gorm:"column:project_id"`
		ProjectCode      string     `gorm:"column:project_code"`
		ProjectName      string     `gorm:"column:project_name"`
		Status           string     `gorm:"column:status"`
		ProgressPct      float64    `gorm:"column:progress_pct"`
		StartDate        *time.Time `gorm:"column:start_date"`
		EndDate          *time.Time `gorm:"column:end_date"`
		BudgetPlan       float64    `gorm:"column:budget_plan"`
		BudgetActual     float64    `gorm:"column:budget_actual"`
		HealthClass      *string    `gorm:"column:health_class"`
		Province         *string    `gorm:"column:province"`
		PriorityScore    *float64   `gorm:"column:priority_score"`
		PriorityCategory *string    `gorm:"column:priority_category"`
	}

	q := s.db.Table("projects p").
		Select(`
			p.id AS project_id,
			p.code AS project_code,
			p.name AS project_name,
			p.status,
			p.progress_pct,
			p.start_date,
			p.end_date,
			COALESCE(SUM(pb.planned),0) AS budget_plan,
			COALESCE(SUM(pb.actual),0) AS budget_actual,
			lhs.health_class,
			p.province,
			pps.total_score AS priority_score,
			pps.score_category AS priority_category
		`).
		Joins("LEFT JOIN project_budgets pb ON pb.project_id = p.id AND pb.deleted_at IS NULL").
		Joins(`LEFT JOIN (
			SELECT hs.project_id, hs.health_class
			FROM health_snapshots hs
			JOIN (
				SELECT project_id, MAX(calculated_at) AS latest
				FROM health_snapshots
				GROUP BY project_id
			) lh ON lh.project_id = hs.project_id AND lh.latest = hs.calculated_at
		) lhs ON lhs.project_id = p.id`).
		Joins(`LEFT JOIN (
			SELECT pps2.project_id, pps2.total_score, pps2.score_category
			FROM project_priority_scores pps2
			JOIN (
				SELECT project_id, MAX(calculated_at) AS latest
				FROM project_priority_scores
				WHERE organization_id = ?
				GROUP BY project_id
			) lp ON lp.project_id = pps2.project_id AND lp.latest = pps2.calculated_at
		) pps ON pps.project_id = p.id`, orgID).
		Where("p.organization_id = ? AND p.deleted_at IS NULL", orgID).
		Group("p.id, lhs.health_class, pps.total_score, pps.score_category").
		Order("p.code ASC")

	if f.Status != nil && *f.Status != "" {
		q = q.Where("p.status = ?", *f.Status)
	}
	if f.Province != nil && *f.Province != "" {
		q = q.Where("p.province = ?", *f.Province)
	}
	if f.PeriodStart != nil {
		q = q.Where("p.start_date >= ?", *f.PeriodStart)
	}
	if f.PeriodEnd != nil {
		q = q.Where("p.end_date <= ?", *f.PeriodEnd)
	}

	var rows []rawRow
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]ProjectPerformanceRow, len(rows))
	for i, r := range rows {
		usagePct := 0.0
		if r.BudgetPlan > 0 {
			usagePct = r.BudgetActual / r.BudgetPlan * 100
		}
		result[i] = ProjectPerformanceRow{
			ProjectID:        r.ProjectID,
			ProjectCode:      r.ProjectCode,
			ProjectName:      r.ProjectName,
			Status:           r.Status,
			ProgressPct:      r.ProgressPct,
			StartDate:        r.StartDate,
			EndDate:          r.EndDate,
			BudgetPlan:       r.BudgetPlan,
			BudgetActual:     r.BudgetActual,
			BudgetUsagePct:   usagePct,
			HealthClass:      r.HealthClass,
			Province:         r.Province,
			PriorityScore:    r.PriorityScore,
			PriorityCategory: r.PriorityCategory,
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Dataset: Risk & Issue
// ---------------------------------------------------------------------------

func (s *Service) GetRiskIssue(orgID uuid.UUID, f DatasetFilter) ([]RiskIssueRow, error) {
	type rawRow struct {
		ProjectID      uuid.UUID `gorm:"column:project_id"`
		ProjectCode    string    `gorm:"column:project_code"`
		ProjectName    string    `gorm:"column:project_name"`
		TotalRisks     int       `gorm:"column:total_risks"`
		OpenRisks      int       `gorm:"column:open_risks"`
		HighRisks      int       `gorm:"column:high_risks"`
		CriticalRisks  int       `gorm:"column:critical_risks"`
		TotalIssues    int       `gorm:"column:total_issues"`
		OpenIssues     int       `gorm:"column:open_issues"`
		HighIssues     int       `gorm:"column:high_issues"`
		CriticalIssues int       `gorm:"column:critical_issues"`
	}

	q := s.db.Table("projects p").
		Select(`
			p.id AS project_id,
			p.code AS project_code,
			p.name AS project_name,
			COUNT(DISTINCT r.id) AS total_risks,
			SUM(CASE WHEN r.status NOT IN ('CLOSED','RESOLVED') THEN 1 ELSE 0 END) AS open_risks,
			SUM(CASE WHEN r.severity = 'HIGH' AND r.status NOT IN ('CLOSED','RESOLVED') THEN 1 ELSE 0 END) AS high_risks,
			SUM(CASE WHEN r.severity = 'CRITICAL' AND r.status NOT IN ('CLOSED','RESOLVED') THEN 1 ELSE 0 END) AS critical_risks,
			COUNT(DISTINCT i.id) AS total_issues,
			SUM(CASE WHEN i.status NOT IN ('CLOSED','RESOLVED') THEN 1 ELSE 0 END) AS open_issues,
			SUM(CASE WHEN i.severity = 'HIGH' AND i.status NOT IN ('CLOSED','RESOLVED') THEN 1 ELSE 0 END) AS high_issues,
			SUM(CASE WHEN i.severity = 'CRITICAL' AND i.status NOT IN ('CLOSED','RESOLVED') THEN 1 ELSE 0 END) AS critical_issues
		`).
		Joins("LEFT JOIN risks r ON r.project_id = p.id AND r.deleted_at IS NULL").
		Joins("LEFT JOIN issues i ON i.project_id = p.id AND i.deleted_at IS NULL").
		Where("p.organization_id = ? AND p.deleted_at IS NULL", orgID).
		Group("p.id").
		Order("p.code ASC")

	if f.Status != nil && *f.Status != "" {
		q = q.Where("p.status = ?", *f.Status)
	}

	var rows []rawRow
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]RiskIssueRow, len(rows))
	for i, r := range rows {
		result[i] = RiskIssueRow{
			ProjectID:      r.ProjectID,
			ProjectCode:    r.ProjectCode,
			ProjectName:    r.ProjectName,
			TotalRisks:     r.TotalRisks,
			OpenRisks:      r.OpenRisks,
			HighRisks:      r.HighRisks,
			CriticalRisks:  r.CriticalRisks,
			TotalIssues:    r.TotalIssues,
			OpenIssues:     r.OpenIssues,
			HighIssues:     r.HighIssues,
			CriticalIssues: r.CriticalIssues,
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Dataset: Budget
// ---------------------------------------------------------------------------

func (s *Service) GetBudget(orgID uuid.UUID, f DatasetFilter) ([]BudgetRow, error) {
	type rawRow struct {
		ProjectID    uuid.UUID `gorm:"column:project_id"`
		ProjectCode  string    `gorm:"column:project_code"`
		ProjectName  string    `gorm:"column:project_name"`
		Status       string    `gorm:"column:status"`
		BudgetPlan   float64   `gorm:"column:budget_plan"`
		BudgetActual float64   `gorm:"column:budget_actual"`
	}

	q := s.db.Table("projects p").
		Select(`
			p.id AS project_id,
			p.code AS project_code,
			p.name AS project_name,
			p.status,
			COALESCE(SUM(pb.planned),0) AS budget_plan,
			COALESCE(SUM(pb.actual),0) AS budget_actual
		`).
		Joins("LEFT JOIN project_budgets pb ON pb.project_id = p.id AND pb.deleted_at IS NULL").
		Where("p.organization_id = ? AND p.deleted_at IS NULL", orgID).
		Group("p.id").
		Order("p.code ASC")

	if f.Status != nil && *f.Status != "" {
		q = q.Where("p.status = ?", *f.Status)
	}

	var rows []rawRow
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]BudgetRow, len(rows))
	for i, r := range rows {
		variance := r.BudgetActual - r.BudgetPlan
		usagePct := 0.0
		if r.BudgetPlan > 0 {
			usagePct = r.BudgetActual / r.BudgetPlan * 100
		}
		result[i] = BudgetRow{
			ProjectID:    r.ProjectID,
			ProjectCode:  r.ProjectCode,
			ProjectName:  r.ProjectName,
			Status:       r.Status,
			BudgetPlan:   r.BudgetPlan,
			BudgetActual: r.BudgetActual,
			Variance:     variance,
			UsagePct:     usagePct,
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Dataset: Benefits
// ---------------------------------------------------------------------------

func (s *Service) GetBenefits(orgID uuid.UUID, f DatasetFilter) ([]BenefitRow, error) {
	type rawRow struct {
		ProjectID         uuid.UUID `gorm:"column:project_id"`
		ProjectCode       string    `gorm:"column:project_code"`
		ProjectName       string    `gorm:"column:project_name"`
		IndicatorID       uuid.UUID `gorm:"column:indicator_id"`
		IndicatorName     string    `gorm:"column:indicator_name"`
		Unit              string    `gorm:"column:unit"`
		Target            float64   `gorm:"column:target"`
		Actual            float64   `gorm:"column:actual"`
		AggregationMethod string    `gorm:"column:aggregation_method"`
	}

	q := s.db.Table("benefit_indicators bi").
		Select(`
			p.id         AS project_id,
			p.code       AS project_code,
			p.name       AS project_name,
			bi.id        AS indicator_id,
			bi.name      AS indicator_name,
			bi.unit,
			COALESCE(bm_agg.target_val, 0) AS target,
			COALESCE(bm_agg.actual_val, 0) AS actual,
			bi.aggregation_method
		`).
		Joins("JOIN projects p ON p.id = bi.project_id").
		Joins(`LEFT JOIN (
			SELECT
				bm.indicator_id,
				CASE bi2.aggregation_method
					WHEN 'SUM'     THEN SUM(bm.target)
					WHEN 'AVERAGE' THEN AVG(bm.target)
					ELSE (
						SELECT bm_t.target FROM benefit_measurements bm_t
						WHERE bm_t.indicator_id = bm.indicator_id
						  AND bm_t.deleted_at IS NULL
						ORDER BY bm_t.period_year DESC, bm_t.period_month DESC LIMIT 1
					)
				END AS target_val,
				CASE bi2.aggregation_method
					WHEN 'SUM'     THEN SUM(bm.actual)
					WHEN 'AVERAGE' THEN AVG(bm.actual)
					ELSE (
						SELECT bm_a.actual FROM benefit_measurements bm_a
						WHERE bm_a.indicator_id = bm.indicator_id
						  AND bm_a.deleted_at IS NULL
						ORDER BY bm_a.period_year DESC, bm_a.period_month DESC LIMIT 1
					)
				END AS actual_val
			FROM benefit_measurements bm
			JOIN benefit_indicators bi2 ON bi2.id = bm.indicator_id
			WHERE bm.deleted_at IS NULL
			GROUP BY bm.indicator_id, bi2.aggregation_method
		) bm_agg ON bm_agg.indicator_id = bi.id`).
		Where("p.organization_id = ? AND p.deleted_at IS NULL AND bi.deleted_at IS NULL", orgID).
		Order("p.code ASC, bi.name ASC")

	var rows []rawRow
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]BenefitRow, len(rows))
	for i, r := range rows {
		achPct := 0.0
		if r.Target > 0 {
			achPct = r.Actual / r.Target * 100
		}
		result[i] = BenefitRow{
			ProjectID:         r.ProjectID,
			ProjectCode:       r.ProjectCode,
			ProjectName:       r.ProjectName,
			IndicatorID:       r.IndicatorID,
			IndicatorName:     r.IndicatorName,
			Unit:              r.Unit,
			Target:            r.Target,
			Actual:            r.Actual,
			AchievementPct:    achPct,
			AggregationMethod: r.AggregationMethod,
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Dataset: Priority
// ---------------------------------------------------------------------------

func (s *Service) GetPriority(orgID uuid.UUID) ([]PriorityRow, error) {
	type rawRow struct {
		ProjectID    uuid.UUID `gorm:"column:project_id"`
		ProjectCode  string    `gorm:"column:project_code"`
		ProjectName  string    `gorm:"column:project_name"`
		TotalScore   float64   `gorm:"column:total_score"`
		Category     string    `gorm:"column:score_category"`
		CalculatedAt time.Time `gorm:"column:calculated_at"`
	}

	var rows []rawRow
	if err := s.db.Table("project_priority_scores pps").
		Select(`
			p.id AS project_id,
			p.code AS project_code,
			p.name AS project_name,
			pps.total_score,
			pps.score_category,
			pps.calculated_at
		`).
		Joins(`JOIN (
			SELECT project_id, MAX(calculated_at) AS latest
			FROM project_priority_scores
			WHERE organization_id = ?
			GROUP BY project_id
		) lp ON lp.project_id = pps.project_id AND lp.latest = pps.calculated_at`, orgID).
		Joins("JOIN projects p ON p.id = pps.project_id").
		Where("pps.organization_id = ? AND p.deleted_at IS NULL", orgID).
		Order("pps.total_score DESC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]PriorityRow, len(rows))
	for i, r := range rows {
		result[i] = PriorityRow{
			ProjectID:    r.ProjectID,
			ProjectCode:  r.ProjectCode,
			ProjectName:  r.ProjectName,
			TotalScore:   r.TotalScore,
			Category:     r.Category,
			CalculatedAt: r.CalculatedAt,
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Power BI config (env-driven, no secrets exposed)
// ---------------------------------------------------------------------------

func (s *Service) GetPowerBIConfig() PowerBIConfig {
	workspaceID := os.Getenv("POWERBI_WORKSPACE_ID")
	reportID := os.Getenv("POWERBI_REPORT_ID")
	tenantID := os.Getenv("POWERBI_TENANT_ID")
	embedURL := os.Getenv("POWERBI_EMBED_URL")
	authMethod := os.Getenv("POWERBI_AUTH_METHOD") // "service_principal" | "master_user"

	configured := workspaceID != "" && reportID != "" && tenantID != ""

	if authMethod == "" && configured {
		authMethod = "service_principal"
	}

	return PowerBIConfig{
		Configured:  configured,
		WorkspaceID: workspaceID,
		ReportID:    reportID,
		TenantID:    tenantID,
		EmbedURL:    embedURL,
		AuthMethod:  authMethod,
	}
}

// ---------------------------------------------------------------------------
// Export request
// ---------------------------------------------------------------------------

// CreateExportRequest inserts a new PENDING export request.
func (s *Service) CreateExportRequest(orgID uuid.UUID, userID uuid.UUID, input CreateExportRequestInput) (*ReportExportRequest, error) {
	paramsJSON, err := json.Marshal(input.Parameters)
	if err != nil {
		paramsJSON = []byte("{}")
	}

	req := &ReportExportRequest{
		OrganizationID: orgID,
		ReportID:       input.ReportID,
		DatasetKey:     input.DatasetKey,
		Format:         input.Format,
		Status:         ExportStatusPending,
		Parameters:     paramsJSON,
		RequestedBy:    userID,
	}

	if err := s.db.Create(req).Error; err != nil {
		return nil, fmt.Errorf("create export request: %w", err)
	}
	return req, nil
}

// ListExportRequests returns recent export requests for the org (latest 50).
func (s *Service) ListExportRequests(orgID uuid.UUID) ([]ReportExportRequest, error) {
	var reqs []ReportExportRequest
	err := s.db.
		Where("organization_id = ?", orgID).
		Order("created_at DESC").
		Limit(50).
		Find(&reqs).Error
	return reqs, err
}

// GetExportRequest returns a single export request, verifying org ownership.
func (s *Service) GetExportRequest(orgID uuid.UUID, reqID uuid.UUID) (*ReportExportRequest, error) {
	var req ReportExportRequest
	err := s.db.
		Where("id = ? AND organization_id = ?", reqID, orgID).
		First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// ---------------------------------------------------------------------------
// ProcessExportRequest — generate real file (CSV or XLSX) synchronously.
// UAT-002: called inline after CreateExportRequest so export completes in the
// same HTTP request cycle. This keeps the implementation simple; a worker
// queue can replace it later without changing the API contract.
// ---------------------------------------------------------------------------

func (s *Service) ProcessExportRequest(req *ReportExportRequest, userID uuid.UUID) {
	storageRoot := reportStoragePath()

	// Mark PROCESSING
	now := time.Now().UTC()
	_ = s.db.Model(req).Updates(map[string]interface{}{
		"status":     ExportStatusProcessing,
		"started_at": now,
	}).Error
	s.auditLog("report.export.requested", "report_export_requests",
		req.ID, req.OrganizationID, userID,
		map[string]interface{}{
			"dataset_key": req.DatasetKey,
			"format":      string(req.Format),
		})

	// Parse optional filter params from JSONB
	var params map[string]interface{}
	_ = json.Unmarshal(req.Parameters, &params)
	filter := parseFilterParams(params)

	// Build rows for the requested dataset
	rows, err := s.buildDatasetRows(req.OrganizationID, req.DatasetKey, filter)
	if err != nil {
		s.markExportFailed(req, userID, fmt.Sprintf("build dataset rows: %v", err))
		return
	}

	// Generate file in requested format
	var gf *GeneratedFile
	switch req.Format {
	case ExportFormatCSV:
		gf, err = generateCSV(storageRoot, req.OrganizationID, req.DatasetKey, rows)
	case ExportFormatXLSX:
		gf, err = generateXLSX(storageRoot, req.OrganizationID, req.DatasetKey, rows)
	default:
		// PDF not yet implemented — fall back to CSV
		gf, err = generateCSV(storageRoot, req.OrganizationID, req.DatasetKey, rows)
	}
	if err != nil {
		s.markExportFailed(req, userID, fmt.Sprintf("generate file: %v", err))
		return
	}

	// Mark COMPLETED with file metadata
	completedAt := time.Now().UTC()
	fileName := gf.FileName
	storageKey := gf.StorageKey
	mimeType := gf.MimeType
	fileSize := gf.FileSizeB

	if err := s.db.Model(req).Updates(map[string]interface{}{
		"status":       ExportStatusCompleted,
		"completed_at": completedAt,
		"file_name":    fileName,
		"storage_key":  storageKey,
		"mime_type":    mimeType,
		"file_size":    fileSize,
		"generated_at": gf.GeneratedAt,
	}).Error; err != nil {
		s.markExportFailed(req, userID, fmt.Sprintf("update completed: %v", err))
		return
	}
	// Refresh in-memory struct for handler response
	req.Status = ExportStatusCompleted
	req.FileName = &fileName
	req.StorageKey = &storageKey
	req.MimeType = &mimeType
	req.FileSize = &fileSize
	req.CompletedAt = &completedAt
	req.GeneratedAt = &gf.GeneratedAt

	s.auditLog("report.export.completed", "report_export_requests",
		req.ID, req.OrganizationID, userID,
		map[string]interface{}{
			"dataset_key": req.DatasetKey,
			"format":      string(req.Format),
			"file_name":   fileName,
			"file_size":   fileSize,
		})
}

func (s *Service) markExportFailed(req *ReportExportRequest, userID uuid.UUID, msg string) {
	errMsg := msg
	_ = s.db.Model(req).Updates(map[string]interface{}{
		"status":        ExportStatusFailed,
		"error_message": errMsg,
	}).Error
	req.Status = ExportStatusFailed
	req.ErrorMessage = &errMsg
	s.auditLog("report.export.failed", "report_export_requests",
		req.ID, req.OrganizationID, userID,
		map[string]interface{}{
			"dataset_key":   req.DatasetKey,
			"error_message": msg,
		})
}

// buildDatasetRows returns header+data rows for a given dataset key.
func (s *Service) buildDatasetRows(orgID uuid.UUID, datasetKey string, filter DatasetFilter) ([][]string, error) {
	switch datasetKey {
	case "executive-summary":
		data, err := s.GetExecutiveSummary(orgID, filter)
		if err != nil {
			return nil, err
		}
		return executiveSummaryRows(data), nil

	case "project-performance":
		data, err := s.GetProjectPerformance(orgID, filter)
		if err != nil {
			return nil, err
		}
		return projectPerformanceRows(data), nil

	case "risk-issue":
		data, err := s.GetRiskIssue(orgID, filter)
		if err != nil {
			return nil, err
		}
		return riskIssueRows(data), nil

	case "budget":
		data, err := s.GetBudget(orgID, filter)
		if err != nil {
			return nil, err
		}
		return budgetRows(data), nil

	case "benefits":
		data, err := s.GetBenefits(orgID, filter)
		if err != nil {
			return nil, err
		}
		return benefitRows(data), nil

	case "priority":
		data, err := s.GetPriority(orgID)
		if err != nil {
			return nil, err
		}
		return priorityRows(data), nil

	default:
		return nil, fmt.Errorf("unknown dataset key: %s", datasetKey)
	}
}

// parseFilterParams converts the JSONB params map into a DatasetFilter.
func parseFilterParams(params map[string]interface{}) DatasetFilter {
	f := DatasetFilter{}
	if params == nil {
		return f
	}
	if v, ok := params["status"].(string); ok && v != "" {
		f.Status = &v
	}
	if v, ok := params["province"].(string); ok && v != "" {
		f.Province = &v
	}
	if v, ok := params["program_id"].(string); ok && v != "" {
		f.ProgramID = &v
	}
	if v, ok := params["period_start"].(string); ok && v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.PeriodStart = &t
		}
	}
	if v, ok := params["period_end"].(string); ok && v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.PeriodEnd = &t
		}
	}
	return f
}

// ---------------------------------------------------------------------------
// DownloadExportFile — serve the physical file, tenant-safe.
// Returns: absPath, mimeType, fileName, error.
// Audit event report.export.downloaded is recorded here.
// ---------------------------------------------------------------------------

func (s *Service) DownloadExportFile(orgID uuid.UUID, reqID uuid.UUID, userID uuid.UUID) (absPath string, mimeType string, fileName string, err error) {
	req, err := s.GetExportRequest(orgID, reqID)
	if err != nil {
		return "", "", "", err
	}
	if req.Status != ExportStatusCompleted || req.StorageKey == nil {
		return "", "", "", fmt.Errorf("export not ready: status=%s", req.Status)
	}

	storageRoot := reportStoragePath()
	abs, err := resolveAbsPath(storageRoot, *req.StorageKey)
	if err != nil {
		return "", "", "", err
	}
	if _, err := os.Stat(abs); err != nil {
		return "", "", "", fmt.Errorf("file not found on disk")
	}

	fn := ""
	if req.FileName != nil {
		fn = *req.FileName
	}
	mt := "application/octet-stream"
	if req.MimeType != nil {
		mt = *req.MimeType
	}

	s.auditLog("report.export.downloaded", "report_export_requests",
		req.ID, req.OrganizationID, userID,
		map[string]interface{}{
			"dataset_key": req.DatasetKey,
			"file_name":   fn,
		})

	return abs, mt, fn, nil
}
