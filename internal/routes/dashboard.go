package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"

	"github.com/gin-gonic/gin"
)

// RegisterDashboardRoutes 注册工作台路由
func RegisterDashboardRoutes(r *gin.RouterGroup, res *resource.Manager) {
	ctl := controller.NewDashboardController(res)
	dash := r.Group("/dashboard")
	{
		dash.GET("/overview", ctl.GetOverview)
	}
}
