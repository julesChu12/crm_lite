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

	// 用户-角色管理路由 (Casbin g policies)
	userRoles := rg.Group("/user-roles")
	{
		userRoles.POST("/assign", pc.AssignRoleToUser)
		userRoles.POST("/remove", pc.RemoveRoleFromUser)
		userRoles.GET("/users/:role", pc.GetUsersForRole)
		userRoles.GET("/roles/:user_id", pc.GetRolesForUser)
	}
}
