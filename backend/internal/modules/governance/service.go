package governance

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/audit"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------------

var (
	ErrSubmissionNotFound   = errors.New("data submission not found")
	ErrInvalidTransition    = errors.New("invalid submission status transition")
	ErrInvalidDatasetType   = errors.New("invalid dataset_type")
	ErrInvalidSourceType    = errors.New("invalid source_type")
	ErrInvalidEntityType    = errors.New("invalid entity_type")
	ErrEmptyItems           = errors.New("submission must contain at least one item")
	ErrRejectionReason      = errors.New("rejection_reason is required")
	ErrInvalidItems         = errors.New("submission has invalid items; cannot approve")
	ErrEntityIDRequired     = errors.New("entity_id is required for this item action")
	ErrEntityMismatch       = errors.New("source entity does not belong to this organization")
	ErrEntityDeleted        = errors.New("source entity is soft-deleted")
	ErrEntityNotFound       = errors.New("source entity not found")
	ErrPendingMatchNotReady = errors.New("government mapping is PENDING_MATCH and cannot be submitted as official data")
	ErrMappingRejected      = errors.New("government mapping is REJECTED and cannot be submitted as official data")
	ErrUnknownMatchStatus   = errors.New("government mapping has unknown match_status and cannot be submitted as official data")
	ErrLockedPeriod         = errors.New("period is locked; changes are not allowed")
	ErrLockPeriodNotFound   = errors.New("lock period not found")
	ErrLockPeriodConflict   = errors.New("a lock period already exists for this dataset/period")
)

// AllowedEntityTypes is the set of entity types accepted as submission items.
// Unknown/free-text entity_type values are INVALID (never silently VALID).
var AllowedEntityTypes = map[string]bool{
	"project":            true,
	"projects":           true,
	"vendor":             true,
	"vendors":            true,
	"government_mapping": true,
}

