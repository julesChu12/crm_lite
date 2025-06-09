package resource

import (
	"context"
	"fmt"
	"sync"
)

//////////////////////////////////////////////////////////////////////
// 基础定义
//////////////////////////////////////////////////////////////////////

// ServiceKey 标识不同资源类型
type ServiceKey string

const (
	DBServiceKey    ServiceKey = "db"
	CacheServiceKey ServiceKey = "cache"
	ESServiceKey    ServiceKey = "es"
)

// Resource 抽象所有可托管资源
type Resource interface {
	Initialize(ctx context.Context) error
	Close(ctx context.Context) error
}

//////////////////////////////////////////////////////////////////////
// Manager
//////////////////////////////////////////////////////////////////////

type Manager struct {
	res   map[ServiceKey]Resource // 已注册资源
	order []ServiceKey            // 注册顺序，用于按序关闭
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		res: make(map[ServiceKey]Resource),
	}
}

// Register 把资源纳入管理（仅注册，不初始化）
func (m *Manager) Register(key ServiceKey, r Resource) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.res[key]; exists {
		return fmt.Errorf("resource %s already registered", key)
	}
	m.res[key] = r
	m.order = append(m.order, key)
	return nil
}

// InitAll 统一初始化（按注册顺序）
func (m *Manager) InitAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, k := range m.order {
		if err := m.res[k].Initialize(ctx); err != nil {
			return fmt.Errorf("init %s: %w", k, err)
		}
	}
	return nil
}

// CloseAll 统一关闭（按注册逆序）
func (m *Manager) CloseAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i := len(m.order) - 1; i >= 0; i-- {
		k := m.order[i]
		if err := m.res[k].Close(ctx); err != nil {
			return fmt.Errorf("close %s: %w", k, err)
		}
	}
	return nil
}

//////////////////////////////////////////////////////////////////////
// 泛型便捷获取（Go1.18+）
//////////////////////////////////////////////////////////////////////

func Get[T Resource](mgr *Manager, key ServiceKey) (T, error) {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	if r, ok := mgr.res[key]; ok {
		if cast, ok := r.(T); ok {
			return cast, nil
		}
	}
	var zero T
	return zero, fmt.Errorf("resource %s not found or type mismatch", key)
}
