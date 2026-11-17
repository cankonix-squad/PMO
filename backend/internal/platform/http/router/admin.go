package router

import (
	"github.com/gin-gonic/gin"
	"github.com/harmanto-49/cankora/internal/governance/auditlog"
	governance "github.com/harmanto-49/cankora/internal/governance/submissions"
	"github.com/harmanto-49/cankora/internal/identity/auth"
	"github.com/harmanto-49/cankora/internal/identity/rbac"
	"github.com/harmanto-49/cankora/internal/identity/user"
	"github.com/harmanto-49/cankora/internal/notifications"
	"github.com/harmanto-49/cankora/internal/platform/http/middleware"
	"github.com/harmanto-49/cankora/internal/shared/constants"
)

func RegisterAdmin(v1 *gin.RouterGroup, uh *user.Handler, gh *governance.Handler, ah *auditlog.Handler, nh *notifications.Handler, t *auth.TokenService, a *auth.Service, p rbac.Repository) {
	u := v1.Group("/users")
	u.Use(middleware.AuthRequired(t, a))
	u.GET("", middleware.RequirePermission(p, constants.ResourceUsers, constants.ActionView), uh.List)
	u.POST("", middleware.RequirePermission(p, constants.ResourceUsers, constants.ActionCreate), uh.Create)
	u.GET("/:id", middleware.RequirePermission(p, constants.ResourceUsers, constants.ActionView), uh.GetByID)
	u.PUT("/:id", middleware.RequirePermission(p, constants.ResourceUsers, constants.ActionUpdate), uh.Update)
	u.POST("/:id/deactivate", middleware.RequirePermission(p, constants.ResourceUsers, constants.ActionUpdate), uh.Deactivate)
	g := v1.Group("/governance")
	g.Use(middleware.AuthRequired(t, a))
	g.GET("/submissions", middleware.RequirePermission(p, constants.ResourceDataGovernance, constants.ActionView), gh.ListSubmissions)
	g.POST("/submissions", middleware.RequirePermission(p, constants.ResourceDataGovernance, constants.ActionCreate), gh.CreateSubmission)
	g.GET("/submissions/:id", middleware.RequirePermission(p, constants.ResourceDataGovernance, constants.ActionView), gh.GetSubmission)
	g.POST("/submissions/:id/submit", middleware.RequirePermission(p, constants.ResourceDataGovernance, constants.ActionUpdate), gh.Submit)
	g.POST("/submissions/:id/review", middleware.RequirePermission(p, constants.ResourceDataGovernance, constants.ActionApprove), gh.StartReview)
	g.POST("/submissions/:id/approve", middleware.RequirePermission(p, constants.ResourceDataGovernance, constants.ActionApprove), gh.Approve)
	g.POST("/submissions/:id/reject", middleware.RequirePermission(p, constants.ResourceDataGovernance, constants.ActionApprove), gh.Reject)
	g.POST("/submissions/:id/lock", middleware.RequirePermission(p, constants.ResourceDataGovernance, constants.ActionUpdate), gh.Lock)
	g.POST("/submissions/:id/cancel", middleware.RequirePermission(p, constants.ResourceDataGovernance, constants.ActionUpdate), gh.Cancel)
	g.GET("/lock-periods", middleware.RequirePermission(p, constants.ResourceDataGovernance, constants.ActionView), gh.ListLockPeriods)
	g.POST("/lock-periods", middleware.RequirePermission(p, constants.ResourceDataGovernance, constants.ActionCreate), gh.CreateLockPeriod)
	g.POST("/lock-periods/:id/lock", middleware.RequirePermission(p, constants.ResourceDataGovernance, constants.ActionUpdate), gh.LockPeriod)
	al := v1.Group("/audit-logs")
	al.Use(middleware.AuthRequired(t, a), middleware.RequirePermission(p, constants.ResourceAuditLogs, constants.ActionView))
	ah.RegisterRoutes(al)
	n := v1.Group("/notifications")
	n.Use(middleware.AuthRequired(t, a), middleware.RequirePermission(p, constants.ResourceNotification, constants.ActionView))
	nh.RegisterRoutes(n)
	// Org-wide listing: administrators only (user management permission).
	na := v1.Group("/notifications")
	na.Use(middleware.AuthRequired(t, a), middleware.RequirePermission(p, constants.ResourceUsers, constants.ActionView))
	nh.RegisterAdminRoutes(na)
}
