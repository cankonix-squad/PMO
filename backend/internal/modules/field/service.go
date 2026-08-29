package field

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/audit"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrNotFound = errors.New("field inspection not found")
var ErrStorage = errors.New("field evidence storage error")
var ErrInvalidEvidence = errors.New("evidence file type is not allowed")

type Service struct {
	db      *gorm.DB
	audit   *audit.Writer
	root    string
	maxSize int64
}

func NewService(db *gorm.DB, writer *audit.Writer, root string, maxSize int64) *Service {
	if root == "" {
		root = "storage/documents"
	}
	if maxSize <= 0 {
		maxSize = 20 * 1024 * 1024
	}
	return &Service{db: db, audit: writer, root: root, maxSize: maxSize}
}

func (s *Service) projectExists(orgID, projectID uuid.UUID) bool {
	var count int64
	return s.db.Table("projects").Where("id = ? AND organization_id = ? AND deleted_at IS NULL", projectID, orgID).Count(&count).Error == nil && count == 1
}
func (s *Service) List(orgID, projectID uuid.UUID) ([]Inspection, error) {
	if !s.projectExists(orgID, projectID) {
		return nil, ErrNotFound
	}
	var list []Inspection
	err := s.db.Where("organization_id = ? AND project_id = ? AND deleted_at IS NULL", orgID, projectID).Preload("Evidence", "deleted_at IS NULL").Order("inspected_at DESC").Find(&list).Error
	return list, err
}
func (s *Service) Get(orgID, projectID, id uuid.UUID) (*Inspection, error) {
	var item Inspection
	err := s.db.Where("organization_id = ? AND project_id = ? AND id = ? AND deleted_at IS NULL", orgID, projectID, id).Preload("Evidence", "deleted_at IS NULL").First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &item, err
}
func (s *Service) Create(orgID, actorID, projectID uuid.UUID, req CreateInspectionRequest, file *multipart.FileHeader) (*Inspection, error) {
	if !s.projectExists(orgID, projectID) {
		return nil, ErrNotFound
	}
	item := &Inspection{OrganizationID: orgID, ProjectID: projectID, InspectedAt: req.InspectedAt, Latitude: req.Latitude, Longitude: req.Longitude, InspectorID: actorID, Notes: req.Notes, VerificationStatus: "PENDING"}
	if err := s.db.Create(item).Error; err != nil {
		return nil, err
	}
	if file != nil {
		if err := s.addEvidence(orgID, actorID, projectID, item, file, req.Latitude, req.Longitude); err != nil {
			_ = s.db.Delete(item).Error
			return nil, err
		}
	}
	s.auditRecord(orgID, actorID, item.ID, "field_inspection.created")
	return s.Get(orgID, projectID, item.ID)
}
func (s *Service) addEvidence(orgID, actorID, projectID uuid.UUID, inspection *Inspection, file *multipart.FileHeader, lat, lon *float64) error {
	if file.Size > s.maxSize {
		return errors.New("evidence file exceeds maximum allowed size")
	}
	src, err := file.Open()
	if err != nil {
		return ErrStorage
	}
	defer src.Close()
	data, err := io.ReadAll(io.LimitReader(src, s.maxSize+1))
	if err != nil || int64(len(data)) > s.maxSize {
		return errors.New("evidence file exceeds maximum allowed size")
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".pdf": true, ".txt": true}
	if !allowed[ext] {
		return ErrInvalidEvidence
	}
	mime := http.DetectContentType(data)
	checksum := fmt.Sprintf("%x", sha256.Sum256(data))
	evidenceID := uuid.New()
	key := filepath.ToSlash(filepath.Join("field", orgID.String(), projectID.String(), inspection.ID.String(), evidenceID.String()+ext))
	full := filepath.Join(s.root, key)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return ErrStorage
	}
	if err := os.WriteFile(full, data, 0644); err != nil {
		return ErrStorage
	}
	evidence := &Evidence{ID: evidenceID, OrganizationID: orgID, ProjectID: projectID, InspectionID: inspection.ID, FileName: file.Filename, StorageKey: key, MimeType: mime, FileSize: int64(len(data)), ChecksumSHA256: checksum, Latitude: lat, Longitude: lon, VerificationStatus: "PENDING"}
	if err := s.db.Create(evidence).Error; err != nil {
		_ = os.Remove(full)
		return err
	}
	s.auditRecord(orgID, actorID, evidence.ID, "field_evidence.uploaded")
	return nil
}

// AddEvidence uploads a new evidence file to an existing inspection.
func (s *Service) AddEvidence(orgID, actorID, projectID, inspectionID uuid.UUID, file *multipart.FileHeader, lat, lon *float64) (*Inspection, error) {
	inspection, err := s.Get(orgID, projectID, inspectionID)
	if err != nil {
		return nil, err
	}
	if err := s.addEvidence(orgID, actorID, projectID, inspection, file, lat, lon); err != nil {
		return nil, err
	}
	s.auditRecord(orgID, actorID, inspectionID, "field_inspection.evidence_added")
	return s.Get(orgID, projectID, inspectionID)
}

func (s *Service) Verify(orgID, actorID, projectID, id uuid.UUID, req VerifyRequest) (*Inspection, error) {
	item, err := s.Get(orgID, projectID, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if err := s.db.Model(item).Updates(map[string]interface{}{"verification_status": req.Status, "verified_by": actorID, "verified_at": now}).Error; err != nil {
		return nil, err
	}
	s.auditRecord(orgID, actorID, item.ID, "field_inspection."+strings.ToLower(req.Status))
	return s.Get(orgID, projectID, id)
}
func (s *Service) OpenEvidence(orgID, projectID, inspectionID, evidenceID uuid.UUID) (*Evidence, string, error) {
	var item Evidence
	err := s.db.Where("organization_id = ? AND project_id = ? AND inspection_id = ? AND id = ? AND deleted_at IS NULL", orgID, projectID, inspectionID, evidenceID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", ErrNotFound
	}
	if err != nil {
		return nil, "", err
	}
	full := filepath.Join(s.root, item.StorageKey)
	if _, err = os.Stat(full); err != nil {
		return nil, "", ErrNotFound
	}
	return &item, full, nil
}
func (s *Service) Delete(orgID, actorID, projectID, id uuid.UUID) error {
	item, err := s.Get(orgID, projectID, id)
	if err != nil {
		return err
	}
	for _, e := range item.Evidence {
		_ = os.Remove(filepath.Join(s.root, e.StorageKey))
		_ = s.db.Delete(&e).Error
	}
	if err = s.db.Delete(item).Error; err == nil {
		s.auditRecord(orgID, actorID, id, "field_inspection.deleted")
	}
	return err
}
func (s *Service) auditRecord(orgID, actorID, id uuid.UUID, action string) {
	if s.audit != nil {
		s.audit.Record(audit.WriteRequest{OrganizationID: orgID, ActorID: &actorID, Action: action, EntityType: "field_inspection", EntityID: id.String()})
	}
}
