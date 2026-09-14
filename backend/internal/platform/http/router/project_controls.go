package router

import (
	"github.com/gin-gonic/gin"
	"github.com/harmanto-49/cankora/internal/identity/auth"
	"github.com/harmanto-49/cankora/internal/identity/rbac"
	"github.com/harmanto-49/cankora/internal/platform/http/middleware"
	"github.com/harmanto-49/cankora/internal/project"
	"github.com/harmanto-49/cankora/internal/shared/constants"
)

// RegisterProjectControls mounts issue and risk endpoints under a project.
func RegisterProjectControls(v1 *gin.RouterGroup, h *project.Handler, tokens *auth.TokenService, authService *auth.Service, perms rbac.Repository) {
	issues := v1.Group("/projects/:id/issues")
	issues.Use(middleware.AuthRequired(tokens, authService))
	issues.GET("", middleware.RequirePermission(perms, constants.ResourceIssues, constants.ActionView), h.ListIssues)
	issues.POST("", middleware.RequirePermission(perms, constants.ResourceIssues, constants.ActionCreate), h.CreateIssue)
	issues.GET("/:issueID", middleware.RequirePermission(perms, constants.ResourceIssues, constants.ActionView), h.GetIssue)
	issues.PUT("/:issueID", middleware.RequirePermission(perms, constants.ResourceIssues, constants.ActionUpdate), h.UpdateIssue)
	issues.DELETE("/:issueID", middleware.RequirePermission(perms, constants.ResourceIssues, constants.ActionDelete), h.DeleteIssue)

	risks := v1.Group("/projects/:id/risks")
	risks.Use(middleware.AuthRequired(tokens, authService))
	risks.GET("", middleware.RequirePermission(perms, constants.ResourceRisks, constants.ActionView), h.ListRisks)
	risks.POST("", middleware.RequirePermission(perms, constants.ResourceRisks, constants.ActionCreate), h.CreateRisk)
	risks.GET("/:riskID", middleware.RequirePermission(perms, constants.ResourceRisks, constants.ActionView), h.GetRisk)
	risks.PUT("/:riskID", middleware.RequirePermission(perms, constants.ResourceRisks, constants.ActionUpdate), h.UpdateRisk)
	risks.DELETE("/:riskID", middleware.RequirePermission(perms, constants.ResourceRisks, constants.ActionDelete), h.DeleteRisk)
}
