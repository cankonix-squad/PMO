package router

import (
	"github.com/gin-gonic/gin"
	"github.com/harmanto-49/cankora/internal/modules/project"
	"github.com/harmanto-49/cankora/internal/platform/http/middleware"
	"github.com/harmanto-49/cankora/internal/shared/constants"
	"github.com/harmanto-49/cankora/internal/core/rbac"
	"github.com/harmanto-49/cankora/internal/core/auth"
)

// RegisterProjects mounts the project aggregate's top-level HTTP endpoints.
func RegisterProjects(v1 *gin.RouterGroup, handler *project.Handler, tokens *auth.TokenService, authService *auth.Service, rbacRepo rbac.Repository) {
	g := v1.Group("/projects")
	g.Use(middleware.AuthRequired(tokens, authService))
	g.GET("", middleware.RequirePermission(rbacRepo, constants.ResourceProjects, constants.ActionView), handler.List)
	g.POST("", middleware.RequirePermission(rbacRepo, constants.ResourceProjects, constants.ActionCreate), handler.Create)
	g.GET("/:id", middleware.RequirePermission(rbacRepo, constants.ResourceProjects, constants.ActionView), handler.GetByID)
	g.PUT("/:id", middleware.RequirePermission(rbacRepo, constants.ResourceProjects, constants.ActionUpdate), handler.Update)
	g.DELETE("/:id", middleware.RequirePermission(rbacRepo, constants.ResourceProjects, constants.ActionDelete), handler.Delete)
	g.POST("/:id/transition", middleware.RequirePermission(rbacRepo, constants.ResourceProjects, constants.ActionUpdate), handler.Transition)
	g.GET("/:id/progress-history", middleware.RequirePermission(rbacRepo, constants.ResourceProjects, constants.ActionView), handler.GetProgressHistory)
}
