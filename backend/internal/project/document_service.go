package project

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ConfigureDocumentStorage sets the local storage root and max upload size
// (bytes). It is called once during wiring; defaults remain until then.
func (s *Service) ConfigureDocumentStorage(root string, maxBytes int64) {
	if root != "" {
		s.storageRoot = root
	}
	if maxBytes > 0 {
		s.maxFileSize = maxBytes
	}
}

// UploadDocument persists a document's file to storage and its metadata to the
// database, verifying the parent project's org ownership. The file content is
// already validated (size + MIME + extension) by the handler; this method
// writes the bytes and stores a relative storage key in FileURL.
func (s *Service) UploadDocument(ctx context.Context, projectID, orgID, uploadedBy uuid.UUID, name, category, version, safeName, mime string, size int64, data []byte) (*ProjectDocument, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	if size > s.maxFileSize {
		return nil, ErrFileTooLarge
	}

	d := &ProjectDocument{
		ID:         uuid.New(),
		ProjectID:  projectID,
		Name:       truncateDocumentTitle(sanitizeFilename(name), maxDocumentNameLen),
		Category:   strings.TrimSpace(category),
		Version:    strings.TrimSpace(version),
		FileSize:   size,
		MimeType:   mime,
		UploadedBy: uploadedBy,
	}
	safeName = sanitizeFilename(safeName)
	key := buildDocumentStorageKey(orgID.String(), projectID.String(), d.ID.String(), safeName)
	if err := saveDocumentFile(s.storageRoot, key, data); err != nil {
		s.log.Error("document: save file", zap.Error(err))
		return nil, ErrDocumentStorage
	}
	d.FileURL = key
	if err := s.repo.CreateDocument(ctx, d); err != nil {
		return nil, fmt.Errorf("document: create metadata: %w", err)
	}
	created, err := s.repo.FindDocumentByID(ctx, projectID, d.ID, orgID)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// ListDocuments returns paginated documents for a project, verifying ownership.
func (s *Service) ListDocuments(ctx context.Context, filter DocumentListFilter) ([]ProjectDocument, int64, error) {
	if _, err := s.GetByID(ctx, filter.ProjectID, filter.OrganizationID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListDocuments(ctx, filter)
}

// GetDocument returns document metadata, verifying project ownership.
func (s *Service) GetDocument(ctx context.Context, projectID, documentID, orgID uuid.UUID) (*ProjectDocument, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	return s.repo.FindDocumentByID(ctx, projectID, documentID, orgID)
}

// UpdateDocument updates document metadata, verifying project ownership.
func (s *Service) UpdateDocument(ctx context.Context, projectID, documentID, orgID uuid.UUID, req *UpdateDocumentRequest) (*ProjectDocument, error) {
	d, err := s.GetDocument(ctx, projectID, documentID, orgID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Name) != "" {
		d.Name = truncateDocumentTitle(sanitizeFilename(req.Name), maxDocumentNameLen)
	}
	if req.Category != "" {
		d.Category = strings.TrimSpace(req.Category)
	}
	if req.Version != "" {
		d.Version = strings.TrimSpace(req.Version)
	}
	if err := s.repo.UpdateDocument(ctx, d); err != nil {
		return nil, fmt.Errorf("document: update metadata: %w", err)
	}
	updated, err := s.repo.FindDocumentByID(ctx, projectID, documentID, orgID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// DeleteDocument soft-deletes a document's metadata, verifying project
// ownership. The physical file is intentionally retained for audit/dev
// retention.
func (s *Service) DeleteDocument(ctx context.Context, projectID, documentID, orgID uuid.UUID) error {
	doc, err := s.GetDocument(ctx, projectID, documentID, orgID)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteDocument(ctx, projectID, documentID, orgID); err != nil {
		return err
	}
	// Remove the physical file. A soft-deleted row can never be downloaded
	// again, so this is safe. Failure to remove is logged but not fatal.
	if doc.FileURL != "" {
		if derr := deleteDocumentFile(s.storageRoot, doc.FileURL); derr != nil {
			s.log.Warn("document: delete physical file", zap.String("document_id", documentID.String()), zap.Error(derr))
		}
	}
	return nil
}

// OpenDocument returns the absolute file path + safe display filename for a
// document, verifying project ownership. Used by the download handler.
func (s *Service) OpenDocument(ctx context.Context, projectID, documentID, orgID uuid.UUID) (doc *ProjectDocument, absPath string, err error) {
	doc, err = s.GetDocument(ctx, projectID, documentID, orgID)
	if err != nil {
		return nil, "", err
	}
	absPath, _, oerr := openDocumentFile(s.storageRoot, doc.FileURL)
	if oerr != nil {
		return nil, "", ErrDocumentStorage
	}
	return doc, absPath, nil
}

// truncateDocumentTitle caps a display name to maxLen runes safely.
func truncateDocumentTitle(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
