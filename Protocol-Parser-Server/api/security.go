package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityMiddleware provides a request id, conservative browser security
// headers and an explicit CORS allow-list. It must be installed before routes.
func SecurityMiddleware() gin.HandlerFunc {
	allowed := make(map[string]struct{})
	raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if raw == "" && os.Getenv("GIN_MODE") != "release" {
		raw = "http://localhost:5173,http://localhost:5174,http://127.0.0.1:5173,http://127.0.0.1:5174"
	}
	for _, origin := range strings.Split(raw, ",") {
		if value := strings.TrimSpace(origin); value != "" {
			allowed[value] = struct{}{}
		}
	}
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if requestID == "" {
			var bytes [12]byte
			if _, err := rand.Read(bytes[:]); err == nil {
				requestID = hex.EncodeToString(bytes[:])
			} else {
				requestID = time.Now().UTC().Format("20060102150405.000000000")
			}
		}
		c.Header("X-Request-ID", requestID)
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "same-origin")
		c.Header("X-Frame-Options", "SAMEORIGIN")

		origin := c.GetHeader("Origin")
		if _, ok := allowed[origin]; origin != "" && ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			if origin != "" {
				if _, ok := allowed[origin]; !ok {
					c.AbortWithStatus(http.StatusForbidden)
					return
				}
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

type rateBucket struct {
	count   int
	expires time.Time
}

// NewRateLimitMiddleware returns a small in-memory limiter suitable for
// sensitive, low-volume endpoints such as login and captcha generation.
func NewRateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	buckets := make(map[string]rateBucket)
	return func(c *gin.Context) {
		now := time.Now()
		key := c.ClientIP()
		mu.Lock()
		bucket := buckets[key]
		if now.After(bucket.expires) {
			bucket = rateBucket{expires: now.Add(window)}
		}
		bucket.count++
		buckets[key] = bucket
		blocked := bucket.count > limit
		mu.Unlock()
		if blocked {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"success": false, "error": "请求过于频繁，请稍后再试"})
			return
		}
		c.Next()
	}
}
