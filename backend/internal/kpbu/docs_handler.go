package kpbu

import (
	"fmt"
	notification "github.com/harmanto-49/cankora/internal/notifications/delivery"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"github.com/harmanto-49/cankora/internal/shared/types"
)

// ListRequirements GET /kpbu/document-requirements?include_inactive=true
func (h *Handler) ListRequirements(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	list, err := h.svc.ListRequirements(claims.OrganizationID, c.Query("include_inactive") == "true")
	if err != nil {
		h.writeErr(c, err, "")
		return
	}
	response.OK(c, "KPBU document requirements retrieved", list)
}

// CreateRequirement POST /kpbu/document-requirements
func (h *Handler) CreateRequirement(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	var req RequirementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	r, err := h.svc.CreateRequirement(claims.OrganizationID, req)
	if err != nil {
		h.writeErr(c, err, "")
		return
	}
	h.record(c, claims, "kpbu_document_requirement.created", "kpbu_document_requirement", r.ID, r.Code+" — "+r.Name, r)
	response.Created(c, "KPBU document requirement created", r)
}

// UpdateRequirement PUT /kpbu/document-requirements/:reqID
func (h *Handler) UpdateRequirement(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "reqID")
	if !ok {
		return
	}
	var req RequirementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	r, err := h.svc.UpdateRequirement(claims.OrganizationID, id, req)
	if err != nil {
		h.writeErr(c, err, "document requirement not found")
		return
	}
	h.record(c, claims, "kpbu_document_requirement.updated", "kpbu_document_requirement", r.ID, r.Code+" — "+r.Name, r)
	response.OK(c, "KPBU document requirement updated", r)
}

// DeleteRequirement DELETE /kpbu/document-requirements/:reqID
func (h *Handler) DeleteRequirement(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "reqID")
	if !ok {
		return
	}
	r, err := h.svc.DeleteRequirement(claims.OrganizationID, id)
	if err != nil {
		h.writeErr(c, err, "document requirement not found")
		return
	}
	h.record(c, claims, "kpbu_document_requirement.deleted", "kpbu_document_requirement", r.ID, r.Code+" — "+r.Name, nil)
	response.NoContent(c)
}

// GetChecklist GET /kpbu/projects/:id/checklist
func (h *Handler) GetChecklist(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	cl, err := h.svc.Checklist(claims.OrganizationID, id)
	if err != nil {
		h.writeErr(c, err, "KPBU project not found")
		return
	}
	response.OK(c, "KPBU document checklist retrieved", cl)
}

// UploadDocument POST /kpbu/projects/:id/documents (multipart)
func (h *Handler) UploadDocument(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	fh, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}
	if fh.Size > h.svc.maxSize {
		response.BadRequest(c, fmt.Sprintf("file exceeds maximum size of %d MB", h.svc.maxSize>>20))
		return
	}
	f, err := fh.Open()
	if err != nil {
		response.BadRequest(c, "cannot read file")
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, h.svc.maxSize+1))
	if err != nil {
		response.BadRequest(c, "cannot read file")
		return
	}
	in := DocumentUpload{
		Filename: fh.Filename, Data: data, Name: c.PostForm("name"),
		DocumentNumber: c.PostForm("document_number"), Notes: c.PostForm("notes"),
	}
	if v := strings.TrimSpace(c.PostForm("requirement_id")); v != "" {
		rid, err := uuid.Parse(v)
		if err != nil {
			response.BadRequest(c, "invalid requirement_id")
			return
		}
		in.RequirementID = &rid
	}
	if v := strings.TrimSpace(c.PostForm("document_date")); v != "" {
		var ft types.FlexTime
		if err := ft.UnmarshalJSON([]byte(`"` + v + `"`)); err != nil {
			response.BadRequest(c, "invalid document_date")
			return
		}
		in.DocumentDate = &ft
	}
	d, err := h.svc.UploadDocument(claims.OrganizationID, claims.UserID, id, in)
	if err != nil {
		h.writeErr(c, err, "KPBU project not found")
		return
	}
	h.record(c, claims, "kpbu_document.uploaded", "kpbu_document", d.ID, d.Name, d)
	if isPortal(c) {
		h.svc.notify(claims.OrganizationID, h.svc.internalReviewers(claims.OrganizationID), notification.PriorityNormal,
			"Dokumen badan usaha KPBU menunggu verifikasi",
			fmt.Sprintf("Badan usaha mengunggah dokumen \"%s\" melalui portal. Periksa di checklist dokumen proyek KPBU.", d.Name),
			"kpbu_document", d.ID)
	}
	response.Created(c, "KPBU document uploaded", d)
}

