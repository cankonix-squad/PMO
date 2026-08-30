package imports

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/audit"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a job does not exist for the organisation.
var ErrNotFound = errors.New("import job not found")

// ErrForbidden is returned when a job belongs to another tenant.
var ErrForbidden = errors.New("forbidden")

// ErrInvalidTransition is returned when a status transition is not allowed.
var ErrInvalidTransition = errors.New("invalid status transition")

// Service handles all import job business logic.
type Service struct {
	db          *gorm.DB
	auditWriter *audit.Writer
	log         *zap.Logger
}

// NewService creates a new import Service.
func NewService(db *gorm.DB, auditWriter *audit.Writer, log *zap.Logger) *Service {
	return &Service{db: db, auditWriter: auditWriter, log: log}
}

// --- Job CRUD ---

// ListJobs returns paginated import jobs for an organisation.
func (s *Service) ListJobs(ctx context.Context, orgID uuid.UUID, f ListJobsFilter) ([]Job, int64, error) {
	q := s.db.WithContext(ctx).Model(&Job{}).Where("organization_id = ?", orgID)
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

	var jobs []Job
	err := q.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&jobs).Error
	return jobs, total, err
}

// GetJob returns a single job, enforcing tenant ownership.
func (s *Service) GetJob(ctx context.Context, orgID, jobID uuid.UUID) (*Job, error) {
	var job Job
	err := s.db.WithContext(ctx).Where("id = ? AND organization_id = ?", jobID, orgID).First(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &job, err
}

// CreateJob persists a new UPLOADED job record (file bytes already saved by handler).
func (s *Service) CreateJob(ctx context.Context, orgID, userID uuid.UUID, datasetType, fileName, mimeType string, fileSize int64) (*Job, error) {
	// validate dataset type
	if _, ok := TemplateByType(datasetType); !ok {
		return nil, fmt.Errorf("unsupported dataset_type: %s", datasetType)
	}

	job := &Job{
		ID:             uuid.New(),
		OrganizationID: orgID,
		DatasetType:    datasetType,
		FileName:       fileName,
		FileSize:       fileSize,
		MIMEType:       mimeType,
		Status:         StatusUploaded,
		ErrorSummary:   "[]",
		UploadedBy:     userID,
	}
	if err := s.db.WithContext(ctx).Create(job).Error; err != nil {
		return nil, err
	}

	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &userID,
		Action:         "import.job.created",
		EntityType:     "import_job",
		EntityID:       job.ID.String(),
		EntityLabel:    fileName,
		NewValues:      job,
	})
	return job, nil
}

