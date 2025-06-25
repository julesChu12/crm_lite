package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"

	"github.com/gin-gonic/gin"
)

func registerRoleRoutes(rg *gin.RouterGroup, resManager *resource.Manager) {
	roleController := controller.NewRoleController(resManager)

	roles := rg.Group("/roles")
	{
		roles.POST("", roleController.CreateRole)
		roles.GET("", roleController.ListRoles)
		roles.GET("/:id", roleController.GetRoleByID)
		roles.PUT("/:id", roleController.UpdateRole)
		roles.DELETE("/:id", roleController.DeleteRole)
	}
}
