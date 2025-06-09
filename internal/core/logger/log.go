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

// Logger aotemanager a basic logger
type Logger struct {
	*zap.Logger
	SugaredLogger *zap.SugaredLogger
}

var (
	globalLogger *Logger
	once         sync.Once
)

// InitGlobalLogger init a global logger
func InitGlobalLogger(opts *config.LogOptions) {
	once.Do(func() {
		encoder := getEncoder()
		writer := getWriter(opts)
		core := zapcore.NewCore(encoder, writer, zapcore.Level(opts.Level))
		zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

		globalLogger = &Logger{
			Logger:        zapLogger,
			SugaredLogger: zapLogger.Sugar(),
		}
	})
}

// GetGlobalLogger get a global logger
func GetGlobalLogger() *Logger {
	return globalLogger
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
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
