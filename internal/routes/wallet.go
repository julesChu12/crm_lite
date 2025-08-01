package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/middleware"
	"crm_lite/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterWalletRoutes(router *gin.RouterGroup, resourceManager *resource.Manager) {
	// 初始化 Service 和 Controller
	walletSvc := service.NewWalletService(resourceManager)
	walletCtl := controller.NewWalletController(walletSvc)

	// 创建一个 "wallets" 路由组
	walletRoutes := router.Group("/customers/:id").Use(middleware.NewSimpleCustomerAccessMiddleware(resourceManager))
	{
		// GET /v1/customers/:id/wallet
		walletRoutes.GET("/wallet", walletCtl.GetWalletByCustomerID)
		// POST /v1/customers/:id/wallet/transactions
		walletRoutes.POST("/wallet/transactions", walletCtl.CreateTransaction)
	}
}
