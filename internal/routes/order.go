package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"

	"github.com/gin-gonic/gin"
)

func registerOrderRoutes(rg *gin.RouterGroup, rm *resource.Manager) {
	orderController := controller.NewOrderController(rm)

	orders := rg.Group("/orders")
	{
		orders.POST("", orderController.CreateOrder)
		orders.GET("", orderController.ListOrders)
		orders.GET("/:id", orderController.GetOrder)

	}
}
