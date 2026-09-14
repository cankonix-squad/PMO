package router

import (
	"github.com/gin-gonic/gin"
	"github.com/harmanto-49/cankora/internal/identity/auth"
	"github.com/harmanto-49/cankora/internal/identity/rbac"
	"github.com/harmanto-49/cankora/internal/platform/http/middleware"
	"github.com/harmanto-49/cankora/internal/project"
	"github.com/harmanto-49/cankora/internal/shared/constants"
)

func RegisterProjectGovernance(v1 *gin.RouterGroup, h *project.Handler, t *auth.TokenService, a *auth.Service, p rbac.Repository) {
	c := v1.Group("/projects/:id/corrective-actions")
	c.Use(middleware.AuthRequired(t, a))
	c.GET("", middleware.RequirePermission(p, constants.ResourceCorrectiveActions, constants.ActionView), h.ListCorrectiveActions)
	c.POST("", middleware.RequirePermission(p, constants.ResourceCorrectiveActions, constants.ActionCreate), h.CreateCorrectiveAction)
	c.GET("/:caID", middleware.RequirePermission(p, constants.ResourceCorrectiveActions, constants.ActionView), h.GetCorrectiveAction)
	c.PUT("/:caID", middleware.RequirePermission(p, constants.ResourceCorrectiveActions, constants.ActionUpdate), h.UpdateCorrectiveAction)
	c.DELETE("/:caID", middleware.RequirePermission(p, constants.ResourceCorrectiveActions, constants.ActionDelete), h.DeleteCorrectiveAction)
	c.POST("/:caID/transition", middleware.RequirePermission(p, constants.ResourceCorrectiveActions, constants.ActionUpdate), h.TransitionCorrectiveAction)
	v := v1.Group("/vendors")
	v.Use(middleware.AuthRequired(t, a))
	v.GET("", middleware.RequirePermission(p, constants.ResourceVendors, constants.ActionView), h.ListVendors)
	v.POST("", middleware.RequirePermission(p, constants.ResourceVendors, constants.ActionCreate), h.CreateVendor)
	v.GET("/:vendorID", middleware.RequirePermission(p, constants.ResourceVendors, constants.ActionView), h.GetVendor)
	v.PUT("/:vendorID", middleware.RequirePermission(p, constants.ResourceVendors, constants.ActionUpdate), h.UpdateVendor)
	v.DELETE("/:vendorID", middleware.RequirePermission(p, constants.ResourceVendors, constants.ActionDelete), h.DeleteVendor)
}
