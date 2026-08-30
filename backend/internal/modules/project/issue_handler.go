package project

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/workflow"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"github.com/harmanto-49/cankora/internal/shared/types"
)

// RegisterIssueRoutes mounts issue routes under /projects/:id/issues.
func (h *Handler) RegisterIssueRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.ListIssues)
	rg.POST("", h.CreateIssue)
	rg.GET("/:issueID", h.GetIssue)
	rg.PUT("/:issueID", h.UpdateIssue)
	rg.DELETE("/:issueID", h.DeleteIssue)
}

// ListIssues godoc
// GET /api/v1/projects/:id/issues
func (h *Handler) ListIssues(c *gin.Context) {
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

	var assignedTo *uuid.UUID
	if raw := c.Query("assigned_to"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			assignedTo = &id
		}
	}

	filter := IssueListFilter{
		OrganizationID: claims.OrganizationID,
		ProjectID:      projectID,
		Status:         c.Query("status"),
		Severity:       c.Query("severity"),
		AssignedTo:     assignedTo,
		Search:         c.Query("search"),
		Page:           page,
		PageSize:       pageSize,
	}

	issues, total, err := h.svc.ListIssues(c.Request.Context(), filter)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OKPaginated(c, "issues retrieved", issues, types.NewPaginationMeta(page, pageSize, total))
}

// CreateIssue godoc
// POST /api/v1/projects/:id/issues
func (h *Handler) CreateIssue(c *gin.Context) {
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

	var req CreateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	issue, err := h.svc.CreateIssue(c.Request.Context(), projectID, claims.OrganizationID, claims.UserID, &req)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrTaskNotFound) {
			response.NotFound(c, "task not found")
			return
		}
		response.InternalError(c)
		return
	}

	h.recordIssueAudit(c, "issue.created", issue)
	response.Created(c, "issue created", issue)
}

// GetIssue godoc
// GET /api/v1/projects/:id/issues/:issueID
func (h *Handler) GetIssue(c *gin.Context) {
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

	issueID, err := uuid.Parse(c.Param("issueID"))
	if err != nil {
		response.BadRequest(c, "invalid issue id")
		return
	}

	issue, err := h.svc.GetIssue(c.Request.Context(), projectID, issueID, claims.OrganizationID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrIssueNotFound) {
			response.NotFound(c, "issue not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "ok", issue)
}

// UpdateIssue godoc
// PUT /api/v1/projects/:id/issues/:issueID
func (h *Handler) UpdateIssue(c *gin.Context) {
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

	issueID, err := uuid.Parse(c.Param("issueID"))
	if err != nil {
		response.BadRequest(c, "invalid issue id")
		return
	}

	var req UpdateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	issue, err := h.svc.UpdateIssue(c.Request.Context(), projectID, issueID, claims.OrganizationID, &req)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrIssueNotFound) {
			response.NotFound(c, "issue not found")
			return
		}
		if errors.Is(err, ErrTaskNotFound) {
			response.NotFound(c, "task not found")
			return
		}
		if errors.Is(err, workflow.ErrTransitionNotAllowed) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}

	h.recordIssueAudit(c, "issue.updated", issue)
	response.OK(c, "ok", issue)
}

// DeleteIssue godoc
// DELETE /api/v1/projects/:id/issues/:issueID
func (h *Handler) DeleteIssue(c *gin.Context) {
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

	issueID, err := uuid.Parse(c.Param("issueID"))
	if err != nil {
		response.BadRequest(c, "invalid issue id")
		return
	}

	if err := h.svc.DeleteIssue(c.Request.Context(), projectID, issueID, claims.OrganizationID); err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrIssueNotFound) {
			response.NotFound(c, "issue not found")
			return
		}
		response.InternalError(c)
		return
	}

	h.recordIssueAudit(c, "issue.deleted", &Issue{ID: issueID, OrganizationID: claims.OrganizationID, Title: "issue " + issueID.String()})
	response.NoContent(c)
}
