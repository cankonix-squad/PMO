package primavera

import (
	"errors"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
)

// Handler exposes Primavera P6 sync run endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new Primavera Handler.
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

// RegisterRoutes wires all Primavera integration routes onto the provided RouterGroup.
// The group should already have AuthRequired middleware applied.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/runs", h.ListRuns)
	rg.POST("/runs", h.CreateRun)
	rg.GET("/runs/:id", h.GetRun)
	rg.POST("/runs/:id/process", h.ProcessRun)
	rg.POST("/runs/:id/cancel", h.CancelRun)
	rg.GET("/runs/:id/mappings", h.ListMappings)
}

// ---------------------------------------------------------------------------
// GET /api/v1/integrations/primavera/runs
// ---------------------------------------------------------------------------

func (h *Handler) ListRuns(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	f := ListRunsFilter{
		ProjectID: c.Query("project_id"),
		Status:    c.Query("status"),
		Format:    c.Query("format"),
	}
	f.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	f.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	runs, total, err := h.svc.ListRuns(c.Request.Context(), cl.OrganizationID, f)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "sync runs retrieved", gin.H{
		"data": runs,
		"meta": gin.H{
			"total":     total,
			"page":      f.Page,
			"page_size": f.PageSize,
		},
	})
}

// ---------------------------------------------------------------------------
// POST /api/v1/integrations/primavera/runs
// Multipart: file=<XER/PMXML>, project_id=<uuid>, format=XER|PMXML (optional)
// ---------------------------------------------------------------------------

func (h *Handler) CreateRun(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	// Parse multipart form (max 50 MB)
	if err := c.Request.ParseMultipartForm(MaxSyncFileSizeBytes); err != nil {
		response.BadRequest(c, "multipart form parse error: "+err.Error())
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}
	if fileHeader.Size > MaxSyncFileSizeBytes {
		response.BadRequest(c, "file exceeds maximum allowed size (50 MB)")
		return
	}
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// project_id is required
	projectIDStr := strings.TrimSpace(c.PostForm("project_id"))
	if projectIDStr == "" {
		response.BadRequest(c, "project_id is required")
		return
	}
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		response.BadRequest(c, "project_id is not a valid UUID")
		return
	}

	// optional: format override
	format := strings.ToUpper(strings.TrimSpace(c.PostForm("format")))
	if format == "" {
		// infer from extension
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		switch ext {
		case ".xml", ".pmxml":
			format = FormatPMXML
		default:
			format = FormatXER
		}
	}
	if format != FormatXER && format != FormatPMXML {
		response.BadRequest(c, "format must be XER or PMXML")
		return
	}

	// optional: lineage metadata from form fields
	lineage := LineageMeta{
		SourceProjectID: c.PostForm("source_project_id"),
		ExportedAt:      c.PostForm("exported_at"),
		P6Version:       c.PostForm("p6_version"),
		Operator:        c.PostForm("operator"),
	}

	run, err := h.svc.CreateRun(
		c.Request.Context(),
		cl.OrganizationID,
		cl.UserID,
		&projectID,
		fileHeader.Filename,
		mimeType,
		format,
		fileHeader.Size,
		lineage,
	)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(c, "project not found or access denied")
			return
		}
		response.InternalError(c)
		return
	}

	// Open file and immediately process (synchronous for MVP; can be made async later)
	f, err := fileHeader.Open()
	if err != nil {
		response.InternalError(c)
		return
	}
	defer f.Close()

	processed, err := h.svc.ProcessRun(c.Request.Context(), cl.OrganizationID, cl.UserID, run.ID, f)
	if err != nil {
		// Return the run record (which now has status=FAILED) with 422 so client can see details.
		// Re-fetch the run if ProcessRun didn't return it (e.g. failRun itself errored).
		if processed == nil {
			fetched, fetchErr := h.svc.GetRun(c.Request.Context(), cl.OrganizationID, run.ID)
			if fetchErr == nil {
				processed = fetched
			}
		}
		if processed != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"success": false,
				"message": "sync run failed during processing",
				"data":    processed,
			})
			return
		}
		response.InternalError(c)
		return
	}

	response.Created(c, "sync run completed", processed)
}

// ---------------------------------------------------------------------------
// GET /api/v1/integrations/primavera/runs/:id
// ---------------------------------------------------------------------------

func (h *Handler) GetRun(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid run id")
		return
	}

	run, err := h.svc.GetRun(c.Request.Context(), cl.OrganizationID, runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(c, "sync run not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "sync run retrieved", run)
}

// ---------------------------------------------------------------------------
// POST /api/v1/integrations/primavera/runs/:id/process
// Re-process a PENDING run by uploading the file again (or retry after fix).
// Multipart: file=<XER/PMXML>
// ---------------------------------------------------------------------------

func (h *Handler) ProcessRun(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid run id")
		return
	}

	if err := c.Request.ParseMultipartForm(MaxSyncFileSizeBytes); err != nil {
		response.BadRequest(c, "multipart form parse error: "+err.Error())
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		response.InternalError(c)
		return
	}
	defer f.Close()

	run, err := h.svc.ProcessRun(c.Request.Context(), cl.OrganizationID, cl.UserID, runID, f)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(c, "sync run not found")
			return
		}
		if errors.Is(err, ErrInvalidTransition) {
			response.BadRequest(c, err.Error())
			return
		}
		if run != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"success": false,
				"message": "sync run failed during processing",
				"data":    run,
			})
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "sync run processed", run)
}

// ---------------------------------------------------------------------------
// POST /api/v1/integrations/primavera/runs/:id/cancel
// ---------------------------------------------------------------------------

func (h *Handler) CancelRun(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid run id")
		return
	}

	run, err := h.svc.CancelRun(c.Request.Context(), cl.OrganizationID, cl.UserID, runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(c, "sync run not found")
			return
		}
		if errors.Is(err, ErrInvalidTransition) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "sync run cancelled", run)
}

// ---------------------------------------------------------------------------
// GET /api/v1/integrations/primavera/runs/:id/mappings
// ---------------------------------------------------------------------------

func (h *Handler) ListMappings(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid run id")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	mappings, total, err := h.svc.ListMappings(c.Request.Context(), cl.OrganizationID, runID, page, pageSize)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "activity mappings retrieved", gin.H{
		"data": mappings,
		"meta": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}
