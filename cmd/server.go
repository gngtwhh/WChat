package main

// @title           WChat API
// @version         1.0
// @description     WChat 即时通讯系统后端 API 文档
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wchat/internal/app/http_server"
	"wchat/pkg/zlog"

	"go.uber.org/zap"
)

func main() {
	server := http_server.NewServer()

	go func() {
		server.Run()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	zlog.Info("http_server closing")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		zlog.Error("http_server shutdown failed", zap.Error(err))
	}
	zlog.Info("http_server closed")
}
