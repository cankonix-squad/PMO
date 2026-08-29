package reporting

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"github.com/harmanto-49/cankora/internal/shared/constants"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Handler
// ---------------------------------------------------------------------------

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// NewService + NewHandler convenience wiring (used by server.go).
func NewServiceAndHandler(db *gorm.DB) *Handler {
	return NewHandler(NewService(db))
}

// ---------------------------------------------------------------------------
// claimsFromCtx — extract auth claims
// ---------------------------------------------------------------------------

func claimsFromCtx(c *gin.Context) *auth.Claims {
	val, exists := c.Get(string(auth.ContextKeyClaims))
	if !exists {
		return nil
	}
	claims, ok := val.(*auth.Claims)
	if !ok {
		return nil
	}
	return claims
}

// ---------------------------------------------------------------------------
// RegisterRoutes
// ---------------------------------------------------------------------------

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	// Catalog
	rg.GET("/catalog", h.GetCatalog)

	// Datasets
	ds := rg.Group("/datasets")
	{
		ds.GET("/executive-summary", h.GetExecutiveSummary)
		ds.GET("/project-performance", h.GetProjectPerformance)
		ds.GET("/risk-issue", h.GetRiskIssue)
		ds.GET("/budget", h.GetBudget)
		ds.GET("/benefits", h.GetBenefits)
		ds.GET("/priority", h.GetPriority)
	}

	// Power BI
	rg.GET("/powerbi/config", h.GetPowerBIConfig)

	// Export requests
	exp := rg.Group("/export")
	{
		exp.POST("/request", h.CreateExportRequest)
		exp.GET("/requests", h.ListExportRequests)
		exp.GET("/requests/:requestID", h.GetExportRequest)
		exp.GET("/requests/:requestID/download", h.DownloadExportFile)
	}
}

// ---------------------------------------------------------------------------
// GET /api/v1/analytics/reports/catalog
// ---------------------------------------------------------------------------

