package resource

import (
	"context"
	"crm_lite/internal/core/config"
	"fmt"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
)

type CasbinResource struct {
	Enforcer *casbin.Enforcer
	manager  *Manager
	conf     config.RbacOptions
}

// NewCasbinResource 创建一个新的 CasbinResource 实例
// 注意：此时不进行初始化，仅传递依赖
func NewCasbinResource(manager *Manager, conf config.RbacOptions) *CasbinResource {
	return &CasbinResource{
		manager: manager,
		conf:    conf,
	}
}

// Initialize 真正执行初始化 Casbin Enforcer 的逻辑
func (r *CasbinResource) Initialize(ctx context.Context) error {
	dbRes, err := Get[*DBResource](r.manager, DBServiceKey)
	if err != nil {
		return fmt.Errorf("casbin resource depends on db resource, but failed to get it: %w", err)
	}

	adapter, err := gormadapter.NewAdapterByDB(dbRes.DB)
	if err != nil {
		return fmt.Errorf("failed to create casbin gorm adapter: %w", err)
	}

	enforcer, err := casbin.NewEnforcer(r.conf.ModelFile, adapter)
	if err != nil {
		return fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	// 从数据库加载策略
	if err := enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("failed to load casbin policy: %w", err)
	}

	r.Enforcer = enforcer
	fmt.Println("Casbin enforcer initialized successfully.")
	return nil
}

// Close CasbinResource 不需要显式关闭资源
func (r *CasbinResource) Close(ctx context.Context) error {
	return nil
}

// GetEnforcer 暴露一个获取 enforcer 的方法
func (r *CasbinResource) GetEnforcer() *casbin.Enforcer {
	return r.Enforcer
}
