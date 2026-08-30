package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/harmanto-49/cankora/internal/core/rbac"
	"github.com/harmanto-49/cankora/internal/platform/response"
)

// RequirePermission returns a Gin middleware that checks whether the authenticated
// user has the given resource+action permission.
// Must be placed after AuthRequired in the middleware chain.
func RequirePermission(rbacRepo rbac.Repository, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := ClaimsFromContext(c)
		if claims == nil {
			response.Unauthorized(c, "")
			c.Abort()
			return
		}

		perms, err := rbacRepo.GetUserPermissions(c.Request.Context(), claims.UserID)
		if err != nil {
			response.InternalError(c)
			c.Abort()
			return
		}

		for _, p := range perms {
			if p.Resource == resource && p.Action == action {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "")
		c.Abort()
	}
}

// RequireAnyPermission passes if the user has at least one of the given resource+action pairs.
func RequireAnyPermission(rbacRepo rbac.Repository, pairs [][2]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := ClaimsFromContext(c)
		if claims == nil {
			response.Unauthorized(c, "")
			c.Abort()
			return
		}

		perms, err := rbacRepo.GetUserPermissions(c.Request.Context(), claims.UserID)
		if err != nil {
			response.InternalError(c)
			c.Abort()
			return
		}

		for _, pair := range pairs {
			for _, p := range perms {
				if p.Resource == pair[0] && p.Action == pair[1] {
					c.Next()
					return
				}
			}
		}

		response.Forbidden(c, "")
		c.Abort()
	}
}
