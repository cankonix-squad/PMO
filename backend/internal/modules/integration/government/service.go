package government

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/audit"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------------

var ErrNotFound = errors.New("government sync run not found")
var ErrForbidden = errors.New("forbidden")
var ErrInvalidConnector = errors.New("invalid connector key")
var ErrInvalidDatasetType = errors.New("invalid dataset type for connector")
var ErrInvalidMode = errors.New("invalid sync mode")
var ErrInvalidTransition = errors.New("invalid status transition")
var ErrDuplicateRun = errors.New("a run with this idempotency key already exists")

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// Service handles government connector business logic.
type Service struct {
	db          *gorm.DB
	auditWriter *audit.Writer
	log         *zap.Logger
	cfg         map[string]ConnectorConfig
}

// NewService creates a new government integration Service.
func NewService(db *gorm.DB, auditWriter *audit.Writer, log *zap.Logger) *Service {
	return &Service{
		db:          db,
		auditWriter: auditWriter,
		log:         log,
		cfg:         LoadConfig(),
	}
}

// ---------------------------------------------------------------------------
// Connector registry
// ---------------------------------------------------------------------------

// ListConnectors returns all registered connector definitions with current state.
func (s *Service) ListConnectors() []ConnectorDefinition {
	return ListConnectors(s.cfg)
}

// GetConnector returns a single connector definition by key.
func (s *Service) GetConnector(key string) (ConnectorDefinition, error) {
	c, ok := GetConnector(key, s.cfg)
	if !ok {
		return ConnectorDefinition{}, ErrInvalidConnector
	}
	return c, nil
}

// GetConfig returns non-secret connector configuration metadata.
func (s *Service) GetConfig() map[string]ConnectorConfig {
	// Return a copy to avoid mutation; strip any fields that could leak secrets.
	result := make(map[string]ConnectorConfig, len(s.cfg))
	for k, v := range s.cfg {
		result[k] = ConnectorConfig{
			ConnectorKey: v.ConnectorKey,
			Enabled:      v.Enabled,
			BaseURL:      v.BaseURL, // safe: URL only, no credentials
			State:        v.State,
			SandboxMode:  v.SandboxMode,
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// List / Get runs
// ---------------------------------------------------------------------------

// ListRuns returns paginated sync runs for an organisation.
func (s *Service) ListRuns(ctx context.Context, orgID uuid.UUID, f ListRunsFilter) ([]SyncRun, int64, error) {
	q := s.db.WithContext(ctx).Model(&SyncRun{}).Where("organization_id = ?", orgID)
	if f.ConnectorKey != "" {
		q = q.Where("connector_key = ?", f.ConnectorKey)
	}
	if f.DatasetType != "" {
		q = q.Where("dataset_type = ?", f.DatasetType)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
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
	if err := q.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&runs).Error; err != nil {
		return nil, 0, err
	}
	return runs, total, nil
}

// GetRun returns a single sync run scoped to the organisation.
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

// ---------------------------------------------------------------------------
// Create / Cancel runs
// ---------------------------------------------------------------------------

// CreateRun creates a new PENDING sync run after validating inputs.
func (s *Service) CreateRun(ctx context.Context, orgID, userID uuid.UUID, req CreateRunRequest) (*SyncRun, error) {
	// Validate connector key
	if !AllowedConnectorKeys[req.ConnectorKey] {
		return nil, ErrInvalidConnector
	}

	// Validate dataset type is supported by this connector
	allowedDS, _ := DatasetTypesForConnector(req.ConnectorKey)
	if !allowedDS[req.DatasetType] {
		return nil, ErrInvalidDatasetType
	}

	// Validate mode
	if !AllowedModes[req.Mode] {
		return nil, ErrInvalidMode
	}

	// Build idempotency key if not supplied
	idemKey := req.IdempotencyKey
	if idemKey == "" {
		raw := fmt.Sprintf("%s:%s:%s:%s", orgID, req.ConnectorKey, req.DatasetType, req.Mode)
		idemKey = fmt.Sprintf("%x", sha256.Sum256([]byte(raw)))
	}

	// Check for duplicate run with same idempotency key
	var existing SyncRun
	err := s.db.WithContext(ctx).
		Where("organization_id = ? AND idempotency_key = ? AND status IN (?)",
			orgID, idemKey, []string{StatusPending, StatusRunning}).
		First(&existing).Error
	if err == nil {
		return &existing, ErrDuplicateRun
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Compute source hash from request
	reqBytes, _ := json.Marshal(req)
	sourceHash := fmt.Sprintf("%x", sha256.Sum256(reqBytes))

	now := time.Now().UTC()
	run := SyncRun{
		ID:             uuid.New(),
		OrganizationID: orgID,
		ConnectorKey:   req.ConnectorKey,
		DatasetType:    req.DatasetType,
		Mode:           req.Mode,
		Status:         StatusPending,
		StartedBy:      userID,
		ErrorSummary:   []byte("[]"),
		SourceHash:     sourceHash,
		IdempotencyKey: idemKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.db.WithContext(ctx).Create(&run).Error; err != nil {
		return nil, fmt.Errorf("government: create run: %w", err)
	}

	actor := userID
	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actor,
		Action:         "government.sync.started",
		EntityType:     "government_sync_run",
		EntityID:       run.ID.String(),
		EntityLabel:    run.ConnectorKey + "/" + run.DatasetType,
		NewValues: map[string]interface{}{
			"connector_key": run.ConnectorKey,
			"dataset_type":  run.DatasetType,
			"mode":          run.Mode,
		},
	})

	return &run, nil
}

// ProcessRun transitions a PENDING run to RUNNING and executes ingestion.
func (s *Service) ProcessRun(ctx context.Context, orgID, runID uuid.UUID) (*SyncRun, error) {
	run, err := s.GetRun(ctx, orgID, runID)
	if err != nil {
		return nil, err
	}

	if run.Status != StatusPending {
		return nil, ErrInvalidTransition
	}

	// Transition to RUNNING
	now := time.Now().UTC()
	if err := s.db.WithContext(ctx).Model(run).Updates(map[string]interface{}{
		"status":     StatusRunning,
		"started_at": now,
		"updated_at": now,
	}).Error; err != nil {
		return nil, fmt.Errorf("government: start run: %w", err)
	}
	run.Status = StatusRunning
	run.StartedAt = &now

	// Execute ingestor
	result, ingestErr := Ingest(ctx, s.db, run)

	finishedAt := time.Now().UTC()
	finalStatus := StatusSucceeded
	if ingestErr != nil {
		finalStatus = StatusFailed
		s.log.Error("government ingest failed", zap.String("run_id", runID.String()), zap.Error(ingestErr))
	}

	errBytes, _ := json.Marshal(result.ErrorSummary)

	updates := map[string]interface{}{
		"status":           finalStatus,
		"finished_at":      finishedAt,
		"total_records":    result.TotalRecords,
		"accepted_records": result.AcceptedRecords,
		"rejected_records": result.RejectedRecords,
		"error_summary":    errBytes,
		"updated_at":       finishedAt,
	}
	if err2 := s.db.WithContext(ctx).Model(run).Updates(updates).Error; err2 != nil {
		return nil, fmt.Errorf("government: finish run: %w", err2)
	}
	run.Status = finalStatus
	run.FinishedAt = &finishedAt
	run.TotalRecords = result.TotalRecords
	run.AcceptedRecords = result.AcceptedRecords
	run.RejectedRecords = result.RejectedRecords
	run.ErrorSummary = errBytes

	auditAction := "government.sync.succeeded"
	if finalStatus == StatusFailed {
		auditAction = "government.sync.failed"
	}
	startedBy := run.StartedBy
	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &startedBy,
		Action:         auditAction,
		EntityType:     "government_sync_run",
		EntityID:       run.ID.String(),
		EntityLabel:    run.ConnectorKey + "/" + run.DatasetType,
		NewValues: map[string]interface{}{
			"connector_key":    run.ConnectorKey,
			"dataset_type":     run.DatasetType,
			"mode":             run.Mode,
			"total_records":    result.TotalRecords,
			"accepted_records": result.AcceptedRecords,
			"rejected_records": result.RejectedRecords,
		},
	})

	return run, nil
}

// CancelRun cancels a PENDING run.
func (s *Service) CancelRun(ctx context.Context, orgID, runID, userID uuid.UUID) (*SyncRun, error) {
	run, err := s.GetRun(ctx, orgID, runID)
	if err != nil {
		return nil, err
	}

	if run.Status != StatusPending {
		return nil, ErrInvalidTransition
	}

	now := time.Now().UTC()
	if err := s.db.WithContext(ctx).Model(run).Updates(map[string]interface{}{
		"status":     StatusCancelled,
		"updated_at": now,
	}).Error; err != nil {
		return nil, fmt.Errorf("government: cancel run: %w", err)
	}
	run.Status = StatusCancelled

	cancelActor := userID
	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &cancelActor,
		Action:         "government.sync.cancelled",
		EntityType:     "government_sync_run",
		EntityID:       run.ID.String(),
		EntityLabel:    run.ConnectorKey + "/" + run.DatasetType,
		NewValues: map[string]interface{}{
			"connector_key": run.ConnectorKey,
			"dataset_type":  run.DatasetType,
		},
	})

	return run, nil
}

