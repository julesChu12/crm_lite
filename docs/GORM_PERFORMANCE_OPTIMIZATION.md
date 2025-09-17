# GORM查询性能优化指南

## 当前性能分析

### 1. 潜在的N+1查询问题

**问题位置**: Orders查询时可能存在N+1查询
```go
// 当前可能存在的问题
orders, err := q.Order.WithContext(ctx).Find()
for _, order := range orders {
    // 每次查询都会触发一次数据库查询
    items, _ := q.OrderItem.WithContext(ctx).Where(q.OrderItem.OrderID.Eq(order.ID)).Find()
}
```

**优化方案**: 使用Preload或Join
```go
// 优化后的查询
type OrderWithItems struct {
    *model.Order
    Items []*model.OrderItem `gorm:"foreignKey:OrderID"`
}

// 使用预加载
orders, err := q.Order.WithContext(ctx).Preload("Items").Find()

// 或者使用Join（更高效）
var results []OrderWithItems
err := db.WithContext(ctx).
    Table("orders").
    Select("orders.*, order_items.*").
    Joins("LEFT JOIN order_items ON orders.id = order_items.order_id").
    Scan(&results)
```

### 2. 索引优化建议

**当前索引分析**:
- ✅ customers表已有phone唯一索引
- ✅ orders表已有customer_id + created_at复合索引
- ✅ wallet_transactions表已有wallet_id + created_at复合索引

**建议新增索引**:
```sql
-- 订单状态查询优化
CREATE INDEX idx_orders_status_created ON orders(status, created_at);

-- 产品分类查询优化
CREATE INDEX idx_products_category_active ON products(category, is_active);

-- 钱包交易类型查询优化
CREATE INDEX idx_wallet_txn_type_created ON wallet_transactions(type, created_at);
```

### 3. 批量操作优化

**问题**: 单条记录插入效率低
```go
// 低效的单条插入
for _, item := range items {
    q.OrderItem.WithContext(ctx).Create(item)
}
```

**优化**: 批量操作
```go
// 高效的批量插入
err = q.OrderItem.WithContext(ctx).CreateInBatches(items, 100)

// 或者使用事务批量操作
err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    txQuery := query.Use(tx)
    return txQuery.OrderItem.Create(items...)
})
```

### 4. 查询优化建议

#### 4.1 分页查询优化
```go
// 当前可能的实现
func (s *Service) ListOrders(ctx context.Context, page, pageSize int) ([]Order, error) {
    offset := (page - 1) * pageSize
    return s.q.Order.WithContext(ctx).Offset(offset).Limit(pageSize).Find()
}

// 优化后的实现（避免大offset性能问题）
func (s *Service) ListOrdersCursor(ctx context.Context, lastID int64, pageSize int) ([]Order, error) {
    query := s.q.Order.WithContext(ctx).Where(s.q.Order.ID.Gt(lastID))
    return query.Order(s.q.Order.ID).Limit(pageSize).Find()
}
```

#### 4.2 条件查询优化
```go
// 避免全表扫描
func (s *Service) SearchCustomers(ctx context.Context, keyword string) ([]*model.Customer, error) {
    if keyword == "" {
        // 返回空结果而不是全表查询
        return []*model.Customer{}, nil
    }

    // 使用索引友好的查询
    q := s.q.Customer.WithContext(ctx)

    // 优先使用索引字段
    if isPhoneNumber(keyword) {
        return q.Where(s.q.Customer.Phone.Like(keyword + "%")).Find()
    }

    // 名称搜索使用全文索引（需要创建）
    return q.Where(s.q.Customer.Name.Like("%" + keyword + "%")).Limit(100).Find()
}
```

### 5. 缓存策略

#### 5.1 查询结果缓存
```go
type CachedQueryService struct {
    q     *query.Query
    cache *redis.Client
}

func (s *CachedQueryService) GetCustomerByID(ctx context.Context, id int64) (*model.Customer, error) {
    cacheKey := fmt.Sprintf("customer:%d", id)

    // 先尝试从缓存获取
    cached, err := s.cache.Get(ctx, cacheKey).Result()
    if err == nil {
        var customer model.Customer
        if json.Unmarshal([]byte(cached), &customer) == nil {
            return &customer, nil
        }
    }

    // 缓存未命中，查询数据库
    customer, err := s.q.Customer.WithContext(ctx).Where(s.q.Customer.ID.Eq(id)).First()
    if err != nil {
        return nil, err
    }

    // 写入缓存
    data, _ := json.Marshal(customer)
    s.cache.Set(ctx, cacheKey, data, 10*time.Minute)

    return customer, nil
}
```

#### 5.2 统计数据缓存
```go
func (s *Service) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
    cacheKey := "dashboard:stats"

    // 尝试从缓存获取
    if cached := s.getFromCache(cacheKey); cached != nil {
        return cached, nil
    }

    // 并发查询多个统计数据
    var (
        totalCustomers int64
        totalOrders    int64
        totalRevenue   float64
    )

    eg := errgroup.Group{}

    eg.Go(func() error {
        var err error
        totalCustomers, err = s.q.Customer.WithContext(ctx).Count()
        return err
    })

    eg.Go(func() error {
        var err error
        totalOrders, err = s.q.Order.WithContext(ctx).Count()
        return err
    })

    eg.Go(func() error {
        return s.q.Order.WithContext(ctx).
            Select("SUM(final_amount)").
            Where(s.q.Order.Status.Eq("paid")).
            Scan(&totalRevenue)
    })

    if err := eg.Wait(); err != nil {
        return nil, err
    }

    stats := &DashboardStats{
        TotalCustomers: totalCustomers,
        TotalOrders:    totalOrders,
        TotalRevenue:   totalRevenue,
    }

    // 缓存5分钟
    s.setToCache(cacheKey, stats, 5*time.Minute)
    return stats, nil
}
```

### 6. 连接池优化

```go
// 数据库连接池配置优化
func setupDBPool(db *gorm.DB) {
    sqlDB, _ := db.DB()

    // 设置最大打开连接数
    sqlDB.SetMaxOpenConns(25)

    // 设置最大空闲连接数
    sqlDB.SetMaxIdleConns(10)

    // 设置连接的最大生命周期
    sqlDB.SetConnMaxLifetime(5 * time.Minute)

    // 设置连接的最大空闲时间
    sqlDB.SetConnMaxIdleTime(time.Minute)
}
```

### 7. 实施建议

1. **立即实施**:
   - 添加missing的索引
   - 修复明显的N+1查询
   - 实现批量操作

2. **中期实施**:
   - 引入查询结果缓存
   - 优化分页查询
   - 实现游标分页

3. **长期优化**:
   - 数据库读写分离
   - 分片策略
   - 异步处理重型查询

### 8. 监控指标

建议监控以下关键指标：
- 慢查询日志 (>100ms)
- 数据库连接池使用率
- 缓存命中率
- QPS和响应时间
- 索引使用情况