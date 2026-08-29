package bim

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/audit"
	"gorm.io/gorm"
)

// projectBelongsToOrg verifies that a project exists, belongs to the given
// organization, and has not been soft-deleted.
func (s *Service) projectBelongsToOrg(orgID, projectID uuid.UUID) (bool, error) {
	var count int64
	err := s.db.Raw(
		`SELECT COUNT(*) FROM projects
		 WHERE id = ? AND organization_id = ? AND deleted_at IS NULL`,
		projectID, orgID,
	).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Service handles business logic for BIM/digital twin integration.
type Service struct {
	db          *gorm.DB
	auditWriter *audit.Writer
}

// NewService constructs a Service with the given database handle and audit writer.
func NewService(db *gorm.DB, auditWriter *audit.Writer) *Service {
	return &Service{db: db, auditWriter: auditWriter}
}

// ---------------------------------------------------------------------------
// BIM Model CRUD
// ---------------------------------------------------------------------------

// ListModels returns all non-deleted BIM models for an organization.
func (s *Service) ListModels(orgID uuid.UUID) ([]BIMModel, int64, error) {
	var models []BIMModel
	var total int64

	q := s.db.Model(&BIMModel{}).Where("organization_id = ? AND deleted_at IS NULL", orgID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}
	return models, total, nil
}

// GetModel returns a single BIM model by ID, scoped to the organization.
func (s *Service) GetModel(orgID, modelID uuid.UUID) (*BIMModel, error) {
	var m BIMModel
	err := s.db.
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL", modelID, orgID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &m, err
}

// CreateModel registers a new BIM model for the organization.
func (s *Service) CreateModel(orgID, actorID uuid.UUID, req CreateBIMModelRequest) (*BIMModel, error) {
	meta, err := marshalMetadata(req.Metadata)
	if err != nil {
		return nil, fmt.Errorf("invalid metadata: %w", err)
	}

	m := BIMModel{
		ID:              uuid.New(),
		OrganizationID:  orgID,
		Name:            req.Name,
		Description:     req.Description,
		Discipline:      req.Discipline,
		Provider:        req.Provider,
		ExternalModelID: req.ExternalModelID,
		ViewerURL:       req.ViewerURL,
		Status:          ModelStatusDraft,
		Metadata:        meta,
		CreatedBy:       actorID,
	}

	if err := s.db.Create(&m).Error; err != nil {
		return nil, err
	}

	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		Action:         "bim.model.created",
		EntityType:     "bim_model",
		EntityID:       m.ID.String(),
		EntityLabel:    m.Name,
		NewValues:      map[string]interface{}{"discipline": m.Discipline, "provider": m.Provider, "status": m.Status},
	})

	return &m, nil
}

// UpdateModel patches a BIM model's mutable fields.
func (s *Service) UpdateModel(orgID, actorID, modelID uuid.UUID, req UpdateBIMModelRequest) (*BIMModel, error) {
	m, err := s.GetModel(orgID, modelID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, nil
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
		m.Name = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Discipline != nil {
		updates["discipline"] = *req.Discipline
	}
	if req.ViewerURL != nil {
		updates["viewer_url"] = *req.ViewerURL
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Metadata != nil {
		meta, merr := marshalMetadata(*req.Metadata)
		if merr != nil {
			return nil, fmt.Errorf("invalid metadata: %w", merr)
		}
		updates["metadata"] = meta
	}

	if len(updates) == 0 {
		return m, nil
	}

	if err := s.db.Model(m).Updates(updates).Error; err != nil {
		return nil, err
	}

	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		Action:         "bim.model.updated",
		EntityType:     "bim_model",
		EntityID:       m.ID.String(),
		EntityLabel:    m.Name,
		NewValues:      updates,
	})

	return m, nil
}

// DeleteModel soft-deletes a BIM model (sets deleted_at).
func (s *Service) DeleteModel(orgID, actorID, modelID uuid.UUID) error {
	m, err := s.GetModel(orgID, modelID)
	if err != nil {
		return err
	}
	if m == nil {
		return nil
	}

	now := time.Now()
	if err := s.db.Model(m).Update("deleted_at", now).Error; err != nil {
		return err
	}

	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		Action:         "bim.model.deleted",
		EntityType:     "bim_model",
		EntityID:       m.ID.String(),
		EntityLabel:    m.Name,
		NewValues:      map[string]interface{}{"deleted_at": now},
	})

	return nil
}

// ---------------------------------------------------------------------------
// BIM Model Versions
// ---------------------------------------------------------------------------

