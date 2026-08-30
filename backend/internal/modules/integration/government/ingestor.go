package government

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Ingestor — SAMPLE / DRY_RUN / COMMIT logic
// ---------------------------------------------------------------------------
// The ingestor processes synthetic (mock) records for each dataset type.
// In a production integration, the fetch step would call the government API
// using credentials stored in a secure secrets manager — never in this file.
// ---------------------------------------------------------------------------

// IngestResult summarises the outcome of one ingest pass.
type IngestResult struct {
	TotalRecords    int
	AcceptedRecords int
	RejectedRecords int
	Records         []SyncRecord
	ErrorSummary    []ErrorEntry
}

// ErrorEntry is a single validation or processing error.
type ErrorEntry struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	ExternalID string `json:"external_id,omitempty"`
	Row        int    `json:"row,omitempty"`
}

// sampleRecord is a raw government data record used for mock/sandbox ingestion.
type sampleRecord struct {
	ExternalID string                 `json:"external_id"`
	Payload    map[string]interface{} `json:"payload"`
}

// ---------------------------------------------------------------------------
// Ingest is the main entry point called by the service.
// ---------------------------------------------------------------------------

// Ingest fetches (or generates sample) records for the given connector and
// dataset type, validates them, and — if mode == COMMIT — writes lineage
// mappings into the database.
//
// Rules:
//   - SAMPLE: return first 5 mock records only, never write anything.
//   - DRY_RUN: validate all mock records, return full preview, never write.
//   - COMMIT: validate + upsert government_external_mappings, write records log.
func Ingest(
	ctx context.Context,
	db *gorm.DB,
	run *SyncRun,
) (*IngestResult, error) {
	records := generateSampleRecords(run.ConnectorKey, run.DatasetType)

	// SAMPLE mode: cap at 5 records, never write
	if run.Mode == ModeSample && len(records) > 5 {
		records = records[:5]
	}

	result := &IngestResult{
		TotalRecords: len(records),
		Records:      make([]SyncRecord, 0, len(records)),
		ErrorSummary: []ErrorEntry{},
	}

	for i, rec := range records {
		validated, errs := validateRecord(rec, run.DatasetType)
		rawBytes, _ := json.Marshal(rec.Payload)
		errBytes, _ := json.Marshal(errs)

		sr := SyncRecord{
			ID:               uuid.New(),
			SyncRunID:        run.ID,
			OrganizationID:   run.OrganizationID,
			ExternalID:       rec.ExternalID,
			DatasetType:      run.DatasetType,
			ValidationErrors: errBytes,
			RawPayload:       rawBytes,
			CreatedAt:        time.Now().UTC(),
		}

		if len(errs) > 0 {
			sr.Status = RecordRejected
			sr.Action = ActionSkip
			result.RejectedRecords++
			for _, e := range errs {
				e.ExternalID = rec.ExternalID
				e.Row = i + 1
				result.ErrorSummary = append(result.ErrorSummary, e)
			}
		} else {
			sr.Status = RecordAccepted
			result.AcceptedRecords++

			// Determine CREATE vs UPDATE based on existing mapping
			if run.Mode == ModeCommit {
				action, err := upsertMapping(ctx, db, run, rec, validated)
				if err != nil {
					sr.Status = RecordRejected
					sr.Action = ActionSkip
					result.AcceptedRecords--
					result.RejectedRecords++
					result.ErrorSummary = append(result.ErrorSummary, ErrorEntry{
						Code:       "E_MAPPING",
						Message:    err.Error(),
						ExternalID: rec.ExternalID,
						Row:        i + 1,
					})
				} else {
					sr.Action = action
				}
			} else {
				// DRY_RUN or SAMPLE: check if mapping exists for preview only
				exists, _ := mappingExists(ctx, db, run.OrganizationID, run.ConnectorKey, run.DatasetType, rec.ExternalID)
				if exists {
					sr.Action = ActionUpdate
				} else {
					sr.Action = ActionCreate
				}
			}
		}

		result.Records = append(result.Records, sr)
	}

	// Persist record log only on COMMIT
	if run.Mode == ModeCommit && len(result.Records) > 0 {
		if err := db.WithContext(ctx).Create(&result.Records).Error; err != nil {
			return result, fmt.Errorf("ingestor: persist records: %w", err)
		}
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func validateRecord(rec sampleRecord, datasetType string) (map[string]interface{}, []ErrorEntry) {
	var errs []ErrorEntry
	if rec.ExternalID == "" {
		errs = append(errs, ErrorEntry{Code: "E_REQUIRED", Message: "external_id is required"})
	}
	if rec.Payload == nil {
		errs = append(errs, ErrorEntry{Code: "E_REQUIRED", Message: "payload is empty"})
	}
	switch datasetType {
	case DatasetProjects:
		if _, ok := rec.Payload["name"]; !ok {
			errs = append(errs, ErrorEntry{Code: "E_REQUIRED", Message: "field 'name' is required"})
		}
	case DatasetBudgetAllocation:
		if _, ok := rec.Payload["amount"]; !ok {
			errs = append(errs, ErrorEntry{Code: "E_REQUIRED", Message: "field 'amount' is required"})
		}
	case DatasetLocations:
		if _, ok := rec.Payload["province_code"]; !ok {
			errs = append(errs, ErrorEntry{Code: "E_REQUIRED", Message: "field 'province_code' is required"})
		}
	case DatasetVendors:
		if _, ok := rec.Payload["npwp"]; !ok {
			errs = append(errs, ErrorEntry{Code: "E_REQUIRED", Message: "field 'npwp' is required"})
		}
	}
	return rec.Payload, errs
}

func payloadHash(payload map[string]interface{}) string {
	b, _ := json.Marshal(payload)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

func mappingExists(ctx context.Context, db *gorm.DB, orgID uuid.UUID, connectorKey, datasetType, externalID string) (bool, error) {
	var count int64
	err := db.WithContext(ctx).Model(&ExternalMapping{}).
		Where("organization_id = ? AND connector_key = ? AND dataset_type = ? AND external_id = ?",
			orgID, connectorKey, datasetType, externalID).
		Count(&count).Error
	return count > 0, err
}

func upsertMapping(ctx context.Context, db *gorm.DB, run *SyncRun, rec sampleRecord, payload map[string]interface{}) (string, error) {
	hash := payloadHash(payload)

	var existing ExternalMapping
	err := db.WithContext(ctx).
		Where("organization_id = ? AND connector_key = ? AND dataset_type = ? AND external_id = ?",
			run.OrganizationID, run.ConnectorKey, run.DatasetType, rec.ExternalID).
		First(&existing).Error

	now := time.Now().UTC()

	if err == gorm.ErrRecordNotFound {
		// CREATE
		// InternalEntityID is intentionally nil: the government sync creates a
		// PENDING_MATCH record because no automated resolution to an internal
		// CANKORA entity has taken place yet.  A separate matching/reconciliation
		// step must set InternalEntityID and promote MatchStatus to MATCHED.
		// NEVER use uuid.New() as a placeholder — that would fabricate a lineage link.
		mapping := ExternalMapping{
			ID:                 uuid.New(),
			OrganizationID:     run.OrganizationID,
			ConnectorKey:       run.ConnectorKey,
			DatasetType:        run.DatasetType,
			ExternalID:         rec.ExternalID,
			InternalEntityType: datasetTypeToEntityType(run.DatasetType),
			InternalEntityID:   nil,
			MatchStatus:        MatchStatusPendingMatch,
			SourcePayloadHash:  hash,
			LastSeenAt:         now,
			SyncRunID:          &run.ID,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if err2 := db.WithContext(ctx).Create(&mapping).Error; err2 != nil {
			return "", err2
		}
		return ActionCreate, nil
	}

	if err != nil {
		return "", err
	}

	// UPDATE: only write if payload changed
	updates := map[string]interface{}{
		"source_payload_hash": hash,
		"last_seen_at":        now,
		"sync_run_id":         run.ID,
		"updated_at":          now,
	}
	if err2 := db.WithContext(ctx).Model(&existing).Updates(updates).Error; err2 != nil {
		return "", err2
	}
	return ActionUpdate, nil
}

func datasetTypeToEntityType(datasetType string) string {
	switch datasetType {
	case DatasetProjects:
		return "project"
	case DatasetBudgetAllocation:
		return "budget"
	case DatasetLocations:
		return "location"
	case DatasetVendors:
		return "vendor"
	default:
		return datasetType
	}
}

// generateSampleRecords returns mock records for a given connector + dataset type.
// In production, this would be replaced by an HTTP call to the government API.
func generateSampleRecords(connectorKey, datasetType string) []sampleRecord {
	switch datasetType {
	case DatasetProjects:
		return []sampleRecord{
			{ExternalID: "SIRUP-2024-0001", Payload: map[string]interface{}{"name": "Pembangunan Bendungan Ciawi", "fiscal_year": 2024, "satker_code": "033.01"}},
			{ExternalID: "SIRUP-2024-0002", Payload: map[string]interface{}{"name": "Rehabilitasi Jaringan Irigasi DI Rentang", "fiscal_year": 2024, "satker_code": "033.02"}},
			{ExternalID: "SIRUP-2024-0003", Payload: map[string]interface{}{"name": "Normalisasi Kali Bekasi", "fiscal_year": 2024, "satker_code": "033.03"}},
			{ExternalID: "SIRUP-2024-0004", Payload: map[string]interface{}{"name": "Pembangunan Embung Kuwil", "fiscal_year": 2024, "satker_code": "033.04"}},
			{ExternalID: "SIRUP-2024-0005", Payload: map[string]interface{}{"name": "Peningkatan Kapasitas SPAM IKK Woja", "fiscal_year": 2024, "satker_code": "033.05"}},
			{ExternalID: "SIRUP-2024-0006", Payload: map[string]interface{}{"name": "Pembangunan Tanggul Pantai Muara Baru", "fiscal_year": 2024, "satker_code": "033.06"}},
			{ExternalID: "SIRUP-2024-0007", Payload: map[string]interface{}{"name": "Pemeliharaan Bendung Cengkrong", "fiscal_year": 2024, "satker_code": "033.07"}},
			{ExternalID: "SIRUP-2024-0008", Payload: map[string]interface{}{"name": "Revitalisasi SPAM Perkotaan Gresik", "fiscal_year": 2024, "satker_code": "033.08"}},
		}
	case DatasetBudgetAllocation:
		return []sampleRecord{
			{ExternalID: "OMSPAN-2024-BA-001", Payload: map[string]interface{}{"project_ref": "SIRUP-2024-0001", "amount": 125000000000, "currency": "IDR", "fiscal_year": 2024}},
			{ExternalID: "OMSPAN-2024-BA-002", Payload: map[string]interface{}{"project_ref": "SIRUP-2024-0002", "amount": 87500000000, "currency": "IDR", "fiscal_year": 2024}},
			{ExternalID: "OMSPAN-2024-BA-003", Payload: map[string]interface{}{"project_ref": "SIRUP-2024-0003", "amount": 45000000000, "currency": "IDR", "fiscal_year": 2024}},
			{ExternalID: "OMSPAN-2024-BA-004", Payload: map[string]interface{}{"project_ref": "SIRUP-2024-0004", "amount": 32000000000, "currency": "IDR", "fiscal_year": 2024}},
			{ExternalID: "OMSPAN-2024-BA-005", Payload: map[string]interface{}{"project_ref": "SIRUP-2024-0005", "amount": 18500000000, "currency": "IDR", "fiscal_year": 2024}},
		}
	case DatasetLocations:
		return []sampleRecord{
			{ExternalID: "LOC-32", Payload: map[string]interface{}{"province_code": "32", "province_name": "Jawa Barat", "level": "province"}},
			{ExternalID: "LOC-31", Payload: map[string]interface{}{"province_code": "31", "province_name": "DKI Jakarta", "level": "province"}},
			{ExternalID: "LOC-33", Payload: map[string]interface{}{"province_code": "33", "province_name": "Jawa Tengah", "level": "province"}},
			{ExternalID: "LOC-35", Payload: map[string]interface{}{"province_code": "35", "province_name": "Jawa Timur", "level": "province"}},
			{ExternalID: "LOC-36", Payload: map[string]interface{}{"province_code": "36", "province_name": "Banten", "level": "province"}},
		}
	case DatasetVendors:
		return []sampleRecord{
			{ExternalID: "SIRUP-VENDOR-001", Payload: map[string]interface{}{"npwp": "01.234.567.8-901.000", "name": "PT. Waskita Karya (Persero) Tbk", "category": "BUMN"}},
			{ExternalID: "SIRUP-VENDOR-002", Payload: map[string]interface{}{"npwp": "02.345.678.9-012.000", "name": "PT. Wijaya Karya (Persero) Tbk", "category": "BUMN"}},
			{ExternalID: "SIRUP-VENDOR-003", Payload: map[string]interface{}{"npwp": "03.456.789.0-123.000", "name": "PT. Brantas Abipraya (Persero)", "category": "BUMN"}},
			{ExternalID: "SIRUP-VENDOR-004", Payload: map[string]interface{}{"npwp": "04.567.890.1-234.000", "name": "PT. Adhi Karya (Persero) Tbk", "category": "BUMN"}},
			{ExternalID: "SIRUP-VENDOR-005", Payload: map[string]interface{}{"npwp": "05.678.901.2-345.000", "name": "PT. Hutama Karya (Persero)", "category": "BUMN"}},
		}
	default:
		return []sampleRecord{}
	}
}
