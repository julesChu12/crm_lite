package controller

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/domains/analytics"
	"crm_lite/internal/domains/analytics/impl"
	"crm_lite/internal/dto"
	"crm_lite/pkg/resp"

	"github.com/gin-gonic/gin"
)

// DashboardController 工作台相关接口
// 已迁移到 analytics 域服务
type DashboardController struct {
	dashboardService analytics.DashboardService
}

// NewDashboardController 注入资源管理器
// 完全使用 analytics 域服务
func NewDashboardController(resManager *resource.Manager) *DashboardController {
	// 创建Analytics领域服务
	dbRes, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource for DashboardController: " + err.Error())
	}

	// 创建仪表盘服务
	dashboardService := impl.NewDashboardService(dbRes.DB)

	return &DashboardController{
		dashboardService: dashboardService,
	}
}

// Overview godoc
// @Summary      工作台总览数据
// @Description  获取客户、订单、收入等汇总统计
// @Tags         Dashboard
// @Accept       json
// @Produce      json
// @Param        date_range  query     string  false  "时间范围: today/week/month/quarter/year"
// @Param        timezone    query     string  false  "时区"
// @Success      200  {object}  resp.Response{data=dto.DashboardOverviewResponse}
// @Failure      400  {object}  resp.Response
// @Router       /dashboard/overview [get]
func (dc *DashboardController) Overview(c *gin.Context) {
	var req dto.DashboardRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}

	// 调用Analytics领域服务
	overview, err := dc.dashboardService.GetOverview(c.Request.Context())
	if err != nil {
		resp.SystemError(c, err)
		return
	}

	// 转换为DTO格式
	overviewResponse := &dto.DashboardOverviewResponse{
		TotalCustomers:      overview.TotalCustomers,
		TotalOrders:         overview.TotalOrders,
		TotalProducts:       overview.TotalProducts,
		TotalRevenue:        overview.TotalRevenue,
		MonthlyNewCustomers: overview.MonthlyCustomers,
		MonthlyOrders:       overview.MonthlyOrders,
		MonthlyRevenue:      overview.MonthlyRevenue,
		CustomerGrowthRate:  overview.GrowthRate,
		OrderGrowthRate:     overview.GrowthRate,
		RevenueGrowthRate:   overview.GrowthRate,
	}

	resp.Success(c, overviewResponse)
}