// normalizeEntityType maps plural/alias forms to the canonical singular key.
func normalizeEntityType(et string) (string, bool) {
	switch et {
	case "project", "projects":
		return "project", true
	case "vendor", "vendors":
		return "vendor", true
	case "government_mapping":
		return "government_mapping", true
	}
	return "", false
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// Service implements the data governance official validation workflow.
type Service struct {
	db      *gorm.DB
	auditor *audit.Writer
	log     *zap.Logger
}

// NewService creates a governance Service.
func NewService(db *gorm.DB, auditor *audit.Writer, log *zap.Logger) *Service {
	return &Service{db: db, auditor: auditor, log: log}
}

// ---------------------------------------------------------------------------
// List / Get submissions
// ---------------------------------------------------------------------------

// ListSubmissions returns paginated submissions for an organisation.
func (s *Service) ListSubmissions(orgID uuid.UUID, f ListSubmissionsFilter) ([]Submission, int64, error) {
	q := s.db.Model(&Submission{}).Where("organization_id = ? AND deleted_at IS NULL", orgID)
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.DatasetType != "" {
		q = q.Where("dataset_type = ?", f.DatasetType)
	}
	if f.SourceType != "" {
		q = q.Where("source_type = ?", f.SourceType)
	}
	if f.PeriodYear > 0 {
		q = q.Where("period_year = ?", f.PeriodYear)
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

	var list []Submission
	if err := q.Order("updated_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetSubmission returns a submission with its items, tenant-scoped.
func (s *Service) GetSubmission(orgID, id uuid.UUID) (*SubmissionDetail, error) {
	var sub Submission
	if err := s.db.Where("organization_id = ? AND id = ? AND deleted_at IS NULL", orgID, id).First(&sub).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSubmissionNotFound
		}
		return nil, err
	}
	var items []SubmissionItem
	if err := s.db.Where("submission_id = ?", sub.ID).Order("created_at ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return &SubmissionDetail{Submission: &sub, Items: items}, nil
}

// ---------------------------------------------------------------------------
// Create submission
// ---------------------------------------------------------------------------

// CreateSubmission creates a DRAFT submission with items.
func (s *Service) CreateSubmission(orgID, actorID uuid.UUID, req CreateSubmissionRequest) (*Submission, error) {
	if !AllowedDatasetTypes[req.DatasetType] {
		return nil, ErrInvalidDatasetType
	}
	if !AllowedSourceTypes[req.SourceType] {
		return nil, ErrInvalidSourceType
	}
	if req.PeriodYear <= 0 {
		return nil, errors.New("period_year must be a positive year")
	}
	if len(req.Items) == 0 {
		return nil, ErrEmptyItems
	}

	// Reject if the period is already locked for this dataset
	if s.isPeriodLocked(orgID, req.DatasetType, req.PeriodYear, monthPtrOrZero(req.PeriodMonth)) {
		return nil, ErrLockedPeriod
	}

	// Validate source_entity_id if provided (tenant-scoped, matches source_type)
	sourceEntityType := strings.ToLower(strings.TrimSpace(req.SourceEntityType))
	if req.SourceEntityID != "" {
		if sourceEntityType == "" {
			return nil, errors.New("source_entity_type is required when source_entity_id is provided")
		}
		if _, ok := normalizeEntityType(sourceEntityType); !ok {
			return nil, ErrInvalidEntityType
		}
		srcID, err := uuid.Parse(req.SourceEntityID)
		if err != nil {
			return nil, errors.New("invalid source_entity_id")
		}
		if err := s.validateSourceEntity(orgID, sourceEntityType, srcID); err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC()
	sub := Submission{
		ID:               uuid.New(),
		OrganizationID:   orgID,
		Source:           req.SourceType,
		SourceType:       req.SourceType,
		DatasetType:      req.DatasetType,
		SourceEntityType: req.SourceEntityType,
		PeriodYear:       req.PeriodYear,
		Status:           StatusDraft,
		CompletenessPct:  100,
		CreatedBy:        &actorID,
		// SubmittedBy/SubmittedAt are set ONLY at Submit time, not at creation.
		Lineage:   json.RawMessage(`{}`),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if req.SourceEntityID != "" {
		srcID, err := uuid.Parse(req.SourceEntityID)
		if err != nil {
			return nil, errors.New("invalid source_entity_id")
		}
		sub.SourceEntityID = &srcID
	}
	sub.PeriodMonth = req.PeriodMonth
	if req.SourceReference != "" {
		sub.SourceReference = req.SourceReference
	}

	// Build items
	items := make([]SubmissionItem, 0, len(req.Items))
	for _, it := range req.Items {
		if !validItemAction(it.Action) {
			return nil, fmt.Errorf("invalid item action: %s", it.Action)
		}
		// UPDATE/DELETE/UPSERT mutate an existing entity, so entity_id is
		// mandatory. CREATE may omit it (brand-new entity); VALIDATE_ONLY may
		// omit it when the payload alone is enough to validate.
		if actionRequiresEntityID(it.Action) && strings.TrimSpace(it.EntityID) == "" {
			return nil, fmt.Errorf("%w: action %s", ErrEntityIDRequired, it.Action)
		}
		// Unknown entity_type is INVALID — never allow free-text/typo to pass.
		entityType := strings.ToLower(strings.TrimSpace(it.EntityType))
		if _, ok := normalizeEntityType(entityType); !ok {
			return nil, fmt.Errorf("%w: %q", ErrInvalidEntityType, it.EntityType)
		}
		item := SubmissionItem{
			ID:               uuid.New(),
			SubmissionID:     sub.ID,
			EntityType:       it.EntityType,
			Action:           it.Action,
			ValidationStatus: ItemValidationPending,
			ValidationErrors: json.RawMessage(`[]`),
		}
		if it.EntityID != "" {
			eid, err := uuid.Parse(it.EntityID)
			if err != nil {
				return nil, errors.New("invalid item entity_id")
			}
			item.EntityID = &eid
		}
		if it.PayloadAfter != nil {
			b, err := json.Marshal(it.PayloadAfter)
			if err != nil {
				return nil, errors.New("invalid item payload_after")
			}
			item.PayloadAfter = b
		}
		if it.PayloadBefore != nil {
			b, err := json.Marshal(it.PayloadBefore)
			if err != nil {
				return nil, errors.New("invalid item payload_before")
			}
			item.PayloadBefore = b
		}
		if len(item.PayloadAfter) == 0 {
			item.PayloadAfter = json.RawMessage(`{}`)
		}
		items = append(items, item)
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&sub).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].SubmissionID = sub.ID
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("governance: create submission: %w", err)
	}

	s.auditor.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		Action:         "governance.submission.created",
		EntityType:     "data_submission",
		EntityID:       sub.ID.String(),
		EntityLabel:    fmt.Sprintf("%s/%d/%s", req.DatasetType, req.PeriodYear, req.SourceType),
		NewValues: map[string]interface{}{
			"dataset_type": req.DatasetType,
			"source_type":  req.SourceType,
			"period_year":  req.PeriodYear,
			"status":       StatusDraft,
			"item_count":   len(items),
		},
	})

	return &sub, nil
}

// ---------------------------------------------------------------------------
// State transitions
// ---------------------------------------------------------------------------

// transitionAllowed returns true if the state transition is permitted.
// FSM: DRAFT→SUBMITTED, SUBMITTED→IN_REVIEW, IN_REVIEW→APPROVED|REJECTED,
//
//	APPROVED→LOCKED, DRAFT/SUBMITTED→CANCELLED, REJECTED→DRAFT.
func transitionAllowed(from, to string) bool {
	switch to {
	case StatusSubmitted:
		return from == StatusDraft
	case StatusInReview:
		return from == StatusSubmitted
	case StatusApproved:
		return from == StatusInReview
	case StatusRejected:
		return from == StatusInReview
	case StatusLocked:
		return from == StatusApproved
	case StatusCancelled:
		return from == StatusDraft || from == StatusSubmitted
	case StatusDraft:
		return from == StatusRejected
	}
	return false
}

// Submit moves a DRAFT submission to SUBMITTED.
func (s *Service) Submit(orgID, actorID, id uuid.UUID) (*Submission, error) {
	sub, err := s.GetSubmission(orgID, id)
	if err != nil {
		return nil, err
	}
	if !transitionAllowed(sub.Status, StatusSubmitted) {
		return nil, ErrInvalidTransition
	}

	// Cannot submit into a locked period
	if s.isPeriodLocked(orgID, sub.DatasetType, sub.PeriodYear, monthPtrOrZero(sub.PeriodMonth)) {
		return nil, ErrLockedPeriod
	}

	// Cannot submit empty items
	if len(sub.Items) == 0 {
		return nil, ErrEmptyItems
	}

	// Re-validate source entity (tenant-scoped) before submit — if it became
	// invalid (deleted / cross-tenant / mapping no longer MATCHED) refuse submit.
	if sub.SourceEntityID != nil {
		srcType := strings.ToLower(strings.TrimSpace(sub.SourceEntityType))
		if srcType != "" {
			if err := s.validateSourceEntity(orgID, srcType, *sub.SourceEntityID); err != nil {
				return nil, err
			}
		}
	}

	now := time.Now().UTC()
	updates := map[string]interface{}{
		"status":       StatusSubmitted,
		"submitted_by": actorID,
		"submitted_at": now,
		"updated_at":   now,
	}
	if err := s.db.Model(&Submission{}).Where("id = ? AND organization_id = ?", id, orgID).Updates(updates).Error; err != nil {
		return nil, err
	}
	sub.Status = StatusSubmitted

	s.auditSubmission(actorID, orgID, sub.Submission, "governance.submission.submitted", map[string]interface{}{"status": StatusSubmitted})
	return sub.Submission, nil
}

// StartReview moves a SUBMITTED submission to IN_REVIEW and runs per-item validation.
func (s *Service) StartReview(orgID, actorID, id uuid.UUID, req ReviewRequest) (*SubmissionDetail, error) {
	detail, err := s.GetSubmission(orgID, id)
	if err != nil {
		return nil, err
	}
	sub := detail.Submission
	if !transitionAllowed(sub.Status, StatusInReview) {
		return nil, ErrInvalidTransition
	}

	// Validate source entity (tenant-scoped) before review — a submission whose
	// source entity is gone/cross-tenant/mapping-not-MATCHED cannot be reviewed.
	if sub.SourceEntityID != nil {
		srcType := strings.ToLower(strings.TrimSpace(sub.SourceEntityType))
		if srcType != "" {
			if err := s.validateSourceEntity(orgID, srcType, *sub.SourceEntityID); err != nil {
				return nil, err
			}
		}
	}

	// Validate each item against source entity ownership + soft-delete + match status.
	validationErr := s.validateItems(orgID, detail.Items)

	now := time.Now().UTC()
	reviewNotes := req.ReviewNotes

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Submission{}).Where("id = ? AND organization_id = ?", id, orgID).Updates(map[string]interface{}{
			"status":       StatusInReview,
			"reviewed_by":  actorID,
			"reviewed_at":  now,
			"review_notes": reviewNotes,
			"updated_at":   now,
		}).Error; err != nil {
			return err
		}
		// Persist per-item validation results
		for i := range detail.Items {
			item := &detail.Items[i]
			if err := tx.Model(&SubmissionItem{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
				"validation_status": item.ValidationStatus,
				"validation_errors": item.ValidationErrors,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sub.Status = StatusInReview
	s.auditor.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		Action:         "governance.submission.review_started",
		EntityType:     "data_submission",
		EntityID:       id.String(),
		EntityLabel:    sub.DatasetType + "/" + sub.SourceType,
		NewValues:      map[string]interface{}{"status": StatusInReview, "validation_error": validationErr},
	})
	return detail, nil
}

// Approve moves an IN_REVIEW submission to APPROVED.
// Rule: all items must be VALID — and the approval MUST re-validate every item
// and the source entity right now. A submission whose entities changed after
// review (deleted / cross-tenant / mapping lost MATCHED) must fail approve with
// 409/400 and the latest INVALID statuses persisted.
func (s *Service) Approve(orgID, actorID, id uuid.UUID) (*Submission, error) {
	detail, err := s.GetSubmission(orgID, id)
	if err != nil {
		return nil, err
	}
	sub := detail.Submission
	if !transitionAllowed(sub.Status, StatusApproved) {
		return nil, ErrInvalidTransition
	}

	// 1. Re-validate source entity at approve time (tenant-scoped).
	if sub.SourceEntityID != nil {
		srcType := strings.ToLower(strings.TrimSpace(sub.SourceEntityType))
		if srcType != "" {
			if err := s.validateSourceEntity(orgID, srcType, *sub.SourceEntityID); err != nil {
				return nil, fmt.Errorf("%w: %v", ErrInvalidItems, err)
			}
		}
	}

	// 2. Re-validate every item NOW — do not trust stale ValidationStatus.
	validationErr := s.validateItems(orgID, detail.Items)

	// 3. Persist the latest per-item validation results regardless of outcome.
	persistErr := s.db.Transaction(func(tx *gorm.DB) error {
		for i := range detail.Items {
			item := &detail.Items[i]
			if err := tx.Model(&SubmissionItem{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
				"validation_status": item.ValidationStatus,
				"validation_errors": item.ValidationErrors,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if persistErr != nil {
		return nil, persistErr
	}

	// 4. Any INVALID item now blocks approval.
	if validationErr != "" {
		// Items were invalidated — record the failure in the audit trail.
		s.auditor.Record(audit.WriteRequest{
			OrganizationID: orgID,
			ActorID:        &actorID,
			Action:         "governance.submission.approve_rejected",
			EntityType:     "data_submission",
			EntityID:       id.String(),
			EntityLabel:    sub.DatasetType + "/" + sub.SourceType,
			NewValues:      map[string]interface{}{"status": StatusInReview, "reason": validationErr},
		})
		return nil, ErrInvalidItems
	}

	now := time.Now().UTC()
	if err := s.db.Model(&Submission{}).Where("id = ? AND organization_id = ?", id, orgID).Updates(map[string]interface{}{
		"status":      StatusApproved,
		"approved_by": actorID,
		"approved_at": now,
		"updated_at":  now,
	}).Error; err != nil {
		return nil, err
	}
	sub.Status = StatusApproved

	s.auditor.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		Action:         "governance.submission.approved",
		EntityType:     "data_submission",
		EntityID:       id.String(),
		EntityLabel:    sub.DatasetType + "/" + sub.SourceType,
		NewValues:      map[string]interface{}{"status": StatusApproved, "approved_by": actorID.String()},
	})
	return sub, nil
}

// Reject moves an IN_REVIEW submission to REJECTED. Rejection requires a reason.
func (s *Service) Reject(orgID, actorID, id uuid.UUID, req RejectRequest) (*Submission, error) {
	if strings.TrimSpace(req.RejectionReason) == "" {
		return nil, ErrRejectionReason
	}
	detail, err := s.GetSubmission(orgID, id)
	if err != nil {
		return nil, err
	}
	sub := detail.Submission
	if !transitionAllowed(sub.Status, StatusRejected) {
		return nil, ErrInvalidTransition
	}

	now := time.Now().UTC()
	if err := s.db.Model(&Submission{}).Where("id = ? AND organization_id = ?", id, orgID).Updates(map[string]interface{}{
		"status":           StatusRejected,
		"rejection_reason": req.RejectionReason,
		"updated_at":       now,
	}).Error; err != nil {
		return nil, err
	}
	sub.Status = StatusRejected

	s.auditor.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		Action:         "governance.submission.rejected",
		EntityType:     "data_submission",
		EntityID:       id.String(),
		EntityLabel:    sub.DatasetType + "/" + sub.SourceType,
		NewValues:      map[string]interface{}{"status": StatusRejected, "rejection_reason": req.RejectionReason},
	})
	return sub, nil
}

// Lock moves an APPROVED submission to LOCKED and ensures the period lock row
// exists (created or locked) for (org, dataset, year, month). After locking,
// creating/submitting a NEW submission for the same period returns 409.
func (s *Service) Lock(orgID, actorID, id uuid.UUID, req LockRequest) (*Submission, error) {
	detail, err := s.GetSubmission(orgID, id)
	if err != nil {
		return nil, err
	}
	sub := detail.Submission
	if !transitionAllowed(sub.Status, StatusLocked) {
		return nil, ErrInvalidTransition
	}

	now := time.Now().UTC()
	// Ensure period lock row exists for this submission's period. The unique
	// expression index (migration 000034) prevents duplicates incl. full-year.
	periodMonth := monthPtrOrZero(sub.PeriodMonth)
	if periodMonth > 0 {
		// monthly period — lock (org, dataset, year, month)
		if err := s.ensureLockPeriod(orgID, actorID, sub.DatasetType, sub.PeriodYear, &periodMonth, req.LockReason, now); err != nil {
			return nil, err
		}
	} else {
		// full-year period (period_month NULL) — lock (org, dataset, year, 0-key)
		if err := s.ensureLockPeriod(orgID, actorID, sub.DatasetType, sub.PeriodYear, nil, req.LockReason, now); err != nil {
			return nil, err
		}
	}

	if err := s.db.Model(&Submission{}).Where("id = ? AND organization_id = ?", id, orgID).Updates(map[string]interface{}{
		"status":     StatusLocked,
		"locked_by":  actorID,
		"locked_at":  now,
		"updated_at": now,
	}).Error; err != nil {
		return nil, err
	}
	sub.Status = StatusLocked

	s.auditor.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		Action:         "governance.submission.locked",
		EntityType:     "data_submission",
		EntityID:       id.String(),
		EntityLabel:    sub.DatasetType + "/" + sub.SourceType,
		NewValues:      map[string]interface{}{"status": StatusLocked, "lock_reason": req.LockReason},
	})
	return sub, nil
}

