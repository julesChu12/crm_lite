package routes

import (
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/middleware"
	"crm_lite/pkg/resp"

	"github.com/gin-gonic/gin"
)

func NewRouter(resManager *resource.Manager) *gin.Engine {
	// 1. 设置Gin模式
	gin.SetMode(string(config.GetInstance().Server.Mode))
	router := gin.New()

	// 2. 注册通用中间件
	router.Use(middleware.GinLogger(), gin.Recovery())

	// 3. 设置 API 路由组
	v1 := router.Group("/api/v1")

	// 4. 注册各个模块的路由
	registerAuthRoutes(v1, resManager)
	// registerUserRoutes(v1, resManager)      // 待实现时取消注释
	// registerCustomerRoutes(v1, resManager) // 待实现时取消注释

	// 5. 设置一些通用路由
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		resp.Success(c, "ok")
	})

	// 欢迎页
	router.GET("/", func(c *gin.Context) {
		resp.Success(c, "Hello, World!")
	})

	return router
}
