package scheduler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"crm_lite/internal/core/config"
	"crm_lite/internal/core/logger"

	"go.uber.org/zap"
)

// LogCleanupMode 日志清理模式
type LogCleanupMode string

const (
	// InternalScheduler 内置调度器模式 - 适用于本地开发、小团队部署
	InternalScheduler LogCleanupMode = "internal"
	// ExternalScheduler 外部调度器模式 - 适用于多实例部署、正式生产
	ExternalScheduler LogCleanupMode = "external"
)

// LogCleanupConfig 日志清理配置
type LogCleanupConfig struct {
	Mode         LogCleanupMode `yaml:"mode"`         // 清理模式
	Interval     time.Duration  `yaml:"interval"`     // 内置模式的检查间隔
	RetentionDay int            `yaml:"retentionDay"` // 日志保留天数
	LogDir       string         `yaml:"logDir"`       // 日志目录
	DryRun       bool           `yaml:"dryRun"`       // 是否为试运行模式
}

// LogCleaner 日志清理器
type LogCleaner struct {
	config    *LogCleanupConfig
	ctx       context.Context
	cancel    context.CancelFunc
	logger    *logger.Logger
	isRunning bool
}

// NewLogCleaner 创建日志清理器
func NewLogCleaner(config *LogCleanupConfig) *LogCleaner {
	ctx, cancel := context.WithCancel(context.Background())

	// 设置默认值
	if config.Interval == 0 {
		config.Interval = 24 * time.Hour // 默认每天检查一次
	}
	if config.RetentionDay == 0 {
		config.RetentionDay = 7 // 默认保留7天
	}
	if config.LogDir == "" {
		config.LogDir = "logs" // 默认日志目录
	}

	return &LogCleaner{
		config: config,
		ctx:    ctx,
		cancel: cancel,
		logger: logger.GetGlobalLogger(),
	}
}

// Start 启动日志清理器
func (lc *LogCleaner) Start() error {
	if lc.isRunning {
		return fmt.Errorf("日志清理器已在运行中")
	}

	switch lc.config.Mode {
	case InternalScheduler:
		return lc.startInternalScheduler()
	case ExternalScheduler:
		lc.logger.Info("日志清理器配置为外部调度模式，等待外部触发")
		return nil
	default:
		return fmt.Errorf("不支持的日志清理模式: %s", lc.config.Mode)
	}
}

// Stop 停止日志清理器
func (lc *LogCleaner) Stop() {
	if lc.cancel != nil {
		lc.cancel()
	}
	lc.isRunning = false
	lc.logger.Info("日志清理器已停止")
}

// startInternalScheduler 启动内置调度器（适用于本地开发、小团队部署）
func (lc *LogCleaner) startInternalScheduler() error {
	lc.isRunning = true
	lc.logger.Info("启动内置日志清理调度器",
		zap.Duration("间隔", lc.config.Interval),
		zap.Int("保留天数", lc.config.RetentionDay))

	// 立即执行一次清理
	go func() {
		if err := lc.CleanupLogs(); err != nil {
			lc.logger.Error("初始日志清理失败", zap.Error(err))
		}
	}()

	// 启动定时器
	ticker := time.NewTicker(lc.config.Interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-lc.ctx.Done():
				lc.logger.Info("日志清理调度器收到停止信号")
				return
			case <-ticker.C:
				if err := lc.CleanupLogs(); err != nil {
					lc.logger.Error("定时日志清理失败", zap.Error(err))
				}
			}
		}
	}()

	return nil
}

