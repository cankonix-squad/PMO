package organization

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

// Handler serves org unit CRUD endpoints.
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

// RegisterRoutes mounts org-unit endpoints onto the provided router group.
// Expected base path: /api/v1/org-units
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.List)
	rg.POST("", h.Create)
	rg.GET("/:unitID", h.Get)
	rg.PUT("/:unitID", h.Update)
	rg.DELETE("/:unitID", h.Delete)
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// List godoc
// GET /api/v1/org-units
// Query params: include_inactive=true
func (h *Handler) List(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	includeInactive := c.Query("include_inactive") == "true"

	units, err := h.svc.ListOrgUnits(claims.OrganizationID, includeInactive)
	if err != nil {
		h.log.Error("org: list", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.OK(c, "org units retrieved", units)
}

// Get godoc
// GET /api/v1/org-units/:unitID
func (h *Handler) Get(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	unitID, err := uuid.Parse(c.Param("unitID"))
	if err != nil {
		response.BadRequest(c, "invalid unit id")
		return
	}

	unit, err := h.svc.GetOrgUnit(claims.OrganizationID, unitID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "org unit not found")
			return
		}
		h.log.Error("org: get", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.OK(c, "org unit retrieved", unit)
}

// Create godoc
// POST /api/v1/org-units
func (h *Handler) Create(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req CreateOrgUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	unit, err := h.svc.CreateOrgUnit(claims.OrganizationID, req)
	if err != nil {
		h.log.Error("org: create", zap.Error(err))
		response.InternalError(c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "org unit created",
		"data":    unit,
	})
}

// Update godoc
// PUT /api/v1/org-units/:unitID
func (h *Handler) Update(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	unitID, err := uuid.Parse(c.Param("unitID"))
	if err != nil {
		response.BadRequest(c, "invalid unit id")
		return
	}

	var req UpdateOrgUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	unit, err := h.svc.UpdateOrgUnit(claims.OrganizationID, unitID, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "org unit not found")
			return
		}
		h.log.Error("org: update", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.OK(c, "org unit updated", unit)
}

// Delete godoc
// DELETE /api/v1/org-units/:unitID
func (h *Handler) Delete(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	unitID, err := uuid.Parse(c.Param("unitID"))
	if err != nil {
		response.BadRequest(c, "invalid unit id")
		return
	}

	if err := h.svc.DeleteOrgUnit(claims.OrganizationID, unitID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "org unit not found")
			return
		}
		h.log.Error("org: delete", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.OK(c, "org unit deleted", nil)
}
