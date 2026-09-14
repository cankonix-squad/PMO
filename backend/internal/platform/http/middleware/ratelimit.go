package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LoginRateLimit provides a small bounded in-process limiter for public auth
// endpoints. Production deployments should back this with Redis for a shared
// limit across replicas.
func LoginRateLimit(limit int, window time.Duration) gin.HandlerFunc {
	type entry struct {
		count int
		reset time.Time
	}
	var mu sync.Mutex
	clients := make(map[string]entry)
	return func(c *gin.Context) {
		key := c.ClientIP()
		now := time.Now()
		mu.Lock()
		e := clients[key]
		if e.reset.Before(now) {
			e = entry{reset: now.Add(window)}
		}
		e.count++
		clients[key] = e
		mu.Unlock()
		if e.count > limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"success": false, "message": "too many requests"})
			return
		}
		c.Next()
	}
}
