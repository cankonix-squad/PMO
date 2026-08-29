package government

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/audit"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Resolution errors (P3-002)
// ---------------------------------------------------------------------------

var ErrMappingNotFound = errors.New("external mapping not found")
var ErrAlreadyMatched = errors.New("mapping is already matched to an internal entity")
var ErrEntityTypeMismatch = errors.New("internal entity type does not match mapping dataset type")
var ErrCandidateNotFound = errors.New("candidate entity not found in this organisation")

// ---------------------------------------------------------------------------
// Internal helper structs for raw DB lookups
// ---------------------------------------------------------------------------

type projectRow struct {
	ID   uuid.UUID
	Name string
	Code string
}

type vendorRow struct {
	ID    uuid.UUID
	Name  string
	TaxID string // NPWP
}

// ---------------------------------------------------------------------------
// GetMapping — single mapping scoped to org
// ---------------------------------------------------------------------------

// GetMapping returns a single ExternalMapping by id, tenant-scoped.
func (s *Service) GetMapping(ctx context.Context, orgID, mappingID uuid.UUID) (*ExternalMapping, error) {
	var m ExternalMapping
	err := s.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", mappingID, orgID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMappingNotFound
	}
	return &m, err
}

// ---------------------------------------------------------------------------
// ListPendingMappings — paginated PENDING_MATCH queue
// ---------------------------------------------------------------------------