// ValidateJob parses the stored CSV bytes, validates rows per dataset rules,
// saves import_rows, and updates job counters + status.
func (s *Service) ValidateJob(ctx context.Context, orgID, userID, jobID uuid.UUID, fileReader io.Reader) (*Job, error) {
	job, err := s.GetJob(ctx, orgID, jobID)
	if err != nil {
		return nil, err
	}
	if job.Status != StatusUploaded {
		return nil, fmt.Errorf("%w: job must be UPLOADED to validate (current: %s)", ErrInvalidTransition, job.Status)
	}

	parsedRows, parseErr := ParseCSV(fileReader)
	if parseErr != nil {
		// Mark job FAILED
		now := time.Now()
		errSummary := marshalErrors([]string{parseErr.Error()})
		s.db.WithContext(ctx).Model(job).Updates(map[string]any{
			"status":        StatusFailed,
			"error_summary": errSummary,
			"validated_at":  now,
		})
		s.auditWriter.Record(audit.WriteRequest{
			OrganizationID: orgID, ActorID: &userID,
			Action: "import.job.failed", EntityType: "import_job",
			EntityID: job.ID.String(), EntityLabel: job.FileName,
		})
		return nil, fmt.Errorf("parse error: %w", parseErr)
	}

	// Validate rows per dataset type
	validator := s.validatorFor(job.DatasetType)
	rowResults := make([]rowResult, 0, len(parsedRows))
	for _, pr := range parsedRows {
		rr := validator(ctx, orgID, pr)
		rowResults = append(rowResults, rr)
	}

	// Persist rows in a transaction
	validCount, invalidCount := 0, 0
	var topErrors []string

	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete any previous rows for this job (re-validate scenario)
		if err := tx.Where("job_id = ?", job.ID).Delete(&Row{}).Error; err != nil {
			return err
		}

		for _, rr := range rowResults {
			if rr.valid {
				validCount++
			} else {
				invalidCount++
				if len(topErrors) < 10 {
					topErrors = append(topErrors, fmt.Sprintf("row %d: %s", rr.rowNumber, strings.Join(rr.errors, "; ")))
				}
			}

			row := &Row{
				ID:                uuid.New(),
				JobID:             job.ID,
				RowNumber:         rr.rowNumber,
				RawPayload:        marshalJSON(rr.raw),
				NormalizedPayload: marshalJSON(rr.normalized),
				Valid:             rr.valid,
				Errors:            marshalErrors(rr.errors),
				Action:            rr.action,
				TargetEntityID:    rr.targetID,
			}
			if err := tx.Create(row).Error; err != nil {
				return err
			}
		}

		now := time.Now()
		errJSON := marshalErrors(topErrors)
		status := StatusValidated
		if validCount == 0 && invalidCount > 0 {
			status = StatusFailed
		}
		return tx.Model(job).Updates(map[string]any{
			"status":        status,
			"total_rows":    len(rowResults),
			"valid_rows":    validCount,
			"invalid_rows":  invalidCount,
			"error_summary": errJSON,
			"validated_at":  now,
		}).Error
	})
	if txErr != nil {
		return nil, txErr
	}

	job, _ = s.GetJob(ctx, orgID, jobID)

	auditAction := "import.job.validated"
	if job.Status == StatusFailed {
		auditAction = "import.job.failed"
	}
	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID, ActorID: &userID,
		Action: auditAction, EntityType: "import_job",
		EntityID: job.ID.String(), EntityLabel: job.FileName,
		NewValues: map[string]any{"valid_rows": validCount, "invalid_rows": invalidCount},
	})
	return job, nil
}

// CommitJob writes validated rows into target tables.
func (s *Service) CommitJob(ctx context.Context, orgID, userID, jobID uuid.UUID) (*Job, error) {
	job, err := s.GetJob(ctx, orgID, jobID)
	if err != nil {
		return nil, err
	}
	if job.Status != StatusValidated {
		return nil, fmt.Errorf("%w: job must be VALIDATED to commit (current: %s)", ErrInvalidTransition, job.Status)
	}

	// Load valid rows only
	var rows []Row
	if err := s.db.WithContext(ctx).
		Where("job_id = ? AND valid = true", job.ID).
		Order("row_number ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	committer := s.committerFor(job.DatasetType)
	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range rows {
			if err := committer(ctx, tx, orgID, userID, &rows[i]); err != nil {
				return fmt.Errorf("commit row %d: %w", rows[i].RowNumber, err)
			}
		}
		now := time.Now()
		return tx.Model(job).Updates(map[string]any{
			"status":       StatusCommitted,
			"committed_at": now,
		}).Error
	})
	if txErr != nil {
		s.db.WithContext(ctx).Model(job).Updates(map[string]any{"status": StatusFailed})
		s.auditWriter.Record(audit.WriteRequest{
			OrganizationID: orgID, ActorID: &userID,
			Action: "import.job.failed", EntityType: "import_job",
			EntityID: job.ID.String(),
		})
		return nil, txErr
	}

	job, _ = s.GetJob(ctx, orgID, jobID)
	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID, ActorID: &userID,
		Action: "import.job.committed", EntityType: "import_job",
		EntityID: job.ID.String(), EntityLabel: job.FileName,
		NewValues: map[string]any{"committed_rows": job.ValidRows},
	})
	return job, nil
}

