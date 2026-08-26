package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"protocol-parser-server/auth"
	"protocol-parser-server/repository/user"
)

const userIDContextKey = "authenticatedUserID"

type loginRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	CaptchaID   string `json:"captchaId"`
	CaptchaCode string `json:"captchaCode"`
}
type profileRequest struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
}

func RegisterAuthRouter(r *gin.Engine, store user.Store, tokens *auth.Manager, captchas *auth.CaptchaManager) {
	group := r.Group("/api/auth")
	group.GET("/captcha", func(c *gin.Context) {
		result, err := captchas.Create()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "生成验证码失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "captcha": result})
	})
	group.POST("/login", func(c *gin.Context) {
		var request loginRequest
		if c.ShouldBindJSON(&request) != nil || strings.TrimSpace(request.Username) == "" || request.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "请输入用户名和密码"})
			return
		}
		if !captchas.Verify(request.CaptchaID, request.CaptchaCode) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "验证码错误或已过期"})
			return
		}
		result, err := store.Authenticate(c.Request.Context(), strings.TrimSpace(request.Username), request.Password)
		if err != nil {
			status := http.StatusInternalServerError
			if err == user.ErrInvalidCredentials {
				status = http.StatusUnauthorized
			}
			c.JSON(status, gin.H{"success": false, "error": err.Error()})
			return
		}
		token, err := tokens.Issue(result.ID, result.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "生成登录令牌失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "token": token, "user": result})
	})
	protected := group.Group("")
	protected.Use(AuthMiddleware(tokens))
	protected.GET("/me", func(c *gin.Context) {
		result, err := store.GetByID(c.Request.Context(), currentUserID(c))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "用户不存在或已禁用"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "user": result})
	})
	protected.PUT("/profile", func(c *gin.Context) {
		var request profileRequest
		if c.ShouldBindJSON(&request) != nil || strings.TrimSpace(request.DisplayName) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "显示名称不能为空"})
			return
		}
		result, err := store.UpdateProfile(c.Request.Context(), currentUserID(c), strings.TrimSpace(request.DisplayName), strings.TrimSpace(request.Email), strings.TrimSpace(request.Phone))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "更新个人资料失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "user": result})
	})
	protected.POST("/logout", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"success": true}) })
}

func AuthMiddleware(tokens *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "请先登录"})
			return
		}
		claims, err := tokens.Verify(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": err.Error()})
			return
		}
		c.Set(userIDContextKey, claims.UserID)
		c.Next()
	}
}
func currentUserID(c *gin.Context) int64 {
	value, _ := c.Get(userIDContextKey)
	id, _ := value.(int64)
	return id
}
