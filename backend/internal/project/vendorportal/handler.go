package vendorportal

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/governance/audit"
	"github.com/harmanto-49/cankora/internal/identity/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler serves the contractor portal, the Balai review queue, and contractor
// account management.
type Handler struct {
	svc   *Service
	audit *audit.Writer
	log   *zap.Logger
}

func NewHandler(db *gorm.DB, storageRoot string, maxSize int64, auditWriter *audit.Writer, log *zap.Logger, notifier Notifier) *Handler {
	svc := NewService(db, storageRoot, maxSize)
	if notifier != nil {
		svc.SetNotifier(notifier, log)
	}
	return &Handler{svc: svc, audit: auditWriter, log: log}
}

func claimsFromGin(c *gin.Context) *auth.Claims {
	val, ok := c.Get(string(auth.ContextKeyClaims))
	if !ok {
		return nil
	}
	claims, _ := val.(*auth.Claims)
	return claims
}

func (h *Handler) record(c *gin.Context, claims *auth.Claims, action, entityType string, id uuid.UUID, label string, values interface{}) {
	if h.audit == nil {
		return
	}
	actor := claims.UserID
	h.audit.Record(audit.WriteRequest{
		OrganizationID: claims.OrganizationID,
		ActorID:        &actor,
		ActorEmail:     claims.Email,
		Action:         action,
		EntityType:     entityType,
		EntityID:       id.String(),
		EntityLabel:    label,
		NewValues:      values,
		IPAddress:      c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		RequestID:      c.GetString("X-Request-ID"),
	})
}

func (h *Handler) writeErr(c *gin.Context, err error, notFound string) {
	var ve *ValidationError
	switch {
	case errors.Is(err, ErrNotVendor):
		response.Forbidden(c, "this account is not linked to a contractor")
	case errors.Is(err, ErrNotFound):
		response.NotFound(c, notFound)
	case errors.Is(err, ErrDuplicate):
		response.Conflict(c, "a progress report for this period is already submitted or verified")
	case errors.Is(err, ErrAlreadyDecided):
		response.Conflict(c, "this submission has already been reviewed")
	case errors.Is(err, ErrNotContractPPK):
		response.Forbidden(c, "only the contract's PPK (or PMO/Admin) can review this submission")
	case errors.Is(err, ErrEmailTaken):
		response.Conflict(c, "email already in use")
	case errors.As(err, &ve):
		response.BadRequest(c, ve.Msg)
	default:
		h.log.Error("vendorportal: internal error", zap.Error(err))
		response.InternalError(c)
	}
}

func parseID(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		response.BadRequest(c, "invalid "+name)
		return uuid.Nil, false
	}
	return id, true
}

// vendorCtx resolves claims and the caller's vendor; writes the error response on failure.
func (h *Handler) vendorCtx(c *gin.Context) (*auth.Claims, uuid.UUID, bool) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return nil, uuid.Nil, false
	}
	vendorID, err := h.svc.VendorOf(claims.OrganizationID, claims.UserID)
	if err != nil {
		h.writeErr(c, err, "")
		return nil, uuid.Nil, false
	}
	return claims, vendorID, true
}

