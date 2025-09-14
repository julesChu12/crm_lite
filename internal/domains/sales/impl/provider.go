package impl

import (
	"os"

	"crm_lite/internal/common"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/query"
	billingImpl "crm_lite/internal/domains/billing/impl"
	catalogImpl "crm_lite/internal/domains/catalog/impl"
	"crm_lite/internal/domains/sales"
)

// ProvideSales 创建sales服务实例
// 根据环境变量选择使用新的sales实现还是旧的适配器
func ProvideSales(rm *resource.Manager) sales.Service {
	// 获取数据库资源
	dbRes, err := resource.Get[*resource.DBResource](rm, resource.DBServiceKey)
	if err != nil {
		// 如果无法获取数据库资源，返回空实现
		return New()
	}

	// 优先使用新的sales服务实现
	if os.Getenv("USE_LEGACY_ORDER") != "1" {
		// 创建依赖服务
		txManager := common.NewTx(dbRes.DB)
		catalogService := catalogImpl.New(query.Use(dbRes.DB))
		billingService := billingImpl.NewBillingService(dbRes.DB)
		outboxService := common.NewOutboxService(dbRes.DB, txManager)

		// 返回新的sales服务实现
		return NewSalesServiceImpl(dbRes.DB, txManager, catalogService, billingService, outboxService)
	}

	// 如果指定使用旧的实现，返回旧的适配器
	return New()
}
