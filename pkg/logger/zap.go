package logger

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Log 是全局可用的 SugaredLogger，需先通过 Init() 初始化
var Log *zap.SugaredLogger

// Init 初始化全局 Zap 日志。
// level: 支持 "debug" "info" "warn" "error"
// filePath: 若非空，则日志同时写入文件（JSON 格式）；为空则只输出到控制台
func Init(level string, filePath string) {
	if Log != nil {
		// 已经初始化过，避免重复
		return
	}

	// 解析日志级别
	var lvl zapcore.Level
	switch level {
	case "debug":
		lvl = zap.DebugLevel
	case "warn":
		lvl = zap.WarnLevel
	case "error":
		lvl = zap.ErrorLevel
	default:
		lvl = zap.InfoLevel
	}

	// 公共 Encoder 配置
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	encCfg.EncodeCaller = zapcore.ShortCallerEncoder
	encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder // 控制台彩色

	consoleEncoder := zapcore.NewConsoleEncoder(encCfg)

	var cores []zapcore.Core
	// 控制台核心
	cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), lvl))

	// 文件核心（JSON 编码，带按大小滚动）
	if filePath != "" {
		_ = os.MkdirAll(filepath.Dir(filePath), 0o755)

		// 使用 lumberjack 控制文件大小、备份和保留时间
		lj := &lumberjack.Logger{
			Filename:   filePath,
			MaxSize:    100, // 每个文件最大 100MB
			MaxBackups: 7,   // 最多保留 7 个备份
			MaxAge:     14,  // 保留 14 天
			Compress:   true,
		}

		fileEncCfg := encCfg
		fileEncCfg.EncodeLevel = zapcore.CapitalLevelEncoder // 文件用普通大写级别
		fileEncoder := zapcore.NewJSONEncoder(fileEncCfg)

		cores = append(cores, zapcore.NewCore(fileEncoder, zapcore.AddSync(lj), lvl))
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	Log = logger.Sugar()
	// 把标准库 log 输出重定向到 Zap
	writer := zap.NewStdLog(Log.Desugar())
	log.SetOutput(writer.Writer())
}

// Sync 在程序退出前调用，确保日志刷盘
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
