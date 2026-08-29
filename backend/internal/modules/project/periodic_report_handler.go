package project

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/audit"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"github.com/harmanto-49/cankora/internal/shared/types"
	"gorm.io/gorm"
)

// periodicDB returns the underlying *gorm.DB from the handler, used by periodic report functions.
// The Handler stores a direct db reference set via WithDB during wiring.
func (h *Handler) periodicDB() *gorm.DB { return h.db }

// RegisterPeriodicReportRoutes mounts periodic report routes under /projects/:id/periodic-reports.
func (h *Handler) RegisterPeriodicReportRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.ListPeriodicReports)
	rg.POST("", h.CreatePeriodicReport)
	rg.GET("/:reportID", h.GetPeriodicReport)
	rg.PUT("/:reportID", h.UpdatePeriodicReport)
	rg.DELETE("/:reportID", h.DeletePeriodicReport)
}

// ListPeriodicReports godoc
// GET /api/v1/projects/:id/periodic-reports
func (h *Handler) ListPeriodicReports(c *gin.Context) {
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

	// Verify project belongs to org (tenant guard)
	if _, err := h.svc.GetByID(c.Request.Context(), projectID, claims.OrganizationID); err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		response.InternalError(c)
		return
	}

	page, pageSize := parsePagination(c)

	var year, month int
	if y := c.Query("year"); y != "" {
		year, _ = strconv.Atoi(y)
	}
	if m := c.Query("month"); m != "" {
		month, _ = strconv.Atoi(m)
	}

	filter := PeriodicReportListFilter{
		OrganizationID: claims.OrganizationID,
		ProjectID:      projectID,
		Year:           year,
		Month:          month,
		Page:           page,
		PageSize:       pageSize,
	}

	reports, total, err := listPeriodicReports(c.Request.Context(), h.periodicDB(), filter)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OKPaginated(c, "periodic reports retrieved", reports, types.NewPaginationMeta(page, pageSize, total))
}

// CreatePeriodicReport godoc
// POST /api/v1/projects/:id/periodic-reports
func (h *Handler) CreatePeriodicReport(c *gin.Context) {
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

	// Tenant + project guard
	if _, err := h.svc.GetByID(c.Request.Context(), projectID, claims.OrganizationID); err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		response.InternalError(c)
		return
	}

	var req CreatePeriodicReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Validate bounds explicitly in addition to binding tags
	if req.PhysicalProgressPct < 0 || req.PhysicalProgressPct > 100 {
		response.BadRequest(c, "physical_progress_pct must be between 0 and 100")
		return
	}
	if req.FinancialPlanned < 0 || req.FinancialActual < 0 {
		response.BadRequest(c, "financial values must not be negative")
		return
	}

	reportedAt := time.Now()
	if req.ReportedAt != "" {
		if t, err := time.Parse(time.RFC3339, req.ReportedAt); err == nil {
			reportedAt = t
		}
	}

	r := &PeriodicReport{
		ID:                  uuid.New(),
		OrganizationID:      claims.OrganizationID,
		ProjectID:           projectID,
		PeriodYear:          req.PeriodYear,
		PeriodMonth:         req.PeriodMonth,
		PhysicalProgressPct: req.PhysicalProgressPct,
		FinancialPlanned:    req.FinancialPlanned,
		FinancialActual:     req.FinancialActual,
		FinancialPct:        computeFinancialPct(req.FinancialPlanned, req.FinancialActual),
		Notes:               req.Notes,
		ReportedBy:          &claims.UserID,
		ReportedAt:          reportedAt,
	}

	if err := createPeriodicReport(c.Request.Context(), h.periodicDB(), r); err != nil {
		if isUniqueViolation(err) {
			response.Conflict(c, "periodic report already exists for this period")
			return
		}
		response.InternalError(c)
		return
	}

	// Update project.progress_pct to the latest reported physical progress
	h.svc.SyncProjectProgressFromPeriodic(c.Request.Context(), claims.OrganizationID, projectID)

	h.recordPeriodicReportAudit(c, "periodic_report.created", r)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "periodic report created",
		"data":    r,
	})
}