// ListPendingMappings returns paginated mappings whose match_status is PENDING_MATCH.
func (s *Service) ListPendingMappings(ctx context.Context, orgID uuid.UUID, f ListPendingMappingsFilter) ([]ExternalMapping, int64, error) {
	q := s.db.WithContext(ctx).Model(&ExternalMapping{}).
		Where("organization_id = ? AND match_status = ?", orgID, MatchStatusPendingMatch)
	if f.ConnectorKey != "" {
		q = q.Where("connector_key = ?", f.ConnectorKey)
	}
	if f.DatasetType != "" {
		q = q.Where("dataset_type = ?", f.DatasetType)
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
	if err := q.Order("last_seen_at DESC").Limit(pageSize).Offset(offset).Find(&mappings).Error; err != nil {
		return nil, 0, err
	}
	return mappings, total, nil
}

// ---------------------------------------------------------------------------
// GetCandidates — suggest internal entities for a pending mapping
// ---------------------------------------------------------------------------

// GetCandidates returns a ranked list of internal CANKORA entities that are
// plausible matches for the given ExternalMapping. The candidates come from
// the projects or vendors table depending on the mapping's dataset_type.
//
// Payload source: the most recent SyncRecord for this mapping's external_id
// within the same sync run (if available), otherwise falls back to external_id.
//
// Matching logic (no external API calls — purely SQL + normalisation):
//   - dataset_type "projects"  → search projects table by code then name
//   - dataset_type "vendors"   → search vendors table by tax_id (NPWP) then name
//   - other dataset types      → empty list (no candidate source available)
func (s *Service) GetCandidates(ctx context.Context, orgID, mappingID uuid.UUID) ([]ResolutionCandidate, error) {
	m, err := s.GetMapping(ctx, orgID, mappingID)
	if err != nil {
		return nil, err
	}

	// Try to load the raw payload from the most recent SyncRecord for this mapping.
	// This gives us richer field data (code, name, npwp, etc.) beyond the bare external_id.
	var payload map[string]interface{}
	if m.SyncRunID != nil {
		var rec SyncRecord
		dbErr := s.db.WithContext(ctx).
			Where("sync_run_id = ? AND external_id = ? AND organization_id = ?",
				*m.SyncRunID, m.ExternalID, orgID).
			Order("created_at DESC").
			First(&rec).Error
		if dbErr == nil && len(rec.RawPayload) > 0 {
			_ = json.Unmarshal(rec.RawPayload, &payload)
		}
	}

	// If no SyncRecord payload found, seed the payload map with the external_id
	// so the matchers still have something to work with.
	if payload == nil {
		payload = map[string]interface{}{
			"code": m.ExternalID,
			"name": m.ExternalID,
		}
	}

	switch m.DatasetType {
	case DatasetProjects:
		return s.projectCandidates(ctx, orgID, payload)
	case DatasetVendors:
		return s.vendorCandidates(ctx, orgID, payload)
	default:
		return []ResolutionCandidate{}, nil
	}
}

// projectCandidates queries the projects table for candidate matches.
// Priority: EXACT code → EXACT name → PARTIAL name → LOW_CONFIDENCE name.
func (s *Service) projectCandidates(ctx context.Context, orgID uuid.UUID, payload map[string]interface{}) ([]ResolutionCandidate, error) {
	// Extract external code and name from raw payload (field names vary by connector)
	extCode := payloadString(payload, "code", "project_code", "kode", "nomor")
	extName := payloadString(payload, "name", "nama", "project_name", "judul")

	var candidates []ResolutionCandidate
	seen := map[uuid.UUID]bool{}

	// 1. EXACT code match
	if extCode != "" {
		var rows []projectRow
		err := s.db.WithContext(ctx).
			Raw(`SELECT id, name, code FROM projects
			     WHERE organization_id = ? AND code = ? AND deleted_at IS NULL
			     LIMIT 5`, orgID, extCode).
			Scan(&rows).Error
		if err != nil {
			return nil, fmt.Errorf("government: project candidates (code): %w", err)
		}
		for _, r := range rows {
			if seen[r.ID] {
				continue
			}
			seen[r.ID] = true
			conf := ConfidenceExactCode
			candidates = append(candidates, ResolutionCandidate{
				EntityID:   r.ID.String(),
				EntityType: "project",
				Name:       r.Name,
				Code:       r.Code,
				Confidence: conf,
				Reason:     MatchReasonExactCode,
			})
		}
	}

	// 2. EXACT normalised name match
	if extName != "" {
		normExt := normaliseName(extName)
		var rows []projectRow
		err := s.db.WithContext(ctx).
			Raw(`SELECT id, name, code FROM projects
			     WHERE organization_id = ? AND deleted_at IS NULL
			     AND lower(trim(name)) = ?
			     LIMIT 5`, orgID, normExt).
			Scan(&rows).Error
		if err != nil {
			return nil, fmt.Errorf("government: project candidates (exact name): %w", err)
		}
		for _, r := range rows {
			if seen[r.ID] {
				continue
			}
			seen[r.ID] = true
			candidates = append(candidates, ResolutionCandidate{
				EntityID:   r.ID.String(),
				EntityType: "project",
				Name:       r.Name,
				Code:       r.Code,
				Confidence: ConfidenceExactName,
				Reason:     MatchReasonExactName,
			})
		}
	}

	// 3. PARTIAL name match (ILIKE both directions, cap at 10 total)
	if extName != "" && len(candidates) < 10 {
		likeExpr := "%" + strings.ToLower(normaliseName(extName)) + "%"
		var rows []projectRow
		err := s.db.WithContext(ctx).
			Raw(`SELECT id, name, code FROM projects
			     WHERE organization_id = ? AND deleted_at IS NULL
			     AND lower(name) LIKE ?
			     LIMIT 10`, orgID, likeExpr).
			Scan(&rows).Error
		if err != nil {
			return nil, fmt.Errorf("government: project candidates (partial name): %w", err)
		}
		for _, r := range rows {
			if seen[r.ID] {
				continue
			}
			seen[r.ID] = true
			candidates = append(candidates, ResolutionCandidate{
				EntityID:   r.ID.String(),
				EntityType: "project",
				Name:       r.Name,
				Code:       r.Code,
				Confidence: ConfidencePartialName,
				Reason:     MatchReasonPartialName,
			})
		}
	}

	return candidates, nil
}

// vendorCandidates queries the vendors table for candidate matches.
// Priority: EXACT tax_id (NPWP) → EXACT name → PARTIAL name.
func (s *Service) vendorCandidates(ctx context.Context, orgID uuid.UUID, payload map[string]interface{}) ([]ResolutionCandidate, error) {
	extTaxID := payloadString(payload, "tax_id", "npwp", "nib", "id_number")
	extName := payloadString(payload, "name", "nama", "vendor_name", "company_name", "company")

	var candidates []ResolutionCandidate
	seen := map[uuid.UUID]bool{}

	// 1. EXACT NPWP / tax_id match
	if extTaxID != "" {
		normTax := normaliseNPWP(extTaxID)
		var rows []vendorRow
		err := s.db.WithContext(ctx).
			Raw(`SELECT id, name, tax_id FROM vendors
			     WHERE organization_id = ? AND deleted_at IS NULL
			     AND regexp_replace(tax_id, '[^0-9]', '', 'g') = ?
			     LIMIT 5`, orgID, normTax).
			Scan(&rows).Error
		if err != nil {
			return nil, fmt.Errorf("government: vendor candidates (tax_id): %w", err)
		}
		for _, r := range rows {
			if seen[r.ID] {
				continue
			}
			seen[r.ID] = true
			candidates = append(candidates, ResolutionCandidate{
				EntityID:   r.ID.String(),
				EntityType: "vendor",
				Name:       r.Name,
				Code:       r.TaxID,
				Confidence: ConfidenceExactNPWP,
				Reason:     MatchReasonExactNPWP,
			})
		}
	}

	// 2. EXACT normalised name match
	if extName != "" {
		normExt := normaliseName(extName)
		var rows []vendorRow
		err := s.db.WithContext(ctx).
			Raw(`SELECT id, name, tax_id FROM vendors
			     WHERE organization_id = ? AND deleted_at IS NULL
			     AND lower(trim(name)) = ?
			     LIMIT 5`, orgID, normExt).
			Scan(&rows).Error
		if err != nil {
			return nil, fmt.Errorf("government: vendor candidates (exact name): %w", err)
		}
		for _, r := range rows {
			if seen[r.ID] {
				continue
			}
			seen[r.ID] = true
			candidates = append(candidates, ResolutionCandidate{
				EntityID:   r.ID.String(),
				EntityType: "vendor",
				Name:       r.Name,
				Code:       r.TaxID,
				Confidence: ConfidenceExactName,
				Reason:     MatchReasonExactName,
			})
		}
	}

	// 3. PARTIAL name match
	if extName != "" && len(candidates) < 10 {
		likeExpr := "%" + strings.ToLower(normaliseName(extName)) + "%"
		var rows []vendorRow
		err := s.db.WithContext(ctx).
			Raw(`SELECT id, name, tax_id FROM vendors
			     WHERE organization_id = ? AND deleted_at IS NULL
			     AND lower(name) LIKE ?
			     LIMIT 10`, orgID, likeExpr).
			Scan(&rows).Error
		if err != nil {
			return nil, fmt.Errorf("government: vendor candidates (partial name): %w", err)
		}
		for _, r := range rows {
			if seen[r.ID] {
				continue
			}
			seen[r.ID] = true
			candidates = append(candidates, ResolutionCandidate{
				EntityID:   r.ID.String(),
				EntityType: "vendor",
				Name:       r.Name,
				Code:       r.TaxID,
				Confidence: ConfidenceLowConfidence,
				Reason:     MatchReasonLowConfidence,
			})
		}
	}

	return candidates, nil
}

// ---------------------------------------------------------------------------
// MatchMapping — resolve a PENDING_MATCH mapping to an internal entity
// ---------------------------------------------------------------------------

// MatchMapping sets match_status = MATCHED for the given mapping, links it to
// the supplied internal entity, and writes an audit record.
//
// Rules:
//   - Only PENDING_MATCH and REJECTED mappings can be matched (not already MATCHED).
//   - The target entity must belong to the same organisation.
//   - The InternalEntityType in the request must be consistent with the mapping's dataset_type.
func (s *Service) MatchMapping(ctx context.Context, orgID, actorID, mappingID uuid.UUID, req MatchMappingRequest) (*ExternalMapping, error) {
	m, err := s.GetMapping(ctx, orgID, mappingID)
	if err != nil {
		return nil, err
	}

	if m.MatchStatus == MatchStatusMatched {
		return nil, ErrAlreadyMatched
	}

	// Parse and validate the target entity UUID
	targetID, err := uuid.Parse(req.InternalEntityID)
	if err != nil {
		return nil, fmt.Errorf("government: invalid internal_entity_id: %w", err)
	}

	// Verify entity type consistency
	expectedType := datasetToEntityType(m.DatasetType)
	if expectedType != "" && req.InternalEntityType != expectedType {
		return nil, ErrEntityTypeMismatch
	}

	// Verify the target entity exists and belongs to this org
	if err := s.verifyEntityOwnership(ctx, orgID, req.InternalEntityType, targetID); err != nil {
		return nil, err
	}

	// Determine confidence and reason
	confidence := 0
	if req.MatchConfidence != nil {
		confidence = *req.MatchConfidence
	}
	reason := req.MatchReason

	now := time.Now().UTC()
	updates := map[string]interface{}{
		"internal_entity_id":   targetID,
		"internal_entity_type": req.InternalEntityType,
		"match_status":         MatchStatusMatched,
		"match_confidence":     confidence,
		"match_reason":         reason,
		"matched_by":           actorID,
		"matched_at":           now,
		// Clear any previous rejection data
		"rejected_by":   nil,
		"rejected_at":   nil,
		"reject_reason": nil,
		"updated_at":    now,
	}

	if err := s.db.WithContext(ctx).Model(m).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("government: match mapping: %w", err)
	}

	// Reload after update
	m, err = s.GetMapping(ctx, orgID, mappingID)
	if err != nil {
		return nil, err
	}

	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		Action:         "government.mapping.matched",
		EntityType:     "government_external_mapping",
		EntityID:       mappingID.String(),
		EntityLabel:    m.ConnectorKey + "/" + m.DatasetType + "/" + m.ExternalID,
		NewValues: map[string]interface{}{
			"internal_entity_id":   targetID.String(),
			"internal_entity_type": req.InternalEntityType,
			"match_confidence":     confidence,
			"match_reason":         reason,
		},
	})

	return m, nil
}

