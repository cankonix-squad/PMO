package project

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/workflow"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"github.com/harmanto-49/cankora/internal/shared/types"
)

// RegisterRiskRoutes mounts risk routes under /projects/:id/risks.
func (h *Handler) RegisterRiskRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.ListRisks)
	rg.POST("", h.CreateRisk)
	rg.GET("/:riskID", h.GetRisk)
	rg.PUT("/:riskID", h.UpdateRisk)
	rg.DELETE("/:riskID", h.DeleteRisk)
}

// ListRisks godoc
// GET /api/v1/projects/:id/risks
func (h *Handler) ListRisks(c *gin.Context) {
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

	var ownedBy *uuid.UUID
	if raw := c.Query("owned_by"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			ownedBy = &id
		}
	}

	filter := RiskListFilter{
		OrganizationID: claims.OrganizationID,
		ProjectID:      projectID,
		Status:         c.Query("status"),
		Severity:       c.Query("severity"),
		OwnedBy:        ownedBy,
		Search:         c.Query("search"),
		Page:           page,
		PageSize:       pageSize,
	}

	risks, total, err := h.svc.ListRisks(c.Request.Context(), filter)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OKPaginated(c, "risks retrieved", risks, types.NewPaginationMeta(page, pageSize, total))
}

// CreateRisk godoc
// POST /api/v1/projects/:id/risks
func (h *Handler) CreateRisk(c *gin.Context) {
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

	var req CreateRiskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	risk, err := h.svc.CreateRisk(c.Request.Context(), projectID, claims.OrganizationID, claims.UserID, &req)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		response.InternalError(c)
		return
	}

	h.recordRiskAudit(c, "risk.created", risk)
	response.Created(c, "risk created", risk)
}

// GetRisk godoc
// GET /api/v1/projects/:id/risks/:riskID
func (h *Handler) GetRisk(c *gin.Context) {
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

	riskID, err := uuid.Parse(c.Param("riskID"))
	if err != nil {
		response.BadRequest(c, "invalid risk id")
		return
	}

	risk, err := h.svc.GetRisk(c.Request.Context(), projectID, riskID, claims.OrganizationID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrRiskNotFound) {
			response.NotFound(c, "risk not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "ok", risk)
}

// UpdateRisk godoc
// PUT /api/v1/projects/:id/risks/:riskID
func (h *Handler) UpdateRisk(c *gin.Context) {
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

	riskID, err := uuid.Parse(c.Param("riskID"))
	if err != nil {
		response.BadRequest(c, "invalid risk id")
		return
	}

	var req UpdateRiskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	risk, err := h.svc.UpdateRisk(c.Request.Context(), projectID, riskID, claims.OrganizationID, &req)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrRiskNotFound) {
			response.NotFound(c, "risk not found")
			return
		}
		if errors.Is(err, workflow.ErrTransitionNotAllowed) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}

	h.recordRiskAudit(c, "risk.updated", risk)
	response.OK(c, "ok", risk)
}

// DeleteRisk godoc
// DELETE /api/v1/projects/:id/risks/:riskID
func (h *Handler) DeleteRisk(c *gin.Context) {
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

	riskID, err := uuid.Parse(c.Param("riskID"))
	if err != nil {
		response.BadRequest(c, "invalid risk id")
		return
	}

	if err := h.svc.DeleteRisk(c.Request.Context(), projectID, riskID, claims.OrganizationID); err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrRiskNotFound) {
			response.NotFound(c, "risk not found")
			return
		}
		response.InternalError(c)
		return
	}

	h.recordRiskAudit(c, "risk.deleted", &Risk{ID: riskID, OrganizationID: claims.OrganizationID, Title: "risk " + riskID.String()})
	response.NoContent(c)
}
