package projectcategory

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

// Handler serves project category CRUD endpoints.
type Handler struct {
	svc *Service
	log *zap.Logger
}

func NewHandler(db *gorm.DB, log *zap.Logger) *Handler {
	return &Handler{svc: NewService(db, log), log: log}
}

// RegisterRoutes mounts routes — base: /api/v1/project-categories
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.List)
	rg.POST("", h.Create)
	rg.GET("/:id", h.GetByID)
	rg.PUT("/:id", h.Update)
	rg.DELETE("/:id", h.Delete)
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

// List godoc GET /api/v1/project-categories
func (h *Handler) List(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	includeInactive := c.Query("include_inactive") == "true"
	list, err := h.svc.List(claims.OrganizationID, includeInactive)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "project categories retrieved", list)
}

// GetByID godoc GET /api/v1/project-categories/:id
func (h *Handler) GetByID(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	cat, err := h.svc.GetByID(claims.OrganizationID, id)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "project category not found")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "project category retrieved", cat)
}

// Create godoc POST /api/v1/project-categories
func (h *Handler) Create(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	cat, err := h.svc.Create(claims.OrganizationID, claims.UserID, req)
	if errors.Is(err, ErrCodeTaken) {
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": "project category code already in use"})
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Created(c, "project category created", cat)
}

// Update godoc PUT /api/v1/project-categories/:id
func (h *Handler) Update(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	cat, err := h.svc.Update(claims.OrganizationID, id, req)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "project category not found")
		return
	}
	if errors.Is(err, ErrCodeTaken) {
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": "project category code already in use"})
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "project category updated", cat)
}

// Delete godoc DELETE /api/v1/project-categories/:id
func (h *Handler) Delete(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if err := h.svc.Delete(claims.OrganizationID, id); errors.Is(err, ErrNotFound) {
		response.NotFound(c, "project category not found")
		return
	} else if err != nil {
		response.InternalError(c)
		return
	}
	c.Status(http.StatusNoContent)
}
