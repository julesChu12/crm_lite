package service

import (
	"context"
	"time"

	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"

	"gorm.io/gorm"
)

// DashboardService 聚合工作台统计数据的服务
// 大部分统计都是使用数据库简单聚合实现，
// 若后续性能不足可引入缓存层或异步任务更新。

type DashboardService struct {
	q  *query.Query
	db *gorm.DB
}

// NewDashboardService 创建实例并注入资源
func NewDashboardService(resManager *resource.Manager) *DashboardService {
	dbRes, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get DB resource for DashboardService: " + err.Error())
	}
	return &DashboardService{q: query.Use(dbRes.DB), db: dbRes.DB}
}

// GetOverview 返回工作台总览数据
func (s *DashboardService) GetOverview(ctx context.Context) (*dto.DashboardOverviewResponse, error) {
	var resp dto.DashboardOverviewResponse

	// 总体统计
	totalCustomers, _ := s.q.Customer.WithContext(ctx).Count()
	totalOrders, _ := s.q.Order.WithContext(ctx).Count()
	totalProducts, _ := s.q.Product.WithContext(ctx).Count()

	var totalRevenue float64
	_ = s.db.WithContext(ctx).Model(&model.Order{}).Select("COALESCE(SUM(final_amount),0)").Scan(&totalRevenue)

	resp.TotalCustomers = totalCustomers
	resp.TotalOrders = totalOrders
	resp.TotalProducts = totalProducts
	resp.TotalRevenue = totalRevenue

	// 当月统计
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthlyNewCustomers, _ := s.q.Customer.WithContext(ctx).Where(s.q.Customer.CreatedAt.Gte(firstOfMonth)).Count()
	monthlyOrders, _ := s.q.Order.WithContext(ctx).Where(s.q.Order.OrderDate.Gte(firstOfMonth)).Count()

	var monthlyRevenue float64
	_ = s.db.WithContext(ctx).Model(&model.Order{}).Where("order_date >= ?", firstOfMonth).Select("COALESCE(SUM(final_amount),0)").Scan(&monthlyRevenue)

	resp.MonthlyNewCustomers = monthlyNewCustomers
	resp.MonthlyOrders = monthlyOrders
	resp.MonthlyRevenue = monthlyRevenue

	// 今日统计
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayNewCustomers, _ := s.q.Customer.WithContext(ctx).Where(s.q.Customer.CreatedAt.Gte(startOfDay)).Count()
	todayOrders, _ := s.q.Order.WithContext(ctx).Where(s.q.Order.OrderDate.Gte(startOfDay)).Count()

	var todayRevenue float64
	_ = s.db.WithContext(ctx).Model(&model.Order{}).Where("order_date >= ?", startOfDay).Select("COALESCE(SUM(final_amount),0)").Scan(&todayRevenue)

	resp.TodayNewCustomers = todayNewCustomers
	resp.TodayOrders = todayOrders
	resp.TodayRevenue = todayRevenue

	// 活跃度指标
	activeCampaigns, _ := s.q.MarketingCampaign.WithContext(ctx).Where(s.q.MarketingCampaign.Status.Eq("active")).Count()
	pendingActivities, _ := s.q.Activity.WithContext(ctx).Where(s.q.Activity.Status.In("planned", "in_progress")).Count()
	lowStockProducts, _ := s.q.Product.WithContext(ctx).Where(s.q.Product.StockQuantity.LtCol(s.q.Product.MinStockLevel)).Count()

	resp.ActiveMarketingCampaigns = activeCampaigns
	resp.PendingActivities = pendingActivities
	resp.LowStockProducts = lowStockProducts

	// 增长率 (简单计算：与上月对比)
	prevMonth := firstOfMonth.AddDate(0, -1, 0)
	prevMonthStart := prevMonth
	prevMonthEnd := firstOfMonth

	prevCustomers, _ := s.q.Customer.WithContext(ctx).Where(s.q.Customer.CreatedAt.Gte(prevMonthStart), s.q.Customer.CreatedAt.Lt(prevMonthEnd)).Count()
	prevOrders, _ := s.q.Order.WithContext(ctx).Where(s.q.Order.OrderDate.Gte(prevMonthStart), s.q.Order.OrderDate.Lt(prevMonthEnd)).Count()
	var prevRevenue float64
	_ = s.db.WithContext(ctx).Model(&model.Order{}).Where("order_date >= ? AND order_date < ?", prevMonthStart, prevMonthEnd).Select("COALESCE(SUM(final_amount),0)").Scan(&prevRevenue)

	resp.CustomerGrowthRate = calcGrowthRate(float64(prevCustomers), float64(monthlyNewCustomers))
	resp.OrderGrowthRate = calcGrowthRate(float64(prevOrders), float64(monthlyOrders))
	resp.RevenueGrowthRate = calcGrowthRate(prevRevenue, monthlyRevenue)

	return &resp, nil
}

func calcGrowthRate(prev, current float64) float64 {
	if prev == 0 {
		if current == 0 {
			return 0
		}
		return 100 // 从0 增长到 current
	}
	return ((current - prev) / prev) * 100
}
