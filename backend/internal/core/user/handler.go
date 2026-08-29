package user

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"github.com/harmanto-49/cankora/internal/shared/types"
)

// Handler wires the user service to HTTP routes.
type Handler struct {
	svc *Service
}

// NewHandler creates a new user Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts user management routes onto a router group.
// The group should already have AuthRequired applied.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.List)
	rg.POST("", h.Create)
	rg.GET("/:id", h.GetByID)
	rg.PUT("/:id", h.Update)
	rg.POST("/:id/deactivate", h.Deactivate)
}

// claimsFromGin extracts JWT claims from the Gin context.
func claimsFromGin(c *gin.Context) *auth.Claims {
	v, ok := c.Get(string(auth.ContextKeyClaims))
	if !ok {
		return nil
	}
	cl, ok := v.(*auth.Claims)
	if !ok {
		return nil
	}
	return cl
}

// parsePagination extracts page / page_size from query params.
func parsePagination(c *gin.Context) (int, int) {
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if v, err := parseInt(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := parseInt(ps); err == nil && v > 0 && v <= 200 {
			pageSize = v
		}
	}
	return page, pageSize
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// List godoc
// GET /api/v1/users
func (h *Handler) List(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	page, pageSize := parsePagination(c)
	filter := UserListFilter{
		OrganizationID: claims.OrganizationID,
		Search:         c.Query("search"),
		IsActive:       parseBoolPtr(c.Query("is_active")),
		Page:           page,
		PageSize:       pageSize,
	}

	users, total, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OKPaginated(c, "users retrieved", users, types.NewPaginationMeta(page, pageSize, total))
}

// GetByID godoc
// GET /api/v1/users/:id
func (h *Handler) GetByID(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	profile, err := h.svc.GetByIDForOrg(c.Request.Context(), id, claims.OrganizationID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	response.OK(c, "ok", profile)
}

// Create godoc
// POST /api/v1/users
func (h *Handler) Create(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Force org to caller's org
	req.OrganizationID = claims.OrganizationID

	profile, err := h.svc.Create(c.Request.Context(), &req, &claims.UserID)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			response.Conflict(c, "email already in use")
			return
		}
		response.InternalError(c)
		return
	}

	response.Created(c, "user created", profile)
}

// Update godoc
// PUT /api/v1/users/:id
func (h *Handler) Update(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	profile, err := h.svc.Update(c.Request.Context(), id, claims.OrganizationID, &req)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "user updated", profile)
}

// Deactivate godoc
// POST /api/v1/users/:id/deactivate
func (h *Handler) Deactivate(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	if err := h.svc.Deactivate(c.Request.Context(), id, claims.OrganizationID); err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "user deactivated", nil)
}

// parseBoolPtr parses "true"/"false" string into *bool.
func parseBoolPtr(s string) *bool {
	if s == "true" {
		v := true
		return &v
	}
	if s == "false" {
		v := false
		return &v
	}
	return nil
}
