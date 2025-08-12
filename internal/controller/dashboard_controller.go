package controller

import (
    "crm_lite/internal/core/resource"
    "crm_lite/internal/dto"
    "crm_lite/internal/service"
    "crm_lite/pkg/resp"

    "github.com/gin-gonic/gin"
)

type DashboardController struct {
    svc *service.DashboardService
}

func NewDashboardController(res *resource.Manager) *DashboardController {
    return &DashboardController{svc: service.NewDashboardService(res)}
}

// GetOverview godoc
// @Summary      工作台总览
// @Description  返回客户与基础统计的总览数据（MVP 版，不含订单指标）
// @Tags         Dashboard
// @Produce      json
// @Param        query query dto.DashboardRequest false "查询参数"
// @Success      200 {object} resp.Response{data=dto.DashboardOverviewResponse}
// @Router       /dashboard/overview [get]
func (dc *DashboardController) GetOverview(c *gin.Context) {
    var req dto.DashboardRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        resp.Error(c, resp.CodeInvalidParam, err.Error())
        return
    }
    data, err := dc.svc.GetOverview(c.Request.Context(), &req)
    if err != nil {
        resp.SystemError(c, err)
        return
    }
    resp.Success(c, data)
}


