package router

import (
	"time"

	"github.com/gin-gonic/gin"
	coreauth "github.com/harmanto-49/cankora/internal/identity/auth"
	"github.com/harmanto-49/cankora/internal/platform/http/middleware"
)

// RegisterAuth mounts all authentication endpoints and their cross-cutting
// controls in one place.
func RegisterAuth(v1 *gin.RouterGroup, handler *coreauth.Handler, tokens *coreauth.TokenService, service *coreauth.Service) {
	public := v1.Group("/auth")
	public.Use(middleware.LoginRateLimit(10, time.Minute))
	handler.RegisterRoutes(public)

	protected := v1.Group("/auth")
	protected.Use(middleware.AuthRequired(tokens, service))
	handler.RegisterProtectedRoutes(protected)
}