// ensureLockPeriod creates (or locks if OPEN) the lock period row for
// (org, dataset, year, month). month == nil means full-year. Idempotent —
// a duplicate row is never created.
func (s *Service) ensureLockPeriod(orgID, actorID uuid.UUID, datasetType string, year int, month *int, lockReason string, now time.Time) error {
	var existing LockPeriod
	err := s.db.Where("organization_id = ? AND dataset_type = ? AND period_year = ? AND (period_month IS NOT DISTINCT FROM ?) AND deleted_at IS NULL",
		orgID, datasetType, year, month).First(&existing).Error
	switch {
	case err == nil:
		// Row exists. If OPEN → lock it. If already LOCKED → fine (idempotent).
		if existing.Status != LockLocked {
			existing.Status = LockLocked
			existing.LockedBy = &actorID
			existing.LockedAt = &now
			if lockReason != "" {
				existing.LockReason = lockReason
			}
			existing.UpdatedAt = now
			if err := s.db.Save(&existing).Error; err != nil {
				return err
			}
			s.auditor.Record(audit.WriteRequest{
				OrganizationID: orgID,
				ActorID:        &actorID,
				Action:         "governance.lock_period.locked",
				EntityType:     "data_lock_period",
				EntityID:       existing.ID.String(),
				EntityLabel:    fmt.Sprintf("%s/%d/%v", datasetType, year, month),
				NewValues:      map[string]interface{}{"status": LockLocked, "lock_reason": lockReason},
			})
		}
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Create a new LOCKED lock period row.
		period := LockPeriod{
			ID:             uuid.New(),
			OrganizationID: orgID,
			DatasetType:    datasetType,
			PeriodYear:     year,
			PeriodMonth:    month,
			Status:         LockLocked,
			LockedBy:       &actorID,
			LockedAt:       &now,
			LockReason:     lockReason,
			CreatedBy:      actorID,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := s.db.Create(&period).Error; err != nil {
			// Unique violation on the expression index → another concurrent
			// lock won; treat as success (idempotent).
			if isUniqueViolation(err) {
				return nil
			}
			return err
		}
		s.auditor.Record(audit.WriteRequest{
			OrganizationID: orgID,
			ActorID:        &actorID,
			Action:         "governance.lock_period.locked",
			EntityType:     "data_lock_period",
			EntityID:       period.ID.String(),
			EntityLabel:    fmt.Sprintf("%s/%d/%v", datasetType, year, month),
			NewValues:      map[string]interface{}{"status": LockLocked, "lock_reason": lockReason},
		})
		return nil
	default:
		return err
	}
}

// isUniqueViolation reports whether err is a PostgreSQL unique-violation
// (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// Cancel moves a DRAFT or SUBMITTED submission to CANCELLED.
func (s *Service) Cancel(orgID, actorID, id uuid.UUID, req CancelRequest) (*Submission, error) {
	detail, err := s.GetSubmission(orgID, id)
	if err != nil {
		return nil, err
	}
	sub := detail.Submission
	if !transitionAllowed(sub.Status, StatusCancelled) {
		return nil, ErrInvalidTransition
	}

	now := time.Now().UTC()
	if err := s.db.Model(&Submission{}).Where("id = ? AND organization_id = ?", id, orgID).Updates(map[string]interface{}{
		"status":     StatusCancelled,
		"updated_at": now,
	}).Error; err != nil {
		return nil, err
	}
	sub.Status = StatusCancelled

	s.auditor.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		Action:         "governance.submission.cancelled",
		EntityType:     "data_submission",
		EntityID:       id.String(),
		EntityLabel:    sub.DatasetType + "/" + sub.SourceType,
		NewValues:      map[string]interface{}{"status": StatusCancelled, "cancel_reason": req.CancelReason},
	})
	return sub, nil
}

