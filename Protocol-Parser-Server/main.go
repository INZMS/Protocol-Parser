// main.go
package main

import (
	"context"
	"log"
	"time"

	"protocol-parser-server/config"
	"protocol-parser-server/database"
	"protocol-parser-server/protocol"
	"protocol-parser-server/repository/history"
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

	// 初始化HTTP路由
	r := router.InitRouter(historyStore)

	// 启动服务
	r.Run(":8080")

}
