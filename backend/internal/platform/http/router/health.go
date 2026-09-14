package router

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// RegisterHealth mounts liveness endpoints. These endpoints deliberately do
// not depend on authentication or domain handlers.
func RegisterHealth(r *gin.Engine, version, environment string) {
	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "version": version, "env": environment})
	}
	r.GET("/health", handler)
	r.GET("/api/health", handler)
}
