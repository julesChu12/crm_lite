package impl

import (
	"crm_lite/internal/common"
	"crm_lite/internal/domains/marketing"

	"gorm.io/gorm"
)

// NewMarketingService 创建Marketing服务实例
func NewMarketingService(db *gorm.DB) marketing.Service {
	tx := common.NewTx(db)
	return NewMarketingServiceImpl(db, tx)
}

// NewMarketingServiceForController 为控制器创建Marketing服务实例
// 支持Legacy兼容接口，便于现有控制器调用
func NewMarketingServiceForController(db *gorm.DB) *MarketingServiceImpl {
	tx := common.NewTx(db)
	return NewMarketingServiceImpl(db, tx)
}
