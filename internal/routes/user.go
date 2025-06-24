package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"

	"github.com/gin-gonic/gin"
)

// registerUserRoutes 注册用户模块路由 (待实现)
func registerUserRoutes(rg *gin.RouterGroup, rm *resource.Manager) {
	// 从资源管理器中获取 GORM DB 实例
	dbRes, err := resource.Get[*resource.DBResource](rm, resource.DBServiceKey)
	if err != nil {
		panic("failed to get db resource from manager: " + err.Error())
	}
	db := dbRes.DB

	// 实例化 Controller
	userController := controller.NewUserController(db)

	// 定义路由
	userRoutes := rg.Group("/users")
	{
		userRoutes.GET("/:uuid", userController.GetUser)
		// 未来可以添加更多用户相关的路由，例如：
		// userRoutes.GET("", userController.ListUsers)
		// userRoutes.PUT("/:uuid", userController.UpdateUser)
		// userRoutes.DELETE("/:uuid", userController.DeleteUser)
	}
}
