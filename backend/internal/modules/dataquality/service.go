package dataquality

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/audit"
	"github.com/harmanto-49/cankora/internal/modules/monitoring"
	"gorm.io/gorm"
)

var ErrSubmissionNotFound = errors.New("data submission not found")
var ErrSelfValidation = errors.New("submitter cannot validate the same submission")

type Service struct {
	db      *gorm.DB
	auditor *audit.Writer
}

func NewService(db *gorm.DB, auditor *audit.Writer) *Service {
	return &Service{db: db, auditor: auditor}
}

func (s *Service) List(orgID uuid.UUID, status string) ([]Submission, error) {
	var list []Submission
	// Only snapshot-driven validation-queue rows belong to dataquality.
	// Governance rows (snapshot_id IS NULL) are managed by the governance module.
	q := s.db.Where("organization_id = ? AND deleted_at IS NULL AND snapshot_id IS NOT NULL", orgID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	return list, q.Order("CASE WHEN sla_due_at < NOW() THEN 0 ELSE 1 END, submitted_at ASC").Find(&list).Error
}

func (s *Service) Get(orgID, id uuid.UUID) (*Submission, error) {
	var item Submission
	// Restrict to dataquality-owned rows (snapshot-driven) for the same reason.
	if err := s.db.Where("organization_id = ? AND id = ? AND deleted_at IS NULL AND snapshot_id IS NOT NULL", orgID, id).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSubmissionNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (s *Service) Create(orgID, actorID uuid.UUID, req CreateSubmissionRequest) (*Submission, error) {
	snapshotID, err := uuid.Parse(req.SnapshotID)
	if err != nil {
		return nil, errors.New("invalid snapshot_id")
	}
	var submission Submission
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var snap monitoring.Snapshot
		if err := tx.Where("organization_id = ? AND id = ? AND deleted_at IS NULL", orgID, snapshotID).First(&snap).Error; err != nil {
			return ErrSubmissionNotFound
		}
		if snap.Status == StatusDraft {
			now := time.Now()
			snap.Status = StatusSubmitted
			snap.SubmittedAt = &now
			snap.SubmittedBy = &actorID
			if err := tx.Save(&snap).Error; err != nil {
				return err
			}
		} else if snap.Status != StatusSubmitted {
			return errors.New("snapshot must be DRAFT or SUBMITTED")
		}
		if req.CompletenessPct == 0 {
			req.CompletenessPct = 100
		}
		now := time.Now()
		slaHours := req.SLAHours
		if slaHours <= 0 {
			slaHours = 72
		}
		slaDue := now.Add(time.Duration(slaHours) * time.Hour)
		lineage, err := json.Marshal(req.Lineage)
		if err != nil {
			return fmt.Errorf("invalid lineage: %w", err)
		}
		submission = Submission{
			OrganizationID: orgID, ProjectID: snap.ProjectID, SnapshotID: snap.ID,
			Source: snap.Source, SourceReference: req.SourceReference,
			PeriodYear: snap.PeriodYear, PeriodMonth: snap.PeriodMonth,
			Status: StatusSubmitted, CompletenessPct: req.CompletenessPct,
			FreshnessAt: req.FreshnessAt, SLADueAt: &slaDue,
			SubmittedBy: &actorID, SubmittedAt: &now, Lineage: lineage,
		}
		return tx.Create(&submission).Error
	})
	if err != nil {
		return nil, err
	}
	s.recordAudit(actorID, orgID, submission.ID, "data_submission.submitted", submission)
	return &submission, nil
}

func (s *Service) Transition(orgID, actorID, id uuid.UUID, req TransitionRequest) (*Submission, error) {
	item, err := s.Get(orgID, id)
	if err != nil {
		return nil, err
	}
	if item.Status != StatusSubmitted {
		return nil, errors.New("only SUBMITTED submissions can be validated")
	}
	if item.SubmittedBy != nil && *item.SubmittedBy == actorID && requireSoD() {
		return nil, ErrSelfValidation
	}
	if req.Status == StatusRejected && req.RejectionReason == "" {
		return nil, errors.New("rejection_reason is required")
	}
	now := time.Now()
	item.Status = req.Status
	item.ValidatorID = &actorID
	item.ValidatedAt = &now
	item.RejectionReason = req.RejectionReason
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(item).Error; err != nil {
			return err
		}
		return tx.Model(&monitoring.Snapshot{}).
			Where("organization_id = ? AND id = ?", orgID, item.SnapshotID).
			Updates(map[string]interface{}{"status": req.Status, "validated_at": now, "validated_by": actorID, "rejection_reason": req.RejectionReason}).Error
	})
	if err != nil {
		return nil, err
	}
	s.recordAudit(actorID, orgID, item.ID, "data_submission."+strings.ToLower(req.Status), *item)
	return item, nil
}

func requireSoD() bool {
	value, ok := os.LookupEnv("DATA_VALIDATION_REQUIRE_SOD")
	if !ok {
		return true
	}
	result, err := strconv.ParseBool(value)
	return err != nil || result
}

func (s *Service) recordAudit(actorID, orgID, id uuid.UUID, action string, item Submission) {
	if s.auditor == nil {
		return
	}
	s.auditor.Record(audit.WriteRequest{OrganizationID: orgID, ActorID: &actorID, Action: action, EntityType: "data_submission", EntityID: id.String(), EntityLabel: fmt.Sprintf("snapshot %s", item.SnapshotID)})
}
