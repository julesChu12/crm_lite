package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/pkg/scheduler"

	"github.com/gin-gonic/gin"
)

// SetupMaintenanceRoutes 设置维护相关路由
func SetupMaintenanceRoutes(router *gin.RouterGroup, logCleaner *scheduler.LogCleaner) {
	maintenanceController := controller.NewMaintenanceController(logCleaner)

	maintenance := router.Group("/maintenance")
	{
		logs := maintenance.Group("/logs")
		{
			logs.POST("/cleanup", maintenanceController.CleanupLogs)
			logs.GET("/status", maintenanceController.GetLogCleanupStatus)
		}
	}
}
