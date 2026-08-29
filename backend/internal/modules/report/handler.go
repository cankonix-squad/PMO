package report

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"github.com/harmanto-49/cankora/internal/shared/constants"
	"github.com/harmanto-49/cankora/internal/shared/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler serves report snapshot endpoints.
type Handler struct {
	svc *Service
	log *zap.Logger
}

func NewHandler(db *gorm.DB, log *zap.Logger) *Handler {
	return &Handler{
		svc: NewService(db, log),
		log: log,
	}
}

// claimsFromGin extracts auth claims from the Gin context.
func claimsFromGin(c *gin.Context) *auth.Claims {
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
// Route registration
// ---------------------------------------------------------------------------

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	// All routes require ResourceReports permission — enforced via middleware at server level.
	rg.GET("", h.List)
	rg.POST("", h.Create)
	rg.POST("/generate", h.Generate)
	rg.GET("/:reportID", h.Get)
	rg.PUT("/:reportID", h.Update)
	rg.DELETE("/:reportID", h.Delete)
	rg.POST("/:reportID/transition", h.Transition)
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// List godoc
// GET /api/v1/reports
func (h *Handler) List(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	var filter ListReportFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	snapshots, total, err := h.svc.ListReports(claims, filter)
	if err != nil {
		h.log.Error("list reports", zap.Error(err))
		response.InternalError(c)
		return
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	response.OKPaginated(c, "reports retrieved", snapshots, types.NewPaginationMeta(page, pageSize, total))
}

// Create godoc
// POST /api/v1/reports
func (h *Handler) Create(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	var req CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	snap, err := h.svc.CreateReport(claims, req)
	if err != nil {
		h.log.Error("create report", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, "report created", snap)
}

// Generate godoc
// POST /api/v1/reports/generate — computes live metrics and saves a new DRAFT snapshot
func (h *Handler) Generate(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	var req GenerateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	snap, err := h.svc.GenerateReport(claims, req)
	if err != nil {
		h.log.Error("generate report", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, "report generated", snap)
}

// Get godoc
// GET /api/v1/reports/:reportID
func (h *Handler) Get(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	id, err := uuid.Parse(c.Param("reportID"))
	if err != nil {
		response.BadRequest(c, "invalid report id")
		return
	}

	snap, err := h.svc.GetReport(claims, id)
	if err != nil {
		if err == ErrReportNotFound {
			response.NotFound(c, "report not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "ok", snap)
}

// Update godoc
// PUT /api/v1/reports/:reportID
func (h *Handler) Update(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	id, err := uuid.Parse(c.Param("reportID"))
	if err != nil {
		response.BadRequest(c, "invalid report id")
		return
	}

	var req UpdateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	snap, err := h.svc.UpdateReport(claims, id, req)
	if err != nil {
		if err == ErrReportNotFound {
			response.NotFound(c, "report not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "report updated", snap)
}

// Delete godoc
// DELETE /api/v1/reports/:reportID
func (h *Handler) Delete(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	id, err := uuid.Parse(c.Param("reportID"))
	if err != nil {
		response.BadRequest(c, "invalid report id")
		return
	}

	if err := h.svc.DeleteReport(claims, id); err != nil {
		if err == ErrReportNotFound {
			response.NotFound(c, "report not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "report deleted", nil)
}

// Transition godoc
// POST /api/v1/reports/:reportID/transition
func (h *Handler) Transition(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	id, err := uuid.Parse(c.Param("reportID"))
	if err != nil {
		response.BadRequest(c, "invalid report id")
		return
	}

	var req TransitionReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	snap, err := h.svc.TransitionReport(claims, id, req)
	if err != nil {
		if err == ErrReportNotFound {
			response.NotFound(c, "report not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "report transitioned", snap)
}

// Ensure constants package is referenced (used by server.go for middleware).
var _ = constants.ResourceReports
