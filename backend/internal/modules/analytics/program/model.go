package program

import "github.com/google/uuid"

// ProgramKPI adalah agregasi KPI untuk satu program atau sektor.
type ProgramKPI struct {
	GroupID        uuid.UUID `json:"group_id"`
	GroupCode      string    `json:"group_code"`
	GroupName      string    `json:"group_name"`
	GroupType      string    `json:"group_type"` // "program" | "sector"
	TotalProjects  int64     `json:"total_projects"`
	ActiveProjects int64     `json:"active_projects"`
	// Budget agregasi
	TotalBudget    float64 `json:"total_budget"`
	BudgetRealized float64 `json:"budget_realized"`
	BudgetUsagePct float64 `json:"budget_usage_pct"`
	Currency       string  `json:"currency"`
	// Progress fisik (rata-rata dari project_snapshots VALID terbaru)
	AvgPhysicalActual float64 `json:"avg_physical_actual"`
	AvgPhysicalTarget float64 `json:"avg_physical_target"`
	PhysicalVariance  float64 `json:"physical_variance"`
	// Health distribution
	HealthGreen    int64 `json:"health_green"`
	HealthYellow   int64 `json:"health_yellow"`
	HealthRed      int64 `json:"health_red"`
	HealthCritical int64 `json:"health_critical"`
	HealthUnscored int64 `json:"health_unscored"`
	// Risk & issue counts
	OpenRisks      int64 `json:"open_risks"`
	HighRisks      int64 `json:"high_risks"`
	OpenIssues     int64 `json:"open_issues"`
	CriticalIssues int64 `json:"critical_issues"`
	OverdueActions int64 `json:"overdue_actions"`
	// Priority score agregasi
	AvgPriorityScore float64 `json:"avg_priority_score"`
	CriticalPriority int64   `json:"critical_priority_count"`
	// Benefit summary
	BenefitIndicators int64 `json:"benefit_indicators"`
	// Metadata
	AsOf string `json:"as_of"`
}

// ProjectRow adalah ringkasan satu proyek di dalam program, untuk tabel drill-down.
type ProjectRow struct {
	ProjectID      uuid.UUID `json:"project_id"`
	ProjectCode    string    `json:"project_code"`
	ProjectName    string    `json:"project_name"`
	Status         string    `json:"status"`
	PhysicalActual float64   `json:"physical_actual"`
	PhysicalTarget float64   `json:"physical_target"`
	PhysicalVar    float64   `json:"physical_variance"`
	BudgetTotal    float64   `json:"budget_total"`
	BudgetUsagePct float64   `json:"budget_usage_pct"`
	HealthClass    string    `json:"health_class"`
	HealthScore    float64   `json:"health_score"`
	OpenRisks      int64     `json:"open_risks"`
	OpenIssues     int64     `json:"open_issues"`
	PriorityScore  float64   `json:"priority_score"`
	PriorityClass  string    `json:"priority_class"`
}

// TopDeviation adalah proyek dengan deviasi waktu atau biaya tertinggi.
type TopDeviation struct {
	ProjectID   uuid.UUID `json:"project_id"`
	ProjectCode string    `json:"project_code"`
	ProjectName string    `json:"project_name"`
	Value       float64   `json:"value"` // variance % atau hari
	Label       string    `json:"label"` // e.g. "-12.5% fisik" atau "+45 hari"
	GroupName   string    `json:"group_name"`
}

// ProgramDashboard adalah response lengkap untuk satu program/sektor.
type ProgramDashboard struct {
	KPI              ProgramKPI     `json:"kpi"`
	Projects         []ProjectRow   `json:"projects"`
	TopPhysicalDev   []TopDeviation `json:"top_physical_deviation"`
	TopBudgetDev     []TopDeviation `json:"top_budget_deviation"`
	HighRiskProjects []ProjectRow   `json:"high_risk_projects"`
	AsOf             string         `json:"as_of"`
}

// ListResponse adalah response untuk daftar semua program/sektor dengan KPI masing-masing.
type ListResponse struct {
	Groups []ProgramKPI `json:"groups"`
	AsOf   string       `json:"as_of"`
}

// Filter untuk query program dashboard.
type Filter struct {
	PeriodYear  int
	PeriodMonth int
	RegionID    *uuid.UUID
	OrgUnitID   *uuid.UUID
}
