package spatial

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

// Handler serves sector, region, and river basin CRUD endpoints.
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

// RegisterSectorRoutes mounts sector endpoints — base: /api/v1/sectors
func (h *Handler) RegisterSectorRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.ListSectors)
	rg.POST("", h.CreateSector)
	rg.GET("/:sectorID", h.GetSector)
	rg.PUT("/:sectorID", h.UpdateSector)
	rg.DELETE("/:sectorID", h.DeleteSector)
}

// RegisterRegionRoutes mounts region endpoints — base: /api/v1/regions
func (h *Handler) RegisterRegionRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.ListRegions)
	rg.POST("", h.CreateRegion)
	rg.GET("/:regionID", h.GetRegion)
	rg.PUT("/:regionID", h.UpdateRegion)
	rg.DELETE("/:regionID", h.DeleteRegion)
}

// RegisterRiverBasinRoutes mounts river basin endpoints — base: /api/v1/river-basins
func (h *Handler) RegisterRiverBasinRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.ListRiverBasins)
	rg.POST("", h.CreateRiverBasin)
	rg.GET("/:riverBasinID", h.GetRiverBasin)
	rg.PUT("/:riverBasinID", h.UpdateRiverBasin)
	rg.DELETE("/:riverBasinID", h.DeleteRiverBasin)
}

// ---------------------------------------------------------------------------
// Sector handlers
// ---------------------------------------------------------------------------

func (h *Handler) ListSectors(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	list, err := h.svc.ListSectors(claims.OrganizationID, c.Query("include_inactive") == "true")
	if err != nil {
		h.log.Error("spatial: list sectors", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, "sectors retrieved", list)
}

func (h *Handler) GetSector(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("sectorID"))
	if err != nil {
		response.BadRequest(c, "invalid sector id")
		return
	}
	sec, err := h.svc.GetSector(claims.OrganizationID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "sector not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, "sector retrieved", sec)
}

func (h *Handler) CreateSector(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	var req CreateSectorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	sec, err := h.svc.CreateSector(claims.OrganizationID, claims.UserID, req)
	if err != nil {
		h.log.Error("spatial: create sector", zap.Error(err))
		response.InternalError(c)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "sector created", "data": sec})
}

func (h *Handler) UpdateSector(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("sectorID"))
	if err != nil {
		response.BadRequest(c, "invalid sector id")
		return
	}
	var req UpdateSectorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	sec, err := h.svc.UpdateSector(claims.OrganizationID, id, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "sector not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, "sector updated", sec)
}

func (h *Handler) DeleteSector(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("sectorID"))
	if err != nil {
		response.BadRequest(c, "invalid sector id")
		return
	}
	if err := h.svc.DeleteSector(claims.OrganizationID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "sector not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, "sector deleted", nil)
}

// ---------------------------------------------------------------------------
// Region handlers
// ---------------------------------------------------------------------------

func (h *Handler) ListRegions(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	list, err := h.svc.ListRegions(claims.OrganizationID, c.Query("include_inactive") == "true")
	if err != nil {
		h.log.Error("spatial: list regions", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, "regions retrieved", list)
}

func (h *Handler) GetRegion(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("regionID"))
	if err != nil {
		response.BadRequest(c, "invalid region id")
		return
	}
	reg, err := h.svc.GetRegion(claims.OrganizationID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "region not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, "region retrieved", reg)
}

func (h *Handler) CreateRegion(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	var req CreateRegionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	reg, err := h.svc.CreateRegion(claims.OrganizationID, claims.UserID, req)
	if err != nil {
		h.log.Error("spatial: create region", zap.Error(err))
		response.InternalError(c)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "region created", "data": reg})
}

func (h *Handler) UpdateRegion(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("regionID"))
	if err != nil {
		response.BadRequest(c, "invalid region id")
		return
	}
	var req UpdateRegionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	reg, err := h.svc.UpdateRegion(claims.OrganizationID, id, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "region not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, "region updated", reg)
}

func (h *Handler) DeleteRegion(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("regionID"))
	if err != nil {
		response.BadRequest(c, "invalid region id")
		return
	}
	if err := h.svc.DeleteRegion(claims.OrganizationID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "region not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, "region deleted", nil)
}

// ---------------------------------------------------------------------------
// River basin handlers
// ---------------------------------------------------------------------------

func (h *Handler) ListRiverBasins(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	list, err := h.svc.ListRiverBasins(claims.OrganizationID, c.Query("include_inactive") == "true")
	if err != nil {
		h.log.Error("spatial: list river basins", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, "river basins retrieved", list)
}

func (h *Handler) GetRiverBasin(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("riverBasinID"))
	if err != nil {
		response.BadRequest(c, "invalid river basin id")
		return
	}
	rb, err := h.svc.GetRiverBasin(claims.OrganizationID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "river basin not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, "river basin retrieved", rb)
}

func (h *Handler) CreateRiverBasin(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	var req CreateRiverBasinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	rb, err := h.svc.CreateRiverBasin(claims.OrganizationID, claims.UserID, req)
	if err != nil {
		h.log.Error("spatial: create river basin", zap.Error(err))
		response.InternalError(c)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "river basin created", "data": rb})
}

func (h *Handler) UpdateRiverBasin(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("riverBasinID"))
	if err != nil {
		response.BadRequest(c, "invalid river basin id")
		return
	}
	var req UpdateRiverBasinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	rb, err := h.svc.UpdateRiverBasin(claims.OrganizationID, id, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "river basin not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, "river basin updated", rb)
}

func (h *Handler) DeleteRiverBasin(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("riverBasinID"))
	if err != nil {
		response.BadRequest(c, "invalid river basin id")
		return
	}
	if err := h.svc.DeleteRiverBasin(claims.OrganizationID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "river basin not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, "river basin deleted", nil)
}
