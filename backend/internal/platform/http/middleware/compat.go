// Package middleware is the HTTP boundary for cross-cutting request controls.
// It currently delegates to the legacy implementation while route consumers
// are migrated. No business/domain package should depend on this package.
package middleware

import (
	coreauth "github.com/harmanto-49/cankora/internal/core/auth"
	legacy "github.com/harmanto-49/cankora/internal/platform/middleware"
	"github.com/gin-gonic/gin"
	"github.com/harmanto-49/cankora/internal/core/rbac"
	"go.uber.org/zap"
	"time"
)

func RequestID() gin.HandlerFunc { return legacy.RequestID() }
func Logger(log *zap.Logger) gin.HandlerFunc { return legacy.Logger(log) }
func CORS(origins []string) gin.HandlerFunc { return legacy.CORS(origins) }
func AuthRequired(ts *coreauth.TokenService, svc *coreauth.Service) gin.HandlerFunc {
	return legacy.AuthRequired(ts, svc)
}
func RequirePermission(repo rbac.Repository, resource, action string) gin.HandlerFunc {
	return legacy.RequirePermission(repo, resource, action)
}
func RequireAnyPermission(repo rbac.Repository, pairs [][2]string) gin.HandlerFunc {
	return legacy.RequireAnyPermission(repo, pairs)
}
func LoginRateLimit(limit int, window time.Duration) gin.HandlerFunc {
	return legacy.LoginRateLimit(limit, window)
}