// ---------------------------------------------------------------------------
// Lock periods
// ---------------------------------------------------------------------------

// ListLockPeriods returns paginated lock periods for an organisation.
func (s *Service) ListLockPeriods(orgID uuid.UUID, f ListLockPeriodsFilter) ([]LockPeriod, int64, error) {
	q := s.db.Model(&LockPeriod{}).Where("organization_id = ? AND deleted_at IS NULL", orgID)
	if f.DatasetType != "" {
		q = q.Where("dataset_type = ?", f.DatasetType)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.PeriodYear > 0 {
		q = q.Where("period_year = ?", f.PeriodYear)
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
	var list []LockPeriod
	if err := q.Order("period_year DESC, period_month DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// CreateLockPeriod creates a lock period. If req.LockNow is true it becomes LOCKED.
func (s *Service) CreateLockPeriod(orgID, actorID uuid.UUID, req CreateLockPeriodRequest) (*LockPeriod, error) {
	if !AllowedDatasetTypes[req.DatasetType] {
		return nil, ErrInvalidDatasetType
	}
	if req.PeriodYear <= 0 {
		return nil, errors.New("period_year must be a positive year")
	}

	// Unique per (org, dataset, year, month-key) — month-key via COALESCE to
	// also cover full-year locks (period_month NULL → key 0). This mirrors the
	// expression unique index from migration 000034.
	var existing LockPeriod
	err := s.db.Where("organization_id = ? AND dataset_type = ? AND period_year = ? AND COALESCE(period_month, 0) = ? AND deleted_at IS NULL",
		orgID, req.DatasetType, req.PeriodYear, monthPtrOrZero(req.PeriodMonth)).First(&existing).Error
	if err == nil {
		return nil, ErrLockPeriodConflict
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	now := time.Now().UTC()
	period := LockPeriod{
		ID:             uuid.New(),
		OrganizationID: orgID,
		DatasetType:    req.DatasetType,
		PeriodYear:     req.PeriodYear,
		PeriodMonth:    req.PeriodMonth,
		Status:         LockOpen,
		CreatedBy:      actorID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if req.LockReason != "" {
		period.LockReason = req.LockReason
	}
	if req.LockNow {
		period.Status = LockLocked
		period.LockedBy = &actorID
		period.LockedAt = &now
	}
	if err := s.db.Create(&period).Error; err != nil {
		return nil, err
	}

	action := "governance.lock_period.created"
	if req.LockNow {
		action = "governance.lock_period.locked"
	}
	s.auditor.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		Action:         action,
		EntityType:     "data_lock_period",
		EntityID:       period.ID.String(),
		EntityLabel:    fmt.Sprintf("%s/%d/%v", req.DatasetType, req.PeriodYear, req.PeriodMonth),
		NewValues:      map[string]interface{}{"dataset_type": req.DatasetType, "period_year": req.PeriodYear, "status": period.Status, "lock_reason": req.LockReason},
	})
	return &period, nil
}

// LockPeriodById locks an OPEN lock period.
func (s *Service) LockPeriodByID(orgID, actorID, id uuid.UUID, req LockRequest) (*LockPeriod, error) {
	var period LockPeriod
	if err := s.db.Where("organization_id = ? AND id = ? AND deleted_at IS NULL", orgID, id).First(&period).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLockPeriodNotFound
		}
		return nil, err
	}
	if period.Status == LockLocked {
		return nil, ErrInvalidTransition
	}
	now := time.Now().UTC()
	period.Status = LockLocked
	period.LockedBy = &actorID
	period.LockedAt = &now
	if req.LockReason != "" {
		period.LockReason = req.LockReason
	}
	period.UpdatedAt = now
	if err := s.db.Save(&period).Error; err != nil {
		return nil, err
	}
	s.auditor.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		Action:         "governance.lock_period.locked",
		EntityType:     "data_lock_period",
		EntityID:       period.ID.String(),
		EntityLabel:    fmt.Sprintf("%s/%d/%v", period.DatasetType, period.PeriodYear, period.PeriodMonth),
		NewValues:      map[string]interface{}{"status": LockLocked, "lock_reason": req.LockReason},
	})
	return &period, nil
}

