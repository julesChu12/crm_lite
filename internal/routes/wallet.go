package routes

import (
	"crm_lite/internal/controller"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterWalletRoutes(router *gin.RouterGroup, resourceManager *resource.Manager) {
	// 初始化 Controller (现已迁移到 billing 域服务)
	walletCtl := controller.NewWalletController(resourceManager)

	// 创建一个 "wallets" 路由组
	walletRoutes := router.Group("/customers/:id").Use(middleware.NewSimpleCustomerAccessMiddleware(resourceManager))
	{
		// GET /v1/customers/:id/wallet
		walletRoutes.GET("/wallet", walletCtl.GetWalletByCustomerID)
		// POST /v1/customers/:id/wallet/transactions
		walletRoutes.POST("/wallet/transactions", walletCtl.CreateTransaction)
		// GET /v1/customers/:id/wallet/transactions
		walletRoutes.GET("/wallet/transactions", walletCtl.GetTransactions)
		// POST /v1/customers/:id/wallet/refund
		walletRoutes.POST("/wallet/refund", walletCtl.ProcessRefund)
	}
}
