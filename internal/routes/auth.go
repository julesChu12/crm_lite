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
		// 登录使用动态风控中间件（Redis 存储），首次不要求验证码，失败后要求
		auth.POST("/login", middleware.SimpleCaptchaGuard(resManager), authController.Login)

		// 需要 Turnstile 验证码的路由组
		turnstileGroup := auth.Group("", middleware.TurnstileMiddleware())
		{
			turnstileGroup.POST("/register", authController.Register)
			turnstileGroup.POST("/forgot-password", authController.ForgotPassword)
			turnstileGroup.POST("/reset-password", authController.ResetPassword)
		}

		// 其他认证相关路由
		auth.POST("/refresh", authController.RefreshToken)
		auth.POST("/logout", authController.Logout)
		auth.GET("/profile", authController.GetProfile)
		auth.PUT("/profile", authController.UpdateProfile)
		auth.PUT("/password", authController.ChangePassword)
	}
}
