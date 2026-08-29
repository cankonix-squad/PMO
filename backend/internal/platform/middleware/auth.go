package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
)

const (
	headerAuthorization = "Authorization"
	bearerPrefix        = "Bearer "
)

// AuthRequired is a Gin middleware that validates the JWT access token.
// On success it injects auth.Claims into the context under the key auth.ContextKeyClaims.
func AuthRequired(tokenSvc *auth.TokenService, authSvc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(headerAuthorization)
		if header == "" || !strings.HasPrefix(header, bearerPrefix) {
			response.Unauthorized(c, "Missing or malformed Authorization header")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(header, bearerPrefix)
		claims, err := tokenSvc.ValidateAccessToken(tokenStr)
		if err != nil {
			if err == auth.ErrTokenExpired {
				response.Unauthorized(c, "Token has expired")
			} else {
				response.Unauthorized(c, "Invalid token")
			}
			c.Abort()
			return
		}

		// Check if session has been revoked (logout)
		if authSvc.IsSessionRevoked(c.Request.Context(), claims.JTI) {
			response.Unauthorized(c, "Session has been revoked")
			c.Abort()
			return
		}

		c.Set(string(auth.ContextKeyClaims), claims)
		c.Next()
	}
}

// ClaimsFromContext extracts the auth.Claims from the Gin context.
// Returns nil if not present (should not happen behind AuthRequired).
func ClaimsFromContext(c *gin.Context) *auth.Claims {
	v, exists := c.Get(string(auth.ContextKeyClaims))
	if !exists {
		return nil
	}
	claims, _ := v.(*auth.Claims)
	return claims
}
