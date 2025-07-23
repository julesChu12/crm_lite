package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"

	"github.com/gin-gonic/gin"
)

// RegisterDashboardRoutes 注册Dashboard工作台路由
func RegisterDashboardRoutes(r *gin.RouterGroup, res *resource.Manager) {
	dashboardController := controller.NewDashboardController(res)

	// Dashboard工作台路由组
	dashboard := r.Group("/dashboard")
	{
		// 总览数据
		dashboard.GET("/overview", dashboardController.GetOverview) // 获取工作台总览数据

		// 快速统计（实时数据）
		dashboard.GET("/quick-stats", dashboardController.GetQuickStats) // 获取快速统计数据

		// 分析数据路由组
		analytics := dashboard.Group("/analytics")
		{
			analytics.GET("/customers", dashboardController.GetCustomerAnalytics)  // 获取客户分析数据
			analytics.GET("/sales", dashboardController.GetSalesAnalytics)         // 获取销售分析数据
			analytics.GET("/products", dashboardController.GetProductAnalytics)    // 获取产品分析数据
			analytics.GET("/marketing", dashboardController.GetMarketingAnalytics) // 获取营销分析数据
			analytics.GET("/financial", dashboardController.GetFinancialAnalytics) // 获取财务分析数据
		}

		// 活动相关数据
		dashboard.GET("/activities", dashboardController.GetActivitySummary) // 获取活动摘要数据
	}
}
