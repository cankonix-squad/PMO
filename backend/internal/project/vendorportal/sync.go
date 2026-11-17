package vendorportal

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	notification "github.com/harmanto-49/cankora/internal/notifications/delivery"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Notifier is the subset of the notification service used by the portal.
type Notifier interface {
	Enqueue(req notification.EnqueueRequest) error
}

// SetNotifier enables in-app notifications (optional; nil disables them).
func (s *Service) SetNotifier(n Notifier, log *zap.Logger) {
	s.notifier, s.log = n, log
}

// periodicNote marks rows created from verified contractor reports.
const periodicNote = "Otomatis dari laporan progres penyedia terverifikasi"

type contractProgress struct {
	ContractID    uuid.UUID
	ContractValue float64
	Pct           *float64
}

// projectPhysicalPct computes the project's physical progress for a period as
// the contract-value-weighted average of each active contract's latest
// VERIFIED cumulative progress up to that period. Contracts without a verified
// report count as 0%. If every contract value is 0, contracts weigh equally.
func projectPhysicalPct(rows []contractProgress) float64 {
	if len(rows) == 0 {
		return 0
	}
	var total, weighted float64
	for _, r := range rows {
		total += r.ContractValue
	}
	for _, r := range rows {
		pct := 0.0
		if r.Pct != nil {
			pct = *r.Pct
		}
		w := r.ContractValue
		if total <= 0 {
			w = 1
		}
		weighted += w * pct
	}
	if total <= 0 {
		total = float64(len(rows))
	}
	v := weighted / total
	if v > 100 {
		v = 100
	}
	return float64(int64(v*100+0.5)) / 100
}

// syncPeriodic writes the verified progress into project_periodic_reports
// (physical % only; financial figures of an existing row are kept) and then
// refreshes projects.progress_pct from the latest periodic row.
func syncPeriodic(tx *gorm.DB, orgID, projectID uuid.UUID, year, month int, reviewer uuid.UUID, now time.Time) (float64, error) {
	var rows []contractProgress
	err := tx.Raw(`
SELECT c.id AS contract_id, COALESCE(c.contract_value, 0) AS contract_value,
       (SELECT r.physical_progress_pct FROM vendor_progress_reports r
         WHERE r.contract_id = c.id AND r.organization_id = c.organization_id
           AND r.status = 'VERIFIED' AND r.deleted_at IS NULL
           AND (r.period_year < ? OR (r.period_year = ? AND r.period_month <= ?))
         ORDER BY r.period_year DESC, r.period_month DESC LIMIT 1) AS pct
FROM contracts c
WHERE c.organization_id = ? AND c.project_id = ? AND c.deleted_at IS NULL
  AND c.status NOT IN ('DRAFT', 'TERMINATED')`, year, year, month, orgID, projectID).Scan(&rows).Error
	if err != nil {
		return 0, err
	}
	pct := projectPhysicalPct(rows)

	var existingID *uuid.UUID
	var ids []uuid.UUID
	if err := tx.Table("project_periodic_reports").
		Where("organization_id = ? AND project_id = ? AND period_year = ? AND period_month = ? AND deleted_at IS NULL", orgID, projectID, year, month).
		Limit(1).Pluck("id", &ids).Error; err != nil {
		return 0, err
	}
	if len(ids) > 0 {
		existingID = &ids[0]
	}
	if existingID != nil {
		if err := tx.Exec(`UPDATE project_periodic_reports SET physical_progress_pct = ?, updated_at = ? WHERE id = ?`, pct, now, *existingID).Error; err != nil {
			return 0, err
		}
	} else if err := tx.Exec(`
INSERT INTO project_periodic_reports
  (organization_id, project_id, period_year, period_month, physical_progress_pct,
   financial_planned, financial_actual, financial_pct, notes, reported_by, reported_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, 0, 0, 0, ?, ?, ?, ?, ?)`, orgID, projectID, year, month, pct, periodicNote, reviewer, now, now, now).Error; err != nil {
		return 0, err
	}

	// Same rule as project.SyncProjectProgressFromPeriodic: latest period wins.
	err = tx.Exec(`
UPDATE projects SET progress_pct = sub.pct, updated_at = ?
FROM (SELECT physical_progress_pct AS pct FROM project_periodic_reports
      WHERE organization_id = ? AND project_id = ? AND deleted_at IS NULL
      ORDER BY period_year DESC, period_month DESC LIMIT 1) sub
WHERE projects.id = ? AND projects.organization_id = ? AND projects.deleted_at IS NULL AND sub.pct > 0`,
		now, orgID, projectID, projectID, orgID).Error
	return pct, err
}

// notify sends best-effort in-app notifications; failures are only logged.
func (s *Service) notify(orgID uuid.UUID, recipients []uuid.UUID, priority, subject, body, sourceType string, sourceID uuid.UUID) {
	if s.notifier == nil {
		return
	}
	seen := map[uuid.UUID]bool{}
	for _, uid := range recipients {
		if uid == uuid.Nil || seen[uid] {
			continue
		}
		seen[uid] = true
		u := uid
		err := s.notifier.Enqueue(notification.EnqueueRequest{
			OrganizationID: orgID, RecipientUserID: &u, Priority: priority,
			Subject: subject, Body: body, SourceType: sourceType, SourceID: sourceID.String(),
		})
		if err != nil && s.log != nil {
			s.log.Warn("vendor portal notification failed", zap.Error(err))
		}
	}
}

// reviewersFor returns the contract PPK, or — when no PPK is set — active
// SUPER_ADMIN/ADMIN/PMO users of the organization.
func (s *Service) reviewersFor(orgID, contractID uuid.UUID) []uuid.UUID {
	var ppk []uuid.UUID
	s.db.Table("contracts").Where("id = ? AND organization_id = ? AND ppk_user_id IS NOT NULL", contractID, orgID).Pluck("ppk_user_id", &ppk)
	if len(ppk) > 0 {
		return ppk
	}
	var ids []uuid.UUID
	s.db.Table("users u").Distinct("u.id").
		Joins("JOIN user_roles ur ON ur.user_id = u.id").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Where("u.organization_id = ? AND u.is_active = TRUE AND u.deleted_at IS NULL AND r.code IN ?", orgID, PrivilegedReviewRoles).
		Pluck("u.id", &ids)
	return ids
}

func (s *Service) contextLabel(orgID, contractID uuid.UUID) string {
	ctx, _ := s.contexts(orgID, []uuid.UUID{contractID})
	c := ctx[contractID]
	return fmt.Sprintf("%s — kontrak %s (%s)", c.ProjectName, c.ContractNumber, c.VendorName)
}

func periodLabel(year, month int) string {
	months := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	if month < 1 || month > 12 {
		return fmt.Sprintf("%d", year)
	}
	return fmt.Sprintf("%s %d", months[month-1], year)
}
