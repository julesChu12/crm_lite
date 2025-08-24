package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"

	"github.com/gin-gonic/gin"
)

// RegisterDashboardRoutes 注册工作台相关路由
func RegisterDashboardRoutes(rg *gin.RouterGroup, resManager *resource.Manager) {
	dashboardController := controller.NewDashboardController(resManager)

	dashboard := rg.Group("/dashboard")
	{
		dashboard.GET("/overview", dashboardController.Overview)
	}
}
