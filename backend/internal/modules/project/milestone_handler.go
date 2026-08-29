package project

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/workflow"
	"github.com/harmanto-49/cankora/internal/platform/response"
)

// RegisterMilestoneRoutes mounts milestone routes under /projects/:id/milestones.
func (h *Handler) RegisterMilestoneRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.ListMilestones)
	rg.POST("", h.CreateMilestone)
	rg.GET("/:milestoneID", h.GetMilestone)
	rg.PUT("/:milestoneID", h.UpdateMilestone)
	rg.DELETE("/:milestoneID", h.DeleteMilestone)
}

// ListMilestones godoc
// GET /api/v1/projects/:id/milestones
func (h *Handler) ListMilestones(c *gin.Context) {
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

	milestones, err := h.svc.ListMilestones(c.Request.Context(), projectID, claims.OrganizationID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "ok", milestones)
}

// CreateMilestone godoc
// POST /api/v1/projects/:id/milestones
func (h *Handler) CreateMilestone(c *gin.Context) {
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

	var req CreateMilestoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	m, err := h.svc.CreateMilestone(c.Request.Context(), projectID, claims.OrganizationID, claims.UserID, &req)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.Created(c, "milestone created", m)
}

// GetMilestone godoc
// GET /api/v1/projects/:id/milestones/:milestoneID
func (h *Handler) GetMilestone(c *gin.Context) {
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

	milestoneID, err := uuid.Parse(c.Param("milestoneID"))
	if err != nil {
		response.BadRequest(c, "invalid milestone id")
		return
	}

	m, err := h.svc.GetMilestone(c.Request.Context(), projectID, milestoneID, claims.OrganizationID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrMilestoneNotFound) {
			response.NotFound(c, "milestone not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "ok", m)
}

// UpdateMilestone godoc
// PUT /api/v1/projects/:id/milestones/:milestoneID
func (h *Handler) UpdateMilestone(c *gin.Context) {
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

	milestoneID, err := uuid.Parse(c.Param("milestoneID"))
	if err != nil {
		response.BadRequest(c, "invalid milestone id")
		return
	}

	var req UpdateMilestoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	m, err := h.svc.UpdateMilestone(c.Request.Context(), projectID, milestoneID, claims.OrganizationID, &req)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrMilestoneNotFound) {
			response.NotFound(c, "milestone not found")
			return
		}
		if errors.Is(err, workflow.ErrTransitionNotAllowed) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "ok", m)
}

// DeleteMilestone godoc
// DELETE /api/v1/projects/:id/milestones/:milestoneID
func (h *Handler) DeleteMilestone(c *gin.Context) {
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

	milestoneID, err := uuid.Parse(c.Param("milestoneID"))
	if err != nil {
		response.BadRequest(c, "invalid milestone id")
		return
	}

	if err := h.svc.DeleteMilestone(c.Request.Context(), projectID, milestoneID, claims.OrganizationID); err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrMilestoneNotFound) {
			response.NotFound(c, "milestone not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.NoContent(c)
}
