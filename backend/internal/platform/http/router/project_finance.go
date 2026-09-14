package router

import (
	"github.com/gin-gonic/gin"
	"github.com/harmanto-49/cankora/internal/identity/auth"
	"github.com/harmanto-49/cankora/internal/identity/rbac"
	"github.com/harmanto-49/cankora/internal/platform/http/middleware"
	"github.com/harmanto-49/cankora/internal/project"
	"github.com/harmanto-49/cankora/internal/shared/constants"
)

func RegisterProjectFinance(v1 *gin.RouterGroup, h *project.Handler, t *auth.TokenService, a *auth.Service, p rbac.Repository) {
	b := v1.Group("/projects/:id/budgets")
	b.Use(middleware.AuthRequired(t, a))
	b.GET("", middleware.RequirePermission(p, constants.ResourceBudgets, constants.ActionView), h.ListBudgets)
	b.POST("", middleware.RequirePermission(p, constants.ResourceBudgets, constants.ActionCreate), h.CreateBudget)
	b.GET("/:budgetID", middleware.RequirePermission(p, constants.ResourceBudgets, constants.ActionView), h.GetBudget)
	b.PUT("/:budgetID", middleware.RequirePermission(p, constants.ResourceBudgets, constants.ActionUpdate), h.UpdateBudget)
	b.DELETE("/:budgetID", middleware.RequirePermission(p, constants.ResourceBudgets, constants.ActionDelete), h.DeleteBudget)
	c := v1.Group("/projects/:id/contracts")
	c.Use(middleware.AuthRequired(t, a))
	c.GET("", middleware.RequirePermission(p, constants.ResourceContracts, constants.ActionView), h.ListContracts)
	c.POST("", middleware.RequirePermission(p, constants.ResourceContracts, constants.ActionCreate), h.CreateContract)
	c.GET("/:contractID", middleware.RequirePermission(p, constants.ResourceContracts, constants.ActionView), h.GetContract)
	c.PUT("/:contractID", middleware.RequirePermission(p, constants.ResourceContracts, constants.ActionUpdate), h.UpdateContract)
	c.DELETE("/:contractID", middleware.RequirePermission(p, constants.ResourceContracts, constants.ActionDelete), h.DeleteContract)
}