// ---------------------------------------------------------------------------
// UnmatchMapping — revert MATCHED → PENDING_MATCH
// ---------------------------------------------------------------------------

// UnmatchMapping clears the entity link and reverts match_status to PENDING_MATCH.
func (s *Service) UnmatchMapping(ctx context.Context, orgID, actorID, mappingID uuid.UUID) (*ExternalMapping, error) {
	m, err := s.GetMapping(ctx, orgID, mappingID)
	if err != nil {
		return nil, err
	}

	if m.MatchStatus != MatchStatusMatched {
		return nil, ErrInvalidTransition
	}

	now := time.Now().UTC()
	updates := map[string]interface{}{
		"internal_entity_id": nil,
		"match_status":       MatchStatusPendingMatch,
		"match_confidence":   nil,
		"match_reason":       nil,
		"matched_by":         nil,
		"matched_at":         nil,
		"updated_at":         now,
	}

	if err := s.db.WithContext(ctx).Model(m).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("government: unmatch mapping: %w", err)
	}

	m, err = s.GetMapping(ctx, orgID, mappingID)
	if err != nil {
		return nil, err
	}

	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		Action:         "government.mapping.unmatched",
		EntityType:     "government_external_mapping",
		EntityID:       mappingID.String(),
		EntityLabel:    m.ConnectorKey + "/" + m.DatasetType + "/" + m.ExternalID,
		NewValues:      map[string]interface{}{"match_status": MatchStatusPendingMatch},
	})

	return m, nil
}

