package project

import (
	"context"

	"github.com/google/uuid"
)

func (r *postgresRepository) SyncProjectProgressFromPeriodic(ctx context.Context, orgID, projectID uuid.UUID) error {
	// Fetch the latest physical_progress_pct from periodic reports
	type latestRow struct{ PhysicalProgressPct float64 }
	var row latestRow
	err := r.db.WithContext(ctx).
		Table("project_periodic_reports").
		Select("physical_progress_pct").
		Where("organization_id = ? AND project_id = ? AND deleted_at IS NULL", orgID, projectID).
		Order("period_year DESC, period_month DESC").
		Limit(1).
		Scan(&row).Error
	if err != nil {
		return err
	}
	if row.PhysicalProgressPct <= 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).
		Model(&Project{}).
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL", projectID, orgID).
		Update("progress_pct", row.PhysicalProgressPct).Error; err != nil {
		return err
	}
	return nil
}
