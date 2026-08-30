package project

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/audit"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"github.com/harmanto-49/cankora/internal/shared/types"
)

// RegisterDocumentRoutes mounts document routes under /projects/:id/documents.
func (h *Handler) RegisterDocumentRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.ListDocuments)
	rg.POST("", h.UploadDocument)
	rg.GET("/:documentID", h.GetDocument)
	rg.GET("/:documentID/download", h.DownloadDocument)
	rg.PUT("/:documentID", h.UpdateDocument)
	rg.DELETE("/:documentID", h.DeleteDocument)
}

// ListDocuments godoc
// GET /api/v1/projects/:id/documents
func (h *Handler) ListDocuments(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	page, pageSize := parsePagination(c)

	filter := DocumentListFilter{
		OrganizationID: claims.OrganizationID,
		ProjectID:      projectID,
		Category:       c.Query("category"),
		Search:         c.Query("search"),
		Page:           page,
		PageSize:       pageSize,
	}

	docs, total, err := h.svc.ListDocuments(c.Request.Context(), filter)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OKPaginated(c, "documents retrieved", docs, types.NewPaginationMeta(page, pageSize, total))
}

// UploadDocument godoc
// POST /api/v1/projects/:id/documents (multipart/form-data)
// Fields: file (required), name (optional), category (optional), version (optional)
func (h *Handler) UploadDocument(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	var form UploadDocumentRequest
	if err := c.ShouldBind(&form); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "no file uploaded")
		return
	}

	// Size guard on the multipart header.
	if fileHeader.Size > h.svc.maxFileSize {
		response.BadRequest(c, "file exceeds maximum allowed size")
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		response.InternalError(c)
		return
	}
	defer src.Close()

	data, err := io.ReadAll(io.LimitReader(src, h.svc.maxFileSize+1))
	if err != nil {
		response.InternalError(c)
		return
	}
	if int64(len(data)) > h.svc.maxFileSize {
		response.BadRequest(c, "file exceeds maximum allowed size")
		return
	}

	// Validate extension allowlist + sniffed MIME (magic bytes).
	originalName := fileHeader.Filename
	mime, err := validateDocumentFile(data, originalName)
	if err != nil {
		if errors.Is(err, ErrInvalidFileType) {
			response.BadRequest(c, "file type is not allowed")
			return
		}
		response.InternalError(c)
		return
	}

	name := strings.TrimSpace(form.Name)
	if name == "" {
		name = sanitizeFilename(originalName)
	}

	doc, err := h.svc.UploadDocument(c.Request.Context(), projectID, claims.OrganizationID, claims.UserID,
		name, form.Category, form.Version, originalName, mime, int64(len(data)), data)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrFileTooLarge) {
			response.BadRequest(c, "file exceeds maximum allowed size")
			return
		}
		if errors.Is(err, ErrDocumentStorage) {
			response.InternalError(c)
			return
		}
		response.InternalError(c)
		return
	}

	h.recordDocumentAudit(c, "document.uploaded", doc)
	response.Created(c, "document uploaded", doc)
}

// GetDocument godoc
// GET /api/v1/projects/:id/documents/:documentID
func (h *Handler) GetDocument(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	documentID, err := uuid.Parse(c.Param("documentID"))
	if err != nil {
		response.BadRequest(c, "invalid document id")
		return
	}

	doc, err := h.svc.GetDocument(c.Request.Context(), projectID, documentID, claims.OrganizationID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) || errors.Is(err, ErrDocumentNotFound) {
			response.NotFound(c, "document not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "ok", doc)
}

// DownloadDocument godoc
// GET /api/v1/projects/:id/documents/:documentID/download
// Streams the physical file with safe Content-Type and Content-Disposition.
func (h *Handler) DownloadDocument(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	documentID, err := uuid.Parse(c.Param("documentID"))
	if err != nil {
		response.BadRequest(c, "invalid document id")
		return
	}

	doc, absPath, err := h.svc.OpenDocument(c.Request.Context(), projectID, documentID, claims.OrganizationID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) || errors.Is(err, ErrDocumentNotFound) {
			response.NotFound(c, "document not found")
			return
		}
		response.InternalError(c)
		return
	}

	// Content-Type from stored metadata (fall back to octet-stream).
	contentType := doc.MimeType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Safe filename for Content-Disposition (RFC 5987 UTF-8).
	fileName := sanitizeFilename(doc.Name)
	if fileName == "" {
		fileName = "document"
	}
	quoted := strings.ReplaceAll(fileName, `"`, `'`)
	encoded := url.PathEscape(fileName)

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", quoted, encoded))
	c.Header("X-Content-Type-Options", "nosniff")
	c.File(absPath)

	h.recordDocumentAudit(c, "document.downloaded", doc)
}

// UpdateDocument godoc
// PUT /api/v1/projects/:id/documents/:documentID
func (h *Handler) UpdateDocument(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	documentID, err := uuid.Parse(c.Param("documentID"))
	if err != nil {
		response.BadRequest(c, "invalid document id")
		return
	}

	var req UpdateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	doc, err := h.svc.UpdateDocument(c.Request.Context(), projectID, documentID, claims.OrganizationID, &req)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) || errors.Is(err, ErrDocumentNotFound) {
			response.NotFound(c, "document not found")
			return
		}
		response.InternalError(c)
		return
	}

	h.recordDocumentAudit(c, "document.updated", doc)
	response.OK(c, "ok", doc)
}

// DeleteDocument godoc
// DELETE /api/v1/projects/:id/documents/:documentID
func (h *Handler) DeleteDocument(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	documentID, err := uuid.Parse(c.Param("documentID"))
	if err != nil {
		response.BadRequest(c, "invalid document id")
		return
	}

	doc, getErr := h.svc.GetDocument(c.Request.Context(), projectID, documentID, claims.OrganizationID)
	if getErr != nil {
		if errors.Is(getErr, ErrProjectNotFound) || errors.Is(getErr, ErrDocumentNotFound) {
			response.NotFound(c, "document not found")
			return
		}
		response.InternalError(c)
		return
	}

	if err := h.svc.DeleteDocument(c.Request.Context(), projectID, documentID, claims.OrganizationID); err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			response.NotFound(c, "document not found")
			return
		}
		response.InternalError(c)
		return
	}

	h.recordDocumentAudit(c, "document.deleted", doc)
	response.NoContent(c)
}

// recordDocumentAudit writes an asynchronous audit entry for document events.
func (h *Handler) recordDocumentAudit(c *gin.Context, action string, doc *ProjectDocument) {
	if h.audit == nil || doc == nil {
		return
	}
	claims := claimsFromGin(c)
	actorID := uuid.Nil
	orgID := uuid.Nil
	if claims != nil {
		actorID = claims.UserID
		// ProjectDocument has no organization_id column; tenant boundary is
		// proven by parent project ownership, which the service verifies before
		// this point. claims.OrganizationID is therefore the correct audit scope.
		orgID = claims.OrganizationID
	}
	h.audit.Record(audit.WriteRequest{
		OrganizationID: orgID,
		ActorID:        &actorID,
		ActorEmail:     actorEmailFromClaims(claims),
		Action:         action,
		EntityType:     "document",
		EntityID:       doc.ID.String(),
		EntityLabel:    doc.Name,
		IPAddress:      c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		RequestID:      c.GetString("X-Request-ID"),
	})
}
