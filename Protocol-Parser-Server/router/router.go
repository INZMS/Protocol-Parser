// 路由初始化
package router

import (
	"github.com/gin-gonic/gin"

	"protocol-parser-server/api"
	"protocol-parser-server/repository/history"
)

func InitRouter(stores ...history.Store) *gin.Engine {

	r := gin.Default()

	api.RegisterParserRouter(r, stores...)

	return r

}