// ---------------------------------------------------------------------------
// Validation helpers
// ---------------------------------------------------------------------------

// validateItems validates each submission item:
//   - entity ownership (cross-tenant rejected)
//   - soft-deleted entities rejected
//   - government PENDING_MATCH mappings cannot be official
func (s *Service) validateItems(orgID uuid.UUID, items []SubmissionItem) string {
	firstErr := ""
	for i := range items {
		item := &items[i]
		// Default: VALID
		item.ValidationStatus = ItemValidationValid
		item.ValidationErrors = json.RawMessage(`[]`)

		errs := s.validateItem(orgID, *item)
		if len(errs) > 0 {
			item.ValidationStatus = ItemValidationInvalid
			b, _ := json.Marshal(errs)
			item.ValidationErrors = b
			if firstErr == "" {
				firstErr = errs[0]
			}
		}
	}
	return firstErr
}

func (s *Service) validateItem(orgID uuid.UUID, item SubmissionItem) []string {
	// Unknown entity_type is INVALID — never silently VALID.
	et, ok := normalizeEntityType(strings.ToLower(strings.TrimSpace(item.EntityType)))
	if !ok {
		return []string{fmt.Sprintf("%v: %q", ErrInvalidEntityType, item.EntityType)}
	}

	if item.EntityID == nil {
		// CREATE may target a brand-new entity and VALIDATE_ONLY may validate a
		// payload without referencing an existing entity. UPDATE/DELETE/UPSERT
		// cannot be applied without an entity id → INVALID.
		if actionRequiresEntityID(item.Action) {
			return []string{fmt.Sprintf("%v: action %s", ErrEntityIDRequired, item.Action)}
		}
		return nil
	}
	var errs []string
	switch et {
	case "project":
		var count int64
		if err := s.db.Raw(`SELECT count(*) FROM projects WHERE id = ? AND organization_id = ? AND deleted_at IS NULL`, *item.EntityID, orgID).Scan(&count).Error; err != nil {
			return []string{"internal error validating entity"}
		}
		if count == 0 {
			var deleted int64
			s.db.Raw(`SELECT count(*) FROM projects WHERE id = ? AND organization_id = ? AND deleted_at IS NOT NULL`, *item.EntityID, orgID).Scan(&deleted)
			if deleted > 0 {
				errs = append(errs, ErrEntityDeleted.Error())
			} else {
				errs = append(errs, ErrEntityNotFound.Error())
			}
		}
	case "vendor":
		var count int64
		if err := s.db.Raw(`SELECT count(*) FROM vendors WHERE id = ? AND organization_id = ? AND deleted_at IS NULL`, *item.EntityID, orgID).Scan(&count).Error; err != nil {
			return []string{"internal error validating entity"}
		}
		if count == 0 {
			var deleted int64
			s.db.Raw(`SELECT count(*) FROM vendors WHERE id = ? AND organization_id = ? AND deleted_at IS NOT NULL`, *item.EntityID, orgID).Scan(&deleted)
			if deleted > 0 {
				errs = append(errs, ErrEntityDeleted.Error())
			} else {
				errs = append(errs, ErrEntityNotFound.Error())
			}
		}
	case "government_mapping":
		// Government external mappings must be MATCHED before they can be official.
		// Mapping not found → INVALID. Cross-tenant → INVALID.
		// Only match_status = MATCHED is VALID. PENDING_MATCH/REJECTED/empty/
		// unknown → INVALID. Never trust Raw().Scan() without checking the row.
		var m struct {
			MatchStatus string
		}
		err := s.db.Raw(`SELECT match_status FROM government_external_mappings WHERE id = ? AND organization_id = ?`, *item.EntityID, orgID).Row().Scan(&m.MatchStatus)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
				errs = append(errs, ErrEntityNotFound.Error())
			} else {
				errs = append(errs, "internal error validating government mapping")
			}
			return errs
		}
		switch m.MatchStatus {
		case "MATCHED":
			// valid
		case "PENDING_MATCH":
			errs = append(errs, ErrPendingMatchNotReady.Error())
		case "REJECTED":
			errs = append(errs, ErrMappingRejected.Error())
		default:
			errs = append(errs, ErrUnknownMatchStatus.Error())
		}
	}
	return errs
}

