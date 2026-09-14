package project

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"github.com/harmanto-49/cankora/internal/shared/types"
)

// RegisterBudgetRoutes mounts budget routes under /projects/:id/budgets.
func (h *Handler) RegisterBudgetRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.ListBudgets)
	rg.POST("", h.CreateBudget)
	rg.GET("/:budgetID", h.GetBudget)
	rg.PUT("/:budgetID", h.UpdateBudget)
	rg.DELETE("/:budgetID", h.DeleteBudget)
}

// ListBudgets godoc
// GET /api/v1/projects/:id/budgets
func (h *Handler) ListBudgets(c *gin.Context) {
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

	filter := BudgetListFilter{
		OrganizationID: claims.OrganizationID,
		ProjectID:      projectID,
		Category:       c.Query("category"),
		Search:         c.Query("search"),
		Page:           page,
		PageSize:       pageSize,
	}

	lines, total, err := h.svc.ListBudgets(c.Request.Context(), filter)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OKPaginated(c, "budgets retrieved", lines, types.NewPaginationMeta(page, pageSize, total))
}

// CreateBudget godoc
// POST /api/v1/projects/:id/budgets
func (h *Handler) CreateBudget(c *gin.Context) {
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

	var req CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	line, err := h.svc.CreateBudget(c.Request.Context(), projectID, claims.OrganizationID, claims.UserID, &req)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		response.InternalError(c)
		return
	}

	h.recordBudgetAudit(c, "budget.created", line)
	response.Created(c, "budget created", line)
}

// GetBudget godoc
// GET /api/v1/projects/:id/budgets/:budgetID
func (h *Handler) GetBudget(c *gin.Context) {
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

	budgetID, err := uuid.Parse(c.Param("budgetID"))
	if err != nil {
		response.BadRequest(c, "invalid budget id")
		return
	}

	line, err := h.svc.GetBudget(c.Request.Context(), projectID, budgetID, claims.OrganizationID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrBudgetNotFound) {
			response.NotFound(c, "budget not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "ok", line)
}

// UpdateBudget godoc
// PUT /api/v1/projects/:id/budgets/:budgetID
func (h *Handler) UpdateBudget(c *gin.Context) {
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

	budgetID, err := uuid.Parse(c.Param("budgetID"))
	if err != nil {
		response.BadRequest(c, "invalid budget id")
		return
	}

	var req UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	line, err := h.svc.UpdateBudget(c.Request.Context(), projectID, budgetID, claims.OrganizationID, &req)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrBudgetNotFound) {
			response.NotFound(c, "budget not found")
			return
		}
		response.InternalError(c)
		return
	}

	h.recordBudgetAudit(c, "budget.updated", line)
	response.OK(c, "ok", line)
}

// DeleteBudget godoc
// DELETE /api/v1/projects/:id/budgets/:budgetID
func (h *Handler) DeleteBudget(c *gin.Context) {
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

	budgetID, err := uuid.Parse(c.Param("budgetID"))
	if err != nil {
		response.BadRequest(c, "invalid budget id")
		return
	}

	if err := h.svc.DeleteBudget(c.Request.Context(), projectID, budgetID, claims.OrganizationID); err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrBudgetNotFound) {
			response.NotFound(c, "budget not found")
			return
		}
		response.InternalError(c)
		return
	}

	h.recordBudgetAudit(c, "budget.deleted", &BudgetLine{ID: budgetID, ProjectID: projectID, Category: "budget " + budgetID.String()})
	response.NoContent(c)
}