// CancelJob transitions an UPLOADED or VALIDATED job to CANCELLED.
func (s *Service) CancelJob(ctx context.Context, orgID, userID, jobID uuid.UUID) (*Job, error) {
	job, err := s.GetJob(ctx, orgID, jobID)
	if err != nil {
		return nil, err
	}
	if job.Status != StatusUploaded && job.Status != StatusValidated {
		return nil, fmt.Errorf("%w: only UPLOADED or VALIDATED jobs can be cancelled", ErrInvalidTransition)
	}
	if err := s.db.WithContext(ctx).Model(job).Update("status", StatusCancelled).Error; err != nil {
		return nil, err
	}
	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID, ActorID: &userID,
		Action: "import.job.cancelled", EntityType: "import_job",
		EntityID: job.ID.String(),
	})
	return job, nil
}

// ListRows returns all rows for a job with optional valid/invalid filter.
func (s *Service) ListRows(ctx context.Context, orgID, jobID uuid.UUID, validOnly *bool, page, pageSize int) ([]Row, int64, error) {
	// verify ownership
	if _, err := s.GetJob(ctx, orgID, jobID); err != nil {
		return nil, 0, err
	}
	q := s.db.WithContext(ctx).Model(&Row{}).Where("job_id = ?", jobID)
	if validOnly != nil {
		q = q.Where("valid = ?", *validOnly)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize
	var rows []Row
	err := q.Order("row_number ASC").Limit(pageSize).Offset(offset).Find(&rows).Error
	return rows, total, err
}

// --- Internal validation helpers ---

type rowResult struct {
	rowNumber  int
	raw        map[string]string
	normalized map[string]any
	valid      bool
	errors     []string
	action     string
	targetID   *uuid.UUID
}

// validatorFor returns the appropriate row validator for a dataset type.
func (s *Service) validatorFor(datasetType string) func(ctx context.Context, orgID uuid.UUID, pr ParsedRow) rowResult {
	switch datasetType {
	case DatasetProjectProgress:
		return s.validateProjectProgressRow
	case DatasetProjectBudgets:
		return s.validateProjectBudgetRow
	case DatasetRisks:
		return s.validateRiskRow
	case DatasetIssues:
		return s.validateIssueRow
	case DatasetBenefitMeasurements:
		return s.validateBenefitMeasurementRow
	default:
		return func(_ context.Context, _ uuid.UUID, pr ParsedRow) rowResult {
			return rowResult{rowNumber: pr.RowNumber, raw: pr.Data, valid: false, errors: []string{"unsupported dataset type"}, action: ActionSkip}
		}
	}
}

// lookupProject finds a non-deleted project by code within an org.
func (s *Service) lookupProject(ctx context.Context, orgID uuid.UUID, code string) (uuid.UUID, error) {
	type row struct{ ID uuid.UUID }
	var r row
	err := s.db.WithContext(ctx).
		Table("projects").
		Select("id").
		Where("organization_id = ? AND code = ? AND deleted_at IS NULL", orgID, code).
		First(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, fmt.Errorf("project code '%s' not found in organisation", code)
	}
	return r.ID, err
}

func (s *Service) validateProjectProgressRow(ctx context.Context, orgID uuid.UUID, pr ParsedRow) rowResult {
	rr := rowResult{rowNumber: pr.RowNumber, raw: pr.Data, action: ActionUpdate}
	normalized := map[string]any{}

	code := pr.Data["project_code"]
	if code == "" {
		rr.errors = append(rr.errors, "project_code is required")
	} else {
		pid, err := s.lookupProject(ctx, orgID, code)
		if err != nil {
			rr.errors = append(rr.errors, err.Error())
		} else {
			normalized["project_id"] = pid.String()
			rr.targetID = &pid
		}
	}

	pctStr := pr.Data["progress_pct"]
	if pctStr == "" {
		rr.errors = append(rr.errors, "progress_pct is required")
	} else {
		pct, err := strconv.ParseFloat(pctStr, 64)
		if err != nil || pct < 0 || pct > 100 {
			rr.errors = append(rr.errors, "progress_pct must be a number between 0 and 100")
		} else {
			normalized["progress_pct"] = pct
		}
	}
	if v := pr.Data["period_date"]; v != "" {
		normalized["period_date"] = v
	}
	if v := pr.Data["note"]; v != "" {
		normalized["note"] = v
	}

	rr.normalized = normalized
	rr.valid = len(rr.errors) == 0
	return rr
}

func (s *Service) validateProjectBudgetRow(ctx context.Context, orgID uuid.UUID, pr ParsedRow) rowResult {
	rr := rowResult{rowNumber: pr.RowNumber, raw: pr.Data, action: ActionCreate}
	normalized := map[string]any{}

	code := pr.Data["project_code"]
	if code == "" {
		rr.errors = append(rr.errors, "project_code is required")
	} else {
		pid, err := s.lookupProject(ctx, orgID, code)
		if err != nil {
			rr.errors = append(rr.errors, err.Error())
		} else {
			normalized["project_id"] = pid.String()
			rr.targetID = &pid
		}
	}

	category := pr.Data["category"]
	if category == "" {
		rr.errors = append(rr.errors, "category is required")
	} else {
		normalized["category"] = strings.ToUpper(category)
	}

	if v := pr.Data["planned"]; v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f < 0 {
			rr.errors = append(rr.errors, "planned must be a non-negative number")
		} else {
			normalized["planned"] = f
		}
	}
	if v := pr.Data["actual"]; v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f < 0 {
			rr.errors = append(rr.errors, "actual must be a non-negative number")
		} else {
			normalized["actual"] = f
		}
	}
	currency := pr.Data["currency"]
	if currency == "" {
		currency = "IDR"
	}
	normalized["currency"] = strings.ToUpper(currency)

	// Check if row is CREATE or UPDATE
	if rr.targetID != nil && category != "" {
		var exists int64
		s.db.WithContext(ctx).Table("project_budgets").
			Where("project_id = ? AND category = ? AND deleted_at IS NULL", rr.targetID, strings.ToUpper(category)).
			Count(&exists)
		if exists > 0 {
			rr.action = ActionUpdate
		}
	}

	rr.normalized = normalized
	rr.valid = len(rr.errors) == 0
	return rr
}

