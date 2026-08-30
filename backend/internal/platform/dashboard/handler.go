package dashboard

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// StatsRow holds aggregated dashboard stats.
type StatsRow struct {
	// Projects
	TotalProjects  int64 `json:"total_projects"`
	ActiveProjects int64 `json:"active_projects"`
	OnHoldProjects int64 `json:"on_hold_projects"`
	ClosedProjects int64 `json:"closed_projects"`

	// Tasks
	TotalTasks      int64 `json:"total_tasks"`
	TodoTasks       int64 `json:"todo_tasks"`
	InProgressTasks int64 `json:"in_progress_tasks"`
	DoneTasks       int64 `json:"done_tasks"`
	OverdueTasks    int64 `json:"overdue_tasks"`

	// Milestones
	TotalMilestones   int64 `json:"total_milestones"`
	PendingMilestones int64 `json:"pending_milestones"`
	DoneMilestones    int64 `json:"done_milestones"`
	OverdueMilestones int64 `json:"overdue_milestones"`

	// Users
	TotalUsers  int64 `json:"total_users"`
	ActiveUsers int64 `json:"active_users"`

	EarlyWarnings []EarlyWarning `json:"early_warnings"`
}

// EarlyWarning is a compact executive-facing alert for dashboard risk signals.
type EarlyWarning struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Severity    string     `json:"severity"`
	Title       string     `json:"title"`
	Message     string     `json:"message"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty"`
	ProjectName string     `json:"project_name,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Value       float64    `json:"value,omitempty"`
	Threshold   float64    `json:"threshold,omitempty"`
}

// Handler serves dashboard aggregation endpoints.
type Handler struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewHandler creates a new dashboard Handler.
func NewHandler(db *gorm.DB, log *zap.Logger) *Handler {
	return &Handler{db: db, log: log}
}

// claimsFromGin extracts JWT claims from the Gin context.
func claimsFromGin(c *gin.Context) *auth.Claims {
	v, ok := c.Get(string(auth.ContextKeyClaims))
	if !ok {
		return nil
	}
	cl, ok := v.(*auth.Claims)
	if !ok {
		return nil
	}
	return cl
}

// Get godoc
// GET /api/v1/dashboard
func (h *Handler) Get(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	stats, err := h.aggregate(c.Request.Context(), claims.OrganizationID)
	if err != nil {
		h.log.Error("dashboard: aggregate", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.OK(c, "ok", stats)
}

func (h *Handler) aggregate(ctx context.Context, orgID uuid.UUID) (*StatsRow, error) {
	today := beginningOfDay(time.Now())
	var s StatsRow

	// ---- Projects -----------------------------------------------------------
	type statusCount struct {
		Status string
		Count  int64
	}
	var projectCounts []statusCount
	if err := h.db.WithContext(ctx).
		Table("projects").
		Select("status, count(*) as count").
		Where("organization_id = ? AND deleted_at IS NULL", orgID).
		Group("status").
		Scan(&projectCounts).Error; err != nil {
		return nil, err
	}
	for _, pc := range projectCounts {
		s.TotalProjects += pc.Count
		switch pc.Status {
		case "IN_PROGRESS", "ACTIVE":
			s.ActiveProjects += pc.Count
		case "ON_HOLD":
			s.OnHoldProjects += pc.Count
		case "CLOSED", "DONE", "COMPLETED":
			s.ClosedProjects += pc.Count
		}
	}

	// ---- Tasks --------------------------------------------------------------
	// Only count tasks whose parent project is active (not soft-deleted).
	var taskCounts []statusCount
	if err := h.db.WithContext(ctx).
		Table("tasks t").
		Select("t.status, count(*) as count").
		Joins("JOIN projects p ON p.id = t.project_id AND p.organization_id = t.organization_id AND p.deleted_at IS NULL").
		Where("t.organization_id = ? AND t.deleted_at IS NULL", orgID).
		Group("t.status").
		Scan(&taskCounts).Error; err != nil {
		return nil, err
	}
	for _, tc := range taskCounts {
		s.TotalTasks += tc.Count
		switch tc.Status {
		case "TODO":
			s.TodoTasks += tc.Count
		case "IN_PROGRESS":
			s.InProgressTasks += tc.Count
		case "DONE", "COMPLETED":
			s.DoneTasks += tc.Count
		}
	}

	// Overdue tasks — due_date in past and not done; parent project must be active
	if err := h.db.WithContext(ctx).
		Table("tasks t").
		Joins("JOIN projects p ON p.id = t.project_id AND p.organization_id = t.organization_id AND p.deleted_at IS NULL").
		Where("t.organization_id = ? AND t.deleted_at IS NULL AND t.due_date < ? AND t.status NOT IN ('DONE','COMPLETED','CANCELLED')", orgID, today).
		Count(&s.OverdueTasks).Error; err != nil {
		return nil, err
	}

	// ---- Milestones ---------------------------------------------------------
	// Only count milestones whose parent project is active (not soft-deleted).
	var msCounts []statusCount
	if err := h.db.WithContext(ctx).
		Table("milestones m").
		Select("m.status, count(*) as count").
		Joins("JOIN projects p ON p.id = m.project_id AND p.organization_id = m.organization_id AND p.deleted_at IS NULL").
		Where("m.organization_id = ? AND m.deleted_at IS NULL", orgID).
		Group("m.status").
		Scan(&msCounts).Error; err != nil {
		return nil, err
	}
	for _, mc := range msCounts {
		s.TotalMilestones += mc.Count
		switch mc.Status {
		case "PENDING":
			s.PendingMilestones += mc.Count
		case "DONE", "COMPLETED":
			s.DoneMilestones += mc.Count
		}
	}

	// Overdue milestones — parent project must be active (not soft-deleted)
	if err := h.db.WithContext(ctx).
		Table("milestones m").
		Joins("JOIN projects p ON p.id = m.project_id AND p.organization_id = m.organization_id AND p.deleted_at IS NULL").
		Where("m.organization_id = ? AND m.deleted_at IS NULL AND m.due_date < ? AND m.status NOT IN ('DONE','COMPLETED','CANCELLED')", orgID, today).
		Count(&s.OverdueMilestones).Error; err != nil {
		return nil, err
	}

	// ---- Users --------------------------------------------------------------
	if err := h.db.WithContext(ctx).
		Table("users").
		Where("organization_id = ? AND deleted_at IS NULL", orgID).
		Count(&s.TotalUsers).Error; err != nil {
		return nil, err
	}
	if err := h.db.WithContext(ctx).
		Table("users").
		Where("organization_id = ? AND deleted_at IS NULL AND is_active = true", orgID).
		Count(&s.ActiveUsers).Error; err != nil {
		return nil, err
	}

	warnings, err := h.earlyWarnings(ctx, orgID, today)
	if err != nil {
		return nil, err
	}
	s.EarlyWarnings = warnings

	return &s, nil
}

func beginningOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func (h *Handler) earlyWarnings(ctx context.Context, orgID uuid.UUID, now time.Time) ([]EarlyWarning, error) {
	warnings := make([]EarlyWarning, 0, 12)

	taskWarnings, err := h.overdueTaskWarnings(ctx, orgID, now)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, taskWarnings...)

	milestoneWarnings, err := h.overdueMilestoneWarnings(ctx, orgID, now)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, milestoneWarnings...)

	progressWarnings, err := h.lowProgressWarnings(ctx, orgID, now)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, progressWarnings...)

	budgetWarnings, err := h.budgetThresholdWarnings(ctx, orgID)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, budgetWarnings...)

	riskWarnings, err := h.riskRegisterWarnings(ctx, orgID)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, riskWarnings...)

	return warnings, nil
}

// riskRegisterWarnings surfaces real risk-register entries (from the risks
// table) as executive-facing warnings, ordered by computed risk_score descending.
// This makes the dashboard "Risiko Utama" read from the actual risk register
// instead of only generic early-warning signals. Only open risks on active
// (non-soft-deleted) parent projects are included.
func (h *Handler) riskRegisterWarnings(ctx context.Context, orgID uuid.UUID) ([]EarlyWarning, error) {
	type row struct {
		ID          uuid.UUID
		Title       string
		ProjectID   uuid.UUID
		ProjectName string
		RiskScore   int
		Severity    string
		DueDate     *time.Time
	}
	var rows []row
	if err := h.db.WithContext(ctx).
		Table("risks r").
		Select("r.id, r.title, r.project_id, p.name as project_name, r.risk_score, r.severity, r.due_date").
		Joins("JOIN projects p ON p.id = r.project_id AND p.organization_id = r.organization_id AND p.deleted_at IS NULL").
		Where("r.organization_id = ? AND r.deleted_at IS NULL AND r.status NOT IN ('CLOSED','ACCEPTED','MITIGATED')", orgID).
		Order("r.risk_score DESC").
		Limit(6).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	warnings := make([]EarlyWarning, 0, len(rows))
	for _, r := range rows {
		projectID := r.ProjectID
		severity := r.Severity
		if severity == "" {
			severity = "MEDIUM"
		}
		riskID := r.ID
		dueDate := r.DueDate
		warnings = append(warnings, EarlyWarning{
			ID:          "risk-register-" + riskID.String(),
			Type:        "RISK_REGISTER",
			Severity:    severity,
			Title:       r.Title,
			Message:     "Open risk from the project risk register.",
			ProjectID:   &projectID,
			ProjectName: r.ProjectName,
			DueDate:     dueDate,
			Value:       float64(r.RiskScore),
			Threshold:   0,
		})
	}

	return warnings, nil
}

func (h *Handler) overdueTaskWarnings(ctx context.Context, orgID uuid.UUID, now time.Time) ([]EarlyWarning, error) {
	type row struct {
		ID          uuid.UUID
		Title       string
		ProjectID   uuid.UUID
		ProjectName string
		DueDate     time.Time
	}
	var rows []row
	if err := h.db.WithContext(ctx).
		Table("tasks t").
		Select("t.id, t.title, t.project_id, p.name as project_name, t.due_date").
		Joins("JOIN projects p ON p.id = t.project_id AND p.organization_id = t.organization_id AND p.deleted_at IS NULL").
		Where("t.organization_id = ? AND t.deleted_at IS NULL AND t.due_date < ? AND t.status NOT IN ('DONE','COMPLETED','CANCELLED')", orgID, now).
		Order("t.due_date ASC").
		Limit(5).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	warnings := make([]EarlyWarning, 0, len(rows))
	for _, r := range rows {
		dueDate := r.DueDate
		projectID := r.ProjectID
		warnings = append(warnings, EarlyWarning{
			ID:          "task-overdue-" + r.ID.String(),
			Type:        "OVERDUE_TASK",
			Severity:    "HIGH",
			Title:       r.Title,
			Message:     "Task is past due and still open.",
			ProjectID:   &projectID,
			ProjectName: r.ProjectName,
			DueDate:     &dueDate,
		})
	}

	return warnings, nil
}

func (h *Handler) overdueMilestoneWarnings(ctx context.Context, orgID uuid.UUID, now time.Time) ([]EarlyWarning, error) {
	type row struct {
		ID          uuid.UUID
		Title       string
		ProjectID   uuid.UUID
		ProjectName string
		DueDate     time.Time
	}
	var rows []row
	if err := h.db.WithContext(ctx).
		Table("milestones m").
		Select("m.id, m.title, m.project_id, p.name as project_name, m.due_date").
		Joins("JOIN projects p ON p.id = m.project_id AND p.organization_id = m.organization_id AND p.deleted_at IS NULL").
		Where("m.organization_id = ? AND m.deleted_at IS NULL AND m.due_date < ? AND m.status NOT IN ('DONE','COMPLETED','CANCELLED')", orgID, now).
		Order("m.due_date ASC").
		Limit(5).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	warnings := make([]EarlyWarning, 0, len(rows))
	for _, r := range rows {
		dueDate := r.DueDate
		projectID := r.ProjectID
		warnings = append(warnings, EarlyWarning{
			ID:          "milestone-overdue-" + r.ID.String(),
			Type:        "OVERDUE_MILESTONE",
			Severity:    "HIGH",
			Title:       r.Title,
			Message:     "Milestone is past due and still open.",
			ProjectID:   &projectID,
			ProjectName: r.ProjectName,
			DueDate:     &dueDate,
		})
	}

	return warnings, nil
}

func (h *Handler) lowProgressWarnings(ctx context.Context, orgID uuid.UUID, now time.Time) ([]EarlyWarning, error) {
	const nearEndDays = 14
	const progressThreshold = 80.0

	type row struct {
		ID          uuid.UUID
		Name        string
		EndDate     time.Time
		ProgressPct float64
	}
	var rows []row
	if err := h.db.WithContext(ctx).
		Table("projects").
		Select("id, name, end_date, progress_pct").
		Where(
			"organization_id = ? AND deleted_at IS NULL AND end_date IS NOT NULL AND end_date BETWEEN ? AND ? AND progress_pct < ? AND status NOT IN ('COMPLETED','CLOSED','DONE','CANCELLED')",
			orgID,
			now,
			now.AddDate(0, 0, nearEndDays),
			progressThreshold,
		).
		Order("end_date ASC").
		Limit(5).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	warnings := make([]EarlyWarning, 0, len(rows))
	for _, r := range rows {
		dueDate := r.EndDate
		projectID := r.ID
		warnings = append(warnings, EarlyWarning{
			ID:          "project-low-progress-" + r.ID.String(),
			Type:        "LOW_PROGRESS_NEAR_END",
			Severity:    "MEDIUM",
			Title:       r.Name,
			Message:     "Project is near its end date with progress below target.",
			ProjectID:   &projectID,
			ProjectName: r.Name,
			DueDate:     &dueDate,
			Value:       r.ProgressPct,
			Threshold:   progressThreshold,
		})
	}

	return warnings, nil
}

// TrendPoint holds one month's aggregated physical and financial progress.
type TrendPoint struct {
	Month        string  `json:"month"`         // "YYYY-MM"
	PhysicalPct  float64 `json:"physical_pct"`  // avg progress % across active projects
	FinancialPct float64 `json:"financial_pct"` // actual/planned * 100 for the month's budgets
	Planned      float64 `json:"planned"`       // total planned budget recorded up to month
	Actual       float64 `json:"actual"`        // total actual budget recorded up to month
	DataType     string  `json:"data_type"`     // "OPERATIONAL" or "SNAPSHOT"
}

// TrendResponse is the payload for GET /api/v1/dashboard/trend.
type TrendResponse struct {
	Points   []TrendPoint `json:"points"`
	DataType string       `json:"data_type"` // "OPERATIONAL" — no official snapshot yet
}

// GetTrend godoc
// GET /api/v1/dashboard/trend
// Returns last 12 months of physical progress (from project_progress_history) and
// financial progress (from project_budgets updated_at per month), aggregated per month.
// Tenant-safe: all queries scope to orgID via projects.organization_id.
// Label: data_type = "OPERATIONAL" (no official snapshot required).
func (h *Handler) GetTrend(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	result, err := h.trendData(c.Request.Context(), claims.OrganizationID, 12)
	if err != nil {
		h.log.Error("dashboard: trendData", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.OK(c, "ok", result)
}

// trendData aggregates progress and budget data per calendar month for the last nMonths.
// Priority: periodic reports (project_periodic_reports) → progress history + budget → operational fallback.
func (h *Handler) trendData(ctx context.Context, orgID uuid.UUID, nMonths int) (*TrendResponse, error) {
	// Build the list of last nMonths calendar months (oldest first).
	now := time.Now()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	months := make([]string, nMonths)
	for i := 0; i < nMonths; i++ {
		t := currentMonth.AddDate(0, -(nMonths - 1 - i), 0)
		months[i] = t.Format("2006-01")
	}
	startMonth := currentMonth.AddDate(0, -(nMonths - 1), 0)

	// ---- 1. Try periodic reports first (PMO-DASH-002) ----------------
	// Aggregate per month: avg physical, sum(financial_actual)/sum(financial_planned)*100.
	// Only include reports whose parent project is active (not soft-deleted).
	type periodicRow struct {
		Month     string
		AvgPhys   float64
		SumPlan   float64
		SumActual float64
	}
	var periodicRows []periodicRow
	if err := h.db.WithContext(ctx).
		Table("project_periodic_reports pr").
		Select(`TO_CHAR(DATE_TRUNC('month', MAKE_DATE(pr.period_year::int, pr.period_month::int, 1)), 'YYYY-MM') AS month,
			AVG(pr.physical_progress_pct) AS avg_phys,
			COALESCE(SUM(pr.financial_planned), 0) AS sum_plan,
			COALESCE(SUM(pr.financial_actual), 0) AS sum_actual`).
		Joins("JOIN projects p ON p.id = pr.project_id AND p.organization_id = ? AND p.deleted_at IS NULL", orgID).
		Where("pr.organization_id = ? AND pr.deleted_at IS NULL AND MAKE_DATE(pr.period_year::int, pr.period_month::int, 1) >= ?", orgID, startMonth).
		Group("TO_CHAR(DATE_TRUNC('month', MAKE_DATE(pr.period_year::int, pr.period_month::int, 1)), 'YYYY-MM')").
		Order("month ASC").
		Scan(&periodicRows).Error; err != nil {
		return nil, err
	}

	if len(periodicRows) > 0 {
		// Have periodic report data — use it as the primary source.
		type periodAgg struct {
			physPct float64
			planned float64
			actual  float64
			finPct  float64
		}
		periodMap := make(map[string]periodAgg, len(periodicRows))
		for _, r := range periodicRows {
			finPct := 0.0
			if r.SumPlan > 0 {
				finPct = (r.SumActual / r.SumPlan) * 100
			}
			periodMap[r.Month] = periodAgg{
				physPct: r.AvgPhys,
				planned: r.SumPlan,
				actual:  r.SumActual,
				finPct:  finPct,
			}
		}

		points := make([]TrendPoint, 0, nMonths)
		var lastPhys, lastFin, lastPlan, lastActual float64
		for _, m := range months {
			phys := lastPhys
			fin := lastFin
			plan := lastPlan
			act := lastActual
			if v, ok := periodMap[m]; ok {
				phys = v.physPct
				fin = v.finPct
				plan = v.planned
				act = v.actual
				lastPhys = phys
				lastFin = fin
				lastPlan = plan
				lastActual = act
			}
			points = append(points, TrendPoint{
				Month:        m,
				PhysicalPct:  roundTwo(phys),
				FinancialPct: roundTwo(fin),
				Planned:      roundTwo(plan),
				Actual:       roundTwo(act),
				DataType:     "PERIODIC_REPORT",
			})
		}
		return &TrendResponse{
			Points:   points,
			DataType: "PERIODIC_REPORT",
		}, nil
	}

	// ---- 2. Fallback: progress history + budget (existing logic) ----------

	// --- Physical progress: avg progress_pct from project_progress_history per month ---
	type physRow struct {
		Month       string
		AvgProgress float64
	}
	var physRows []physRow
	if err := h.db.WithContext(ctx).
		Table("project_progress_history pph").
		Select("TO_CHAR(pph.recorded_at, 'YYYY-MM') AS month, AVG(pph.progress_pct) AS avg_progress").
		Joins("JOIN projects p ON p.id = pph.project_id AND p.organization_id = ? AND p.deleted_at IS NULL", orgID).
		Where("pph.recorded_at >= ?", startMonth).
		Group("TO_CHAR(pph.recorded_at, 'YYYY-MM')").
		Order("month ASC").
		Scan(&physRows).Error; err != nil {
		return nil, err
	}
	physMap := make(map[string]float64, len(physRows))
	for _, r := range physRows {
		physMap[r.Month] = r.AvgProgress
	}

	// --- Financial progress: SUM(planned) / SUM(actual) from project_budgets per month ---
	type budgetRow struct {
		Month   string
		Planned float64
		Actual  float64
	}
	var budgetRows []budgetRow
	if err := h.db.WithContext(ctx).
		Table("project_budgets pb").
		Select("TO_CHAR(pb.updated_at, 'YYYY-MM') AS month, COALESCE(SUM(pb.planned), 0) AS planned, COALESCE(SUM(pb.actual), 0) AS actual").
		Joins("JOIN projects p ON p.id = pb.project_id AND p.organization_id = ? AND p.deleted_at IS NULL", orgID).
		Where("pb.deleted_at IS NULL AND pb.updated_at >= ?", startMonth).
		Group("TO_CHAR(pb.updated_at, 'YYYY-MM')").
		Order("month ASC").
		Scan(&budgetRows).Error; err != nil {
		return nil, err
	}
	type budgetAgg struct {
		planned float64
		actual  float64
	}
	budgetMap := make(map[string]budgetAgg, len(budgetRows))
	for _, r := range budgetRows {
		budgetMap[r.Month] = budgetAgg{planned: r.Planned, actual: r.Actual}
	}

	if len(physMap) == 0 && len(budgetMap) == 0 {
		return h.currentOperationalTrend(ctx, orgID, months)
	}
	if len(physMap) == 0 {
		currentProgress, err := h.currentAverageProgress(ctx, orgID)
		if err != nil {
			return nil, err
		}
		for i, month := range months {
			physMap[month] = currentProgress * (float64(i+1) / float64(len(months)))
		}
	}
	if len(budgetMap) == 0 {
		currentBudget, err := h.currentBudgetAggregate(ctx, orgID)
		if err != nil {
			return nil, err
		}
		for i, month := range months {
			ramp := float64(i+1) / float64(len(months))
			budgetMap[month] = budgetAgg{
				planned: currentBudget.Planned * ramp,
				actual:  currentBudget.Actual * ramp,
			}
		}
	}

	// Build result: one point per month. Carry forward last known values for
	// months that have no new data (so the chart isn't artificially empty).
	points := make([]TrendPoint, 0, nMonths)
	var lastPhysical, lastFinancial, lastPlanned, lastActual float64
	for _, m := range months {
		physical := lastPhysical
		if v, ok := physMap[m]; ok {
			physical = v
			lastPhysical = v
		}

		financial := lastFinancial
		planned := lastPlanned
		actual := lastActual
		if b, ok := budgetMap[m]; ok {
			planned = b.planned
			actual = b.actual
			if b.planned > 0 {
				financial = (b.actual / b.planned) * 100
			}
			lastFinancial = financial
			lastPlanned = planned
			lastActual = actual
		}

		points = append(points, TrendPoint{
			Month:        m,
			PhysicalPct:  roundTwo(physical),
			FinancialPct: roundTwo(financial),
			Planned:      roundTwo(planned),
			Actual:       roundTwo(actual),
			DataType:     "OPERATIONAL",
		})
	}

	return &TrendResponse{
		Points:   points,
		DataType: "OPERATIONAL",
	}, nil
}

func (h *Handler) currentOperationalTrend(ctx context.Context, orgID uuid.UUID, months []string) (*TrendResponse, error) {
	avgProgress, err := h.currentAverageProgress(ctx, orgID)
	if err != nil {
		return nil, err
	}

	budget, err := h.currentBudgetAggregate(ctx, orgID)
	if err != nil {
		return nil, err
	}

	financialPct := 0.0
	if budget.Planned > 0 {
		financialPct = (budget.Actual / budget.Planned) * 100
	}

	points := make([]TrendPoint, 0, len(months))
	for i, month := range months {
		ramp := float64(i+1) / float64(len(months))
		points = append(points, TrendPoint{
			Month:        month,
			PhysicalPct:  roundTwo(avgProgress * ramp),
			FinancialPct: roundTwo(financialPct * ramp),
			Planned:      roundTwo(budget.Planned * ramp),
			Actual:       roundTwo(budget.Actual * ramp),
			DataType:     "OPERATIONAL",
		})
	}

	return &TrendResponse{
		Points:   points,
		DataType: "OPERATIONAL",
	}, nil
}

func (h *Handler) currentAverageProgress(ctx context.Context, orgID uuid.UUID) (float64, error) {
	type progressRow struct {
		AvgProgress float64
	}
	var progress progressRow
	if err := h.db.WithContext(ctx).
		Table("projects").
		Select("COALESCE(AVG(progress_pct), 0) AS avg_progress").
		Where("organization_id = ? AND deleted_at IS NULL AND status <> ?", orgID, "CANCELLED").
		Scan(&progress).Error; err != nil {
		return 0, err
	}
	return progress.AvgProgress, nil
}

type currentBudgetRow struct {
	Planned float64
	Actual  float64
}

func (h *Handler) currentBudgetAggregate(ctx context.Context, orgID uuid.UUID) (currentBudgetRow, error) {
	var budget currentBudgetRow
	err := h.db.WithContext(ctx).
		Table("project_budgets pb").
		Select("COALESCE(SUM(pb.planned), 0) AS planned, COALESCE(SUM(pb.actual), 0) AS actual").
		Joins("JOIN projects p ON p.id = pb.project_id AND p.organization_id = ? AND p.deleted_at IS NULL", orgID).
		Where("pb.deleted_at IS NULL AND p.status <> ?", "CANCELLED").
		Scan(&budget).Error
	return budget, err
}

// roundTwo rounds a float64 to 2 decimal places.
func roundTwo(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func (h *Handler) budgetThresholdWarnings(ctx context.Context, orgID uuid.UUID) ([]EarlyWarning, error) {
	const budgetThreshold = 90.0

	type row struct {
		ProjectID   uuid.UUID
		ProjectName string
		Planned     float64
		Actual      float64
	}
	var rows []row
	if err := h.db.WithContext(ctx).
		Table("project_budgets pb").
		Select("p.id as project_id, p.name as project_name, COALESCE(SUM(pb.planned), 0) as planned, COALESCE(SUM(pb.actual), 0) as actual").
		Joins("JOIN projects p ON p.id = pb.project_id AND p.organization_id = ? AND p.deleted_at IS NULL", orgID).
		Where("pb.deleted_at IS NULL").
		Group("p.id, p.name").
		Having("COALESCE(SUM(pb.planned), 0) > 0 AND (COALESCE(SUM(pb.actual), 0) / COALESCE(SUM(pb.planned), 0)) * 100 >= ?", budgetThreshold).
		Order("(COALESCE(SUM(pb.actual), 0) / COALESCE(SUM(pb.planned), 0)) DESC").
		Limit(5).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	warnings := make([]EarlyWarning, 0, len(rows))
	for _, r := range rows {
		usagePct := (r.Actual / r.Planned) * 100
		projectID := r.ProjectID
		severity := "MEDIUM"
		if usagePct >= 100 {
			severity = "HIGH"
		}
		warnings = append(warnings, EarlyWarning{
			ID:          "budget-threshold-" + r.ProjectID.String(),
			Type:        "BUDGET_THRESHOLD",
			Severity:    severity,
			Title:       r.ProjectName,
			Message:     "Budget usage has crossed the monitoring threshold (operational input, not yet validated).",
			ProjectID:   &projectID,
			ProjectName: r.ProjectName,
			Value:       usagePct,
			Threshold:   budgetThreshold,
		})
	}

	return warnings, nil
}
