package imports

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
)

// Handler exposes import job endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new import Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// claims extracts JWT claims from gin context.
func (h *Handler) claims(c *gin.Context) (*auth.Claims, bool) {
	v, ok := c.Get(string(auth.ContextKeyClaims))
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return nil, false
	}
	cl, ok := v.(*auth.Claims)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return nil, false
	}
	return cl, true
}

// RegisterRoutes wires all import routes onto the provided RouterGroup.
// The group should already have AuthRequired middleware applied.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/templates", h.ListTemplates)
	rg.GET("/jobs", h.ListJobs)
	rg.POST("/jobs", h.CreateJob)
	rg.GET("/jobs/:id", h.GetJob)
	rg.POST("/jobs/:id/validate", h.ValidateJob)
	rg.POST("/jobs/:id/commit", h.CommitJob)
	rg.POST("/jobs/:id/cancel", h.CancelJob)
	rg.GET("/jobs/:id/rows", h.ListRows)
}

// GET /api/v1/imports/templates
func (h *Handler) ListTemplates(c *gin.Context) {
	_, ok := h.claims(c)
	if !ok {
		return
	}
	response.OK(c, "import templates retrieved", Templates())
}

// GET /api/v1/imports/jobs
func (h *Handler) ListJobs(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	f := ListJobsFilter{
		DatasetType: c.Query("dataset_type"),
		Status:      c.Query("status"),
	}
	f.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	f.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	jobs, total, err := h.svc.ListJobs(c.Request.Context(), cl.OrganizationID, f)
	if err != nil {
		response.InternalError(c)
		return
	}

	var result []JobResponse
	for i := range jobs {
		result = append(result, toJobResponse(&jobs[i]))
	}
	if result == nil {
		result = []JobResponse{}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "import jobs retrieved",
		"data":    result,
		"meta": gin.H{
			"total":     total,
			"page":      f.Page,
			"page_size": f.PageSize,
		},
	})
}

// POST /api/v1/imports/jobs  (multipart/form-data)
// Form fields:
//   - dataset_type: string (required)
//   - file: the CSV/XLSX file (required)
func (h *Handler) CreateJob(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	datasetType := strings.TrimSpace(c.PostForm("dataset_type"))
	if datasetType == "" {
		response.BadRequest(c, "dataset_type is required")
		return
	}
	if _, found := TemplateByType(datasetType); !found {
		response.BadRequest(c, "unsupported dataset_type: "+datasetType)
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}

	// Size check
	if fh.Size > MaxFileSizeBytes {
		response.BadRequest(c, "file exceeds maximum allowed size of 10 MB")
		return
	}

	// MIME type check — use filename extension as fallback
	mimeType := fh.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	// Strip parameters e.g. "text/csv; charset=utf-8"
	if idx := strings.Index(mimeType, ";"); idx >= 0 {
		mimeType = strings.TrimSpace(mimeType[:idx])
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !AllowedMIMETypes[mimeType] && ext != ".csv" && ext != ".xlsx" && ext != ".xls" {
		response.BadRequest(c, "unsupported file type; allowed: CSV, XLS, XLSX")
		return
	}

	// Sanitize filename — no path traversal
	safeName := filepath.Base(fh.Filename)
	if safeName == "." || safeName == "/" {
		safeName = "upload.csv"
	}

	// Read file into memory for parsing (max 10 MB already checked above)
	f, err := fh.Open()
	if err != nil {
		response.InternalError(c)
		return
	}
	defer f.Close()
	fileBytes, err := io.ReadAll(f)
	if err != nil {
		response.InternalError(c)
		return
	}

	// Create job record
	job, err := h.svc.CreateJob(
		c.Request.Context(),
		cl.OrganizationID, cl.UserID,
		datasetType, safeName, mimeType, int64(len(fileBytes)),
	)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Auto-validate immediately after upload
	job, err = h.svc.ValidateJob(
		c.Request.Context(),
		cl.OrganizationID, cl.UserID, job.ID,
		bytes.NewReader(fileBytes),
	)
	if err != nil {
		// Job was already marked FAILED inside ValidateJob — return the job with error
		if getJob, gErr := h.svc.GetJob(c.Request.Context(), cl.OrganizationID, job.ID); gErr == nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"success": false,
				"message": "file parsed but validation failed: " + err.Error(),
				"data":    toJobResponse(getJob),
			})
			return
		}
		response.InternalError(c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "import job created and validated",
		"data":    toJobResponse(job),
	})
}