func (s *Service) validateRiskRow(ctx context.Context, orgID uuid.UUID, pr ParsedRow) rowResult {
	rr := rowResult{rowNumber: pr.RowNumber, raw: pr.Data, action: ActionCreate}
	normalized := map[string]any{}

	code := pr.Data["project_code"]
	if code == "" {
		rr.errors = append(rr.errors, "project_code is required")
	} else {
		pid, err := s.lookupProject(ctx, orgID, code)
		if err != nil {
			rr.errors = append(rr.errors, err.Error())
		} else {
			normalized["project_id"] = pid.String()
			rr.targetID = &pid
		}
	}

	if pr.Data["title"] == "" {
		rr.errors = append(rr.errors, "title is required")
	} else {
		normalized["title"] = pr.Data["title"]
	}

	probStr := pr.Data["probability"]
	if probStr == "" {
		rr.errors = append(rr.errors, "probability is required")
	} else {
		prob, err := strconv.Atoi(probStr)
		if err != nil || prob < 1 || prob > 5 {
			rr.errors = append(rr.errors, "probability must be an integer 1-5")
		} else {
			normalized["probability"] = prob
		}
	}

	impactStr := pr.Data["impact"]
	if impactStr == "" {
		rr.errors = append(rr.errors, "impact is required")
	} else {
		imp, err := strconv.Atoi(impactStr)
		if err != nil || imp < 1 || imp > 5 {
			rr.errors = append(rr.errors, "impact must be an integer 1-5")
		} else {
			normalized["impact"] = imp
		}
	}

	// Backend computes risk_score — never trust file input
	if prob, ok := normalized["probability"].(int); ok {
		if imp, ok2 := normalized["impact"].(int); ok2 {
			normalized["risk_score"] = prob * imp
			switch {
			case prob*imp >= 20:
				normalized["severity"] = "CRITICAL"
			case prob*imp >= 12:
				normalized["severity"] = "HIGH"
			case prob*imp >= 6:
				normalized["severity"] = "MEDIUM"
			default:
				normalized["severity"] = "LOW"
			}
		}
	}

	for _, k := range []string{"description", "mitigation", "owner", "due_date"} {
		if v := pr.Data[k]; v != "" {
			normalized[k] = v
		}
	}

	rr.normalized = normalized
	rr.valid = len(rr.errors) == 0
	return rr
}

