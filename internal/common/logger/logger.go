package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger

// Init 初始化 zap 日志记录器
// level: 日志级别 (debug, info, warn, error, panic, fatal)
// filePath: 日志文件路径，如果为空则输出到控制台
func Init(level string, filePath string) {
	var zapLevel zapcore.Level
	err := zapLevel.UnmarshalText([]byte(level))
	if err != nil {
		zapLevel = zapcore.InfoLevel // 默认级别
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var core zapcore.Core
	if filePath != "" {
		fileWriter := zapcore.AddSync(&zapcore.BufferedWriteSyncer{
			WS: zapcore.AddSync(mustOpenFile(filePath)),
			//BufferSize: 256 * 1024, // 256KB
			//FlushInterval: 30 * time.Second,
		})
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			fileWriter,
			zapLevel,
		)
	} else {
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.Lock(os.Stdout),
			zapLevel,
		)
	}

	rawLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)) // AddCallerSkip(1) 以便封装后调用栈正确
	Log = rawLogger.Sugar()
	Log.Info("Logger initialized", "level", zapLevel.String(), "filePath", filePath)
}

func mustOpenFile(filePath string) *os.File {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic("Failed to open log file: " + err.Error())
	}
	return file
}

// Sync flushes any buffered log entries.
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
