package impl

import (
	"crm_lite/internal/common"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/domains/billing"

	"gorm.io/gorm"
)

// NewBillingServiceForController 为控制器创建 billing 服务实例
// 支持 Legacy 兼容接口，便于现有控制器调用
func NewBillingServiceForController(db *gorm.DB) *BillingServiceImpl {
	q := query.Use(db)
	tx := common.NewTx(db)
	return &BillingServiceImpl{
		db: db,
		q:  q,
		tx: tx,
	}
}

// NewBillingServiceWithTx 创建带事务管理的 billing 服务实例
// 用于跨域事务协调
func NewBillingServiceWithTx(db *gorm.DB, tx common.Tx) billing.Service {
	return NewBillingServiceImpl(db, tx)
}