// GET /api/v1/imports/jobs/:id
func (h *Handler) GetJob(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid job id")
		return
	}
	job, err := h.svc.GetJob(c.Request.Context(), cl.OrganizationID, jobID)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "import job not found")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "import job retrieved", toJobResponse(job))
}

// POST /api/v1/imports/jobs/:id/validate  (re-validate with new file upload)
func (h *Handler) ValidateJob(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid job id")
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required for re-validation")
		return
	}
	if fh.Size > MaxFileSizeBytes {
		response.BadRequest(c, "file exceeds maximum allowed size of 10 MB")
		return
	}
	f, err := fh.Open()
	if err != nil {
		response.InternalError(c)
		return
	}
	defer f.Close()

	job, err := h.svc.ValidateJob(c.Request.Context(), cl.OrganizationID, cl.UserID, jobID, f)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "import job not found")
		return
	}
	if errors.Is(err, ErrInvalidTransition) {
		response.BadRequest(c, err.Error())
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "import job validated", toJobResponse(job))
}

// POST /api/v1/imports/jobs/:id/commit
func (h *Handler) CommitJob(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid job id")
		return
	}
	job, err := h.svc.CommitJob(c.Request.Context(), cl.OrganizationID, cl.UserID, jobID)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "import job not found")
		return
	}
	if errors.Is(err, ErrInvalidTransition) {
		response.BadRequest(c, err.Error())
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "import job committed", toJobResponse(job))
}

// POST /api/v1/imports/jobs/:id/cancel
func (h *Handler) CancelJob(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid job id")
		return
	}
	job, err := h.svc.CancelJob(c.Request.Context(), cl.OrganizationID, cl.UserID, jobID)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "import job not found")
		return
	}
	if errors.Is(err, ErrInvalidTransition) {
		response.BadRequest(c, err.Error())
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "import job cancelled", toJobResponse(job))
}

// GET /api/v1/imports/jobs/:id/rows
func (h *Handler) ListRows(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid job id")
		return
	}

	var validOnly *bool
	if v := c.Query("valid"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			validOnly = &b
		}
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	rows, total, err := h.svc.ListRows(c.Request.Context(), cl.OrganizationID, jobID, validOnly, page, pageSize)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "import job not found")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}

	var result []RowResponse
	for i := range rows {
		result = append(result, toRowResponse(&rows[i]))
	}
	if result == nil {
		result = []RowResponse{}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "import rows retrieved",
		"data":    result,
		"meta": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// --- helpers ---

func toJobResponse(j *Job) JobResponse {
	var errSummary []string
	_ = json.Unmarshal([]byte(j.ErrorSummary), &errSummary)
	if errSummary == nil {
		errSummary = []string{}
	}
	return JobResponse{
		ID:             j.ID,
		OrganizationID: j.OrganizationID,
		DatasetType:    j.DatasetType,
		FileName:       j.FileName,
		FileSize:       j.FileSize,
		MIMEType:       j.MIMEType,
		Status:         j.Status,
		TotalRows:      j.TotalRows,
		ValidRows:      j.ValidRows,
		InvalidRows:    j.InvalidRows,
		ErrorSummary:   errSummary,
		UploadedBy:     j.UploadedBy,
		ValidatedAt:    j.ValidatedAt,
		CommittedAt:    j.CommittedAt,
		CreatedAt:      j.CreatedAt,
		UpdatedAt:      j.UpdatedAt,
	}
}

func toRowResponse(r *Row) RowResponse {
	var raw, normalized any
	_ = json.Unmarshal([]byte(r.RawPayload), &raw)
	_ = json.Unmarshal([]byte(r.NormalizedPayload), &normalized)
	var errs []string
	_ = json.Unmarshal([]byte(r.Errors), &errs)
	if errs == nil {
		errs = []string{}
	}
	return RowResponse{
		ID:                r.ID,
		JobID:             r.JobID,
		RowNumber:         r.RowNumber,
		RawPayload:        raw,
		NormalizedPayload: normalized,
		Valid:             r.Valid,
		Errors:            errs,
		Action:            r.Action,
		TargetEntityID:    r.TargetEntityID,
	}
}
