package service

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

// SimpleHierarchyService 专为中小企业设计的极简上下级关系查询服务
// 实现了 Redis 缓存以优化性能。
type SimpleHierarchyService struct {
	q     *query.Query
	db    *gorm.DB
	cache *redis.Client
}

// NewSimpleHierarchyService 创建实例，并注入数据库和 Redis 资源
func NewSimpleHierarchyService(resManager *resource.Manager) *SimpleHierarchyService {
	dbRes, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get DB resource for HierarchyService: " + err.Error())
	}
	cacheRes, err := resource.Get[*resource.CacheResource](resManager, resource.CacheServiceKey)
	if err != nil {
		panic("Failed to get Cache resource for HierarchyService: " + err.Error())
	}
	return &SimpleHierarchyService{
		q:     query.Use(dbRes.DB),
		db:    dbRes.DB,
		cache: cacheRes.Client,
	}
}

// GetSubordinates 返回 managerID 的所有下属（含多级），层级深度不超过 5。
// 优先从 Redis 缓存获取，失败则查询数据库并回填缓存。
func (s *SimpleHierarchyService) GetSubordinates(ctx context.Context, managerID int64) ([]int64, error) {
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

// CanAccessCustomer 判断 operator 是否可以访问 customer
// 规则：1) 如果 operator 是 customer 的 assigned_to 负责人，则允许
//  2. 如果 operator 是负责人的上级（递归），则允许
func (s *SimpleHierarchyService) CanAccessCustomer(ctx context.Context, operatorID int64, customerID int64) (bool, error) {
	// 查询客户负责人
	customer, err := s.q.Customer.WithContext(ctx).Where(s.q.Customer.ID.Eq(customerID)).First()
	if err != nil {
		return false, err
	}

	// 负责人本人
	if customer.AssignedTo == operatorID {
		return true, nil
	}

	// 判断是否上级
	subs, err := s.GetSubordinates(ctx, operatorID)
	if err != nil {
		return false, err
	}
	for _, id := range subs {
		if id == customer.AssignedTo {
			return true, nil
		}
	}
	return false, nil
}

// --- 缓存实现 ---

func (s *SimpleHierarchyService) getCachedSubordinates(ctx context.Context, key string) ([]int64, error) {
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

func (s *SimpleHierarchyService) cacheSubordinates(ctx context.Context, key string, ids []int64, ttl time.Duration) {
	bytes, err := json.Marshal(ids)
	if err != nil {
		log.Printf("Error marshalling subordinates for key %s: %v", key, err)
		return
	}
	if err := s.cache.Set(ctx, key, bytes, ttl).Err(); err != nil {
		log.Printf("Error setting cache for key %s: %v", key, err)
	}
}