// ---------------------------------------------------------------------------
// Records / Mappings
// ---------------------------------------------------------------------------

// ListRecords returns paginated sync records for a given run.
func (s *Service) ListRecords(ctx context.Context, orgID, runID uuid.UUID, f ListRecordsFilter) ([]SyncRecord, int64, error) {
	// Verify run belongs to org
	if _, err := s.GetRun(ctx, orgID, runID); err != nil {
		return nil, 0, err
	}

	q := s.db.WithContext(ctx).Model(&SyncRecord{}).Where("sync_run_id = ?", runID)
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Action != "" {
		q = q.Where("action = ?", f.Action)
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

	var records []SyncRecord
	if err := q.Order("created_at ASC").Limit(pageSize).Offset(offset).Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

// ListMappings returns paginated external mappings for an organisation.
func (s *Service) ListMappings(ctx context.Context, orgID uuid.UUID, f ListMappingsFilter) ([]ExternalMapping, int64, error) {
	q := s.db.WithContext(ctx).Model(&ExternalMapping{}).Where("organization_id = ?", orgID)
	if f.ConnectorKey != "" {
		q = q.Where("connector_key = ?", f.ConnectorKey)
	}
	if f.DatasetType != "" {
		q = q.Where("dataset_type = ?", f.DatasetType)
	}
	if f.InternalEntityType != "" {
		q = q.Where("internal_entity_type = ?", f.InternalEntityType)
	}
	if f.MatchStatus != "" {
		q = q.Where("match_status = ?", f.MatchStatus)
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

	var mappings []ExternalMapping
	if err := q.Order("updated_at DESC").Limit(pageSize).Offset(offset).Find(&mappings).Error; err != nil {
		return nil, 0, err
	}
	return mappings, total, nil
}