// ---------------------------------------------------------------------------
// RejectMapping — mark a mapping as REJECTED (no suitable internal entity)
// ---------------------------------------------------------------------------

// RejectMapping sets match_status = REJECTED with an optional reason.
// A rejected mapping can still be re-opened via MatchMapping.
func (s *Service) RejectMapping(ctx context.Context, orgID, actorID, mappingID uuid.UUID, req RejectMappingRequest) (*ExternalMapping, error) {
	m, err := s.GetMapping(ctx, orgID, mappingID)
	if err != nil {
		return nil, err
	}

	if m.MatchStatus == MatchStatusRejected {
		// Idempotent: already rejected — update reason if provided
	}

	now := time.Now().UTC()
	rejectReason := req.RejectReason
	updates := map[string]interface{}{
		"internal_entity_id": nil,
		"match_status":       MatchStatusRejected,
		"match_confidence":   nil,
		"match_reason":       nil,
		"matched_by":         nil,
		"matched_at":         nil,
		"rejected_by":        actorID,
		"rejected_at":        now,
		"reject_reason":      rejectReason,
		"updated_at":         now,
	}

	if err := s.db.WithContext(ctx).Model(m).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("government: reject mapping: %w", err)
	}

	m, err = s.GetMapping(ctx, orgID, mappingID)
	if err != nil {
		return nil, err
	}

	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		Action:         "government.mapping.rejected",
		EntityType:     "government_external_mapping",
		EntityID:       mappingID.String(),
		EntityLabel:    m.ConnectorKey + "/" + m.DatasetType + "/" + m.ExternalID,
		NewValues:      map[string]interface{}{"reject_reason": rejectReason},
	})

	return m, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// verifyEntityOwnership checks that the entity with the given type and id
