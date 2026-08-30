package primavera

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/audit"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------------

// ErrNotFound is returned when a sync run does not exist for the organisation.
var ErrNotFound = errors.New("primavera sync run not found")

// ErrForbidden is returned when a sync run belongs to another tenant.
var ErrForbidden = errors.New("forbidden")

// ErrInvalidTransition is returned when a status transition is not allowed.
var ErrInvalidTransition = errors.New("invalid status transition")

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// Service handles Primavera P6 sync run business logic.
type Service struct {
	db          *gorm.DB
	auditWriter *audit.Writer
	log         *zap.Logger
}

// NewService creates a new Primavera integration Service.
func NewService(db *gorm.DB, auditWriter *audit.Writer, log *zap.Logger) *Service {
	return &Service{db: db, auditWriter: auditWriter, log: log}
}

// ---------------------------------------------------------------------------
// List / Get
// ---------------------------------------------------------------------------

// ListRuns returns paginated sync runs for an organisation.
func (s *Service) ListRuns(ctx context.Context, orgID uuid.UUID, f ListRunsFilter) ([]SyncRun, int64, error) {
	q := s.db.WithContext(ctx).Model(&SyncRun{}).Where("organization_id = ?", orgID)
	if f.ProjectID != "" {
		if pid, err := uuid.Parse(f.ProjectID); err == nil {
			q = q.Where("project_id = ?", pid)
		}
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Format != "" {
		q = q.Where("format = ?", f.Format)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := f.Page
	if page <= 0 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var runs []SyncRun
	err := q.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&runs).Error
	return runs, total, err
}

// GetRun returns a single sync run, enforcing tenant ownership.
func (s *Service) GetRun(ctx context.Context, orgID, runID uuid.UUID) (*SyncRun, error) {
	var run SyncRun
	err := s.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", runID, orgID).
		First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &run, err
}

// ListMappings returns activity mappings for a given sync run, tenant-scoped.
func (s *Service) ListMappings(ctx context.Context, orgID, runID uuid.UUID, page, pageSize int) ([]ActivityMapping, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	q := s.db.WithContext(ctx).Model(&ActivityMapping{}).
		Where("organization_id = ? AND sync_run_id = ?", orgID, runID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var mappings []ActivityMapping
	err := q.Order("created_at ASC").Limit(pageSize).Offset(offset).Find(&mappings).Error
	return mappings, total, err
}

// ---------------------------------------------------------------------------
// Create (upload + queue)
// ---------------------------------------------------------------------------

// CreateRun persists a new PENDING sync run record.
// The caller must separately call ProcessRun to trigger actual parsing/import.
func (s *Service) CreateRun(
	ctx context.Context,
	orgID, userID uuid.UUID,
	projectID *uuid.UUID,
	fileName, mimeType, format string,
	fileSize int64,
	lineageMeta LineageMeta,
) (*SyncRun, error) {
	// normalise format
	format = normaliseFormat(format, fileName)

	// validate project belongs to this tenant before inserting (FK constraint)
	if projectID != nil {
		var projectExists int64
		s.db.WithContext(ctx).Table("projects").
			Where("id = ? AND organization_id = ? AND deleted_at IS NULL", *projectID, orgID).
			Count(&projectExists)
		if projectExists == 0 {
			return nil, fmt.Errorf("%w: project not found or cross-tenant access denied", ErrNotFound)
		}
	}

	lineageJSON, _ := json.Marshal(lineageMeta)

	run := &SyncRun{
		ID:             uuid.New(),
		OrganizationID: orgID,
		ProjectID:      projectID,
		Direction:      DirectionImport,
		SourceFileName: fileName,
		SourceFileSize: fileSize,
		SourceMIMEType: mimeType,
		Format:         format,
		Status:         StatusPending,
		ErrorSummary:   "[]",
		ConflictReport: "[]",
		Lineage:        string(lineageJSON),
		TriggeredBy:    userID,
	}

	if err := s.db.WithContext(ctx).Create(run).Error; err != nil {
		return nil, err
	}

	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &userID,
		Action:         "primavera.sync.created",
		EntityType:     "primavera_sync_run",
		EntityID:       run.ID.String(),
		EntityLabel:    fileName,
		NewValues:      run,
	})
	return run, nil
}

// ---------------------------------------------------------------------------
// ProcessRun — parse + import in one transaction
// ---------------------------------------------------------------------------

