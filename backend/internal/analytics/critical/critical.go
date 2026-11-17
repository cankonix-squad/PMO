// Package critical holds the single definition of a "critical project" used
// by both the operational Dashboard and the Executive (Level 1) dashboard.
package critical

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Joins attaches the latest health snapshot (hs) and priority score (pps) to
// projects p. Callers must select from "projects p".
const Joins = `
		LEFT JOIN LATERAL (
			SELECT health_class FROM health_snapshots
			WHERE project_id = p.id ORDER BY calculated_at DESC LIMIT 1
		) hs ON true
		LEFT JOIN LATERAL (
			SELECT total_score, score_category FROM project_priority_scores
			WHERE project_id = p.id ORDER BY calculated_at DESC LIMIT 1
		) pps ON true`

// Predicate: latest health is CRITICAL/RED, or latest priority category is
// CRITICAL/HIGH. Use with Joins.
const Predicate = `(COALESCE(hs.health_class, 'UNSCORED') IN ('CRITICAL', 'RED')
			OR COALESCE(pps.score_category, '') IN ('CRITICAL', 'HIGH'))`

// Count returns the number of critical, non-deleted projects of an organization.
func Count(ctx context.Context, db *gorm.DB, orgID uuid.UUID) (int64, error) {
	var n int64
	err := db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM projects p`+Joins+`
		WHERE p.organization_id = ? AND p.deleted_at IS NULL AND `+Predicate, orgID).Scan(&n).Error
	return n, err
}
