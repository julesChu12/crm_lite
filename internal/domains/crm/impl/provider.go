package impl

import (
	"crm_lite/internal/dao/query"
	"crm_lite/internal/domains/billing"
	"crm_lite/internal/domains/crm"

	"gorm.io/gorm"
)

// NewCRMServiceWithWallet 创建带钱包依赖的 CRM 服务实例
func NewCRMServiceWithWallet(q *query.Query, walletSvc WalletPort) crm.Service {
	return NewCRMService(q, walletSvc)
}

// NewCRMServiceWithBilling 创建带 billing 域依赖的 CRM 服务实例
// 使用新的 billing 域替代旧的钱包服务
func NewCRMServiceWithBilling(db *gorm.DB, billingSvc billing.Service) crm.Service {
	q := query.Use(db)
	walletAdapter := newBillingAdapter(billingSvc)
	return NewCRMService(q, walletAdapter)
}
