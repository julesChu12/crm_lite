package controller

import (
	"strconv"

	"crm_lite/pkg/resp"
	"crm_lite/pkg/scheduler"

	"github.com/gin-gonic/gin"
)

// MaintenanceController 维护相关的控制器
type MaintenanceController struct {
	logCleaner *scheduler.LogCleaner
}

// NewMaintenanceController 创建维护控制器
func NewMaintenanceController(logCleaner *scheduler.LogCleaner) *MaintenanceController {
	return &MaintenanceController{
		logCleaner: logCleaner,
	}
}

// CleanupLogs 手动触发日志清理
// @Summary 手动触发日志清理
// @Description 手动执行日志清理操作，通常由外部调度器调用
// @Tags 维护管理
// @Accept json
// @Produce json
// @Param dry_run query bool false "是否为试运行模式"
// @Success 200 {object} resp.Response "清理成功"
// @Failure 500 {object} resp.Response "清理失败"
// @Router /api/maintenance/logs/cleanup [post]
func (mc *MaintenanceController) CleanupLogs(c *gin.Context) {
	// 解析查询参数
	dryRun := false
	if dryRunStr := c.Query("dry_run"); dryRunStr != "" {
		if parsed, err := strconv.ParseBool(dryRunStr); err == nil {
			dryRun = parsed
		}
	}

	// 如果是试运行模式，临时修改配置
	originalDryRun := mc.logCleaner.GetConfig().DryRun
	if dryRun {
		mc.logCleaner.GetConfig().DryRun = true
		defer func() {
			mc.logCleaner.GetConfig().DryRun = originalDryRun
		}()
	}

	// 执行日志清理
	if err := mc.logCleaner.CleanupLogs(); err != nil {
		resp.Error(c, resp.CodeInternalError, "日志清理失败: "+err.Error())
		return
	}

	message := "日志清理完成"
	if dryRun {
		message = "日志清理试运行完成"
	}

	resp.Success(c, gin.H{"message": message})
}

// GetLogCleanupStatus 获取日志清理器状态
// @Summary 获取日志清理器状态
// @Description 获取当前日志清理器的运行状态和配置信息
// @Tags 维护管理
// @Accept json
// @Produce json
// @Success 200 {object} resp.Response "获取成功"
// @Router /api/maintenance/logs/status [get]
func (mc *MaintenanceController) GetLogCleanupStatus(c *gin.Context) {
	config := mc.logCleaner.GetConfig()

	status := gin.H{
		"mode":         string(config.Mode),
		"interval":     config.Interval.String(),
		"retentionDay": config.RetentionDay,
		"logDir":       config.LogDir,
		"dryRun":       config.DryRun,
		"isRunning":    mc.logCleaner.IsRunning(),
	}

	resp.Success(c, status)
}
