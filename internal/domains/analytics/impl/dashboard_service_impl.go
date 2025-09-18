package impl

import (
	"context"
	"time"

	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/domains/analytics"

	"gorm.io/gorm"
)

// DashboardServiceImpl 仪表盘服务实现
// 迁移自 internal/service/dashboard_service.go
type DashboardServiceImpl struct {
	q  *query.Query
	db *gorm.DB
}

// NewDashboardService 创建仪表盘服务实例
func NewDashboardService(db *gorm.DB) analytics.DashboardService {
	return &DashboardServiceImpl{
		q:  query.Use(db),
		db: db,
	}
}

// GetOverview 获取业务总览数据
// 从 service.DashboardService.GetOverview 迁移而来
func (s *DashboardServiceImpl) GetOverview(ctx context.Context) (*analytics.Overview, error) {
	var overview analytics.Overview

	// 总体统计 - 保持原有逻辑
	totalCustomers, _ := s.q.Customer.WithContext(ctx).Count()
	totalOrders, _ := s.q.Order.WithContext(ctx).Count()
	totalProducts, _ := s.q.Product.WithContext(ctx).Count()

	var totalRevenue float64
	_ = s.db.WithContext(ctx).Model(&model.Order{}).Select("COALESCE(SUM(final_amount),0)").Scan(&totalRevenue)

	overview.TotalCustomers = totalCustomers
	overview.TotalOrders = totalOrders
	overview.TotalProducts = totalProducts
	overview.TotalRevenue = totalRevenue

	// 当月统计
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	monthlyCustomers, _ := s.q.Customer.WithContext(ctx).
		Where(s.q.Customer.CreatedAt.Gte(firstOfMonth)).Count()

	monthlyOrders, _ := s.q.Order.WithContext(ctx).
		Where(s.q.Order.CreatedAt.Gte(firstOfMonth)).Count()

	var monthlyRevenue float64
	_ = s.db.WithContext(ctx).Model(&model.Order{}).
		Where("created_at >= ?", firstOfMonth.Unix()).
		Select("COALESCE(SUM(final_amount),0)").Scan(&monthlyRevenue)

	overview.MonthlyCustomers = monthlyCustomers
	overview.MonthlyOrders = monthlyOrders
	overview.MonthlyRevenue = monthlyRevenue

	// 计算增长率（简化版本，与上月对比）
	lastMonth := firstOfMonth.AddDate(0, -1, 0)
	lastMonthEnd := firstOfMonth.Add(-time.Second)

	var lastMonthRevenue float64
	_ = s.db.WithContext(ctx).Model(&model.Order{}).
		Where("created_at >= ? AND created_at <= ?", lastMonth.Unix(), lastMonthEnd.Unix()).
		Select("COALESCE(SUM(final_amount),0)").Scan(&lastMonthRevenue)

	if lastMonthRevenue > 0 {
		overview.GrowthRate = (monthlyRevenue - lastMonthRevenue) / lastMonthRevenue * 100
	}

	return &overview, nil
}

// GetSalesAnalysis 获取销售分析
// 新增功能，扩展原有能力
func (s *DashboardServiceImpl) GetSalesAnalysis(ctx context.Context, days int) (*analytics.SalesAnalysis, error) {
	since := time.Now().AddDate(0, 0, -days)

	var analysis analytics.SalesAnalysis

	// 基础销售统计
	orderCount, _ := s.q.Order.WithContext(ctx).
		Where(s.q.Order.CreatedAt.Gte(since)).Count()

	var totalRevenue float64
	_ = s.db.WithContext(ctx).Model(&model.Order{}).
		Where("created_at >= ?", since.Unix()).
		Select("COALESCE(SUM(final_amount),0)").Scan(&totalRevenue)

	analysis.TotalRevenue = totalRevenue
	analysis.OrderCount = orderCount
	if orderCount > 0 {
		analysis.AverageValue = totalRevenue / float64(orderCount)
	}

	// 热销产品 Top 5
	type ProductSalesRow struct {
		ProductID   int64   `gorm:"column:product_id"`
		ProductName string  `gorm:"column:product_name_snapshot"`
		SalesCount  int64   `gorm:"column:sales_count"`
		Revenue     float64 `gorm:"column:revenue"`
	}

	var topProducts []ProductSalesRow
	s.db.WithContext(ctx).
		Table("order_items oi").
		Select("oi.product_id, oi.product_name_snapshot as product_name_snapshot, SUM(oi.quantity) as sales_count, SUM(oi.final_price) as revenue").
		Joins("INNER JOIN orders o ON oi.order_id = o.id").
		Where("o.created_at >= ?", since.Unix()).
		Group("oi.product_id, oi.product_name_snapshot").
		Order("sales_count DESC").
		Limit(5).
		Scan(&topProducts)

	for _, row := range topProducts {
		analysis.TopProducts = append(analysis.TopProducts, analytics.ProductSales{
			ProductID:   row.ProductID,
			ProductName: row.ProductName,
			SalesCount:  row.SalesCount,
			Revenue:     row.Revenue,
			Percentage:  row.Revenue / totalRevenue * 100,
		})
	}

	return &analysis, nil
}

// GetCustomerAnalysis 获取客户分析
func (s *DashboardServiceImpl) GetCustomerAnalysis(ctx context.Context, days int) (*analytics.CustomerAnalysis, error) {
	since := time.Now().AddDate(0, 0, -days)

	var analysis analytics.CustomerAnalysis

	// 基础客户统计
	totalCustomers, _ := s.q.Customer.WithContext(ctx).Count()
	newCustomers, _ := s.q.Customer.WithContext(ctx).
		Where(s.q.Customer.CreatedAt.Gte(since)).Count()

	// 活跃客户（有订单的客户）
	var activeCustomers int64
	s.db.WithContext(ctx).
		Table("customers c").
		Select("COUNT(DISTINCT c.id)").
		Joins("INNER JOIN orders o ON c.id = o.customer_id").
		Where("o.created_at >= ?", since.Unix()).
		Scan(&activeCustomers)

	analysis.TotalCustomers = totalCustomers
	analysis.NewCustomers = newCustomers
	analysis.ActiveCustomers = activeCustomers

	if totalCustomers > 0 {
		analysis.RetentionRate = float64(activeCustomers) / float64(totalCustomers) * 100
	}

	return &analysis, nil
}

// GetRevenueTrend 获取收入趋势
func (s *DashboardServiceImpl) GetRevenueTrend(ctx context.Context, days int) ([]analytics.TimeSeriesData, error) {
	since := time.Now().AddDate(0, 0, -days)

	type DailyRevenue struct {
		Date    string  `gorm:"column:date"`
		Revenue float64 `gorm:"column:revenue"`
	}

	var dailyRevenues []DailyRevenue
	s.db.WithContext(ctx).
		Table("orders").
		Select("DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d') as date, SUM(final_amount) as revenue").
		Where("created_at >= ?", since.Unix()).
		Group("DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d')").
		Order("date ASC").
		Scan(&dailyRevenues)

	var trend []analytics.TimeSeriesData
	for _, row := range dailyRevenues {
		trend = append(trend, analytics.TimeSeriesData{
			Time:  row.Date,
			Value: row.Revenue,
		})
	}

	return trend, nil
}

// GetTopPerformers 获取业绩排行
func (s *DashboardServiceImpl) GetTopPerformers(ctx context.Context, metric string, limit int) ([]interface{}, error) {
	// 简化实现，返回空数组
	// 后续可根据需要扩展具体的业绩指标
	return []interface{}{}, nil
}