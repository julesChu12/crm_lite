package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterContactRoutes(r *gin.RouterGroup, res *resource.Manager) {
	ctl := controller.NewContactController(res)

	grp := r.Group("/customers/:id/contacts").Use(middleware.NewSimpleCustomerAccessMiddleware(res))
	{
		grp.GET("", ctl.ListContacts)                 // 获取客户的联系人列表
		grp.POST("", ctl.CreateContact)               // 创建联系人
		grp.GET("/:contact_id", ctl.GetContact)       // 获取单个联系人详情
		grp.PUT("/:contact_id", ctl.UpdateContact)    // 更新联系人
		grp.DELETE("/:contact_id", ctl.DeleteContact) // 删除联系人
	}
}