// validateSourceEntity validates that source_entity_id exists, belongs to the
// organisation, is not soft-deleted, and (for government_mapping) is MATCHED.
func (s *Service) validateSourceEntity(orgID uuid.UUID, entityType string, id uuid.UUID) error {
	et, ok := normalizeEntityType(strings.ToLower(strings.TrimSpace(entityType)))
	if !ok {
		return ErrInvalidEntityType
	}
	switch et {
	case "project":
		var count int64
		if err := s.db.Raw(`SELECT count(*) FROM projects WHERE id = ? AND organization_id = ? AND deleted_at IS NULL`, id, orgID).Scan(&count).Error; err != nil {
			return fmt.Errorf("internal error validating source entity: %w", err)
		}
		if count == 0 {
			var deleted int64
			s.db.Raw(`SELECT count(*) FROM projects WHERE id = ? AND organization_id = ? AND deleted_at IS NOT NULL`, id, orgID).Scan(&deleted)
			if deleted > 0 {
				return ErrEntityDeleted
			}
			return ErrEntityNotFound
		}
		return nil
	case "vendor":
		var count int64
		if err := s.db.Raw(`SELECT count(*) FROM vendors WHERE id = ? AND organization_id = ? AND deleted_at IS NULL`, id, orgID).Scan(&count).Error; err != nil {
			return fmt.Errorf("internal error validating source entity: %w", err)
		}
		if count == 0 {
			var deleted int64
			s.db.Raw(`SELECT count(*) FROM vendors WHERE id = ? AND organization_id = ? AND deleted_at IS NOT NULL`, id, orgID).Scan(&deleted)
			if deleted > 0 {
				return ErrEntityDeleted
			}
			return ErrEntityNotFound
		}
		return nil
	case "government_mapping":
		var m struct {
			MatchStatus string
		}
		err := s.db.Raw(`SELECT match_status FROM government_external_mappings WHERE id = ? AND organization_id = ?`, id, orgID).Row().Scan(&m.MatchStatus)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
				return ErrEntityNotFound
			}
			return fmt.Errorf("internal error validating government mapping: %w", err)
		}
		switch m.MatchStatus {
		case "MATCHED":
			return nil
		case "PENDING_MATCH":
			return ErrPendingMatchNotReady
		case "REJECTED":
			return ErrMappingRejected
		default:
			return ErrUnknownMatchStatus
		}
	}
	return ErrInvalidEntityType
}