func (s *Service) validateIssueRow(ctx context.Context, orgID uuid.UUID, pr ParsedRow) rowResult {
	rr := rowResult{rowNumber: pr.RowNumber, raw: pr.Data, action: ActionCreate}
	normalized := map[string]any{}

	code := pr.Data["project_code"]
	if code == "" {
		rr.errors = append(rr.errors, "project_code is required")
	} else {
		pid, err := s.lookupProject(ctx, orgID, code)
		if err != nil {
			rr.errors = append(rr.errors, err.Error())
		} else {
			normalized["project_id"] = pid.String()
			rr.targetID = &pid
		}
	}

	if pr.Data["title"] == "" {
		rr.errors = append(rr.errors, "title is required")
	} else {
		normalized["title"] = pr.Data["title"]
	}

	validSeverities := map[string]bool{"CRITICAL": true, "HIGH": true, "MEDIUM": true, "LOW": true}
	if sv := strings.ToUpper(pr.Data["severity"]); sv != "" {
		if !validSeverities[sv] {
			rr.errors = append(rr.errors, "severity must be CRITICAL, HIGH, MEDIUM, or LOW")
		} else {
			normalized["severity"] = sv
		}
	} else {
		normalized["severity"] = "MEDIUM"
	}

	for _, k := range []string{"description", "due_date", "owner"} {
		if v := pr.Data[k]; v != "" {
			normalized[k] = v
		}
	}

	rr.normalized = normalized
	rr.valid = len(rr.errors) == 0
	return rr
}

func (s *Service) validateBenefitMeasurementRow(ctx context.Context, orgID uuid.UUID, pr ParsedRow) rowResult {
	rr := rowResult{rowNumber: pr.RowNumber, raw: pr.Data, action: ActionCreate}
	normalized := map[string]any{}

	code := pr.Data["project_code"]
	if code == "" {
		rr.errors = append(rr.errors, "project_code is required")
	} else {
		pid, err := s.lookupProject(ctx, orgID, code)
		if err != nil {
			rr.errors = append(rr.errors, err.Error())
		} else {
			normalized["project_id"] = pid.String()
			rr.targetID = &pid
		}
	}

	indicatorName := pr.Data["indicator_name"]
	if indicatorName == "" {
		rr.errors = append(rr.errors, "indicator_name is required")
	} else if rr.targetID != nil {
		// Lookup indicator by project + name
		type indRow struct{ ID uuid.UUID }
		var ind indRow
		err := s.db.WithContext(ctx).Table("benefit_indicators").
			Select("id").
			Where("project_id = ? AND name = ? AND deleted_at IS NULL", rr.targetID, indicatorName).
			First(&ind).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			rr.errors = append(rr.errors, fmt.Sprintf("indicator '%s' not found for project", indicatorName))
		} else if err == nil {
			normalized["indicator_id"] = ind.ID.String()
		}
	}

	yearStr := pr.Data["period_year"]
	if yearStr == "" {
		rr.errors = append(rr.errors, "period_year is required")
	} else {
		y, err := strconv.Atoi(yearStr)
		if err != nil || y < 2000 || y > 2100 {
			rr.errors = append(rr.errors, "period_year must be a valid year (2000-2100)")
		} else {
			normalized["period_year"] = y
		}
	}

	month := 1
	if mStr := pr.Data["period_month"]; mStr != "" {
		m, err := strconv.Atoi(mStr)
		if err != nil || m < 1 || m > 12 {
			rr.errors = append(rr.errors, "period_month must be 1-12")
		} else {
			month = m
		}
	}
	normalized["period_month"] = month

	actualStr := pr.Data["actual_value"]
	if actualStr == "" {
		rr.errors = append(rr.errors, "actual_value is required")
	} else {
		f, err := strconv.ParseFloat(actualStr, 64)
		if err != nil {
			rr.errors = append(rr.errors, "actual_value must be a number")
		} else {
			normalized["actual"] = f
		}
	}

	if tv := pr.Data["target_value"]; tv != "" {
		f, err := strconv.ParseFloat(tv, 64)
		if err == nil {
			normalized["target"] = f
		}
	}
	normalized["validation_status"] = "DRAFT"

	rr.normalized = normalized
	rr.valid = len(rr.errors) == 0
	return rr
}

