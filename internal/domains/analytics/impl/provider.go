package impl

import (
	"crm_lite/internal/domains/analytics"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// NewAnalyticsService 创建Analytics服务实例
func NewAnalyticsService(db *gorm.DB, cache *redis.Client) analytics.Service {
	return NewAnalyticsServiceImpl(db, cache)
}

// NewAnalyticsServiceForController 为控制器创建Analytics服务实例
// 支持Legacy兼容接口，便于现有控制器调用
func NewAnalyticsServiceForController(db *gorm.DB, cache *redis.Client) *AnalyticsServiceImpl {
	return NewAnalyticsServiceImpl(db, cache)
}