// ReviewDocument POST /kpbu/projects/:id/documents/:docID/review
func (h *Handler) ReviewDocument(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	docID, ok := parseID(c, "docID")
	if !ok {
		return
	}
	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	d, err := h.svc.ReviewDocument(claims.OrganizationID, claims.UserID, id, docID, req)
	if err != nil {
		h.writeErr(c, err, "KPBU document not found")
		return
	}
	h.record(c, claims, "kpbu_document."+strings.ToLower(d.Status), "kpbu_document", d.ID, d.Name,
		gin.H{"status": d.Status, "review_notes": d.ReviewNotes})
	if d.UploadedBy != claims.UserID {
		subject, prio, body := "Dokumen KPBU disetujui", notification.PriorityNormal, fmt.Sprintf("Dokumen \"%s\" telah disetujui.", d.Name)
		if d.Status == DocRejected {
			subject, prio = "Dokumen KPBU ditolak", notification.PriorityHigh
			body = fmt.Sprintf("Dokumen \"%s\" ditolak. Alasan: %s. Unggah ulang dokumen yang diperbaiki.", d.Name, d.ReviewNotes)
		}
		h.svc.notify(claims.OrganizationID, []uuid.UUID{d.UploadedBy}, prio, subject, body, "kpbu_document", d.ID)
	}
	response.OK(c, "KPBU document reviewed", d)
}

// DownloadDocument GET /kpbu/projects/:id/documents/:docID/download
func (h *Handler) DownloadDocument(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	docID, ok := parseID(c, "docID")
	if !ok {
		return
	}
	d, path, err := h.svc.DocumentFile(claims.OrganizationID, id, docID)
	if err != nil {
		h.writeErr(c, err, "KPBU document not found")
		return
	}
	h.record(c, claims, "kpbu_document.downloaded", "kpbu_document", d.ID, d.Name, nil)
	name := filepath.Base(path)
	ct := d.MimeType
	if ct == "" {
		ct = "application/octet-stream"
	}
	c.Header("Content-Type", ct)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s",
		strings.ReplaceAll(name, `"`, `'`), url.PathEscape(name)))
	c.Header("X-Content-Type-Options", "nosniff")
	c.File(path)
}

// DeleteDocument DELETE /kpbu/projects/:id/documents/:docID
func (h *Handler) DeleteDocument(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	docID, ok := parseID(c, "docID")
	if !ok {
		return
	}
	d, err := h.svc.DeleteDocument(claims.OrganizationID, id, docID)
	if err != nil {
		h.writeErr(c, err, "KPBU document not found")
		return
	}
	h.record(c, claims, "kpbu_document.deleted", "kpbu_document", d.ID, d.Name, nil)
	response.NoContent(c)
}

// ListGuarantees GET /kpbu/projects/:id/guarantees
func (h *Handler) ListGuarantees(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	list, err := h.svc.ListGuarantees(claims.OrganizationID, id)
	if err != nil {
		h.writeErr(c, err, "KPBU project not found")
		return
	}
	response.OK(c, "KPBU guarantees retrieved", list)
}

// CreateGuarantee POST /kpbu/projects/:id/guarantees
func (h *Handler) CreateGuarantee(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req GuaranteeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	g, err := h.svc.CreateGuarantee(claims.OrganizationID, id, req)
	if err != nil {
		h.writeErr(c, err, "KPBU project not found")
		return
	}
	h.record(c, claims, "kpbu_guarantee.created", "kpbu_guarantee", g.ID, g.Type+" "+g.GuaranteeNumber, g)
	response.Created(c, "KPBU guarantee created", g)
}

// UpdateGuarantee PUT /kpbu/projects/:id/guarantees/:guaranteeID
func (h *Handler) UpdateGuarantee(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	gid, ok := parseID(c, "guaranteeID")
	if !ok {
		return
	}
	var req GuaranteeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	g, err := h.svc.UpdateGuarantee(claims.OrganizationID, id, gid, req)
	if err != nil {
		h.writeErr(c, err, "KPBU guarantee not found")
		return
	}
	h.record(c, claims, "kpbu_guarantee.updated", "kpbu_guarantee", g.ID, g.Type+" "+g.GuaranteeNumber, g)
	response.OK(c, "KPBU guarantee updated", g)
}

// DeleteGuarantee DELETE /kpbu/projects/:id/guarantees/:guaranteeID
func (h *Handler) DeleteGuarantee(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	gid, ok := parseID(c, "guaranteeID")
	if !ok {
		return
	}
	g, err := h.svc.DeleteGuarantee(claims.OrganizationID, id, gid)
	if err != nil {
		h.writeErr(c, err, "KPBU guarantee not found")
		return
	}
	h.record(c, claims, "kpbu_guarantee.deleted", "kpbu_guarantee", g.ID, g.Type+" "+g.GuaranteeNumber, nil)
	response.NoContent(c)
}
