// 路由初始化
package router

import (
	"github.com/gin-gonic/gin"

	"protocol-parser-server/api"
	"protocol-parser-server/auth"
	"protocol-parser-server/repository/basicinfo"
	"protocol-parser-server/repository/history"
	"protocol-parser-server/repository/rbac"
	"protocol-parser-server/repository/settings"
	"protocol-parser-server/repository/user"
)

func InitRouter(historyStore history.Store, userStore user.Store, settingsStore settings.Store, rbacStore rbac.Store, basicInfoStore basicinfo.Store, tokens *auth.Manager, captchas *auth.CaptchaManager) *gin.Engine {

	r := gin.Default()
	r.Use(api.SecurityMiddleware())
	r.Static("/uploads", "./uploads")

	api.RegisterAuthRouter(r, userStore, rbacStore, tokens, captchas)
	api.RegisterSettingsRouter(r, settingsStore, userStore, tokens)
	api.RegisterAdminRouter(r, rbacStore, userStore, tokens)
	api.RegisterBasicInfoRouter(r, basicInfoStore, userStore, tokens)
	r.Use(api.AuthMiddleware(tokens))
	api.RegisterProtectedParserRouter(r, historyStore, userStore)

	return r

}
