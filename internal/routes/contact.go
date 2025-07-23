package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"

	"github.com/gin-gonic/gin"
)

func RegisterContactRoutes(r *gin.RouterGroup, res *resource.Manager) {
	ctl := controller.NewContactController(res)

	grp := r.Group("/customers/:id/contacts")
	{
		grp.GET("", ctl.ListContacts)
		grp.POST("", ctl.CreateContact)
		grp.PUT("/:contact_id", ctl.UpdateContact)
		grp.DELETE("/:contact_id", ctl.DeleteContact)
	}
}
