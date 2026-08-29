package portfolio

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

// Handler serves program CRUD endpoints.
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

// RegisterRoutes mounts program endpoints onto the provided router group.
// Expected base path: /api/v1/programs
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.List)
	rg.POST("", h.Create)
	rg.GET("/:programID", h.Get)
	rg.PUT("/:programID", h.Update)
	rg.DELETE("/:programID", h.Delete)
}

// List godoc — GET /api/v1/programs
func (h *Handler) List(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	includeInactive := c.Query("include_inactive") == "true"
	list, err := h.svc.List(claims.OrganizationID, includeInactive)
	if err != nil {
		h.log.Error("portfolio: list", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, "programs retrieved", list)
}

// Get godoc — GET /api/v1/programs/:programID
func (h *Handler) Get(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("programID"))
	if err != nil {
		response.BadRequest(c, "invalid program id")
		return
	}
	p, err := h.svc.Get(claims.OrganizationID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "program not found")
			return
		}
		h.log.Error("portfolio: get", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, "program retrieved", p)
}

// Create godoc — POST /api/v1/programs
func (h *Handler) Create(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	var req CreateProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.svc.Create(claims.OrganizationID, claims.UserID, req)
	if err != nil {
		h.log.Error("portfolio: create", zap.Error(err))
		response.InternalError(c)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "program created", "data": p})
}

// Update godoc — PUT /api/v1/programs/:programID
func (h *Handler) Update(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("programID"))
	if err != nil {
		response.BadRequest(c, "invalid program id")
		return
	}
	var req UpdateProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.svc.Update(claims.OrganizationID, id, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "program not found")
			return
		}
		h.log.Error("portfolio: update", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, "program updated", p)
}

// Delete godoc — DELETE /api/v1/programs/:programID
func (h *Handler) Delete(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("programID"))
	if err != nil {
		response.BadRequest(c, "invalid program id")
		return
	}
	if err := h.svc.Delete(claims.OrganizationID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "program not found")
			return
		}
		h.log.Error("portfolio: delete", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, "program deleted", nil)
}
