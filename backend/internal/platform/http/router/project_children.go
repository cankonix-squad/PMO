package router

import (
	"github.com/gin-gonic/gin"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/core/rbac"
	"github.com/harmanto-49/cankora/internal/modules/project"
	"github.com/harmanto-49/cankora/internal/platform/http/middleware"
	"github.com/harmanto-49/cankora/internal/shared/constants"
)

// RegisterProjectChildren mounts task and milestone endpoints below a project.
func RegisterProjectChildren(v1 *gin.RouterGroup, h *project.Handler, tokens *auth.TokenService, authService *auth.Service, perms rbac.Repository) {
	tasks := v1.Group("/projects/:id/tasks")
	tasks.Use(middleware.AuthRequired(tokens, authService))
	tasks.GET("", middleware.RequirePermission(perms, constants.ResourceTasks, constants.ActionView), h.ListTasks)
	tasks.POST("", middleware.RequirePermission(perms, constants.ResourceTasks, constants.ActionCreate), h.CreateTask)
	tasks.GET("/:taskID", middleware.RequirePermission(perms, constants.ResourceTasks, constants.ActionView), h.GetTask)
	tasks.PUT("/:taskID", middleware.RequirePermission(perms, constants.ResourceTasks, constants.ActionUpdate), h.UpdateTask)
	tasks.DELETE("/:taskID", middleware.RequirePermission(perms, constants.ResourceTasks, constants.ActionDelete), h.DeleteTask)

	milestones := v1.Group("/projects/:id/milestones")
	milestones.Use(middleware.AuthRequired(tokens, authService))
	milestones.GET("", middleware.RequirePermission(perms, constants.ResourceMilestones, constants.ActionView), h.ListMilestones)
	milestones.POST("", middleware.RequirePermission(perms, constants.ResourceMilestones, constants.ActionCreate), h.CreateMilestone)
	milestones.GET("/:milestoneID", middleware.RequirePermission(perms, constants.ResourceMilestones, constants.ActionView), h.GetMilestone)
	milestones.PUT("/:milestoneID", middleware.RequirePermission(perms, constants.ResourceMilestones, constants.ActionUpdate), h.UpdateMilestone)
	milestones.DELETE("/:milestoneID", middleware.RequirePermission(perms, constants.ResourceMilestones, constants.ActionDelete), h.DeleteMilestone)
}
