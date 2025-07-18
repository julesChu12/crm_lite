package routes

import (
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/middleware"
	"crm_lite/pkg/resp"

	"github.com/gin-gonic/gin"

	// docs is generated by Swag CLI
	_ "crm_lite/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(resManager *resource.Manager) *gin.Engine {
	// 1. 设置Gin模式
	gin.SetMode(string(config.GetInstance().Server.Mode))
	router := gin.New()

	// 2. 注册通用中间件
	router.Use(middleware.GinLogger(), gin.Recovery())

	// 3. 创建 /api/v1 路由组并应用安全中间件
	apiV1 := router.Group("/v1")
	apiV1.Use(middleware.NewJWTAuthMiddleware(resManager), middleware.NewCasbinMiddleware(resManager), middleware.NewSimpleCustomerAccessMiddleware(resManager))
	{
		// 所有 v1 路由都在这里注册。
		// 中间件内部的白名单会负责放行登录、注册等公开路由。
		registerAuthRoutes(apiV1, resManager)
		registerUserRoutes(apiV1, resManager)
		registerRoleRoutes(apiV1, resManager)
		registerPermissionRoutes(apiV1, resManager)
		registerCustomerRoutes(apiV1, resManager)
		RegisterContactRoutes(apiV1, resManager)
		registerProductRoutes(apiV1, resManager)
		registerOrderRoutes(apiV1, resManager)
		RegisterWalletRoutes(apiV1, resManager)
	}

	// 5. 设置一些通用路由
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		resp.Success(c, "ok")
	})

	// 欢迎页
	router.GET("/", func(c *gin.Context) {
		resp.Success(c, "Hello, World!")
	})
	// Add Swagger UI route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
