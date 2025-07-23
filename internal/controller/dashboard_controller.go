package controller

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"

	"github.com/gin-gonic/gin"
)

// DashboardController 负责处理工作台数据相关的HTTP请求
type DashboardController struct {
	dashboardService *service.DashboardService
}

// NewDashboardController 创建一个新的DashboardController实例
func NewDashboardController(resManager *resource.Manager) *DashboardController {
	return &DashboardController{
		dashboardService: service.NewDashboardService(resManager),
	}
}

// GetOverview godoc
// @Summary      获取工作台总览数据
// @Description  获取CRM系统工作台的总览统计数据，包括客户、订单、收入等核心指标
// @Tags         Dashboard
// @Produce      json
// @Param        date_range query string false "时间范围" Enums(today,week,month,quarter,year) default(month)
// @Param        timezone query string false "时区" default(Asia/Shanghai)
// @Success      200 {object} resp.Response{data=dto.DashboardOverviewResponse} "获取成功"
// @Failure      400 {object} resp.Response "请求参数错误"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /dashboard/overview [get]
func (dc *DashboardController) GetOverview(c *gin.Context) {
	var req dto.DashboardRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}

	overview, err := dc.dashboardService.GetOverview(c.Request.Context(), &req)
	if err != nil {
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, overview)
}

// GetCustomerAnalytics godoc
// @Summary      获取客户分析数据
// @Description  获取客户相关的分析数据，包括客户等级分布、来源分布、增长趋势等
// @Tags         Dashboard
// @Produce      json
// @Success      200 {object} resp.Response{data=dto.CustomerAnalyticsResponse} "获取成功"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /dashboard/analytics/customers [get]
func (dc *DashboardController) GetCustomerAnalytics(c *gin.Context) {
	analytics, err := dc.dashboardService.GetCustomerAnalytics(c.Request.Context())
	if err != nil {
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, analytics)
}

// GetSalesAnalytics godoc
// @Summary      获取销售分析数据
// @Description  获取销售相关的分析数据，包括销售趋势、订单状态分布、客单价、复购率等
// @Tags         Dashboard
// @Produce      json
// @Success      200 {object} resp.Response{data=dto.SalesAnalyticsResponse} "获取成功"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /dashboard/analytics/sales [get]
func (dc *DashboardController) GetSalesAnalytics(c *gin.Context) {
	analytics, err := dc.dashboardService.GetSalesAnalytics(c.Request.Context())
	if err != nil {
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, analytics)
}

// GetProductAnalytics godoc
// @Summary      获取产品分析数据
// @Description  获取产品相关的分析数据，包括热销产品排行、类别销售分布、库存预警等
// @Tags         Dashboard
// @Produce      json
// @Success      200 {object} resp.Response{data=dto.ProductAnalyticsResponse} "获取成功"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /dashboard/analytics/products [get]
func (dc *DashboardController) GetProductAnalytics(c *gin.Context) {
	analytics, err := dc.dashboardService.GetProductAnalytics(c.Request.Context())
	if err != nil {
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, analytics)
}

// GetMarketingAnalytics godoc
// @Summary      获取营销分析数据
// @Description  获取营销相关的分析数据，包括营销活动概览、渠道效果、ROI排行等
// @Tags         Dashboard
// @Produce      json
// @Success      200 {object} resp.Response{data=dto.MarketingAnalyticsResponse} "获取成功"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /dashboard/analytics/marketing [get]
func (dc *DashboardController) GetMarketingAnalytics(c *gin.Context) {
	analytics, err := dc.dashboardService.GetMarketingAnalytics(c.Request.Context())
	if err != nil {
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, analytics)
}

// GetFinancialAnalytics godoc
// @Summary      获取财务分析数据
// @Description  获取财务相关的分析数据，包括收入概览、钱包统计、收入来源分布等
// @Tags         Dashboard
// @Produce      json
// @Success      200 {object} resp.Response{data=dto.FinancialAnalyticsResponse} "获取成功"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /dashboard/analytics/financial [get]
func (dc *DashboardController) GetFinancialAnalytics(c *gin.Context) {
	analytics, err := dc.dashboardService.GetFinancialAnalytics(c.Request.Context())
	if err != nil {
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, analytics)
}

// GetActivitySummary godoc
// @Summary      获取活动摘要数据
// @Description  获取客户活动相关的摘要数据，包括待办活动统计、活动类型分布、最近活动等
// @Tags         Dashboard
// @Produce      json
// @Success      200 {object} resp.Response{data=dto.ActivitySummaryResponse} "获取成功"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /dashboard/activities [get]
func (dc *DashboardController) GetActivitySummary(c *gin.Context) {
	summary, err := dc.dashboardService.GetActivitySummary(c.Request.Context())
	if err != nil {
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, summary)
}

// GetQuickStats godoc
// @Summary      获取快速统计数据
// @Description  获取实时快速统计数据，用于工作台实时刷新显示
// @Tags         Dashboard
// @Produce      json
// @Success      200 {object} resp.Response{data=dto.QuickStatsResponse} "获取成功"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /dashboard/quick-stats [get]
func (dc *DashboardController) GetQuickStats(c *gin.Context) {
	stats, err := dc.dashboardService.GetQuickStats(c.Request.Context())
	if err != nil {
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, stats)
}
