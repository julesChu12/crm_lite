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
		// 公开路由，无需登录
		auth.POST("/login", authController.Login)
		auth.POST("/register", authController.Register)
		auth.POST("/refresh", authController.RefreshToken)
		auth.POST("/forgot-password", authController.ForgotPassword)
		auth.POST("/reset-password", authController.ResetPassword)

		// 保护路由，需要 JWT 认证
		authed := auth.Group("").Use(middleware.JWTAuthMiddleware())
		{
			authed.POST("/logout", authController.Logout)
			authed.GET("/profile", authController.GetProfile)
			authed.PUT("/profile", authController.UpdateProfile)
			authed.PUT("/password", authController.ChangePassword)
		}
	}
}
