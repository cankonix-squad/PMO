package router

import (
 "github.com/gin-gonic/gin"
 "github.com/harmanto-49/cankora/internal/core/auth"
 "github.com/harmanto-49/cankora/internal/core/rbac"
 "github.com/harmanto-49/cankora/internal/modules/project"
 "github.com/harmanto-49/cankora/internal/platform/http/middleware"
 "github.com/harmanto-49/cankora/internal/shared/constants"
)

func RegisterProjectDocuments(v1 *gin.RouterGroup, h *project.Handler, t *auth.TokenService, a *auth.Service, p rbac.Repository) {
 d := v1.Group("/projects/:id/documents"); d.Use(middleware.AuthRequired(t,a))
 d.GET("", middleware.RequirePermission(p,constants.ResourceDocuments,constants.ActionView),h.ListDocuments)
 d.POST("", middleware.RequirePermission(p,constants.ResourceDocuments,constants.ActionCreate),h.UploadDocument)
 d.GET("/:documentID", middleware.RequirePermission(p,constants.ResourceDocuments,constants.ActionView),h.GetDocument)
 d.GET("/:documentID/download", middleware.RequirePermission(p,constants.ResourceDocuments,constants.ActionView),h.DownloadDocument)
 d.PUT("/:documentID", middleware.RequirePermission(p,constants.ResourceDocuments,constants.ActionUpdate),h.UpdateDocument)
 d.DELETE("/:documentID", middleware.RequirePermission(p,constants.ResourceDocuments,constants.ActionDelete),h.DeleteDocument)
 r := v1.Group("/projects/:id/periodic-reports"); r.Use(middleware.AuthRequired(t,a))
 r.GET("", middleware.RequirePermission(p,constants.ResourcePeriodicReport,constants.ActionView),h.ListPeriodicReports)
 r.POST("", middleware.RequirePermission(p,constants.ResourcePeriodicReport,constants.ActionCreate),h.CreatePeriodicReport)
 r.GET("/:reportID", middleware.RequirePermission(p,constants.ResourcePeriodicReport,constants.ActionView),h.GetPeriodicReport)
 r.PUT("/:reportID", middleware.RequirePermission(p,constants.ResourcePeriodicReport,constants.ActionUpdate),h.UpdatePeriodicReport)
 r.DELETE("/:reportID", middleware.RequirePermission(p,constants.ResourcePeriodicReport,constants.ActionDelete),h.DeletePeriodicReport)
}
