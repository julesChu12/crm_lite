package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"

	"github.com/gin-gonic/gin"
)

func RegisterContactRoutes(r *gin.RouterGroup, res *resource.Manager) {
	ctl := controller.NewContactController(res)

	grp := r.Group("/customers/:customer_id/contacts")
	{
		grp.GET("", ctl.List)
		grp.POST("", ctl.Create)
		grp.PUT(":contact_id", ctl.Update)
		grp.DELETE(":contact_id", ctl.Delete)
	}
}