// GetCatalog godoc
// Returns all report definitions available for the org.
func (h *Handler) GetCatalog(c *gin.Context) {
	claims := claimsFromCtx(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	defs, err := h.svc.GetCatalog(claims.OrganizationID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "ok", defs)
}

// ---------------------------------------------------------------------------
// GET /api/v1/analytics/reports/datasets/executive-summary
// ---------------------------------------------------------------------------

func (h *Handler) GetExecutiveSummary(c *gin.Context) {
	claims := claimsFromCtx(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var f DatasetFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	data, err := h.svc.GetExecutiveSummary(claims.OrganizationID, f)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "ok", data)
}

// ---------------------------------------------------------------------------
// GET /api/v1/analytics/reports/datasets/project-performance
// ---------------------------------------------------------------------------

func (h *Handler) GetProjectPerformance(c *gin.Context) {
	claims := claimsFromCtx(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var f DatasetFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	data, err := h.svc.GetProjectPerformance(claims.OrganizationID, f)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "ok", data)
}

// ---------------------------------------------------------------------------
// GET /api/v1/analytics/reports/datasets/risk-issue
// ---------------------------------------------------------------------------

func (h *Handler) GetRiskIssue(c *gin.Context) {
	claims := claimsFromCtx(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var f DatasetFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	data, err := h.svc.GetRiskIssue(claims.OrganizationID, f)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "ok", data)
}

// ---------------------------------------------------------------------------
// GET /api/v1/analytics/reports/datasets/budget
// ---------------------------------------------------------------------------

func (h *Handler) GetBudget(c *gin.Context) {
	claims := claimsFromCtx(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var f DatasetFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	data, err := h.svc.GetBudget(claims.OrganizationID, f)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "ok", data)
}

// ---------------------------------------------------------------------------
// GET /api/v1/analytics/reports/datasets/benefits
// ---------------------------------------------------------------------------

func (h *Handler) GetBenefits(c *gin.Context) {
	claims := claimsFromCtx(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var f DatasetFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	data, err := h.svc.GetBenefits(claims.OrganizationID, f)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "ok", data)
}

// ---------------------------------------------------------------------------
// GET /api/v1/analytics/reports/datasets/priority
// ---------------------------------------------------------------------------

func (h *Handler) GetPriority(c *gin.Context) {
	claims := claimsFromCtx(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	data, err := h.svc.GetPriority(claims.OrganizationID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "ok", data)
}

// ---------------------------------------------------------------------------
// GET /api/v1/analytics/reports/powerbi/config
// ---------------------------------------------------------------------------

func (h *Handler) GetPowerBIConfig(c *gin.Context) {
	claims := claimsFromCtx(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	cfg := h.svc.GetPowerBIConfig()
	response.OK(c, "ok", cfg)
}

// ---------------------------------------------------------------------------
// POST /api/v1/analytics/reports/export/request
// ---------------------------------------------------------------------------

func (h *Handler) CreateExportRequest(c *gin.Context) {
	claims := claimsFromCtx(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var input CreateExportRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	req, err := h.svc.CreateExportRequest(claims.OrganizationID, claims.UserID, input)
	if err != nil {
		response.InternalError(c)
		return
	}

	// Process synchronously so the export is ready immediately for UAT/demo.
	h.svc.ProcessExportRequest(req, claims.UserID)

	c.JSON(http.StatusCreated, gin.H{"data": req})
}

// ---------------------------------------------------------------------------
// GET /api/v1/analytics/reports/export/requests
// ---------------------------------------------------------------------------

func (h *Handler) ListExportRequests(c *gin.Context) {
	claims := claimsFromCtx(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	reqs, err := h.svc.ListExportRequests(claims.OrganizationID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "ok", reqs)
}

// ---------------------------------------------------------------------------
// GET /api/v1/analytics/reports/export/requests/:requestID
// ---------------------------------------------------------------------------

func (h *Handler) GetExportRequest(c *gin.Context) {
	claims := claimsFromCtx(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	reqIDStr := c.Param("requestID")
	reqID, err := uuid.Parse(reqIDStr)
	if err != nil {
		response.BadRequest(c, "invalid request ID")
		return
	}

	req, err := h.svc.GetExportRequest(claims.OrganizationID, reqID)
	if err != nil {
		if isNotFound(err) {
			response.NotFound(c, "export request not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, "ok", req)
}

// ---------------------------------------------------------------------------
// GET /api/v1/analytics/reports/export/requests/:requestID/download
// ---------------------------------------------------------------------------

func (h *Handler) DownloadExportFile(c *gin.Context) {
	claims := claimsFromCtx(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	reqIDStr := c.Param("requestID")
	reqID, err := uuid.Parse(reqIDStr)
	if err != nil {
		response.BadRequest(c, "invalid request ID")
		return
	}

	absPath, mimeType, fileName, err := h.svc.DownloadExportFile(
		claims.OrganizationID, reqID, claims.UserID,
	)
	if err != nil {
		if isNotFound(err) {
			response.NotFound(c, "export request not found")
			return
		}
		// Not ready / not completed
		response.BadRequest(c, err.Error())
		return
	}

	// Set disposition using just the base filename — safe from path traversal
	safeFileName := filepath.Base(fileName)
	c.Header("Content-Disposition", `attachment; filename="`+safeFileName+`"`)
	c.Header("Content-Type", mimeType)
	c.Header("Cache-Control", "private, no-cache")
	c.File(absPath)
}

// isNotFound checks for gorm.ErrRecordNotFound.
func isNotFound(err error) bool {
	return err != nil && errors.Is(err, gorm.ErrRecordNotFound)
}

// ---------------------------------------------------------------------------
// Permission helper — used by server.go middleware setup
// ---------------------------------------------------------------------------

// RequireReportingView is a convenience alias for middleware wiring.
func RequireReportingView() string {
	return constants.ResourceReport
}
