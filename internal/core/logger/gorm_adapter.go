package logger

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
)

// GormLoggerAdapter 实现了 gorm.io/gorm/logger.Interface 接口
type GormLoggerAdapter struct {
	ZapLogger *zap.Logger
	LogLevel  gormlogger.LogLevel
}

// NewGormLoggerAdapter 创建一个新的GORM日志适配器
func NewGormLoggerAdapter() gormlogger.Interface {
	// 为GORM创建专门的logger，调整调用深度以显示实际的调用位置
	// 注意：这里我们使用 Raw() 获取原始 logger，因为我们需要完全控制 CallerSkip
	gormLogger := GetGlobalLogger().Raw().WithOptions(zap.AddCallerSkip(4))

	return &GormLoggerAdapter{
		ZapLogger: gormLogger,
		LogLevel:  gormlogger.Warn, // 默认日志级别
	}
}

func (l *GormLoggerAdapter) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *GormLoggerAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		l.ZapLogger.Sugar().Infof(msg, data...)
	}
}

func (l *GormLoggerAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		l.ZapLogger.Sugar().Warnf(msg, data...)
	}
}

func (l *GormLoggerAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		l.ZapLogger.Sugar().Errorf(msg, data...)
	}
}

func (l *GormLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && !errors.Is(err, gormlogger.ErrRecordNotFound):
		sql, rows := fc()
		l.ZapLogger.Error("gorm_trace",
			zap.Error(err),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	case elapsed > 200*time.Millisecond && l.LogLevel >= gormlogger.Warn:
		sql, rows := fc()
		l.ZapLogger.Warn("gorm_trace_slow",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	case l.LogLevel >= gormlogger.Info:
		sql, rows := fc()
		l.ZapLogger.Info("gorm_trace",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	}
}