// GetPeriodicReport godoc
// GET /api/v1/projects/:id/periodic-reports/:reportID
func (h *Handler) GetPeriodicReport(c *gin.Context) {
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
	reportID, err := uuid.Parse(c.Param("reportID"))
	if err != nil {
		response.BadRequest(c, "invalid report id")
		return
	}

	r, err := getPeriodicReport(c.Request.Context(), h.periodicDB(), claims.OrganizationID, projectID, reportID)
	if err != nil {
		if errors.Is(err, ErrPeriodicReportNotFound) {
			response.NotFound(c, "periodic report not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "ok", r)
}

// UpdatePeriodicReport godoc
// PUT /api/v1/projects/:id/periodic-reports/:reportID
func (h *Handler) UpdatePeriodicReport(c *gin.Context) {
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
	reportID, err := uuid.Parse(c.Param("reportID"))
	if err != nil {
		response.BadRequest(c, "invalid report id")
		return
	}

	r, err := getPeriodicReport(c.Request.Context(), h.periodicDB(), claims.OrganizationID, projectID, reportID)
	if err != nil {
		if errors.Is(err, ErrPeriodicReportNotFound) {
			response.NotFound(c, "periodic report not found")
			return
		}
		response.InternalError(c)
		return
	}

	var req UpdatePeriodicReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.PhysicalProgressPct != nil {
		if *req.PhysicalProgressPct < 0 || *req.PhysicalProgressPct > 100 {
			response.BadRequest(c, "physical_progress_pct must be between 0 and 100")
			return
		}
		r.PhysicalProgressPct = *req.PhysicalProgressPct
	}
	if req.FinancialPlanned != nil {
		if *req.FinancialPlanned < 0 {
			response.BadRequest(c, "financial_planned must not be negative")
			return
		}
		r.FinancialPlanned = *req.FinancialPlanned
	}
	if req.FinancialActual != nil {
		if *req.FinancialActual < 0 {
			response.BadRequest(c, "financial_actual must not be negative")
			return
		}
		r.FinancialActual = *req.FinancialActual
	}
	if req.Notes != nil {
		r.Notes = *req.Notes
	}
	if req.ReportedAt != nil && *req.ReportedAt != "" {
		if t, err := time.Parse(time.RFC3339, *req.ReportedAt); err == nil {
			r.ReportedAt = t
		}
	}

	// Recompute financial_pct
	r.FinancialPct = computeFinancialPct(r.FinancialPlanned, r.FinancialActual)

	if err := updatePeriodicReport(c.Request.Context(), h.periodicDB(), r); err != nil {
		response.InternalError(c)
		return
	}

	// Sync project progress if this is the latest period
	h.svc.SyncProjectProgressFromPeriodic(c.Request.Context(), claims.OrganizationID, projectID)

	h.recordPeriodicReportAudit(c, "periodic_report.updated", r)

	response.OK(c, "periodic report updated", r)
}

// DeletePeriodicReport godoc
// DELETE /api/v1/projects/:id/periodic-reports/:reportID
func (h *Handler) DeletePeriodicReport(c *gin.Context) {
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
	reportID, err := uuid.Parse(c.Param("reportID"))
	if err != nil {
		response.BadRequest(c, "invalid report id")
		return
	}

	// Fetch for audit before deleting
	r, err := getPeriodicReport(c.Request.Context(), h.periodicDB(), claims.OrganizationID, projectID, reportID)
	if err != nil {
		if errors.Is(err, ErrPeriodicReportNotFound) {
			response.NotFound(c, "periodic report not found")
			return
		}
		response.InternalError(c)
		return
	}

	if err := softDeletePeriodicReport(c.Request.Context(), h.periodicDB(), claims.OrganizationID, projectID, reportID); err != nil {
		if errors.Is(err, ErrPeriodicReportNotFound) {
			response.NotFound(c, "periodic report not found")
			return
		}
		response.InternalError(c)
		return
	}

	h.recordPeriodicReportAudit(c, "periodic_report.deleted", r)
	response.NoContent(c)
}

// recordPeriodicReportAudit writes an async audit entry.
func (h *Handler) recordPeriodicReportAudit(c *gin.Context, action string, r *PeriodicReport) {
	if h.audit == nil || r == nil {
		return
	}
	claims := claimsFromGin(c)
	actorID := uuid.Nil
	if claims != nil {
		actorID = claims.UserID
	}
	h.audit.Record(audit.WriteRequest{
		OrganizationID: r.OrganizationID,
		ActorID:        &actorID,
		ActorEmail:     actorEmailFromClaims(claims),
		Action:         action,
		EntityType:     "periodic_report",
		EntityID:       r.ID.String(),
		EntityLabel:    r.ProjectID.String() + " " + strconv.Itoa(r.PeriodYear) + "-" + strconv.Itoa(r.PeriodMonth),
		IPAddress:      c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		RequestID:      c.GetString("X-Request-ID"),
	})
}
