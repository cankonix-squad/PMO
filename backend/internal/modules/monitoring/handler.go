package monitoring

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler serves baseline and snapshot endpoints.
type Handler struct {
	svc *Service
	log *zap.Logger
}

func NewHandler(db *gorm.DB, log *zap.Logger) *Handler {
	return &Handler{svc: NewService(db, log), log: log}
}

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

// RegisterBaselineRoutes mounts baseline endpoints.
// Expected base: /api/v1/projects/:projectID/baselines
func (h *Handler) RegisterBaselineRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.ListBaselines)
	rg.POST("", h.CreateBaseline)
	rg.GET("/:baselineID", h.GetBaseline)
	rg.PUT("/:baselineID", h.UpdateBaseline)
	rg.DELETE("/:baselineID", h.DeleteBaseline)
}

// RegisterSnapshotRoutes mounts snapshot endpoints.
// Expected base: /api/v1/projects/:projectID/snapshots
func (h *Handler) RegisterSnapshotRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.ListSnapshots)
	rg.POST("", h.CreateSnapshot)
	rg.GET("/:snapshotID", h.GetSnapshot)
	rg.PUT("/:snapshotID", h.UpdateSnapshot)
	rg.PATCH("/:snapshotID/status", h.TransitionSnapshot)
	rg.DELETE("/:snapshotID", h.DeleteSnapshot)
}

// ── helpers ──────────────────────────────────────────────────────────────

func (h *Handler) projectScope(c *gin.Context) (orgID, projectID uuid.UUID, ok bool) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	pid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}
	return claims.OrganizationID, pid, true
}

// ── Baseline handlers ─────────────────────────────────────────────────────

// ListBaselines godoc — GET /api/v1/projects/:projectID/baselines
func (h *Handler) ListBaselines(c *gin.Context) {
	orgID, projectID, ok := h.projectScope(c)
	if !ok {
		return
	}
	list, err := h.svc.ListBaselines(orgID, projectID)
	if err != nil {
		h.log.Error("monitoring: list baselines", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, "baselines retrieved", list)
}

// GetBaseline godoc — GET /api/v1/projects/:projectID/baselines/:baselineID
func (h *Handler) GetBaseline(c *gin.Context) {
	orgID, projectID, ok := h.projectScope(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("baselineID"))
	if err != nil {
		response.BadRequest(c, "invalid baseline id")
		return
	}
	b, err := h.svc.GetBaseline(orgID, projectID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "baseline not found")
			return
		}
		h.log.Error("monitoring: get baseline", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, "baseline retrieved", b)
}

// CreateBaseline godoc — POST /api/v1/projects/:projectID/baselines
func (h *Handler) CreateBaseline(c *gin.Context) {
	orgID, projectID, ok := h.projectScope(c)
	if !ok {
		return
	}
	claims := claimsFromGin(c)

	var req CreateBaselineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	b, err := h.svc.CreateBaseline(orgID, projectID, claims.UserID, req)
	if err != nil {
		h.log.Error("monitoring: create baseline", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "baseline created", "data": b})
}

// UpdateBaseline godoc — PUT /api/v1/projects/:projectID/baselines/:baselineID
func (h *Handler) UpdateBaseline(c *gin.Context) {
	orgID, projectID, ok := h.projectScope(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("baselineID"))
	if err != nil {
		response.BadRequest(c, "invalid baseline id")
		return
	}
	var req UpdateBaselineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	b, err := h.svc.UpdateBaseline(orgID, projectID, id, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "baseline not found")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "baseline updated", b)
}

// DeleteBaseline godoc — DELETE /api/v1/projects/:projectID/baselines/:baselineID
func (h *Handler) DeleteBaseline(c *gin.Context) {
	orgID, projectID, ok := h.projectScope(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("baselineID"))
	if err != nil {
		response.BadRequest(c, "invalid baseline id")
		return
	}
	if err := h.svc.DeleteBaseline(orgID, projectID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "baseline not found")
			return
		}
		h.log.Error("monitoring: delete baseline", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, "baseline deleted", nil)
}

// ── Snapshot handlers ─────────────────────────────────────────────────────

// ListSnapshots godoc — GET /api/v1/projects/:projectID/snapshots
func (h *Handler) ListSnapshots(c *gin.Context) {
	orgID, projectID, ok := h.projectScope(c)
	if !ok {
		return
	}
	status := c.Query("status")
	list, err := h.svc.ListSnapshots(orgID, projectID, status)
	if err != nil {
		h.log.Error("monitoring: list snapshots", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, "snapshots retrieved", list)
}

// GetSnapshot godoc — GET /api/v1/projects/:projectID/snapshots/:snapshotID
func (h *Handler) GetSnapshot(c *gin.Context) {
	orgID, projectID, ok := h.projectScope(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("snapshotID"))
	if err != nil {
		response.BadRequest(c, "invalid snapshot id")
		return
	}
	s, err := h.svc.GetSnapshot(orgID, projectID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "snapshot not found")
			return
		}
		h.log.Error("monitoring: get snapshot", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, "snapshot retrieved", s)
}

// CreateSnapshot godoc — POST /api/v1/projects/:projectID/snapshots
func (h *Handler) CreateSnapshot(c *gin.Context) {
	orgID, projectID, ok := h.projectScope(c)
	if !ok {
		return
	}
	claims := claimsFromGin(c)

	var req CreateSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	s, err := h.svc.CreateSnapshot(orgID, projectID, claims.UserID, req)
	if err != nil {
		h.log.Error("monitoring: create snapshot", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "snapshot created", "data": s})
}

// UpdateSnapshot godoc — PUT /api/v1/projects/:projectID/snapshots/:snapshotID
func (h *Handler) UpdateSnapshot(c *gin.Context) {
	orgID, projectID, ok := h.projectScope(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("snapshotID"))
	if err != nil {
		response.BadRequest(c, "invalid snapshot id")
		return
	}
	var req UpdateSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	s, err := h.svc.UpdateSnapshot(orgID, projectID, id, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "snapshot not found")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "snapshot updated", s)
}

// TransitionSnapshot godoc — PATCH /api/v1/projects/:projectID/snapshots/:snapshotID/status
func (h *Handler) TransitionSnapshot(c *gin.Context) {
	orgID, projectID, ok := h.projectScope(c)
	if !ok {
		return
	}
	claims := claimsFromGin(c)
	id, err := uuid.Parse(c.Param("snapshotID"))
	if err != nil {
		response.BadRequest(c, "invalid snapshot id")
		return
	}
	var req SubmitSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	s, err := h.svc.TransitionSnapshot(orgID, projectID, id, claims.UserID, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "snapshot not found")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "snapshot status updated", s)
}

// DeleteSnapshot godoc — DELETE /api/v1/projects/:projectID/snapshots/:snapshotID
func (h *Handler) DeleteSnapshot(c *gin.Context) {
	orgID, projectID, ok := h.projectScope(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("snapshotID"))
	if err != nil {
		response.BadRequest(c, "invalid snapshot id")
		return
	}
	if err := h.svc.DeleteSnapshot(orgID, projectID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "snapshot not found")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "snapshot deleted", nil)
}
