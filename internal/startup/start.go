package startup

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"crm_lite/internal/core/config"
	"crm_lite/internal/core/logger"
	"crm_lite/pkg/process"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Start 启动HTTP服务器并设置优雅关闭
func Start(router *gin.Engine, cleanup func()) {
	opts := config.GetInstance()

	// 写入PID文件
	if err := process.WritePIDFile(opts.Server.PidFile); err != nil {
		logger.Fatal("Failed to write PID file", zap.Error(err))
	}

	// 1. 创建并启动HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", opts.Server.Host, opts.Server.Port),
		Handler: router,
	}

	go func() {
		logger.GetGlobalLogger().Raw().Info("Server is starting...", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 2. 等待中断信号以进行优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Server is shutting down...")

	// 在程序退出时清理PID文件
	process.CleanupPIDFile(opts.Server.PidFile)

	// 执行传递进来的cleanup函数，关闭所有资源
	if cleanup != nil {
		cleanup()
	}

	// 设置一个超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", zap.Error(err))
	}

	logger.Info("Server exited properly")
}
