package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(resManager *resource.Manager) *gin.Engine {
	// 1. 设置Gin模式
	gin.SetMode(string(config.GetInstance().Server.Mode))
	router := gin.New()

	// 2. 注册中间件
	router.Use(middleware.GinLogger(), gin.Recovery())

	// 3. 设置路由
	// 创建 v1 API 路由组
	v1 := router.Group("/api/v1")
	{
		// 认证相关路由
		authController := controller.NewAuthController(resManager)
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/login", authController.Login)
		}
	}

	// 例如，一个健康检查路由
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// 原有的根路由
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, World!"})
	})

	return router
}
