package projectcontrol

import "github.com/google/uuid"

type Control struct {
	Project  ProjectSummary   `json:"project"`
	AsOf     string           `json:"as_of"`
	Contract ContractSummary  `json:"contract"`
	Snapshot *SnapshotSummary `json:"snapshot,omitempty"`
	Health   *HealthSummary   `json:"health,omitempty"`
	Evidence EvidenceSummary  `json:"evidence"`
	Issues   []Item           `json:"issues"`
	Risks    []Item           `json:"risks"`
	Actions  []Item           `json:"actions"`
}
type ProjectSummary struct {
	ID       uuid.UUID `json:"id"`
	Code     string    `json:"code"`
	Name     string    `json:"name"`
	Status   string    `json:"status"`
	Progress float64   `json:"progress_pct"`
	Budget   float64   `json:"budget_total"`
	Currency string    `json:"currency"`
}
type ContractSummary struct {
	Count       int64   `json:"count"`
	TotalValue  float64 `json:"total_value"`
	ActiveCount int64   `json:"active_count"`
	Currency    string  `json:"currency"`
}
type SnapshotSummary struct {
	PeriodYear            int     `json:"period_year"`
	PeriodMonth           int     `json:"period_month"`
	PhysicalActual        float64 `json:"physical_actual"`
	PhysicalTarget        float64 `json:"physical_target"`
	PhysicalVariance      float64 `json:"physical_variance"`
	FinancialActual       float64 `json:"financial_actual"`
	FinancialTarget       float64 `json:"financial_target"`
	FinancialVariance     float64 `json:"financial_variance"`
	Currency              string  `json:"currency"`
	ScheduleDeviationDays *int    `json:"schedule_deviation_days,omitempty"`
	Status                string  `json:"status"`
	Source                string  `json:"source,omitempty"`
}
type HealthSummary struct {
	Score       float64   `json:"score"`
	Class       string    `json:"class"`
	FormulaID   uuid.UUID `json:"formula_id"`
	Explanation string    `json:"explanation"`
}
type EvidenceSummary struct {
	Inspections         int64 `json:"inspections"`
	VerifiedInspections int64 `json:"verified_inspections"`
	EvidenceFiles       int64 `json:"evidence_files"`
}
type Item struct {
	ID       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	Status   string    `json:"status"`
	Severity string    `json:"severity,omitempty"`
	DueDate  string    `json:"due_date,omitempty"`
}
