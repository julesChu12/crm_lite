package logger

import (
	"testing"

	"crm_lite/internal/core/config"

	"github.com/stretchr/testify/assert"
)

func TestGlobalLogger(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 定义配置
	opts := &config.LogOptions{
		Dir:      tempDir,
		Filename: "test.log",
		Level:    -1, // Debug level
		MaxSize:  1,  // 1 MB
		MaxAge:   1,  // 1 day
		Compress: false,
	}

	// 初始化全局Logger
	InitGlobalLogger(opts)

	// 获取全局Logger
	logger := GetGlobalLogger()

	// 断言
	assert.NotNil(t, logger, "全局logger不应为nil")
	assert.NotNil(t, logger.rawLogger, "内部的原始 zap.Logger 不应为nil")
	assert.NotNil(t, logger.defaultLogger, "内部的默认 zap.Logger 不应为nil")
	assert.NotNil(t, logger.sugaredLogger, "内部的zap.SugaredLogger不应为nil")

	// 尝试记录一条日志
	Info("这是一条测试日志")

	// 验证日志文件是否被创建
	// 注意：由于日志是异步写入的，这里不直接断言文件内容，
	// 只需确保调用不产生panic，并且logger实例有效即可。
	// 在真实的集成测试中，可以加入短暂延时或更复杂的同步机制来验证文件写入。
}
