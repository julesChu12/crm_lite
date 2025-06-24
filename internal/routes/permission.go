package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"

	"github.com/gin-gonic/gin"
)

func registerPermissionRoutes(rg *gin.RouterGroup, rm *resource.Manager) {
	pc := controller.NewPermissionController(rm)

	permissions := rg.Group("/permissions")
	{
		permissions.POST("", pc.AddPermission)
		permissions.DELETE("", pc.RemovePermission)
		permissions.GET("/:role", pc.ListPermissionsByRole)
	}
}
