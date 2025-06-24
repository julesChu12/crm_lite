package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"

	"github.com/gin-gonic/gin"
)

func registerRoleRoutes(rg *gin.RouterGroup, rm *resource.Manager) {
	dbRes, _ := resource.Get[*resource.DBResource](rm, resource.DBServiceKey)
	rc := controller.NewRoleController(dbRes.DB)

	roles := rg.Group("/roles")
	{
		roles.POST("", rc.CreateRole)
		roles.GET("", rc.ListRoles)
		roles.PUT("/:id", rc.UpdateRole)
		roles.DELETE("/:id", rc.DeleteRole)
	}
}
