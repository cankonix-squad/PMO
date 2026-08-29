package project

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/audit"
	"github.com/harmanto-49/cankora/internal/core/workflow"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"github.com/harmanto-49/cankora/internal/shared/types"
)

// RegisterCorrectiveActionRoutes mounts corrective action routes under
// /projects/:id/corrective-actions.
func (h *Handler) RegisterCorrectiveActionRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.ListCorrectiveActions)
	rg.POST("", h.CreateCorrectiveAction)
	rg.GET("/:caID", h.GetCorrectiveAction)
	rg.PUT("/:caID", h.UpdateCorrectiveAction)
	rg.DELETE("/:caID", h.DeleteCorrectiveAction)
	rg.POST("/:caID/transition", h.TransitionCorrectiveAction)
}

// ListCorrectiveActions godoc
// GET /api/v1/projects/:id/corrective-actions
func (h *Handler) ListCorrectiveActions(c *gin.Context) {
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

	page, pageSize := parsePagination(c)

	filter := CorrectiveActionListFilter{
		OrganizationID: claims.OrganizationID,
		ProjectID:      projectID,
		Status:         c.Query("status"),
		SourceType:     c.Query("source_type"),
		Search:         c.Query("search"),
		Page:           page,
		PageSize:       pageSize,
	}

	cas, total, err := h.svc.ListCorrectiveActions(c.Request.Context(), filter)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OKPaginated(c, "corrective actions retrieved", cas, types.NewPaginationMeta(page, pageSize, total))
}

// CreateCorrectiveAction godoc
// POST /api/v1/projects/:id/corrective-actions
func (h *Handler) CreateCorrectiveAction(c *gin.Context) {
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

	var req CreateCorrectiveActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ca, err := h.svc.CreateCorrectiveAction(c.Request.Context(), projectID, claims.OrganizationID, claims.UserID, &req)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		response.InternalError(c)
		return
	}

	h.recordCorrectiveActionAudit(c, "corrective_action.created", ca)
	response.Created(c, "corrective action created", ca)
}

// GetCorrectiveAction godoc
// GET /api/v1/projects/:id/corrective-actions/:caID
func (h *Handler) GetCorrectiveAction(c *gin.Context) {
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

	caID, err := uuid.Parse(c.Param("caID"))
	if err != nil {
		response.BadRequest(c, "invalid corrective action id")
		return
	}

	ca, err := h.svc.GetCorrectiveAction(c.Request.Context(), projectID, caID, claims.OrganizationID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrCorrectiveActionNotFound) {
			response.NotFound(c, "corrective action not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "ok", ca)
}

// UpdateCorrectiveAction godoc
// PUT /api/v1/projects/:id/corrective-actions/:caID
func (h *Handler) UpdateCorrectiveAction(c *gin.Context) {
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

	caID, err := uuid.Parse(c.Param("caID"))
	if err != nil {
		response.BadRequest(c, "invalid corrective action id")
		return
	}

	var req UpdateCorrectiveActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ca, err := h.svc.UpdateCorrectiveAction(c.Request.Context(), projectID, caID, claims.OrganizationID, &req)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrCorrectiveActionNotFound) {
			response.NotFound(c, "corrective action not found")
			return
		}
		if errors.Is(err, workflow.ErrTransitionNotAllowed) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}

	h.recordCorrectiveActionAudit(c, "corrective_action.updated", ca)
	response.OK(c, "ok", ca)
}

// DeleteCorrectiveAction godoc
// DELETE /api/v1/projects/:id/corrective-actions/:caID
func (h *Handler) DeleteCorrectiveAction(c *gin.Context) {
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

	caID, err := uuid.Parse(c.Param("caID"))
	if err != nil {
		response.BadRequest(c, "invalid corrective action id")
		return
	}

	if err := h.svc.DeleteCorrectiveAction(c.Request.Context(), projectID, caID, claims.OrganizationID); err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrCorrectiveActionNotFound) {
			response.NotFound(c, "corrective action not found")
			return
		}
		response.InternalError(c)
		return
	}

	h.recordCorrectiveActionAudit(c, "corrective_action.deleted", &CorrectiveAction{
		ID:             caID,
		OrganizationID: claims.OrganizationID,
		Title:          "corrective_action " + caID.String(),
	})
	response.NoContent(c)
}

// TransitionCorrectiveAction godoc
// POST /api/v1/projects/:id/corrective-actions/:caID/transition
func (h *Handler) TransitionCorrectiveAction(c *gin.Context) {
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

	caID, err := uuid.Parse(c.Param("caID"))
	if err != nil {
		response.BadRequest(c, "invalid corrective action id")
		return
	}

	var req TransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ca, err := h.svc.TransitionCorrectiveAction(c.Request.Context(), projectID, caID, claims.OrganizationID, &req)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrCorrectiveActionNotFound) {
			response.NotFound(c, "corrective action not found")
			return
		}
		if errors.Is(err, workflow.ErrTransitionNotAllowed) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}

	h.recordCorrectiveActionAudit(c, "corrective_action.transitioned", ca)
	response.OK(c, "ok", ca)
}

// recordCorrectiveActionAudit writes an asynchronous audit entry for corrective
// action lifecycle events.
func (h *Handler) recordCorrectiveActionAudit(c *gin.Context, action string, ca *CorrectiveAction) {
	if h.audit == nil || ca == nil {
		return
	}
	claims := claimsFromGin(c)
	actorID := uuid.Nil
	if claims != nil {
		actorID = claims.UserID
	}
	h.audit.Record(audit.WriteRequest{
		OrganizationID: ca.OrganizationID,
		ActorID:        &actorID,
		ActorEmail:     actorEmailFromClaims(claims),
		Action:         action,
		EntityType:     "corrective_action",
		EntityID:       ca.ID.String(),
		EntityLabel:    ca.Title,
		IPAddress:      c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		RequestID:      c.GetString("X-Request-ID"),
	})
}