// ProcessRun reads file bytes, parses activities, and upserts mappings +
// monitoring entities within a single DB transaction.
// It is idempotent: re-running with the same p6_activity_id + entity_type
// for the same project will UPDATE the existing mapping instead of INSERT.
func (s *Service) ProcessRun(
	ctx context.Context,
	orgID, userID, runID uuid.UUID,
	fileReader io.Reader,
) (*SyncRun, error) {
	run, err := s.GetRun(ctx, orgID, runID)
	if err != nil {
		return nil, err
	}
	if run.Status != StatusPending {
		return nil, fmt.Errorf("%w: run is %s, expected PENDING", ErrInvalidTransition, run.Status)
	}

	// mark RUNNING
	now := time.Now().UTC()
	run.Status = StatusRunning
	run.StartedAt = &now
	run.UpdatedAt = now
	if err := s.db.WithContext(ctx).Save(run).Error; err != nil {
		return nil, err
	}

	// buffer entire file (needed for potential retry without re-reading)
	fileBytes, err := io.ReadAll(fileReader)
	if err != nil {
		return s.failRun(ctx, run, []SyncErrorEntry{{Code: "E000", Message: "cannot read file: " + err.Error()}})
	}

	// parse
	parseResult, err := Parse(bytes.NewReader(fileBytes), run.Format)
	if err != nil {
		return s.failRun(ctx, run, []SyncErrorEntry{{Code: "E000", Message: "parse error: " + err.Error()}})
	}
	if len(parseResult.Activities) == 0 && len(parseResult.Errors) > 0 {
		return s.failRun(ctx, run, parseResult.Errors)
	}

	// determine target project — either from run.ProjectID or infer from file
	var targetProjectID uuid.UUID
	if run.ProjectID != nil {
		targetProjectID = *run.ProjectID
	} else {
		// No project specified — multi-project import is not yet supported.
		// Return a clear error rather than silently skipping.
		return s.failRun(ctx, run, []SyncErrorEntry{{
			Code:    "E010",
			Message: "project_id is required for IMPORT runs; multi-project batch not yet supported",
		}})
	}

	// verify project belongs to this tenant
	var projectExists int64
	s.db.WithContext(ctx).Table("projects").
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL", targetProjectID, orgID).
		Count(&projectExists)
	if projectExists == 0 {
		return s.failRun(ctx, run, []SyncErrorEntry{{
			Code:    "E011",
			Message: "project not found or cross-tenant overwrite denied",
		}})
	}

	// process activities inside a transaction
	var (
		importedCount int
		skippedCount  int
		failedCount   int
		conflictCount int
		syncErrors    = parseResult.Errors // carry forward parse errors
		conflicts     []ConflictEntry
	)

	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, act := range parseResult.Activities {
			mapping, conflicts_, err := s.upsertActivityMapping(ctx, tx, orgID, targetProjectID, runID, act)
			if err != nil {
				syncErrors = append(syncErrors, SyncErrorEntry{
					Code:       "E020",
					Message:    err.Error(),
					ActivityID: act.ActivityID,
				})
				failedCount++
				continue
			}
			conflicts = append(conflicts, conflicts_...)
			if len(conflicts_) > 0 {
				conflictCount++
			}
			switch mapping.Action {
			case ActionSkip:
				skippedCount++
			default:
				importedCount++
			}
		}
		return nil
	})
	if txErr != nil {
		return s.failRun(ctx, run, append(syncErrors, SyncErrorEntry{
			Code:    "E099",
			Message: "transaction failed: " + txErr.Error(),
		}))
	}

	// serialise errors and conflicts
	errJSON, _ := json.Marshal(syncErrors)
	conflictJSON, _ := json.Marshal(conflicts)

	finished := time.Now().UTC()
	run.Status = StatusDone
	run.TotalActivities = len(parseResult.Activities)
	run.ImportedActivities = importedCount
	run.SkippedActivities = skippedCount
	run.FailedActivities = failedCount
	run.ConflictCount = conflictCount
	run.ErrorSummary = string(errJSON)
	run.ConflictReport = string(conflictJSON)
	run.FinishedAt = &finished
	run.UpdatedAt = finished

	if err := s.db.WithContext(ctx).Save(run).Error; err != nil {
		return nil, err
	}

	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &userID,
		Action:         "primavera.sync.done",
		EntityType:     "primavera_sync_run",
		EntityID:       run.ID.String(),
		EntityLabel:    run.SourceFileName,
		NewValues:      run,
	})
	return run, nil
}

// ---------------------------------------------------------------------------
// Cancel
// ---------------------------------------------------------------------------

// CancelRun cancels a PENDING sync run.
func (s *Service) CancelRun(ctx context.Context, orgID, userID, runID uuid.UUID) (*SyncRun, error) {
	run, err := s.GetRun(ctx, orgID, runID)
	if err != nil {
		return nil, err
	}
	if run.Status != StatusPending {
		return nil, fmt.Errorf("%w: cannot cancel a run in status %s", ErrInvalidTransition, run.Status)
	}

	now := time.Now().UTC()
	run.Status = StatusCancelled
	run.FinishedAt = &now
	run.UpdatedAt = now

	if err := s.db.WithContext(ctx).Save(run).Error; err != nil {
		return nil, err
	}

	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &userID,
		Action:         "primavera.sync.cancelled",
		EntityType:     "primavera_sync_run",
		EntityID:       run.ID.String(),
		EntityLabel:    run.SourceFileName,
		NewValues:      run,
	})
	return run, nil
}

