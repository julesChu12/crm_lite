package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"

	"github.com/gin-gonic/gin"
)

// registerUserRoutes 注册用户管理模块的路由
func registerUserRoutes(rg *gin.RouterGroup, resManager *resource.Manager) {
	userController := controller.NewUserController(resManager)
	users := rg.Group("/users")
	{
		users.POST("", userController.CreateUser)
		users.GET("", userController.GetUserList)
		users.POST("/batch-get", userController.BatchGetUsers) // 新增路由
		users.GET("/:uuid", userController.GetUserByID)
		users.PUT("/:uuid", userController.UpdateUser)
		users.DELETE("/:uuid", userController.DeleteUser)
	}
}