// CleanupLogs 执行日志清理（可被外部调度器调用）
func (lc *LogCleaner) CleanupLogs() error {
	lc.logger.Info("开始清理过期日志",
		zap.String("目录", lc.config.LogDir),
		zap.Int("保留天数", lc.config.RetentionDay),
		zap.Bool("试运行", lc.config.DryRun))

	cutoffTime := time.Now().AddDate(0, 0, -lc.config.RetentionDay)
	deletedCount := 0
	totalSize := int64(0)

	err := filepath.Walk(lc.config.LogDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 只处理日志文件（.log, .log.* 格式）
		if !lc.isLogFile(info.Name()) {
			return nil
		}

		// 检查文件修改时间
		if info.ModTime().Before(cutoffTime) {
			lc.logger.Info("发现过期日志文件",
				zap.String("文件", path),
				zap.Time("修改时间", info.ModTime()),
				zap.Int64("大小", info.Size()))

			totalSize += info.Size()
			deletedCount++

			if !lc.config.DryRun {
				if err := os.Remove(path); err != nil {
					lc.logger.Error("删除日志文件失败",
						zap.String("文件", path),
						zap.Error(err))
					return err
				}
				lc.logger.Info("成功删除过期日志文件", zap.String("文件", path))
			} else {
				lc.logger.Info("试运行模式：将删除日志文件", zap.String("文件", path))
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("清理日志时发生错误: %w", err)
	}

	// 记录清理结果
	lc.logger.Info("日志清理完成",
		zap.Int("删除文件数", deletedCount),
		zap.String("释放空间", lc.formatSize(totalSize)),
		zap.Bool("试运行", lc.config.DryRun))

	return nil
}

// isLogFile 判断是否为日志文件
func (lc *LogCleaner) isLogFile(filename string) bool {
	// 匹配各种日志文件格式：
	// 1. 普通格式: .log 和 .log.*
	// 2. 按时间轮转格式: filename-20250124.log, filename-20250124-14.log
	// 3. 软链接: current.log

	if strings.HasSuffix(filename, ".log") {
		return true
	}

	if strings.Contains(filename, ".log.") {
		return true
	}

	// 检查按时间轮转的文件格式
	// 格式如: dev-20250124.log 或 dev-20250124-14.log
	if strings.Contains(filename, "-") && strings.HasSuffix(filename, ".log") {
		parts := strings.Split(filename, "-")
		if len(parts) >= 2 {
			// 检查最后一部分是否包含日期格式
			lastPart := parts[len(parts)-1]
			if strings.HasSuffix(lastPart, ".log") {
				datePart := strings.TrimSuffix(lastPart, ".log")
				// 检查是否为8位数字日期格式(YYYYMMDD)或2位小时格式
				if len(datePart) == 8 || len(datePart) == 2 {
					for _, char := range datePart {
						if char < '0' || char > '9' {
							return false
						}
					}
					return true
				}
			}
		}
	}

	return false
}

// GetConfig 获取配置
func (lc *LogCleaner) GetConfig() *LogCleanupConfig {
	return lc.config
}

// IsRunning 检查是否正在运行
func (lc *LogCleaner) IsRunning() bool {
	return lc.isRunning
}

// formatSize 格式化文件大小
func (lc *LogCleaner) formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig(env string, logOpts *config.LogOptions) *LogCleanupConfig {
	var mode LogCleanupMode
	var interval time.Duration
	var retentionDay int

	switch env {
	case "dev", "test":
		// 开发/测试环境：内置调度器，较短的保留期
		mode = InternalScheduler
		interval = 1 * time.Hour // 开发环境更频繁检查
		retentionDay = 1
	case "staging":
		// 预发布环境：内置调度器，中等保留期
		mode = InternalScheduler
		interval = 6 * time.Hour
		retentionDay = 3
	case "prod":
		// 生产环境：外部调度器，较长保留期
		mode = ExternalScheduler
		interval = 24 * time.Hour
		retentionDay = 30
	default:
		// 默认：内置调度器
		mode = InternalScheduler
		interval = 24 * time.Hour
		retentionDay = 7
	}

	return &LogCleanupConfig{
		Mode:         mode,
		Interval:     interval,
		RetentionDay: retentionDay,
		LogDir:       logOpts.Dir,
		DryRun:       false,
	}
}
