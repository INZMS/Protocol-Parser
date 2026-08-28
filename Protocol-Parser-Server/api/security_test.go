package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestSecurityMiddlewareRejectsUnknownPreflightOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("GIN_MODE", "release")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://allowed.example")
	router := gin.New()
	router.Use(SecurityMiddleware())
	router.GET("/ok", func(c *gin.Context) { c.Status(http.StatusOK) })

	request := httptest.NewRequest(http.MethodOptions, "/ok", nil)
	request.Header.Set("Origin", "https://unknown.example")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(NewRateLimitMiddleware(1, time.Minute))
	router.GET("/ok", func(c *gin.Context) { c.Status(http.StatusOK) })

	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/ok", nil))
	second := httptest.NewRecorder()
	router.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/ok", nil))
	if first.Code != http.StatusOK || second.Code != http.StatusTooManyRequests {
		t.Fatalf("unexpected statuses: first=%d second=%d", first.Code, second.Code)
	}
}

func TestPreferenceKeyValidation(t *testing.T) {
	valid := []string{"list-columns-vehicles", "table.columns.device", "page:size"}
	for _, key := range valid {
		if !validPreferenceKey(key) {
			t.Fatalf("expected valid preference key %q", key)
		}
	}
	invalid := []string{"", "../secret", "white space", string(make([]byte, 129))}
	for _, key := range invalid {
		if validPreferenceKey(key) {
			t.Fatalf("expected invalid preference key %q", key)
		}
	}
}
