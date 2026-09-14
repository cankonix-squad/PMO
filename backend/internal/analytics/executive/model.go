package executive

import (
	"github.com/google/uuid"
)

// NationalSummary adalah KPI nasional level eksekutif (FR-CTL1-01).
type NationalSummary struct {
	// Proyek
	TotalProjects  int64 `json:"total_projects"`
	ActiveProjects int64 `json:"active_projects"`
	DraftProjects  int64 `json:"draft_projects"`

	// Anggaran (total dari projects.budget_total)
	TotalBudget    float64 `json:"total_budget"`
	BudgetRealized float64 `json:"budget_realized"`
	BudgetUsagePct float64 `json:"budget_usage_pct"`

	// Progress fisik rata-rata
	AvgPhysicalProgress float64 `json:"avg_physical_progress"`

	// Health distribution
	HealthGreen    int64 `json:"health_green"`
	HealthYellow   int64 `json:"health_yellow"`
	HealthRed      int64 `json:"health_red"`
	HealthCritical int64 `json:"health_critical"`
	HealthUnscored int64 `json:"health_unscored"`

	// Risk & issue
	OpenRisks      int64 `json:"open_risks"`
	HighRisks      int64 `json:"high_risks"`
	OpenIssues     int64 `json:"open_issues"`
	CriticalIssues int64 `json:"critical_issues"`
	OverdueActions int64 `json:"overdue_actions"`

	// Command center
	OpenEscalations  int64 `json:"open_escalations"`
	PendingDecisions int64 `json:"pending_decisions"`
	OverdueDecisions int64 `json:"overdue_decisions"`

	// Benefit
	BenefitIndicators int64 `json:"benefit_indicators"`

	// Metadata
	AsOf string `json:"as_of"`
}

// CriticalProject adalah proyek dengan health CRITICAL/RED atau prioritas tinggi (FR-CTL1-02).
type CriticalProject struct {
	ProjectID      uuid.UUID `json:"project_id"`
	ProjectCode    string    `json:"project_code"`
	ProjectName    string    `json:"project_name"`
	Status         string    `json:"status"`
	HealthClass    string    `json:"health_class"`
	PhysicalActual float64   `json:"physical_actual"`
	BudgetTotal    float64   `json:"budget_total"`
	OpenRisks      int64     `json:"open_risks"`
	OpenIssues     int64     `json:"open_issues"`
	PriorityScore  float64   `json:"priority_score"`
	PriorityClass  string    `json:"priority_class"`
	ProgramName    string    `json:"program_name"`
	SectorName     string    `json:"sector_name"`
}

// EscalationItem adalah satu eskalasi yang perlu perhatian eksekutif.
type EscalationItem struct {
	ID          uuid.UUID  `json:"id"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty"`
	ProjectName string     `json:"project_name,omitempty"`
	Level       string     `json:"level"`
	SourceType  string     `json:"source_type"`
	Reason      string     `json:"reason"`
	Status      string     `json:"status"`
	CreatedAt   string     `json:"created_at"`
}

// DecisionItem adalah satu keputusan eksekutif yang pending.
type DecisionItem struct {
	ID           uuid.UUID  `json:"id"`
	ProjectID    *uuid.UUID `json:"project_id,omitempty"`
	ProjectName  string     `json:"project_name,omitempty"`
	Subject      string     `json:"subject"`
	DecisionText string     `json:"decision_text"`
	Status       string     `json:"status"`
	DueDate      *string    `json:"due_date,omitempty"`
	IsOverdue    bool       `json:"is_overdue"`
	CreatedAt    string     `json:"created_at"`
}

// BenefitSummary adalah ringkasan indikator manfaat nasional (FR-CTL1-04).
type BenefitSummary struct {
	TotalIndicators int64         `json:"total_indicators"`
	OnTrackCount    int64         `json:"on_track_count"`
	BehindCount     int64         `json:"behind_count"`
	Indicators      []BenefitItem `json:"indicators"`
}

// BenefitItem adalah satu indikator manfaat.
type BenefitItem struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Unit              string    `json:"unit"`
	TargetValue       float64   `json:"target_value"`
	ActualValue       float64   `json:"actual_value"`
	AchievementPct    float64   `json:"achievement_pct"`
	AggregationMethod string    `json:"aggregation_method"`
}

// ProgramKPISummary adalah ringkasan KPI per program untuk tabel perbandingan.
type ProgramKPISummary struct {
	ProgramID      uuid.UUID `json:"program_id"`
	ProgramCode    string    `json:"program_code"`
	ProgramName    string    `json:"program_name"`
	TotalProjects  int64     `json:"total_projects"`
	ActiveProjects int64     `json:"active_projects"`
	TotalBudget    float64   `json:"total_budget"`
	AvgProgress    float64   `json:"avg_physical_progress"`
	HealthGreen    int64     `json:"health_green"`
	HealthYellow   int64     `json:"health_yellow"`
	HealthRed      int64     `json:"health_red"`
	HealthCritical int64     `json:"health_critical"`
	OpenRisks      int64     `json:"open_risks"`
	OpenIssues     int64     `json:"open_issues"`
}

// ExecutiveDashboard adalah payload lengkap dashboard Level 1 (FR-CTL1-01 s/d FR-CTL1-05).
type ExecutiveDashboard struct {
	Summary          NationalSummary     `json:"summary"`
	CriticalProjects []CriticalProject   `json:"critical_projects"`
	Escalations      []EscalationItem    `json:"escalations"`
	PendingDecisions []DecisionItem      `json:"pending_decisions"`
	Programs         []ProgramKPISummary `json:"programs"`
	Benefits         BenefitSummary      `json:"benefits"`
	AsOf             string              `json:"as_of"`
}

// Filter untuk periode dan scope.
type Filter struct {
	PeriodYear  int
	PeriodMonth int
}
