package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"crm_lite/internal/core/config"
	"crm_lite/internal/core/logger"
	"crm_lite/internal/core/resource"
	"crm_lite/pkg/scheduler"
	"crm_lite/pkg/validator"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Bootstrap 初始化应用程序，返回资源管理器、日志清理器和清理函数
func Bootstrap() (*resource.Manager, *scheduler.LogCleaner, func(), error) {
	// 1. 初始化配置
	opts := config.GetInstance()
	if opts == nil {
		return nil, nil, nil, errors.New("failed to get configuration instance")
	}

	// 2. 初始化日志记录器
	// 检查配置是否正确读取，如果没有则手动从配置文件设置
	if !opts.Logger.EnableTimeRotation {
		// 从原始配置中读取
		if opts.LogCleanup.Mode == "" && viper.GetBool("logger.enableTimeRotation") {
			opts.Logger.EnableTimeRotation = viper.GetBool("logger.enableTimeRotation")
			opts.Logger.RotationTime = viper.GetString("logger.rotationTime")
			opts.Logger.LinkName = viper.GetString("logger.linkName")
			fmt.Printf("Fixed config from viper: EnableTimeRotation=%v, RotationTime=%s\n",
				opts.Logger.EnableTimeRotation, opts.Logger.RotationTime)
		}
	}
	logger.InitGlobalLogger(&opts.Logger)

	// 注册自定义验证器
	validator.RegisterMobileValidator()

	logger.Info("Bootstrap process started", zap.String("env", getEnvFromConfig()))

	// 3. 初始化资源管理器
	resManager := resource.NewManager()

	// 初始化各个资源
	if err := initResources(resManager, opts); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize resources: %w", err)
	}

	// 4. 初始化日志清理器
	logCleanupConfig := convertToSchedulerConfig(&opts.LogCleanup, &opts.Logger)
	logCleaner := scheduler.NewLogCleaner(logCleanupConfig)

	// 启动日志清理器
	if err := logCleaner.Start(); err != nil {
		logger.Error("Failed to start log cleaner", zap.Error(err))
		// 日志清理器启动失败不应该导致整个应用启动失败
	} else {
		logger.Info("Log cleaner started successfully",
			zap.String("mode", string(logCleanupConfig.Mode)),
			zap.Duration("interval", logCleanupConfig.Interval))
	}

	// 5. 初始化管理员用户和权限系统
	if err := initSuperAdmin(resManager); err != nil {
		log.Printf("Warning: Failed to initialize admin user: %v", err)
	}

	// 6. 定义清理函数
	cleanup := func() {
		logger.Info("Application is shutting down...")

		// 停止日志清理器
		if logCleaner != nil {
			logCleaner.Stop()
		}

		// 关闭资源管理器
		if resManager != nil {
			if err := resManager.CloseAll(context.Background()); err != nil {
				logger.Error("Failed to close resources gracefully", zap.Error(err))
			}
		}

		logger.Info("All resources cleaned up successfully")
	}

	logger.Info("Bootstrap process completed successfully")
	return resManager, logCleaner, cleanup, nil
}

// initResources 初始化所有资源
func initResources(resManager *resource.Manager, opts *config.Options) error {
	// 创建并注册数据库资源
	dbResource := resource.NewDBResource(opts.Database)
	if err := resManager.Register(resource.DBServiceKey, dbResource); err != nil {
		return fmt.Errorf("failed to register db resource: %w", err)
	}

	// 创建并注册缓存资源
	cacheResource := resource.NewCacheResource(opts.Cache)
	if err := resManager.Register(resource.CacheServiceKey, cacheResource); err != nil {
		return fmt.Errorf("failed to register cache resource: %w", err)
	}

	// 创建并注册Casbin资源
	casbinResource := resource.NewCasbinResource(resManager, opts.Auth.RbacOptions)
	if err := resManager.Register(resource.CasbinServiceKey, casbinResource); err != nil {
		return fmt.Errorf("failed to register casbin resource: %w", err)
	}

	// 初始化所有已注册的资源
	logger.Info("Starting to initialize all resources...")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := resManager.InitAll(ctx); err != nil {
		logger.Error("Failed to initialize resources, attempting cleanup...", zap.Error(err))
		if closeErr := resManager.CloseAll(context.Background()); closeErr != nil {
			logger.Error("Failed to cleanup resources after initialization failure", zap.Error(closeErr))
		}
		return err
	}

	logger.Info("All resources initialized successfully")
	return nil
}

// convertToSchedulerConfig 将配置转换为调度器配置
func convertToSchedulerConfig(logCleanupOpts *config.LogCleanupOptions, logOpts *config.LogOptions) *scheduler.LogCleanupConfig {
	// 如果没有配置，使用默认配置
	if logCleanupOpts.Mode == "" {
		env := getEnvFromConfig()
		return scheduler.GetDefaultConfig(env, logOpts)
	}

	mode := scheduler.InternalScheduler
	if logCleanupOpts.Mode == "external" {
		mode = scheduler.ExternalScheduler
	}

	interval := logCleanupOpts.Interval
	if interval == 0 {
		interval = 24 * time.Hour // 默认24小时
	}

	retentionDay := logCleanupOpts.RetentionDay
	if retentionDay == 0 {
		retentionDay = 7 // 默认保留7天
	}

	return &scheduler.LogCleanupConfig{
		Mode:         mode,
		Interval:     interval,
		RetentionDay: retentionDay,
		LogDir:       logOpts.Dir,
		DryRun:       logCleanupOpts.DryRun,
	}
}

// getEnvFromConfig 从配置中获取环境信息
func getEnvFromConfig() string {
	// 从环境变量或配置文件中获取环境信息
	// 这里简单从Logger的文件名推断环境
	opts := config.GetInstance()
	switch opts.Logger.Filename {
	case "dev.log":
		return "dev"
	case "test.log":
		return "test"
	case "prod.log":
		return "prod"
	}
	return "dev" // 默认开发环境
}
