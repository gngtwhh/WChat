package main

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
