package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"protocol-parser-server/auth"
	"protocol-parser-server/repository/settings"
	"protocol-parser-server/repository/user"
)

func RegisterSettingsRouter(r *gin.Engine, store settings.Store, users user.Store, tokens *auth.Manager) {
	r.GET("/api/settings/login-page", func(c *gin.Context) {
		value, err := store.GetLoginPage(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "读取登录页设置失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "settings": value})
	})

	protected := r.Group("/api/settings")
	protected.Use(AuthMiddleware(tokens))
	protected.PUT("/login-page", requirePermission(users, "system:settings:save"), func(c *gin.Context) {
		_, err := users.GetByID(c.Request.Context(), currentUserID(c))
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "无权修改系统设置"})
			return
		}
		var request settings.LoginPage
		if c.ShouldBindJSON(&request) != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "设置内容格式错误"})
			return
		}
		if err := store.SaveLoginPage(c.Request.Context(), request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "settings": request})
	})
	protected.GET("/preferences/:key", func(c *gin.Context) {
		value, err := store.GetUserPreference(c.Request.Context(), currentUserID(c), c.Param("key"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "读取个人设置失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "value": json.RawMessage(value)})
	})
	protected.PUT("/preferences/:key", func(c *gin.Context) {
		var request struct {
			Value json.RawMessage `json:"value"`
		}
		if c.ShouldBindJSON(&request) != nil || !json.Valid(request.Value) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "个人设置格式错误"})
			return
		}
		if err := store.SaveUserPreference(c.Request.Context(), currentUserID(c), c.Param("key"), request.Value); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "value": request.Value})
	})
}
