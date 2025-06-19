package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/middleware"

	"github.com/gin-gonic/gin"
)

// registerAuthRoutes 注册认证模块路由
func registerAuthRoutes(rg *gin.RouterGroup, resManager *resource.Manager) {
	authController := controller.NewAuthController(resManager)
	auth := rg.Group("/auth")
	{
		auth.POST("/login", authController.Login)
		auth.POST("/register", authController.Register)
		// 需要认证的路由
		authed := auth.Group("").Use(middleware.JWTAuthMiddleware())
		{
			authed.PUT("/profile", authController.UpdateProfile)
		}
	}
}
