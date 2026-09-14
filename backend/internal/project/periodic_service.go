package project

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SyncProjectProgressFromPeriodic updates project.progress_pct from the latest
// periodic report for that project. Called after create/update of a periodic report.
// Errors are logged but not surfaced — this is best-effort.
func (s *Service) SyncProjectProgressFromPeriodic(ctx context.Context, orgID, projectID uuid.UUID) {
	if err := s.repo.SyncProjectProgressFromPeriodic(ctx, orgID, projectID); err != nil {
		s.log.Warn("SyncProjectProgressFromPeriodic failed", zap.Error(err))
	}
}