// --- Commit helpers ---

type commitFn func(ctx context.Context, tx *gorm.DB, orgID, userID uuid.UUID, row *Row) error

func (s *Service) committerFor(datasetType string) commitFn {
	switch datasetType {
	case DatasetProjectProgress:
		return s.commitProjectProgressRow
	case DatasetProjectBudgets:
		return s.commitProjectBudgetRow
	case DatasetRisks:
		return s.commitRiskRow
	case DatasetIssues:
		return s.commitIssueRow
	case DatasetBenefitMeasurements:
		return s.commitBenefitMeasurementRow
	default:
		return func(_ context.Context, _ *gorm.DB, _, _ uuid.UUID, row *Row) error {
			return fmt.Errorf("no committer for dataset type")
		}
	}
}

func unmarshalNormalized(row *Row) (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal([]byte(row.NormalizedPayload), &m); err != nil {
		return nil, err
	}
	return m, nil
}

func uuidFromMap(m map[string]any, key string) (uuid.UUID, error) {
	v, ok := m[key]
	if !ok {
		return uuid.Nil, fmt.Errorf("missing %s", key)
	}
	s, ok := v.(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("%s is not a string", key)
	}
	return uuid.Parse(s)
}

func (s *Service) commitProjectProgressRow(_ context.Context, tx *gorm.DB, _, userID uuid.UUID, row *Row) error {
	m, err := unmarshalNormalized(row)
	if err != nil {
		return err
	}
	pid, err := uuidFromMap(m, "project_id")
	if err != nil {
		return err
	}
	pct, _ := m["progress_pct"].(float64)
	updates := map[string]any{
		"progress_pct": pct,
		"updated_at":   time.Now(),
	}
	return tx.Table("projects").
		Where("id = ? AND deleted_at IS NULL", pid).
		Updates(updates).Error
}

