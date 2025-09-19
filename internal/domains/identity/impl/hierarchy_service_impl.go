package impl

import (
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/query"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// HierarchyServiceImpl 组织架构层级服务实现
// 专为中小企业设计的极简上下级关系查询服务，使用Redis缓存优化性能
type HierarchyServiceImpl struct {
	q     *query.Query
	db    *gorm.DB
	cache *redis.Client
}

// NewHierarchyService 创建层级服务实例
func NewHierarchyService(resManager *resource.Manager) *HierarchyServiceImpl {
	dbRes, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get DB resource for HierarchyService: " + err.Error())
	}
	cacheRes, err := resource.Get[*resource.CacheResource](resManager, resource.CacheServiceKey)
	if err != nil {
		panic("Failed to get Cache resource for HierarchyService: " + err.Error())
	}
	return &HierarchyServiceImpl{
		q:     query.Use(dbRes.DB),
		db:    dbRes.DB,
		cache: cacheRes.Client,
	}
}

// GetSubordinates 返回指定管理者的所有下属（含多级），层级深度不超过5层
// 优先从Redis缓存获取，失败则查询数据库并回填缓存
func (s *HierarchyServiceImpl) GetSubordinates(ctx context.Context, managerID int64) ([]int64, error) {
	cacheKey := fmt.Sprintf("subordinates:%d", managerID)

	// 1. 尝试从缓存获取
	cachedIDs, err := s.getCachedSubordinates(ctx, cacheKey)
	if err != nil {
		log.Printf("Error getting subordinates from cache for manager %d: %v", managerID, err)
		// 即使缓存出错，也继续从数据库查询，保证可用性
	}
	if cachedIDs != nil {
		log.Printf("Cache hit for manager %d", managerID)
		return cachedIDs, nil
	}

	log.Printf("Cache miss for manager %d", managerID)

	// 2. 缓存未命中，从数据库查询
	var subordinates []int64
	queue := []int64{managerID}
	visited := make(map[int64]bool)
	const maxDepth = 5

	for depth := 0; depth < maxDepth && len(queue) > 0; depth++ {
		var nextLevel []int64
		for _, current := range queue {
			if visited[current] {
				continue
			}
			visited[current] = true
			var direct []int64
			if err := s.db.Raw("SELECT id FROM admin_users WHERE manager_id = ? AND deleted_at IS NULL", current).Scan(&direct).Error; err != nil {
				return nil, err
			}
			subordinates = append(subordinates, direct...)
			nextLevel = append(nextLevel, direct...)
		}
		queue = nextLevel
	}

	// 3. 异步写入缓存
	go s.cacheSubordinates(context.Background(), cacheKey, subordinates, 5*time.Minute)

	return subordinates, nil
}

// GetDirectReports 获取直接下属用户列表
func (s *HierarchyServiceImpl) GetDirectReports(ctx context.Context, managerID int64) ([]int64, error) {
	var directReports []int64
	if err := s.db.Raw("SELECT id FROM admin_users WHERE manager_id = ? AND deleted_at IS NULL", managerID).Scan(&directReports).Error; err != nil {
		return nil, fmt.Errorf("failed to get direct reports for manager %d: %w", managerID, err)
	}
	return directReports, nil
}

// GetManagerChain 获取用户的完整管理链
func (s *HierarchyServiceImpl) GetManagerChain(ctx context.Context, userID int64) ([]int64, error) {
	var managers []int64
	currentID := userID
	visited := make(map[int64]bool)
	const maxDepth = 10 // 防止循环引用

	for depth := 0; depth < maxDepth; depth++ {
		if visited[currentID] {
			break // 防止循环引用
		}
		visited[currentID] = true

		var managerID int64
		err := s.db.Raw("SELECT manager_id FROM admin_users WHERE id = ? AND deleted_at IS NULL", currentID).Scan(&managerID).Error
		if err != nil {
			return nil, fmt.Errorf("failed to get manager for user %d: %w", currentID, err)
		}

		if managerID == 0 {
			break // 没有上级了
		}

		managers = append(managers, managerID)
		currentID = managerID
	}

	return managers, nil
}

// CanAccessCustomer 判断操作者是否可以访问指定客户
// 规则：1) 如果操作者是客户的assigned_to负责人，则允许
//       2) 如果操作者是负责人的上级（递归），则允许
func (s *HierarchyServiceImpl) CanAccessCustomer(ctx context.Context, operatorID int64, customerID int64) (bool, error) {
	// 查询客户负责人
	customer, err := s.q.Customer.WithContext(ctx).Where(s.q.Customer.ID.Eq(customerID)).First()
	if err != nil {
		return false, fmt.Errorf("failed to get customer %d: %w", customerID, err)
	}

	// 负责人本人
	if customer.AssignedTo == operatorID {
		return true, nil
	}

	// 判断是否上级
	subordinates, err := s.GetSubordinates(ctx, operatorID)
	if err != nil {
		return false, fmt.Errorf("failed to get subordinates for operator %d: %w", operatorID, err)
	}

	for _, id := range subordinates {
		if id == customer.AssignedTo {
			return true, nil
		}
	}
	return false, nil
}

// --- 缓存实现 ---

func (s *HierarchyServiceImpl) getCachedSubordinates(ctx context.Context, key string) ([]int64, error) {
	val, err := s.cache.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // 缓存未命中，不是错误
	} else if err != nil {
		return nil, err // Redis 服务错误
	}

	var ids []int64
	if err := json.Unmarshal([]byte(val), &ids); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached subordinates: %w", err)
	}
	return ids, nil
}

func (s *HierarchyServiceImpl) cacheSubordinates(ctx context.Context, key string, ids []int64, ttl time.Duration) {
	bytes, err := json.Marshal(ids)
	if err != nil {
		log.Printf("Error marshalling subordinates for key %s: %v", key, err)
		return
	}
	if err := s.cache.Set(ctx, key, bytes, ttl).Err(); err != nil {
		log.Printf("Error setting cache for key %s: %v", key, err)
	}
}