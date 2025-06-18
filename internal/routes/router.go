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
	v1 := router.Group("/api/v1")

	// ===== Auth Routes =====
	authController := controller.NewAuthController(resManager)
	auth := v1.Group("/auth")
	{
		auth.POST("/login", authController.Login)
		auth.POST("/register", authController.Register)
		authAuth := auth.Group("").Use(middleware.JWTAuthMiddleware())
		authAuth.PUT("/profile", authController.UpdateProfile)
	}

	// ===== User Routes (待实现) =====
	// userController := controller.NewUserController(resManager)
	// user := v1.Group("/users")
	// {
	// 	user.GET("/", userController.List)
	// 	user.POST("/", userController.Create)
	// }

	// ===== Customer Routes (待实现) =====
	// customerController := controller.NewCustomerController(resManager)
	// customer := v1.Group("/customers")
	// {
	// 	customer.GET("/", customerController.List)
	// }

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