func (s *Service) commitProjectBudgetRow(_ context.Context, tx *gorm.DB, _, userID uuid.UUID, row *Row) error {
	m, err := unmarshalNormalized(row)
	if err != nil {
		return err
	}
	pid, err := uuidFromMap(m, "project_id")
	if err != nil {
		return err
	}
	category, _ := m["category"].(string)

	updates := map[string]any{"updated_at": time.Now()}
	if v, ok := m["planned"].(float64); ok {
		updates["planned"] = v
	}
	if v, ok := m["actual"].(float64); ok {
		updates["actual"] = v
	}

	// Upsert: update existing or create new budget line
	var existingID uuid.UUID
	type budRow struct{ ID uuid.UUID }
	var br budRow
	err = tx.Table("project_budgets").Select("id").
		Where("project_id = ? AND category = ? AND deleted_at IS NULL", pid, category).
		First(&br).Error
	if err == nil {
		existingID = br.ID
	}

	if existingID != uuid.Nil {
		return tx.Table("project_budgets").Where("id = ?", existingID).Updates(updates).Error
	}

	// Create new budget line
	newLine := map[string]any{
		"id":         uuid.New(),
		"project_id": pid,
		"category":   category,
		"currency":   m["currency"],
		"created_by": userID,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
	if v, ok := m["planned"].(float64); ok {
		newLine["planned"] = v
	}
	if v, ok := m["actual"].(float64); ok {
		newLine["actual"] = v
	}
	return tx.Table("project_budgets").Create(newLine).Error
}

func (s *Service) commitRiskRow(_ context.Context, tx *gorm.DB, orgID, userID uuid.UUID, row *Row) error {
	m, err := unmarshalNormalized(row)
	if err != nil {
		return err
	}
	pid, err := uuidFromMap(m, "project_id")
	if err != nil {
		return err
	}

	prob, _ := m["probability"].(float64)
	imp, _ := m["impact"].(float64)
	score := int(prob) * int(imp)
	severity := "LOW"
	switch {
	case score >= 20:
		severity = "CRITICAL"
	case score >= 12:
		severity = "HIGH"
	case score >= 6:
		severity = "MEDIUM"
	}

	newRisk := map[string]any{
		"id":              uuid.New(),
		"organization_id": orgID,
		"project_id":      pid,
		"title":           m["title"],
		"probability":     int(prob),
		"impact":          int(imp),
		"risk_score":      score,
		"severity":        severity,
		"status":          "OPEN",
		"created_by":      userID,
		"created_at":      time.Now(),
		"updated_at":      time.Now(),
	}
	for _, k := range []string{"description", "mitigation", "owner"} {
		if v, ok := m[k].(string); ok && v != "" {
			newRisk[k] = v
		}
	}
	if dd, ok := m["due_date"].(string); ok && dd != "" {
		t, err := time.Parse("2006-01-02", dd)
		if err == nil {
			newRisk["due_date"] = t
		}
	}
	return tx.Table("risks").Create(newRisk).Error
}

func (s *Service) commitIssueRow(_ context.Context, tx *gorm.DB, orgID, userID uuid.UUID, row *Row) error {
	m, err := unmarshalNormalized(row)
	if err != nil {
		return err
	}
	pid, err := uuidFromMap(m, "project_id")
	if err != nil {
		return err
	}

	newIssue := map[string]any{
		"id":              uuid.New(),
		"organization_id": orgID,
		"project_id":      pid,
		"title":           m["title"],
		"severity":        m["severity"],
		"status":          "OPEN",
		"created_by":      userID,
		"created_at":      time.Now(),
		"updated_at":      time.Now(),
	}
	for _, k := range []string{"description", "owner"} {
		if v, ok := m[k].(string); ok && v != "" {
			newIssue[k] = v
		}
	}
	if dd, ok := m["due_date"].(string); ok && dd != "" {
		t, err := time.Parse("2006-01-02", dd)
		if err == nil {
			newIssue["due_date"] = t
		}
	}
	return tx.Table("issues").Create(newIssue).Error
}

func (s *Service) commitBenefitMeasurementRow(_ context.Context, tx *gorm.DB, orgID, userID uuid.UUID, row *Row) error {
	m, err := unmarshalNormalized(row)
	if err != nil {
		return err
	}
	indID, err := uuidFromMap(m, "indicator_id")
	if err != nil {
		return err
	}

	year, _ := m["period_year"].(float64)
	month, _ := m["period_month"].(float64)
	actual, _ := m["actual"].(float64)

	newMeas := map[string]any{
		"id":                uuid.New(),
		"organization_id":   orgID,
		"indicator_id":      indID,
		"period_year":       int(year),
		"period_month":      int(month),
		"actual":            actual,
		"baseline":          0,
		"target":            0,
		"validation_status": "DRAFT",
		"created_by":        userID,
		"created_at":        time.Now(),
		"updated_at":        time.Now(),
	}
	if tv, ok := m["target"].(float64); ok {
		newMeas["target"] = tv
	}
	return tx.Table("benefit_measurements").Create(newMeas).Error
}