// ListVersions returns all versions for a BIM model, newest first.
func (s *Service) ListVersions(orgID, modelID uuid.UUID) ([]BIMModelVersion, error) {
	// Ensure the model belongs to the org.
	m, err := s.GetModel(orgID, modelID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, nil
	}

	var versions []BIMModelVersion
	err = s.db.
		Where("bim_model_id = ? AND organization_id = ?", modelID, orgID).
		Order("created_at DESC").
		Find(&versions).Error
	return versions, err
}

// AddVersion appends an immutable version record to a BIM model.
func (s *Service) AddVersion(orgID, actorID, modelID uuid.UUID, req CreateVersionRequest) (*BIMModelVersion, error) {
	m, err := s.GetModel(orgID, modelID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, nil
	}

	v := BIMModelVersion{
		ID:                uuid.New(),
		BIMModelID:        modelID,
		OrganizationID:    orgID,
		VersionLabel:      req.VersionLabel,
		ExternalVersionID: req.ExternalVersionID,
		ChangeSummary:     req.ChangeSummary,
		FileSizeBytes:     req.FileSizeBytes,
		Checksum:          req.Checksum,
		CreatedBy:         actorID,
	}

	if err := s.db.Create(&v).Error; err != nil {
		return nil, err
	}

	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		Action:         "bim.version.created",
		EntityType:     "bim_model_version",
		EntityID:       v.ID.String(),
		EntityLabel:    fmt.Sprintf("%s / %s", m.Name, v.VersionLabel),
		NewValues:      map[string]interface{}{"version_label": v.VersionLabel, "checksum": v.Checksum},
	})

	return &v, nil
}

// ---------------------------------------------------------------------------
// BIM Project Mappings
// ---------------------------------------------------------------------------

// ListMappings returns all project mappings for a BIM model.
func (s *Service) ListMappings(orgID, modelID uuid.UUID) ([]BIMProjectMapping, error) {
	m, err := s.GetModel(orgID, modelID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, nil
	}

	var mappings []BIMProjectMapping
	err = s.db.
		Where("bim_model_id = ? AND organization_id = ?", modelID, orgID).
		Find(&mappings).Error
	return mappings, err
}

// ErrProjectNotFound is returned when the requested project does not exist
// within the caller's organization (or has been soft-deleted).
var ErrProjectNotFound = errors.New("project not found or not accessible")

// LinkProject creates a mapping between a BIM model and a project.
// Returns an error if the mapping already exists (unique constraint).
// Returns ErrProjectNotFound if the project does not exist or is cross-tenant.
func (s *Service) LinkProject(orgID, actorID, modelID uuid.UUID, req LinkProjectRequest) (*BIMProjectMapping, error) {
	m, err := s.GetModel(orgID, modelID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, nil
	}

	// Tenant safety: verify the project exists, belongs to this org, and is not deleted.
	belongs, err := s.projectBelongsToOrg(orgID, req.ProjectID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, ErrProjectNotFound
	}

	mapping := BIMProjectMapping{
		ID:             uuid.New(),
		OrganizationID: orgID,
		BIMModelID:     modelID,
		ProjectID:      req.ProjectID,
		ModelRole:      req.ModelRole,
		Notes:          req.Notes,
		LinkedBy:       actorID,
		LinkedAt:       time.Now(),
	}

	if err := s.db.Create(&mapping).Error; err != nil {
		return nil, err
	}

	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		Action:         "bim.project.linked",
		EntityType:     "bim_project_mapping",
		EntityID:       mapping.ID.String(),
		EntityLabel:    fmt.Sprintf("%s → %s", m.Name, req.ProjectID),
		NewValues:      map[string]interface{}{"project_id": req.ProjectID, "model_role": req.ModelRole},
	})

	return &mapping, nil
}

// UnlinkProject removes a project mapping from a BIM model.
func (s *Service) UnlinkProject(orgID, actorID, modelID, projectID uuid.UUID) error {
	result := s.db.
		Where("bim_model_id = ? AND project_id = ? AND organization_id = ?", modelID, projectID, orgID).
		Delete(&BIMProjectMapping{})
	if result.Error != nil {
		return result.Error
	}

	s.auditWriter.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		Action:         "bim.project.unlinked",
		EntityType:     "bim_project_mapping",
		EntityID:       modelID.String(),
		EntityLabel:    fmt.Sprintf("model %s → project %s", modelID, projectID),
		NewValues:      map[string]interface{}{"project_id": projectID},
	})

	return nil
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func marshalMetadata(m map[string]interface{}) ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}
