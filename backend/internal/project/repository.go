package project

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// postgresRepository is the GORM implementation.
type postgresRepository struct {
	db *gorm.DB
}

// NewRepository creates a new project Repository.
func NewRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, p *Project) error {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrCodeTaken
		}
		return err
	}
	return nil
}

// isUniqueViolation reports whether err is a PostgreSQL unique constraint violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func (r *postgresRepository) FindByID(ctx context.Context, id, orgID uuid.UUID) (*Project, error) {
	var p Project
	err := r.db.WithContext(ctx).
		Preload("Milestones", "organization_id = ?", orgID).
		Preload("Team").
		Where("projects.id = ? AND projects.organization_id = ?", id, orgID).
		First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProjectNotFound
	}
	// Populate org_unit_name via separate query (avoids Preload conflict with GORM '-' tag)
	if p.OrgUnitID != nil {
		var name string
		if e := r.db.WithContext(ctx).
			Raw("SELECT name FROM org_units WHERE id = ? AND deleted_at IS NULL", p.OrgUnitID).
			Scan(&name).Error; e == nil {
			p.OrgUnitName = name
		}
	}
	return &p, err
}

func (r *postgresRepository) List(ctx context.Context, filter ProjectListFilter) ([]Project, int64, error) {
	query := r.db.WithContext(ctx).Model(&Project{}).
		Where("organization_id = ?", filter.OrganizationID)

	if filter.OrgUnitID != nil {
		query = query.Where("org_unit_id = ?", filter.OrgUnitID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Priority != "" {
		query = query.Where("priority = ?", filter.Priority)
	}
	if filter.ManagerID != nil {
		query = query.Where("manager_id = ?", filter.ManagerID)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ?", s, s)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	var projects []Project
	err := query.Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&projects).Error
	if err != nil {
		return nil, 0, err
	}

	// Batch-resolve org_unit names: collect unique non-nil org_unit_ids,
	// fetch names in one query, then populate OrgUnitName on each project.
	orgUnitIDs := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]bool)
	for _, p := range projects {
		if p.OrgUnitID != nil && !seen[*p.OrgUnitID] {
			orgUnitIDs = append(orgUnitIDs, *p.OrgUnitID)
			seen[*p.OrgUnitID] = true
		}
	}
	if len(orgUnitIDs) > 0 {
		type orgUnitRow struct {
			ID   uuid.UUID
			Name string
		}
		var ouRows []orgUnitRow
		if e := r.db.WithContext(ctx).
			Raw("SELECT id, name FROM org_units WHERE id IN ? AND deleted_at IS NULL", orgUnitIDs).
			Scan(&ouRows).Error; e == nil {
			nameMap := make(map[uuid.UUID]string, len(ouRows))
			for _, ou := range ouRows {
				nameMap[ou.ID] = ou.Name
			}
			for i := range projects {
				if projects[i].OrgUnitID != nil {
					projects[i].OrgUnitName = nameMap[*projects[i].OrgUnitID]
				}
			}
		}
	}
	return projects, total, err
}

func (r *postgresRepository) Update(ctx context.Context, p *Project) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *postgresRepository) MasterDataExists(ctx context.Context, table string, id, orgID uuid.UUID) (bool, error) {
	allowed := map[string]bool{
		"org_units":    true,
		"programs":     true,
		"sectors":      true,
		"regions":      true,
		"river_basins": true,
	}
	if !allowed[table] {
		return false, fmt.Errorf("project: unsupported master table %s", table)
	}

	var count int64
	err := r.db.WithContext(ctx).
		Table(table).
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, orgID).
		Count(&count).Error
	return count > 0, err
}

func (r *postgresRepository) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, orgID).
		Delete(&Project{}).Error
}

// DeleteCascade soft-deletes a project AND all of its business children inside
// a single transaction. Every statement is scoped by project_id (and where the
// child table carries organization_id, also by orgID) so cross-tenant writes
// are impossible. Child rows without a deleted_at column (e.g. progress
// history) are hard-removed because they are historical snapshots, not
// business records.
func (r *postgresRepository) DeleteCascade(ctx context.Context, id, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Scope child tables to the project. Tables with organization_id also
		// scope by orgID as a defensive tenant check.
		if err := tx.Where("project_id = ?", id).Delete(&Task{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&Milestone{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&Issue{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&Risk{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&ProjectBudget{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&Contract{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&ProjectDocument{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&CorrectiveAction{}).Error; err != nil {
			return err
		}
		if err := tx.Exec(
			`UPDATE benefit_measurements bm
			 SET deleted_at = NOW()
			 FROM benefit_indicators bi
			 WHERE bi.id = bm.indicator_id
			   AND bi.project_id = ?
			   AND bi.organization_id = ?
			   AND bm.organization_id = ?
			   AND bm.deleted_at IS NULL`,
			id, orgID, orgID,
		).Error; err != nil {
			return err
		}
		if err := tx.Exec(
			`UPDATE benefit_indicators
			 SET deleted_at = NOW()
			 WHERE project_id = ?
			   AND organization_id = ?
			   AND deleted_at IS NULL`,
			id, orgID,
		).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&TeamMember{}).Error; err != nil {
			return err
		}

		// task_assignments are soft-deleted via their parent tasks in GORM
		// association hooks; we also cascade explicitly for safety.
		if err := tx.Exec(
			"UPDATE task_assignments ta SET deleted_at = NOW() FROM tasks t WHERE t.id = ta.task_id AND t.project_id = ? AND ta.deleted_at IS NULL",
			id,
		).Error; err != nil {
			return err
		}

		// Progress history has no deleted_at — it is an immutable snapshot
		// series. Removing the rows is the correct "soft delete" equivalent
		// for a deleted project (no business row remains accessible).
		if err := tx.Where("project_id = ?", id).Delete(&ProgressHistory{}).Error; err != nil {
			return err
		}

		// Finally soft-delete the project itself.
		if err := tx.Where("id = ? AND organization_id = ?", id, orgID).Delete(&Project{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *postgresRepository) AddTeamMember(ctx context.Context, m *TeamMember) error {
	return r.db.WithContext(ctx).
		Where(TeamMember{ProjectID: m.ProjectID, UserID: m.UserID}).
		FirstOrCreate(m).Error
}

func (r *postgresRepository) RemoveTeamMember(ctx context.Context, projectID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Delete(&TeamMember{}).Error
}

func (r *postgresRepository) GetTeam(ctx context.Context, projectID uuid.UUID) ([]TeamMember, error) {
	var members []TeamMember
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&members).Error
	return members, err
}

func (r *postgresRepository) RecordProgress(ctx context.Context, h *ProgressHistory) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *postgresRepository) GetProgressHistory(ctx context.Context, projectID uuid.UUID) ([]ProgressHistory, error) {
	var history []ProgressHistory
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("recorded_at DESC").
		Find(&history).Error
	return history, err
}
