package gis

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service menyediakan data GIS proyek dari database.
type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// GetProjects mengembalikan semua marker proyek milik org, dengan filter opsional.
func (s *Service) GetProjects(ctx context.Context, orgID uuid.UUID, f GISFilter) ([]GISProjectMarker, error) {
	type row struct {
		ProjectID     uuid.UUID
		ProjectCode   string
		ProjectName   string
		Status        string
		HealthClass   string
		ProgressPct   float64
		BudgetTotal   float64
		PriorityScore float64
		Latitude      *float64
		Longitude     *float64
		Province      string
		City          string
		LocationName  string
		RegionName    string
		OpenRisks     int64
		OpenIssues    int64
	}

	q := s.db.WithContext(ctx).
		Table("projects p").
		Select(`
			p.id                                                        AS project_id,
			p.code                                                      AS project_code,
			p.name                                                      AS project_name,
			p.status                                                    AS status,
			COALESCE(hs.health_class, 'UNSCORED')                      AS health_class,
			p.progress_pct                                              AS progress_pct,
			p.budget_total                                              AS budget_total,
			COALESCE(pps.total_score, 0)                               AS priority_score,
			p.latitude                                                  AS latitude,
			p.longitude                                                 AS longitude,
			COALESCE(p.province, '')                                   AS province,
			COALESCE(p.city, '')                                       AS city,
			COALESCE(p.location_name, '')                              AS location_name,
			COALESCE(r.name, '')                                       AS region_name,
			(SELECT COUNT(*) FROM risks ri
			 WHERE ri.project_id = p.id
			   AND ri.status NOT IN ('CLOSED','RESOLVED')
			   AND ri.deleted_at IS NULL)                              AS open_risks,
			(SELECT COUNT(*) FROM issues i
			 WHERE i.project_id = p.id
			   AND i.status NOT IN ('CLOSED','RESOLVED')
			   AND i.deleted_at IS NULL)                               AS open_issues
		`).
		Joins(`LEFT JOIN (
			SELECT DISTINCT ON (project_id) project_id, health_class
			FROM health_snapshots
			ORDER BY project_id, calculated_at DESC
		) hs ON hs.project_id = p.id`).
		Joins(`LEFT JOIN (
			SELECT DISTINCT ON (project_id) project_id, total_score
			FROM project_priority_scores
			ORDER BY project_id, calculated_at DESC
		) pps ON pps.project_id = p.id`).
		Joins("LEFT JOIN regions r ON r.id = p.region_id AND r.deleted_at IS NULL").
		Where("p.organization_id = ? AND p.deleted_at IS NULL", orgID)

	if f.Status != "" {
		q = q.Where("p.status = ?", f.Status)
	}
	if f.HealthClass != "" {
		q = q.Where("COALESCE(hs.health_class, 'UNSCORED') = ?", f.HealthClass)
	}
	if f.Province != "" {
		q = q.Where("p.province ILIKE ?", "%"+f.Province+"%")
	}

	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	markers := make([]GISProjectMarker, 0, len(rows))
	for _, r := range rows {
		markers = append(markers, GISProjectMarker{
			ProjectID:     r.ProjectID,
			ProjectCode:   r.ProjectCode,
			ProjectName:   r.ProjectName,
			Status:        r.Status,
			HealthClass:   r.HealthClass,
			ProgressPct:   r.ProgressPct,
			BudgetTotal:   r.BudgetTotal,
			PriorityScore: r.PriorityScore,
			Latitude:      r.Latitude,
			Longitude:     r.Longitude,
			Province:      r.Province,
			City:          r.City,
			LocationName:  r.LocationName,
			RegionName:    r.RegionName,
			OpenRisks:     r.OpenRisks,
			OpenIssues:    r.OpenIssues,
		})
	}
	return markers, nil
}

// GetSummary mengembalikan ringkasan statistik GIS untuk panel samping peta.
func (s *Service) GetSummary(ctx context.Context, orgID uuid.UUID) (*GISSummary, error) {
	type countRow struct {
		Total    int64
		Mapped   int64
		AvgPct   float64
		Green    int64
		Yellow   int64
		Red      int64
		Critical int64
	}

	var cr countRow
	err := s.db.WithContext(ctx).
		Table("projects p").
		Select(`
			COUNT(*)                                                         AS total,
			COUNT(*) FILTER (WHERE p.latitude IS NOT NULL AND p.longitude IS NOT NULL) AS mapped,
			COALESCE(AVG(p.progress_pct), 0)                                AS avg_pct,
			COUNT(*) FILTER (WHERE COALESCE(hs.health_class,'UNSCORED') = 'GREEN')    AS green,
			COUNT(*) FILTER (WHERE COALESCE(hs.health_class,'UNSCORED') = 'YELLOW')   AS yellow,
			COUNT(*) FILTER (WHERE COALESCE(hs.health_class,'UNSCORED') = 'RED')      AS red,
			COUNT(*) FILTER (WHERE COALESCE(hs.health_class,'UNSCORED') = 'CRITICAL') AS critical
		`).
		Joins(`LEFT JOIN (
			SELECT DISTINCT ON (project_id) project_id, health_class
			FROM health_snapshots
			ORDER BY project_id, calculated_at DESC
		) hs ON hs.project_id = p.id`).
		Where("p.organization_id = ? AND p.deleted_at IS NULL", orgID).
		Scan(&cr).Error
	if err != nil {
		return nil, err
	}

	unscored := cr.Total - cr.Green - cr.Yellow - cr.Red - cr.Critical
	return &GISSummary{
		TotalProjects:    cr.Total,
		MappedProjects:   cr.Mapped,
		UnmappedProjects: cr.Total - cr.Mapped,
		AvgProgressPct:   cr.AvgPct,
		HealthGreen:      cr.Green,
		HealthYellow:     cr.Yellow,
		HealthRed:        cr.Red,
		HealthCritical:   cr.Critical,
		HealthUnscored:   unscored,
	}, nil
}

// GetProjectDetail mengembalikan detail satu proyek untuk popup marker.
func (s *Service) GetProjectDetail(ctx context.Context, orgID uuid.UUID, projectID uuid.UUID) (*GISProjectMarker, error) {
	markers, err := s.GetProjects(ctx, orgID, GISFilter{})
	if err != nil {
		return nil, err
	}
	for i := range markers {
		if markers[i].ProjectID == projectID {
			return &markers[i], nil
		}
	}
	return nil, nil
}

// nowStr helper untuk metadata as_of timestamp.
func nowStr() string { return time.Now().UTC().Format(time.RFC3339) }
