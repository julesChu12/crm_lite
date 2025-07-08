package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"

	"github.com/gin-gonic/gin"
)

// registerAuthRoutes 注册认证模块路由
func registerAuthRoutes(rg *gin.RouterGroup, resManager *resource.Manager) {
	authController := controller.NewAuthController(resManager)
	auth := rg.Group("/auth")
	{
		// 公开路由 - 会被中间件白名单放行
		auth.POST("/login", authController.Login)
		auth.POST("/register", authController.Register)
		auth.POST("/refresh", authController.RefreshToken)
		auth.POST("/forgot-password", authController.ForgotPassword)
		auth.POST("/reset-password", authController.ResetPassword)

		// 保护路由 - 会经过JWT和Casbin中间件的验证
		auth.POST("/logout", authController.Logout)
		auth.GET("/profile", authController.GetProfile)
		auth.PUT("/profile", authController.UpdateProfile)
		auth.PUT("/password", authController.ChangePassword)
	}
}