// isPeriodLocked checks whether the (org, dataset, year, month) period is locked.
// periodMonth 0 or nil means the full-year lock applies.
func (s *Service) isPeriodLocked(orgID uuid.UUID, datasetType string, year, month int) bool {
	q := s.db.Model(&LockPeriod{}).
		Where("organization_id = ? AND dataset_type = ? AND period_year = ? AND status = ? AND deleted_at IS NULL",
			orgID, datasetType, year, LockLocked)
	// If month is provided, check that specific month OR a full-year lock
	if month > 0 {
		q = q.Where("period_month = ? OR period_month IS NULL", month)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}

// periodMonthOrZero returns the month if valid, else 0.
func periodMonthOrZero(m int) int {
	if m > 0 && m <= 12 {
		return m
	}
	return 0
}

// monthPtrOrZero dereferences a *int month, returning 0 for nil/invalid.
func monthPtrOrZero(m *int) int {
	if m == nil {
		return 0
	}
	return periodMonthOrZero(*m)
}

func validItemAction(action string) bool {
	switch action {
	case ItemActionCreate, ItemActionUpdate, ItemActionDelete, ItemActionUpsert, ItemActionValidateOnly:
		return true
	}
	return false
}

// actionRequiresEntityID reports whether an item action must reference an
// existing entity id. UPDATE/DELETE/UPSERT mutate an existing entity and thus
// require entity_id; CREATE (new entity) and VALIDATE_ONLY (payload-only
// validation) may omit it.
func actionRequiresEntityID(action string) bool {
	switch action {
	case ItemActionUpdate, ItemActionDelete, ItemActionUpsert:
		return true
	}
	return false
}

func (s *Service) auditSubmission(actorID, orgID uuid.UUID, sub *Submission, action string, values map[string]interface{}) {
	if s.auditor == nil {
		return
	}
	s.auditor.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		Action:         action,
		EntityType:     "data_submission",
		EntityID:       sub.ID.String(),
		EntityLabel:    sub.DatasetType + "/" + sub.SourceType,
		NewValues:      values,
	})
}