// belongs to the organisation and has not been soft-deleted.
func (s *Service) verifyEntityOwnership(ctx context.Context, orgID uuid.UUID, entityType string, entityID uuid.UUID) error {
	var count int64
	var err error

	switch entityType {
	case "project":
		err = s.db.WithContext(ctx).
			Raw(`SELECT count(*) FROM projects WHERE id = ? AND organization_id = ? AND deleted_at IS NULL`, entityID, orgID).
			Scan(&count).Error
	case "vendor":
		err = s.db.WithContext(ctx).
			Raw(`SELECT count(*) FROM vendors WHERE id = ? AND organization_id = ? AND deleted_at IS NULL`, entityID, orgID).
			Scan(&count).Error
	default:
		// Unknown entity type — accept without verification
		// (allows future entity types without code changes)
		return nil
	}

	if err != nil {
		return fmt.Errorf("government: verify entity ownership: %w", err)
	}
	if count == 0 {
		return ErrCandidateNotFound
	}
	return nil
}

// datasetToEntityType maps a DatasetType constant to its expected internal entity type.
// Returns empty string for dataset types that have no fixed entity type.
func datasetToEntityType(datasetType string) string {
	switch datasetType {
	case DatasetProjects:
		return "project"
	case DatasetVendors:
		return "vendor"
	default:
		return ""
	}
}

// normaliseName lower-cases and trims a name for fuzzy comparison.
func normaliseName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	// Collapse multiple spaces
	var b strings.Builder
	prevSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !prevSpace {
				b.WriteRune(' ')
			}
			prevSpace = true
		} else {
			b.WriteRune(r)
			prevSpace = false
		}
	}
	return b.String()
}

// normaliseNPWP strips all non-digit characters from a tax ID string
// so that "01.234.567.8-910.000" == "012345678910000".
func normaliseNPWP(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// payloadString looks up a string value from a JSON payload map by trying
// multiple candidate field names in order.
func payloadString(payload map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := payload[k]; ok {
			if s, ok := v.(string); ok {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}
