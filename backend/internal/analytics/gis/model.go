package gis

import "github.com/google/uuid"

// GISProjectMarker adalah data satu proyek untuk ditampilkan sebagai marker di peta.
type GISProjectMarker struct {
	ProjectID     uuid.UUID `json:"project_id"`
	ProjectCode   string    `json:"project_code"`
	ProjectName   string    `json:"project_name"`
	Status        string    `json:"status"`
	HealthClass   string    `json:"health_class"` // GREEN/YELLOW/RED/CRITICAL/UNSCORED
	ProgressPct   float64   `json:"progress_pct"`
	BudgetTotal   float64   `json:"budget_total"`
	PriorityScore float64   `json:"priority_score"`
	Latitude      *float64  `json:"latitude"`
	Longitude     *float64  `json:"longitude"`
	Province      string    `json:"province"`
	City          string    `json:"city"`
	LocationName  string    `json:"location_name"`
	RegionName    string    `json:"region_name"`
	OpenRisks     int64     `json:"open_risks"`
	OpenIssues    int64     `json:"open_issues"`
}

// GISSummary adalah ringkasan statistik untuk panel samping peta.
type GISSummary struct {
	TotalProjects    int64   `json:"total_projects"`
	MappedProjects   int64   `json:"mapped_projects"` // punya lat/lng
	UnmappedProjects int64   `json:"unmapped_projects"`
	AvgProgressPct   float64 `json:"avg_progress_pct"`
	HealthGreen      int64   `json:"health_green"`
	HealthYellow     int64   `json:"health_yellow"`
	HealthRed        int64   `json:"health_red"`
	HealthCritical   int64   `json:"health_critical"`
	HealthUnscored   int64   `json:"health_unscored"`
}

// GISFilter adalah filter opsional untuk endpoint GIS.
type GISFilter struct {
	Status      string // filter status proyek, misal "ACTIVE"
	HealthClass string // filter health class, misal "RED"
	Province    string // filter provinsi
}
