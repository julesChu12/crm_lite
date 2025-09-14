package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterOrderRoutes(rg *gin.RouterGroup, rm *resource.Manager) {
	orderController := controller.NewOrderController(rm)

	orders := rg.Group("/orders").Use(middleware.NewSimpleCustomerAccessMiddleware(rm))
	{
		orders.POST("", orderController.CreateOrder)
		orders.GET("", orderController.ListOrders)
		orders.GET("/:id", orderController.GetOrder)

	}
}
