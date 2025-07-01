package bootstrap

import (
	"context"
	"fmt"
	"time"

	"crm_lite/internal/core/config"
	"crm_lite/internal/core/logger"
	"crm_lite/internal/core/resource"

	"go.uber.org/zap"
)

// Bootstrap an application
func Bootstrap() (*resource.Manager, func(), error) {
	// 1. 创建资源管理器
	resManager := resource.NewManager()

	// 2. 获取全局配置
	opts := config.GetInstance()

	// 3. 初始化全局 Logger (这是第一步，因为所有后续步骤都可能需要日志)
	logger.InitGlobalLogger(&opts.Logger)

	// 4. 创建并注册资源
	// - 数据库
	dbResource := resource.NewDBResource(opts.Database)
	if err := resManager.Register(resource.DBServiceKey, dbResource); err != nil {
		return nil, nil, fmt.Errorf("failed to register db resource: %w", err)
	}

	// - 缓存
	cacheResource := resource.NewCacheResource(opts.Cache)
	if err := resManager.Register(resource.CacheServiceKey, cacheResource); err != nil {
		return nil, nil, fmt.Errorf("failed to register cache resource: %w", err)
	}

	// - Casbin 权限
	casbinResource := resource.NewCasbinResource(resManager, opts.Auth.RbacOptions)
	if err := resManager.Register(resource.CasbinServiceKey, casbinResource); err != nil {
		return nil, nil, fmt.Errorf("failed to register casbin resource: %w", err)
	}

	// - 邮件服务
	emailResource := resource.NewEmailResource(opts.Email)
	if err := resManager.Register(resource.EmailServiceKey, emailResource); err != nil {
		return nil, nil, fmt.Errorf("failed to register email resource: %w", err)
	}

	// ... 未来可以在这里注册更多的资源, e.g., Message Queue, Tracer ...

	// 5. 初始化所有已注册的资源
	logger.Info("Starting to initialize all resources...")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := resManager.InitAll(ctx); err != nil {
		// 如果初始化失败，尝试关闭已经初始化的资源
		logger.Error("Failed to initialize resources, attempting cleanup...", zap.Error(err))
		if closeErr := resManager.CloseAll(context.Background()); closeErr != nil {
			logger.Error("Failed to cleanup resources after initialization failure", zap.Error(closeErr))
		}
		return nil, nil, err
	}

	logger.Info("All resources initialized successfully")

	// 6. 初始化超级管理员与 Casbin
	if err := initSuperAdmin(resManager); err != nil {
		logger.Error("Failed to init super admin", zap.Error(err))
	}

	// 7. 创建并返回一个优雅关闭的函数
	cleanup := func() {
		logger.Info("Starting to close all resources...")
		if err := resManager.CloseAll(context.Background()); err != nil {
			logger.Error("Failed to close resources gracefully", zap.Error(err))
		} else {
			logger.Info("All resources closed successfully")
		}
	}

	return resManager, cleanup, nil
}
