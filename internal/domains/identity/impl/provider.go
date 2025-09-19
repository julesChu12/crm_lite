package impl

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/domains/identity"

	"github.com/casbin/casbin/v2"
	"gorm.io/gorm"
)

// NewIdentityService 创建Identity服务实例
func NewIdentityService(db *gorm.DB, enforcer *casbin.Enforcer, resManager *resource.Manager) identity.Service {
	return NewSimpleIdentityService(db, enforcer, resManager)
}

// NewIdentityServiceForController 为控制器创建Identity服务实例
// 支持Legacy兼容接口，便于现有控制器调用
func NewIdentityServiceForController(db *gorm.DB, enforcer *casbin.Enforcer, resManager *resource.Manager) *SimpleIdentityService {
	return NewSimpleIdentityService(db, enforcer, resManager)
}
