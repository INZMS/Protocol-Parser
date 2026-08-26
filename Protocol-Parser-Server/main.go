// main.go
package main

import (
	"context"
	"log"
	"time"

	"protocol-parser-server/auth"
	"protocol-parser-server/config"
	"protocol-parser-server/database"
	"protocol-parser-server/protocol"
	"protocol-parser-server/repository/basicinfo"
	"protocol-parser-server/repository/history"
	"protocol-parser-server/repository/rbac"
	"protocol-parser-server/repository/settings"
	"protocol-parser-server/repository/user"
	"protocol-parser-server/router"
)

func main() {
	if err := config.LoadEnvFile(".env"); err != nil {
		log.Fatalf("读取配置失败: %v", err)
	}
	// 注册所有协议
	protocol.RegisterProtocols()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := database.OpenMySQLFromEnv(ctx)
	if err != nil {
		log.Fatalf("初始化MySQL失败: %v", err)
	}
	defer db.Close()
	historyStore := history.NewMySQLStore(db)
	userStore := user.NewMySQLStore(db)
	settingsStore := settings.NewMySQLStore(db)
	rbacStore := rbac.NewMySQLStore(db)
	basicInfoStore := basicinfo.NewMySQLStore(db)
	jobCtx, stopJobs := context.WithCancel(context.Background())
	defer stopJobs()
	basicInfoStore.StartServiceJobs(jobCtx)
	tokenManager := auth.NewManagerFromEnv()
	captchaManager := auth.NewCaptchaManager()

	// 初始化HTTP路由
	r := router.InitRouter(historyStore, userStore, settingsStore, rbacStore, basicInfoStore, tokenManager, captchaManager)

	// 启动服务
	r.Run(":8080")

}
