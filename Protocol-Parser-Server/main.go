// main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	server := &http.Server{
		Addr:              envOrDefault("HTTP_ADDR", ":8080"),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       90 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("HTTP服务已启动: %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
		close(serverErrors)
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	select {
	case sig := <-signals:
		log.Printf("收到退出信号%s，正在安全停止服务", sig)
	case err := <-serverErrors:
		if err != nil {
			log.Fatalf("HTTP服务启动失败: %v", err)
		}
	}
	stopJobs()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP服务未能在限定时间内安全停止: %v", err)
	}

}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
