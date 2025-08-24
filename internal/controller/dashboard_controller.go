package controller

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"

	"github.com/gin-gonic/gin"
)

// DashboardController 工作台相关接口

type DashboardController struct {
	svc *service.DashboardService
}

// NewDashboardController 注入资源管理器
func NewDashboardController(resManager *resource.Manager) *DashboardController {
	return &DashboardController{
		svc: service.NewDashboardService(resManager),
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

	data, err := dc.svc.GetOverview(c.Request.Context())
	if err != nil {
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, data)
}
