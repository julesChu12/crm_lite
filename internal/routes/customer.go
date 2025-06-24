package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"

	"github.com/gin-gonic/gin"
)

// registerCustomerRoutes 注册客户模块路由
func registerCustomerRoutes(rg *gin.RouterGroup, rm *resource.Manager) {
	dbRes, _ := resource.Get[*resource.DBResource](rm, resource.DBServiceKey)
	customerController := controller.NewCustomerController(dbRes.DB)

	customers := rg.Group("/customers")
	{
		customers.POST("", customerController.CreateCustomer)
		customers.GET("", customerController.ListCustomers)
		customers.GET("/:id", customerController.GetCustomer)
		customers.PUT("/:id", customerController.UpdateCustomer)
		customers.DELETE("/:id", customerController.DeleteCustomer)
	}
}
