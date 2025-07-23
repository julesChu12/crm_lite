package logger

import (
	"os"
	"path/filepath"
	"sync"

	"crm_lite/internal/core/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 封装了 zap logger，提供了统一的日志记录接口
type Logger struct {
	rawLogger     *zap.Logger // 原始 logger，无 skip，供特殊场景使用
	defaultLogger *zap.Logger // 默认 logger，带 skip(2)，供常规使用
	sugaredLogger *zap.SugaredLogger
}

var (
	globalLogger *Logger
	once         sync.Once
)

// InitGlobalLogger 初始化全局日志记录器
func InitGlobalLogger(opts *config.LogOptions) {
	once.Do(func() {
		encoder := getEncoder(opts)
		writer := getWriter(opts)
		core := zapcore.NewCore(encoder, writer, zapcore.Level(opts.Level))

		// 创建原始 logger，仅添加 caller
		rawLogger := zap.New(core, zap.AddCaller())

		// 基于原始 logger 创建带 skip(2) 的默认 logger，以跳过包函数和Logger方法这两层封装
		defaultLogger := rawLogger.WithOptions(zap.AddCallerSkip(2))

		globalLogger = &Logger{
			rawLogger:     rawLogger,
			defaultLogger: defaultLogger,
			sugaredLogger: defaultLogger.Sugar(), // Sugared Logger 通常也用于常规场景
		}
	})
}

// GetGlobalLogger 获取全局日志记录器实例
func GetGlobalLogger() *Logger {
	if globalLogger == nil {
		// 在测试等场景未显式初始化时，使用一个简易的默认 logger，避免空指针。
		raw := zap.NewExample()
		globalLogger = &Logger{
			rawLogger:     raw,
			defaultLogger: raw,
			sugaredLogger: raw.Sugar(),
		}
	}
	return globalLogger
}

// Raw 返回底层的原始 *zap.Logger，用于需要自定义 Options 的特殊场景（如中间件）
func (log *Logger) Raw() *zap.Logger {
	return log.rawLogger
}

// Sugared 返回底层的 *zap.SugaredLogger
func (log *Logger) Sugared() *zap.SugaredLogger {
	return log.sugaredLogger
}

// Info 记录 Info 级别的日志
func (log *Logger) Info(msg string, fields ...zap.Field) {
	log.defaultLogger.Info(msg, fields...)
}

// Warn 记录 Warn 级别的日志
func (log *Logger) Warn(msg string, fields ...zap.Field) {
	log.defaultLogger.Warn(msg, fields...)
}

// Error 记录 Error 级别的日志
func (log *Logger) Error(msg string, fields ...zap.Field) {
	log.defaultLogger.Error(msg, fields...)
}

// Fatal 记录 Fatal 级别的日志，然后退出程序
func (log *Logger) Fatal(msg string, fields ...zap.Field) {
	log.defaultLogger.Fatal(msg, fields...)
}

func getEncoder(opts *config.LogOptions) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 根据配置决定是否显示调用者信息
	if opts.LineNum {
		encoderConfig.CallerKey = "caller"
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getWriter(opts *config.LogOptions) zapcore.WriteSyncer {
	lumberjackSyncer := &lumberjack.Logger{
		Filename:   filepath.Join(opts.Dir, opts.Filename),
		MaxSize:    opts.MaxSize,
		MaxBackups: opts.MaxBackups,
		MaxAge:     opts.MaxAge,
		Compress:   opts.Compress,
	}
	// 同时输出到控制台和文件
	return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lumberjackSyncer))
}

// ============== 包级快捷函数，便于直接调用 ==============

// Info 记录 Info 级别日志（业务代码可直接调用 logger.Info）
func Info(msg string, fields ...zap.Field) {
	GetGlobalLogger().Info(msg, fields...)
}

// Warn 记录 Warn 级别日志
func Warn(msg string, fields ...zap.Field) {
	GetGlobalLogger().Warn(msg, fields...)
}

// Error 记录 Error 级别日志
func Error(msg string, fields ...zap.Field) {
	GetGlobalLogger().Error(msg, fields...)
}

// Fatal 记录 Fatal 级别日志（记录后程序将退出）
func Fatal(msg string, fields ...zap.Field) {
	GetGlobalLogger().Fatal(msg, fields...)
}
