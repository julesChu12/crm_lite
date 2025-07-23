package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"

	"github.com/gin-gonic/gin"
)

// RegisterMarketingRoutes 注册营销模块路由
func RegisterMarketingRoutes(r *gin.RouterGroup, res *resource.Manager) {
	marketingController := controller.NewMarketingController(res)

	// 营销模块路由组
	marketing := r.Group("/marketing")
	{
		// 营销活动管理路由
		campaigns := marketing.Group("/campaigns")
		{
			campaigns.POST("", marketingController.CreateCampaign)              // 创建营销活动
			campaigns.GET("", marketingController.ListCampaigns)                // 获取营销活动列表
			campaigns.GET("/:id", marketingController.GetCampaign)              // 获取单个营销活动
			campaigns.PUT("/:id", marketingController.UpdateCampaign)           // 更新营销活动
			campaigns.DELETE("/:id", marketingController.DeleteCampaign)        // 删除营销活动
			campaigns.POST("/:id/execute", marketingController.ExecuteCampaign) // 执行营销活动
			campaigns.GET("/:id/stats", marketingController.GetCampaignStats)   // 获取营销活动统计
		}

		// 营销记录管理路由
		marketing.GET("/records", marketingController.ListMarketingRecords) // 获取营销记录列表

		// 客户分群管理路由
		marketing.POST("/customer-segments", marketingController.GetCustomerSegment) // 获取客户分群
	}
}
