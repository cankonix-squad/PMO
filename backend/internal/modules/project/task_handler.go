package project

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/workflow"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"github.com/harmanto-49/cankora/internal/shared/types"
)

// RegisterTaskRoutes mounts task routes under /projects/:id/tasks.
// The router group must have :id as the project ID param.
func (h *Handler) RegisterTaskRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.ListTasks)
	rg.POST("", h.CreateTask)
	rg.GET("/:taskID", h.GetTask)
	rg.PUT("/:taskID", h.UpdateTask)
	rg.DELETE("/:taskID", h.DeleteTask)
}

// ListTasks godoc
// GET /api/v1/projects/:id/tasks
func (h *Handler) ListTasks(c *gin.Context) {
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
		id, err := uuid.Parse(raw)
		if err == nil {
			assignedTo = &id
		}
	}

	var milestoneID *uuid.UUID
	if raw := c.Query("milestone_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err == nil {
			milestoneID = &id
		}
	}

	filter := TaskListFilter{
		OrganizationID: claims.OrganizationID,
		ProjectID:      projectID,
		MilestoneID:    milestoneID,
		Status:         c.Query("status"),
		AssignedTo:     assignedTo,
		Search:         c.Query("search"),
		Page:           page,
		PageSize:       pageSize,
	}

	tasks, total, err := h.svc.ListTasks(c.Request.Context(), filter)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OKPaginated(c, "tasks retrieved", tasks, types.NewPaginationMeta(page, pageSize, total))
}

// CreateTask godoc
// POST /api/v1/projects/:id/tasks
func (h *Handler) CreateTask(c *gin.Context) {
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

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	task, err := h.svc.CreateTask(c.Request.Context(), projectID, claims.OrganizationID, claims.UserID, &req)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrTaskNotFound) {
			response.NotFound(c, "parent task not found")
			return
		}
		if errors.Is(err, ErrMilestoneNotFound) {
			response.NotFound(c, "milestone not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.Created(c, "task created", task)
}

// GetTask godoc
// GET /api/v1/projects/:id/tasks/:taskID
func (h *Handler) GetTask(c *gin.Context) {
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

	taskID, err := uuid.Parse(c.Param("taskID"))
	if err != nil {
		response.BadRequest(c, "invalid task id")
		return
	}

	task, err := h.svc.GetTask(c.Request.Context(), projectID, taskID, claims.OrganizationID)
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

	response.OK(c, "ok", task)
}

// UpdateTask godoc
// PUT /api/v1/projects/:id/tasks/:taskID
func (h *Handler) UpdateTask(c *gin.Context) {
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

	taskID, err := uuid.Parse(c.Param("taskID"))
	if err != nil {
		response.BadRequest(c, "invalid task id")
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	task, err := h.svc.UpdateTask(c.Request.Context(), projectID, taskID, claims.OrganizationID, &req)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrTaskNotFound) {
			response.NotFound(c, "task not found")
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

	response.OK(c, "ok", task)
}

// DeleteTask godoc
// DELETE /api/v1/projects/:id/tasks/:taskID
func (h *Handler) DeleteTask(c *gin.Context) {
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

	taskID, err := uuid.Parse(c.Param("taskID"))
	if err != nil {
		response.BadRequest(c, "invalid task id")
		return
	}

	if err := h.svc.DeleteTask(c.Request.Context(), projectID, taskID, claims.OrganizationID); err != nil {
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

	response.NoContent(c)
}