func (h *Handler) sendFile(c *gin.Context, d *Document, path string) {
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

// ── Contractor side: /api/v1/vendor-portal ──────────────────────────────────

// MyContracts GET /vendor-portal/contracts
func (h *Handler) MyContracts(c *gin.Context) {
	claims, vendorID, ok := h.vendorCtx(c)
	if !ok {
		return
	}
	list, err := h.svc.MyContracts(claims.OrganizationID, vendorID)
	if err != nil {
		h.writeErr(c, err, "")
		return
	}
	response.OK(c, "contracts retrieved", list)
}

// MyContract GET /vendor-portal/contracts/:contractID
func (h *Handler) MyContract(c *gin.Context) {
	claims, vendorID, ok := h.vendorCtx(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "contractID")
	if !ok {
		return
	}
	ct, err := h.svc.MyContract(claims.OrganizationID, vendorID, id)
	if err != nil {
		h.writeErr(c, err, "contract not found")
		return
	}
	response.OK(c, "contract retrieved", ct)
}

// ListReports GET /vendor-portal/contracts/:contractID/progress-reports
func (h *Handler) ListReports(c *gin.Context) {
	claims, vendorID, ok := h.vendorCtx(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "contractID")
	if !ok {
		return
	}
	list, err := h.svc.ListReports(claims.OrganizationID, vendorID, id)
	if err != nil {
		h.writeErr(c, err, "contract not found")
		return
	}
	response.OK(c, "progress reports retrieved", list)
}

// CreateReport POST /vendor-portal/contracts/:contractID/progress-reports
func (h *Handler) CreateReport(c *gin.Context) {
	claims, vendorID, ok := h.vendorCtx(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "contractID")
	if !ok {
		return
	}
	var req CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	r, err := h.svc.CreateReport(claims.OrganizationID, vendorID, claims.UserID, id, req)
	if err != nil {
		h.writeErr(c, err, "contract not found")
		return
	}
	h.record(c, claims, "vendor_progress.submitted", "vendor_progress_report", r.ID,
		fmt.Sprintf("%d-%02d %.2f%%", r.PeriodYear, r.PeriodMonth, r.PhysicalProgressPct), r)
	response.Created(c, "progress report submitted", r)
}

// ListDocuments GET /vendor-portal/contracts/:contractID/documents
func (h *Handler) ListDocuments(c *gin.Context) {
	claims, vendorID, ok := h.vendorCtx(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "contractID")
	if !ok {
		return
	}
	list, err := h.svc.ListDocuments(claims.OrganizationID, vendorID, id)
	if err != nil {
		h.writeErr(c, err, "contract not found")
		return
	}
	response.OK(c, "documents retrieved", list)
}

func optionalInt(v string) (*int, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// UploadDocument POST /vendor-portal/contracts/:contractID/documents (multipart)
func (h *Handler) UploadDocument(c *gin.Context) {
	claims, vendorID, ok := h.vendorCtx(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "contractID")
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
	year, err1 := optionalInt(c.PostForm("period_year"))
	month, err2 := optionalInt(c.PostForm("period_month"))
	if err1 != nil || err2 != nil {
		response.BadRequest(c, "invalid period")
		return
	}
	d, err := h.svc.UploadDocument(claims.OrganizationID, vendorID, claims.UserID, id, UploadInput{
		Filename: fh.Filename, Data: data, Name: c.PostForm("name"), Category: c.PostForm("category"),
		PeriodYear: year, PeriodMonth: month, Notes: c.PostForm("notes"),
	})
	if err != nil {
		h.writeErr(c, err, "contract not found")
		return
	}
	h.record(c, claims, "vendor_document.submitted", "vendor_document", d.ID, d.Category+" — "+d.Name, d)
	response.Created(c, "document submitted", d)
}

// DownloadOwnDocument GET /vendor-portal/documents/:documentID/download
func (h *Handler) DownloadOwnDocument(c *gin.Context) {
	claims, vendorID, ok := h.vendorCtx(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "documentID")
	if !ok {
		return
	}
	d, path, err := h.svc.DocumentFile(claims.OrganizationID, &vendorID, id)
	if err != nil {
		h.writeErr(c, err, "document not found")
		return
	}
	h.sendFile(c, d, path)
}

// ── Balai review side: /api/v1/vendor-submissions ───────────────────────────

// ListSubmissions GET /vendor-submissions?status=&project_id=
func (h *Handler) ListSubmissions(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	f := SubmissionFilter{Status: c.Query("status")}
	if p := c.Query("project_id"); p != "" {
		pid, err := uuid.Parse(p)
		if err != nil {
			response.BadRequest(c, "invalid project_id")
			return
		}
		f.ProjectID = &pid
	}
	out, err := h.svc.ListSubmissions(claims.OrganizationID, f)
	if err != nil {
		h.writeErr(c, err, "")
		return
	}
	response.OK(c, "submissions retrieved", out)
}

// ReviewReport POST /vendor-submissions/progress-reports/:id/review
func (h *Handler) ReviewReport(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	r, err := h.svc.ReviewReport(claims.OrganizationID, Reviewer{UserID: claims.UserID}, id, req)
	if err != nil {
		h.writeErr(c, err, "progress report not found")
		return
	}
	h.record(c, claims, "vendor_progress."+strings.ToLower(r.Status), "vendor_progress_report", r.ID,
		fmt.Sprintf("%d-%02d %.2f%%", r.PeriodYear, r.PeriodMonth, r.PhysicalProgressPct),
		gin.H{"status": r.Status, "review_notes": r.ReviewNotes})
	response.OK(c, "progress report reviewed", r)
}

// ReviewDocument POST /vendor-submissions/documents/:id/review
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
	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	d, err := h.svc.ReviewDocument(claims.OrganizationID, Reviewer{UserID: claims.UserID}, id, req)
	if err != nil {
		h.writeErr(c, err, "document not found")
		return
	}
	h.record(c, claims, "vendor_document."+strings.ToLower(d.Status), "vendor_document", d.ID, d.Category+" — "+d.Name,
		gin.H{"status": d.Status, "review_notes": d.ReviewNotes})
	response.OK(c, "document reviewed", d)
}

// DownloadSubmittedDocument GET /vendor-submissions/documents/:id/download
func (h *Handler) DownloadSubmittedDocument(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	d, path, err := h.svc.DocumentFile(claims.OrganizationID, nil, id)
	if err != nil {
		h.writeErr(c, err, "document not found")
		return
	}
	h.sendFile(c, d, path)
}

// ── Contractor accounts: /api/v1/vendors/:vendorID/accounts ─────────────────

// ListAccounts GET /vendors/:vendorID/accounts
func (h *Handler) ListAccounts(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "vendorID")
	if !ok {
		return
	}
	list, err := h.svc.ListAccounts(claims.OrganizationID, id)
	if err != nil {
		h.writeErr(c, err, "vendor not found")
		return
	}
	response.OK(c, "contractor accounts retrieved", list)
}

// CreateAccount POST /vendors/:vendorID/accounts
func (h *Handler) CreateAccount(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "vendorID")
	if !ok {
		return
	}
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	out, err := h.svc.CreateAccount(claims.OrganizationID, id, claims.UserID, req)
	if err != nil {
		h.writeErr(c, err, "vendor not found")
		return
	}
	// Never write the temporary password to the audit log.
	h.record(c, claims, "vendor_account.created", "user", out.Account.ID, out.Account.Email,
		gin.H{"vendor_id": id, "email": out.Account.Email})
	response.Created(c, "contractor account created", out)
}