// ---------------------------------------------------------------------------
// Private helpers
// ---------------------------------------------------------------------------

// upsertActivityMapping inserts or updates a primavera_activity_mappings row.
// Returns the mapping and any detected field conflicts.
func (s *Service) upsertActivityMapping(
	ctx context.Context,
	tx *gorm.DB,
	orgID, projectID, runID uuid.UUID,
	act ParsedActivity,
) (*ActivityMapping, []ConflictEntry, error) {
	rawJSON := RawPayloadJSON(act.RawPayload)

	// Look up existing mapping for idempotency
	var existing ActivityMapping
	err := tx.Where(
		"organization_id = ? AND project_id = ? AND p6_activity_id = ? AND entity_type = ?",
		orgID, projectID, act.ActivityID, "task",
	).First(&existing).Error

	var conflicts []ConflictEntry

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// CREATE path — generate a stable mapping UUID for this activity.
		// The mapping table itself serves as the lineage record; we do not
		// write directly into the `tasks` table because that table uses a
		// different schema (title, progress_pct, start_date/due_date, etc.)
		// and a separate import flow.
		entityID := uuid.New()

		mapping := &ActivityMapping{
			ID:               uuid.New(),
			OrganizationID:   orgID,
			ProjectID:        projectID,
			SyncRunID:        runID,
			P6ActivityID:     act.ActivityID,
			P6WBSCode:        act.WBSCode,
			P6ActivityName:   act.ActivityName,
			EntityType:       "task",
			EntityID:         entityID,
			Action:           ActionCreate,
			BaselinePhysical: act.BaselinePhysical,
			ActualPhysical:   act.ActualPhysical,
			PlannedStart:     parseDate(act.PlannedStart),
			PlannedEnd:       parseDate(act.PlannedEnd),
			ActualStart:      parseDate(act.ActualStart),
			ActualEnd:        parseDate(act.ActualEnd),
			RawPayload:       rawJSON,
		}
		if err := tx.Create(mapping).Error; err != nil {
			return nil, nil, fmt.Errorf("insert mapping %s: %w", act.ActivityID, err)
		}
		return mapping, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("lookup mapping %s: %w", act.ActivityID, err)
	}

	// UPDATE path — detect conflicts on progress fields
	if existing.ActualPhysical != act.ActualPhysical {
		conflicts = append(conflicts, ConflictEntry{
			ActivityID: act.ActivityID,
			Field:      "actual_physical",
			Existing:   fmt.Sprintf("%g", existing.ActualPhysical),
			Incoming:   fmt.Sprintf("%g", act.ActualPhysical),
		})
	}

	action := ActionUpdate
	if len(conflicts) > 0 {
		action = ActionConflict
	}

	now := time.Now().UTC()
	updates := map[string]interface{}{
		"sync_run_id":       runID,
		"p6_wbs_code":       act.WBSCode,
		"p6_activity_name":  act.ActivityName,
		"action":            action,
		"baseline_physical": act.BaselinePhysical,
		"actual_physical":   act.ActualPhysical,
		"planned_start":     parseDate(act.PlannedStart),
		"planned_end":       parseDate(act.PlannedEnd),
		"actual_start":      parseDate(act.ActualStart),
		"actual_end":        parseDate(act.ActualEnd),
		"raw_payload":       rawJSON,
		"updated_at":        now,
	}
	if err := tx.Model(&existing).Updates(updates).Error; err != nil {
		return nil, conflicts, fmt.Errorf("update mapping %s: %w", act.ActivityID, err)
	}
	existing.Action = action
	return &existing, conflicts, nil
}

// failRun marks a sync run as FAILED and persists the error summary.
func (s *Service) failRun(ctx context.Context, run *SyncRun, errs []SyncErrorEntry) (*SyncRun, error) {
	errJSON, _ := json.Marshal(errs)
	now := time.Now().UTC()
	run.Status = StatusFailed
	run.ErrorSummary = string(errJSON)
	run.FinishedAt = &now
	run.UpdatedAt = now
	_ = s.db.WithContext(ctx).Save(run).Error
	return run, fmt.Errorf("primavera sync failed: %s", errJSON)
}

// normaliseFormat infers format from filename extension when format is empty.
func normaliseFormat(format, fileName string) string {
	if format != "" {
		return format
	}
	lower := toLower(fileName)
	switch {
	case hasSuffix(lower, ".xer"):
		return FormatXER
	case hasSuffix(lower, ".xml") || hasSuffix(lower, ".pmxml"):
		return FormatPMXML
	default:
		return FormatXER // default
	}
}

// parseDate converts an ISO date string "2006-01-02" to *time.Time.
func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		result[i] = c
	}
	return string(result)
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
