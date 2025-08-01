package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/middleware"

	"github.com/gin-gonic/gin"
)

// registerCustomerRoutes 注册客户模块路由
func registerCustomerRoutes(rg *gin.RouterGroup, rm *resource.Manager) {
	customerController := controller.NewCustomerController(rm)

	customers := rg.Group("/customers").Use(middleware.NewSimpleCustomerAccessMiddleware(rm))
	{
		customers.POST("", customerController.CreateCustomer)
		customers.GET("", customerController.ListCustomers)
		customers.POST("/batch-get", customerController.BatchGetCustomers)
		customers.GET("/:id", customerController.GetCustomer)
		customers.PUT("/:id", customerController.UpdateCustomer)
		customers.DELETE("/:id", customerController.DeleteCustomer)
	}
}
