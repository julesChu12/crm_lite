package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"

	"github.com/gin-gonic/gin"
)

func registerProductRoutes(rg *gin.RouterGroup, rm *resource.Manager) {
	productController := controller.NewProductController(rm)

	products := rg.Group("/products")
	{
		products.POST("", productController.CreateProduct)
		products.GET("", productController.ListProducts)
		products.POST("/batch-get", productController.BatchGetProducts)
		products.GET("/:id", productController.GetProduct)
		products.PUT("/:id", productController.UpdateProduct)
		products.DELETE("/:id", productController.DeleteProduct)
	}
}
