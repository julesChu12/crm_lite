package service

import (
	"context"
	"time"

	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"crm_lite/pkg/utils"
)

type DashboardService struct {
	q        *query.Query
	resource *resource.Manager
}

func NewDashboardService(resManager *resource.Manager) *DashboardService {
	db, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource for DashboardService: " + err.Error())
	}
	return &DashboardService{
		q:        query.Use(db.DB),
		resource: resManager,
	}
}

// GetOverview 获取工作台总览数据
func (s *DashboardService) GetOverview(ctx context.Context, req *dto.DashboardRequest) (*dto.DashboardOverviewResponse, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	thisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastMonth := thisMonth.AddDate(0, -1, 0)

	overview := &dto.DashboardOverviewResponse{}

	// 总体统计
	totalCustomers, _ := s.q.Customer.WithContext(ctx).Where(s.q.Customer.DeletedAt.IsNull()).Count()
	totalOrders, _ := s.q.Order.WithContext(ctx).Where(s.q.Order.DeletedAt.IsNull()).Count()
	totalProducts, _ := s.q.Product.WithContext(ctx).Where(s.q.Product.DeletedAt.IsNull()).Count()

	overview.TotalCustomers = totalCustomers
	overview.TotalOrders = totalOrders
	overview.TotalProducts = totalProducts

	// 计算总收入（从已完成订单）
	var totalRevenue float64
	s.q.Order.WithContext(ctx).
		Where(s.q.Order.Status.Eq("completed"), s.q.Order.DeletedAt.IsNull()).
		Select(s.q.Order.TotalAmount.Sum()).
		Scan(&totalRevenue)
	overview.TotalRevenue = totalRevenue

	// 本月数据
	monthlyNewCustomers, _ := s.q.Customer.WithContext(ctx).
		Where(s.q.Customer.CreatedAt.Gte(thisMonth), s.q.Customer.DeletedAt.IsNull()).
		Count()
	monthlyOrders, _ := s.q.Order.WithContext(ctx).
		Where(s.q.Order.CreatedAt.Gte(thisMonth), s.q.Order.DeletedAt.IsNull()).
		Count()

	var monthlyRevenue float64
	s.q.Order.WithContext(ctx).
		Where(s.q.Order.Status.Eq("completed"),
			s.q.Order.CreatedAt.Gte(thisMonth),
			s.q.Order.DeletedAt.IsNull()).
		Select(s.q.Order.TotalAmount.Sum()).
		Scan(&monthlyRevenue)

	overview.MonthlyNewCustomers = monthlyNewCustomers
	overview.MonthlyOrders = monthlyOrders
	overview.MonthlyRevenue = monthlyRevenue

	// 计算增长率（与上月对比）
	lastMonthCustomers, _ := s.q.Customer.WithContext(ctx).
		Where(s.q.Customer.CreatedAt.Gte(lastMonth),
			s.q.Customer.CreatedAt.Lt(thisMonth),
			s.q.Customer.DeletedAt.IsNull()).
		Count()

	lastMonthOrders, _ := s.q.Order.WithContext(ctx).
		Where(s.q.Order.CreatedAt.Gte(lastMonth),
			s.q.Order.CreatedAt.Lt(thisMonth),
			s.q.Order.DeletedAt.IsNull()).
		Count()

	var lastMonthRevenue float64
	s.q.Order.WithContext(ctx).
		Where(s.q.Order.Status.Eq("completed"),
			s.q.Order.CreatedAt.Gte(lastMonth),
			s.q.Order.CreatedAt.Lt(thisMonth),
			s.q.Order.DeletedAt.IsNull()).
		Select(s.q.Order.TotalAmount.Sum()).
		Scan(&lastMonthRevenue)

	overview.CustomerGrowthRate = s.calculateGrowthRate(float64(monthlyNewCustomers), float64(lastMonthCustomers))
	overview.OrderGrowthRate = s.calculateGrowthRate(float64(monthlyOrders), float64(lastMonthOrders))
	overview.RevenueGrowthRate = s.calculateGrowthRate(monthlyRevenue, lastMonthRevenue)

	// 今日数据
	todayNewCustomers, _ := s.q.Customer.WithContext(ctx).
		Where(s.q.Customer.CreatedAt.Gte(today), s.q.Customer.DeletedAt.IsNull()).
		Count()
	todayOrders, _ := s.q.Order.WithContext(ctx).
		Where(s.q.Order.CreatedAt.Gte(today), s.q.Order.DeletedAt.IsNull()).
		Count()

	var todayRevenue float64
	s.q.Order.WithContext(ctx).
		Where(s.q.Order.Status.Eq("completed"),
			s.q.Order.CreatedAt.Gte(today),
			s.q.Order.DeletedAt.IsNull()).
		Select(s.q.Order.TotalAmount.Sum()).
		Scan(&todayRevenue)

	overview.TodayNewCustomers = todayNewCustomers
	overview.TodayOrders = todayOrders
	overview.TodayRevenue = todayRevenue

	// 活跃度指标
	activeCampaigns, _ := s.q.MarketingCampaign.WithContext(ctx).
		Where(s.q.MarketingCampaign.Status.Eq("active")).
		Count()

	pendingActivities, _ := s.q.Activity.WithContext(ctx).
		Where(s.q.Activity.Status.In("planned", "in_progress"), s.q.Activity.DeletedAt.IsNull()).
		Count()

	lowStockProducts, _ := s.q.Product.WithContext(ctx).
		Where(s.q.Product.StockQuantity.LtCol(s.q.Product.MinStockLevel),
			s.q.Product.DeletedAt.IsNull()).
		Count()

	overview.ActiveMarketingCampaigns = activeCampaigns
	overview.PendingActivities = pendingActivities
	overview.LowStockProducts = lowStockProducts

	return overview, nil
}

// GetCustomerAnalytics 获取客户分析数据
func (s *DashboardService) GetCustomerAnalytics(ctx context.Context) (*dto.CustomerAnalyticsResponse, error) {
	analytics := &dto.CustomerAnalyticsResponse{}

	// 客户等级分布
	type LevelCount struct {
		Level string
		Count int64
	}
	var levelCounts []LevelCount
	s.q.Customer.WithContext(ctx).
		Select(s.q.Customer.Level, s.q.Customer.Level.Count().As("count")).
		Where(s.q.Customer.DeletedAt.IsNull()).
		Group(s.q.Customer.Level).
		Scan(&levelCounts)

	totalCustomers, _ := s.q.Customer.WithContext(ctx).
		Where(s.q.Customer.DeletedAt.IsNull()).Count()

	for _, lc := range levelCounts {
		percentage := float64(lc.Count) / float64(totalCustomers) * 100
		item := struct {
			Level      string  `json:"level" example:"VIP"`
			Count      int64   `json:"count" example:"125"`
			Percentage float64 `json:"percentage" example:"10.0"`
		}{
			Level:      lc.Level,
			Count:      lc.Count,
			Percentage: percentage,
		}
		analytics.CustomerLevelDistribution = append(analytics.CustomerLevelDistribution, item)
	}

	// 客户来源分布
	type SourceCount struct {
		Source string
		Count  int64
	}
	var sourceCounts []SourceCount
	s.q.Customer.WithContext(ctx).
		Select(s.q.Customer.Source, s.q.Customer.Source.Count().As("count")).
		Where(s.q.Customer.DeletedAt.IsNull()).
		Group(s.q.Customer.Source).
		Scan(&sourceCounts)

	for _, sc := range sourceCounts {
		percentage := float64(sc.Count) / float64(totalCustomers) * 100
		item := struct {
			Source     string  `json:"source" example:"微信"`
			Count      int64   `json:"count" example:"356"`
			Percentage float64 `json:"percentage" example:"28.5"`
		}{
			Source:     sc.Source,
			Count:      sc.Count,
			Percentage: percentage,
		}
		analytics.CustomerSourceDistribution = append(analytics.CustomerSourceDistribution, item)
	}

	// 客户性别分布
	type GenderCount struct {
		Gender string
		Count  int64
	}
	var genderCounts []GenderCount
	s.q.Customer.WithContext(ctx).
		Select(s.q.Customer.Gender, s.q.Customer.Gender.Count().As("count")).
		Where(s.q.Customer.DeletedAt.IsNull()).
		Group(s.q.Customer.Gender).
		Scan(&genderCounts)

	for _, gc := range genderCounts {
		percentage := float64(gc.Count) / float64(totalCustomers) * 100
		item := struct {
			Gender     string  `json:"gender" example:"male"`
			Count      int64   `json:"count" example:"687"`
			Percentage float64 `json:"percentage" example:"55.0"`
		}{
			Gender:     gc.Gender,
			Count:      gc.Count,
			Percentage: percentage,
		}
		analytics.CustomerGenderDistribution = append(analytics.CustomerGenderDistribution, item)
	}

	// 新客户趋势（最近12个月）
	now := time.Now()
	for i := 11; i >= 0; i-- {
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -i, 0)
		monthEnd := monthStart.AddDate(0, 1, 0)

		newCustomers, _ := s.q.Customer.WithContext(ctx).
			Where(s.q.Customer.CreatedAt.Gte(monthStart),
				s.q.Customer.CreatedAt.Lt(monthEnd),
				s.q.Customer.DeletedAt.IsNull()).
			Count()

		totalCustomersAtMonth, _ := s.q.Customer.WithContext(ctx).
			Where(s.q.Customer.CreatedAt.Lt(monthEnd),
				s.q.Customer.DeletedAt.IsNull()).
			Count()

		item := struct {
			Month          string `json:"month" example:"2024-01"`
			NewCustomers   int64  `json:"new_customers" example:"45"`
			TotalCustomers int64  `json:"total_customers" example:"1205"`
		}{
			Month:          monthStart.Format("2006-01"),
			NewCustomers:   newCustomers,
			TotalCustomers: totalCustomersAtMonth,
		}
		analytics.CustomerTrend = append(analytics.CustomerTrend, item)
	}

	return analytics, nil
}

// GetSalesAnalytics 获取销售分析数据
func (s *DashboardService) GetSalesAnalytics(ctx context.Context) (*dto.SalesAnalyticsResponse, error) {
	analytics := &dto.SalesAnalyticsResponse{}

	// 销售趋势（最近12个月）
	now := time.Now()
	for i := 11; i >= 0; i-- {
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -i, 0)
		monthEnd := monthStart.AddDate(0, 1, 0)

		orders, _ := s.q.Order.WithContext(ctx).
			Where(s.q.Order.CreatedAt.Gte(monthStart),
				s.q.Order.CreatedAt.Lt(monthEnd),
				s.q.Order.DeletedAt.IsNull()).
			Count()

		var revenue float64
		s.q.Order.WithContext(ctx).
			Where(s.q.Order.Status.Eq("completed"),
				s.q.Order.CreatedAt.Gte(monthStart),
				s.q.Order.CreatedAt.Lt(monthEnd),
				s.q.Order.DeletedAt.IsNull()).
			Select(s.q.Order.TotalAmount.Sum()).
			Scan(&revenue)

		analytics.SalesTrend = append(analytics.SalesTrend, struct {
			Month   string  `json:"month" example:"2024-01"`
			Orders  int64   `json:"orders" example:"289"`
			Revenue float64 `json:"revenue" example:"156800.00"`
		}{
			Month:   monthStart.Format("2006-01"),
			Orders:  orders,
			Revenue: revenue,
		})
	}

	// 订单状态分布
	type StatusCount struct {
		Status string
		Count  int64
	}
	var statusCounts []StatusCount
	s.q.Order.WithContext(ctx).
		Select(s.q.Order.Status, s.q.Order.Status.Count().As("count")).
		Where(s.q.Order.DeletedAt.IsNull()).
		Group(s.q.Order.Status).
		Scan(&statusCounts)

	totalOrders, _ := s.q.Order.WithContext(ctx).
		Where(s.q.Order.DeletedAt.IsNull()).Count()

	for _, sc := range statusCounts {
		percentage := float64(sc.Count) / float64(totalOrders) * 100
		analytics.OrderStatusDistribution = append(analytics.OrderStatusDistribution, struct {
			Status     string  `json:"status" example:"completed"`
			Count      int64   `json:"count" example:"2896"`
			Percentage float64 `json:"percentage" example:"83.8"`
		}{
			Status:     sc.Status,
			Count:      sc.Count,
			Percentage: percentage,
		})
	}

	// 客单价分析
	thisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastMonth := thisMonth.AddDate(0, -1, 0)

	var currentAOV, previousAOV float64
	s.q.Order.WithContext(ctx).
		Where(s.q.Order.Status.Eq("completed"),
			s.q.Order.CreatedAt.Gte(thisMonth),
			s.q.Order.DeletedAt.IsNull()).
		Select(s.q.Order.TotalAmount.Avg()).
		Scan(&currentAOV)

	s.q.Order.WithContext(ctx).
		Where(s.q.Order.Status.Eq("completed"),
			s.q.Order.CreatedAt.Gte(lastMonth),
			s.q.Order.CreatedAt.Lt(thisMonth),
			s.q.Order.DeletedAt.IsNull()).
		Select(s.q.Order.TotalAmount.Avg()).
		Scan(&previousAOV)

	aovGrowthRate := s.calculateGrowthRate(currentAOV, previousAOV)

	analytics.AverageOrderValue = struct {
		Current    float64 `json:"current" example:"708.50"`
		Previous   float64 `json:"previous" example:"652.20"`
		GrowthRate float64 `json:"growth_rate" example:"8.6"`
	}{
		Current:    currentAOV,
		Previous:   previousAOV,
		GrowthRate: aovGrowthRate,
	}

	// 复购率分析
	type CustomerOrderCount struct {
		CustomerID int64
		OrderCount int64
	}
	var customerOrderCounts []CustomerOrderCount
	s.q.Order.WithContext(ctx).
		Select(s.q.Order.CustomerID, s.q.Order.CustomerID.Count().As("order_count")).
		Where(s.q.Order.Status.Eq("completed"), s.q.Order.DeletedAt.IsNull()).
		Group(s.q.Order.CustomerID).
		Having(s.q.Order.CustomerID.Count().Gt(1)).
		Scan(&customerOrderCounts)

	repeatCustomers := int64(len(customerOrderCounts))
	totalCustomersWithOrders, _ := s.q.Order.WithContext(ctx).
		Select(s.q.Order.CustomerID.Count().As("distinct_customers")).
		Where(s.q.Order.Status.Eq("completed"), s.q.Order.DeletedAt.IsNull()).
		Group(s.q.Order.CustomerID).
		Count()

	repeatRate := float64(0)
	if totalCustomersWithOrders > 0 {
		repeatRate = float64(repeatCustomers) / float64(totalCustomersWithOrders) * 100
	}

	analytics.RepeatPurchaseRate = struct {
		Rate            float64 `json:"rate" example:"32.5"`
		RepeatCustomers int64   `json:"repeat_customers" example:"406"`
		TotalCustomers  int64   `json:"total_customers" example:"1250"`
	}{
		Rate:            repeatRate,
		RepeatCustomers: repeatCustomers,
		TotalCustomers:  totalCustomersWithOrders,
	}

	return analytics, nil
}

// GetProductAnalytics 获取产品分析数据
func (s *DashboardService) GetProductAnalytics(ctx context.Context) (*dto.ProductAnalyticsResponse, error) {
	analytics := &dto.ProductAnalyticsResponse{}

	// 热销产品排行（Top 10）
	type ProductSales struct {
		ID           int64
		Name         string
		Category     string
		SoldQuantity int32
		Revenue      float64
	}

	var topProducts []ProductSales
	s.q.Product.WithContext(ctx).
		Select(s.q.Product.ID, s.q.Product.Name, s.q.Product.Category,
			s.q.OrderItem.Quantity.Sum().As("sold_quantity"),
			s.q.OrderItem.FinalPrice.Sum().As("revenue")).
		LeftJoin(s.q.OrderItem, s.q.OrderItem.ProductID.EqCol(s.q.Product.ID)).
		LeftJoin(s.q.Order, s.q.Order.ID.EqCol(s.q.OrderItem.OrderID)).
		Where(s.q.Order.Status.Eq("completed"), s.q.Product.DeletedAt.IsNull()).
		Group(s.q.Product.ID).
		Order(s.q.OrderItem.FinalPrice.Sum().Desc()).
		Limit(10).
		Scan(&topProducts)

	for _, p := range topProducts {
		analytics.TopSellingProducts = append(analytics.TopSellingProducts, struct {
			ID           int64   `json:"id" example:"1"`
			Name         string  `json:"name" example:"iPhone 15 Pro"`
			Category     string  `json:"category" example:"数码产品"`
			SoldQuantity int32   `json:"sold_quantity" example:"156"`
			Revenue      float64 `json:"revenue" example:"156000.00"`
		}{
			ID:           p.ID,
			Name:         p.Name,
			Category:     p.Category,
			SoldQuantity: p.SoldQuantity,
			Revenue:      p.Revenue,
		})
	}

	// 产品类别销售分布
	type CategorySales struct {
		Category string
		Products int64
		Revenue  float64
	}

	var categorySales []CategorySales
	s.q.Product.WithContext(ctx).
		Select(s.q.Product.Category,
			s.q.Product.ID.Count().As("products"),
			s.q.OrderItem.FinalPrice.Sum().As("revenue")).
		LeftJoin(s.q.OrderItem, s.q.OrderItem.ProductID.EqCol(s.q.Product.ID)).
		LeftJoin(s.q.Order, s.q.Order.ID.EqCol(s.q.OrderItem.OrderID)).
		Where(s.q.Order.Status.Eq("completed"), s.q.Product.DeletedAt.IsNull()).
		Group(s.q.Product.Category).
		Scan(&categorySales)

	var totalCategoryRevenue float64
	for _, cs := range categorySales {
		totalCategoryRevenue += cs.Revenue
	}

	for _, cs := range categorySales {
		percentage := float64(0)
		if totalCategoryRevenue > 0 {
			percentage = cs.Revenue / totalCategoryRevenue * 100
		}
		analytics.CategorySalesDistribution = append(analytics.CategorySalesDistribution, struct {
			Category   string  `json:"category" example:"数码产品"`
			Products   int64   `json:"products" example:"23"`
			Revenue    float64 `json:"revenue" example:"890000.00"`
			Percentage float64 `json:"percentage" example:"36.3"`
		}{
			Category:   cs.Category,
			Products:   cs.Products,
			Revenue:    cs.Revenue,
			Percentage: percentage,
		})
	}

	// 库存预警产品
	type LowStockProduct struct {
		ID            int64
		Name          string
		CurrentStock  int32
		MinStockLevel int32
		Category      string
	}

	var lowStockProducts []LowStockProduct
	s.q.Product.WithContext(ctx).
		Select(s.q.Product.ID, s.q.Product.Name, s.q.Product.StockQuantity.As("current_stock"),
			s.q.Product.MinStockLevel, s.q.Product.Category).
		Where(s.q.Product.StockQuantity.LtCol(s.q.Product.MinStockLevel),
			s.q.Product.DeletedAt.IsNull()).
		Scan(&lowStockProducts)

	for _, p := range lowStockProducts {
		analytics.LowStockProducts = append(analytics.LowStockProducts, struct {
			ID            int64  `json:"id" example:"45"`
			Name          string `json:"name" example:"MacBook Air M2"`
			CurrentStock  int32  `json:"current_stock" example:"3"`
			MinStockLevel int32  `json:"min_stock_level" example:"10"`
			Category      string `json:"category" example:"数码产品"`
		}{
			ID:            p.ID,
			Name:          p.Name,
			CurrentStock:  p.CurrentStock,
			MinStockLevel: p.MinStockLevel,
			Category:      p.Category,
		})
	}

	// 产品销售趋势（最近12个月）
	now := time.Now()
	for i := 11; i >= 0; i-- {
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -i, 0)
		monthEnd := monthStart.AddDate(0, 1, 0)

		var productsSold int64
		s.q.OrderItem.WithContext(ctx).
			Select(s.q.OrderItem.Quantity.Sum().As("products_sold")).
			LeftJoin(s.q.Order, s.q.Order.ID.EqCol(s.q.OrderItem.OrderID)).
			Where(s.q.Order.Status.Eq("completed"),
				s.q.Order.CreatedAt.Gte(monthStart),
				s.q.Order.CreatedAt.Lt(monthEnd),
				s.q.Order.DeletedAt.IsNull()).
			Scan(&productsSold)

		categoriesActive, _ := s.q.Product.WithContext(ctx).
			Select(s.q.Product.Category.Count().As("categories_active")).
			LeftJoin(s.q.OrderItem, s.q.OrderItem.ProductID.EqCol(s.q.Product.ID)).
			LeftJoin(s.q.Order, s.q.Order.ID.EqCol(s.q.OrderItem.OrderID)).
			Where(s.q.Order.Status.Eq("completed"),
				s.q.Order.CreatedAt.Gte(monthStart),
				s.q.Order.CreatedAt.Lt(monthEnd),
				s.q.Product.DeletedAt.IsNull()).
			Group(s.q.Product.Category).
			Count()

		analytics.ProductSalesTrend = append(analytics.ProductSalesTrend, struct {
			Month            string `json:"month" example:"2024-01"`
			ProductsSold     int64  `json:"products_sold" example:"1234"`
			CategoriesActive int64  `json:"categories_active" example:"12"`
		}{
			Month:            monthStart.Format("2006-01"),
			ProductsSold:     productsSold,
			CategoriesActive: categoriesActive,
		})
	}

	return analytics, nil
}

// GetMarketingAnalytics 获取营销分析数据
func (s *DashboardService) GetMarketingAnalytics(ctx context.Context) (*dto.MarketingAnalyticsResponse, error) {
	analytics := &dto.MarketingAnalyticsResponse{}

	// 营销活动概览
	totalCampaigns, _ := s.q.MarketingCampaign.WithContext(ctx).Count()
	activeCampaigns, _ := s.q.MarketingCampaign.WithContext(ctx).
		Where(s.q.MarketingCampaign.Status.Eq("active")).Count()
	completedCampaigns, _ := s.q.MarketingCampaign.WithContext(ctx).
		Where(s.q.MarketingCampaign.Status.Eq("completed")).Count()

	var totalReach int64
	s.q.MarketingCampaign.WithContext(ctx).
		Select(s.q.MarketingCampaign.TargetCount.Sum()).
		Scan(&totalReach)

	analytics.CampaignOverview = struct {
		TotalCampaigns     int64 `json:"total_campaigns" example:"28"`
		ActiveCampaigns    int64 `json:"active_campaigns" example:"5"`
		CompletedCampaigns int64 `json:"completed_campaigns" example:"20"`
		TotalReach         int64 `json:"total_reach" example:"15600"`
	}{
		TotalCampaigns:     totalCampaigns,
		ActiveCampaigns:    activeCampaigns,
		CompletedCampaigns: completedCampaigns,
		TotalReach:         totalReach,
	}

	// 营销渠道效果（简化实现）
	type ChannelStats struct {
		Channel      string
		Campaigns    int64
		Reach        int64
		DeliveryRate float64
		OpenRate     float64
		ClickRate    float64
	}

	var channelStats []ChannelStats
	s.q.MarketingCampaign.WithContext(ctx).
		Select(s.q.MarketingCampaign.Type.As("channel"),
			s.q.MarketingCampaign.ID.Count().As("campaigns"),
			s.q.MarketingCampaign.TargetCount.Sum().As("reach")).
		Group(s.q.MarketingCampaign.Type).
		Scan(&channelStats)

	for _, cs := range channelStats {
		// 简化计算，实际应该从营销记录中统计
		analytics.ChannelPerformance = append(analytics.ChannelPerformance, struct {
			Channel      string  `json:"channel" example:"sms"`
			Campaigns    int64   `json:"campaigns" example:"12"`
			Reach        int64   `json:"reach" example:"5600"`
			DeliveryRate float64 `json:"delivery_rate" example:"96.5"`
			OpenRate     float64 `json:"open_rate" example:"32.1"`
			ClickRate    float64 `json:"click_rate" example:"8.9"`
		}{
			Channel:      cs.Channel,
			Campaigns:    cs.Campaigns,
			Reach:        cs.Reach,
			DeliveryRate: 95.0 + float64(cs.Campaigns%5),  // 模拟数据
			OpenRate:     25.0 + float64(cs.Campaigns%15), // 模拟数据
			ClickRate:    5.0 + float64(cs.Campaigns%10),  // 模拟数据
		})
	}

	return analytics, nil
}

// GetFinancialAnalytics 获取财务分析数据
func (s *DashboardService) GetFinancialAnalytics(ctx context.Context) (*dto.FinancialAnalyticsResponse, error) {
	analytics := &dto.FinancialAnalyticsResponse{}
	now := time.Now()

	// 收入概览
	var totalRevenue float64
	s.q.Order.WithContext(ctx).
		Where(s.q.Order.Status.Eq("completed"), s.q.Order.DeletedAt.IsNull()).
		Select(s.q.Order.TotalAmount.Sum()).
		Scan(&totalRevenue)

	thisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastMonth := thisMonth.AddDate(0, -1, 0)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var monthlyRevenue, dailyRevenue, lastMonthRevenue float64

	s.q.Order.WithContext(ctx).
		Where(s.q.Order.Status.Eq("completed"),
			s.q.Order.CreatedAt.Gte(thisMonth),
			s.q.Order.DeletedAt.IsNull()).
		Select(s.q.Order.TotalAmount.Sum()).
		Scan(&monthlyRevenue)

	s.q.Order.WithContext(ctx).
		Where(s.q.Order.Status.Eq("completed"),
			s.q.Order.CreatedAt.Gte(today),
			s.q.Order.DeletedAt.IsNull()).
		Select(s.q.Order.TotalAmount.Sum()).
		Scan(&dailyRevenue)

	s.q.Order.WithContext(ctx).
		Where(s.q.Order.Status.Eq("completed"),
			s.q.Order.CreatedAt.Gte(lastMonth),
			s.q.Order.CreatedAt.Lt(thisMonth),
			s.q.Order.DeletedAt.IsNull()).
		Select(s.q.Order.TotalAmount.Sum()).
		Scan(&lastMonthRevenue)

	growthRate := s.calculateGrowthRate(monthlyRevenue, lastMonthRevenue)

	analytics.RevenueOverview = struct {
		TotalRevenue   float64 `json:"total_revenue" example:"2450000.00"`
		MonthlyRevenue float64 `json:"monthly_revenue" example:"186500.00"`
		DailyRevenue   float64 `json:"daily_revenue" example:"8500.00"`
		GrowthRate     float64 `json:"growth_rate" example:"15.2"`
	}{
		TotalRevenue:   totalRevenue,
		MonthlyRevenue: monthlyRevenue,
		DailyRevenue:   dailyRevenue,
		GrowthRate:     growthRate,
	}

	// 钱包统计
	totalWallets, _ := s.q.Wallet.WithContext(ctx).Count()
	activeWallets, _ := s.q.Wallet.WithContext(ctx).
		Where(s.q.Wallet.Balance.Gt(0)).Count()

	var totalBalance, totalRecharge, totalConsumption float64
	s.q.Wallet.WithContext(ctx).
		Select(s.q.Wallet.Balance.Sum()).
		Scan(&totalBalance)

	s.q.WalletTransaction.WithContext(ctx).
		Where(s.q.WalletTransaction.Type.Eq("recharge")).
		Select(s.q.WalletTransaction.Amount.Sum()).
		Scan(&totalRecharge)

	s.q.WalletTransaction.WithContext(ctx).
		Where(s.q.WalletTransaction.Type.Eq("consumption")).
		Select(s.q.WalletTransaction.Amount.Sum()).
		Scan(&totalConsumption)

	analytics.WalletStats = struct {
		TotalWallets     int64   `json:"total_wallets" example:"1180"`
		ActiveWallets    int64   `json:"active_wallets" example:"856"`
		TotalBalance     float64 `json:"total_balance" example:"125600.00"`
		TotalRecharge    float64 `json:"total_recharge" example:"356000.00"`
		TotalConsumption float64 `json:"total_consumption" example:"230400.00"`
	}{
		TotalWallets:     totalWallets,
		ActiveWallets:    activeWallets,
		TotalBalance:     totalBalance,
		TotalRecharge:    totalRecharge,
		TotalConsumption: totalConsumption,
	}

	return analytics, nil
}

// GetActivitySummary 获取活动摘要数据
func (s *DashboardService) GetActivitySummary(ctx context.Context) (*dto.ActivitySummaryResponse, error) {
	summary := &dto.ActivitySummaryResponse{}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 待办活动统计
	total, _ := s.q.Activity.WithContext(ctx).
		Where(s.q.Activity.Status.In("planned", "in_progress"),
			s.q.Activity.DeletedAt.IsNull()).Count()

	highPriority, _ := s.q.Activity.WithContext(ctx).
		Where(s.q.Activity.Status.In("planned", "in_progress"),
			s.q.Activity.Priority.Eq("high"),
			s.q.Activity.DeletedAt.IsNull()).Count()

	overdue, _ := s.q.Activity.WithContext(ctx).
		Where(s.q.Activity.Status.In("planned", "in_progress"),
			s.q.Activity.ScheduledAt.Lt(now),
			s.q.Activity.DeletedAt.IsNull()).Count()

	dueToday, _ := s.q.Activity.WithContext(ctx).
		Where(s.q.Activity.Status.In("planned", "in_progress"),
			s.q.Activity.ScheduledAt.Gte(today),
			s.q.Activity.ScheduledAt.Lt(today.Add(24*time.Hour)),
			s.q.Activity.DeletedAt.IsNull()).Count()

	summary.PendingActivities = struct {
		Total        int64 `json:"total" example:"23"`
		HighPriority int64 `json:"high_priority" example:"5"`
		Overdue      int64 `json:"overdue" example:"3"`
		DueToday     int64 `json:"due_today" example:"8"`
	}{
		Total:        total,
		HighPriority: highPriority,
		Overdue:      overdue,
		DueToday:     dueToday,
	}

	// 活动类型分布
	type TypeCount struct {
		Type  string
		Count int64
	}
	var typeCounts []TypeCount
	s.q.Activity.WithContext(ctx).
		Select(s.q.Activity.Type, s.q.Activity.Type.Count().As("count")).
		Where(s.q.Activity.DeletedAt.IsNull()).
		Group(s.q.Activity.Type).
		Scan(&typeCounts)

	totalActivities, _ := s.q.Activity.WithContext(ctx).
		Where(s.q.Activity.DeletedAt.IsNull()).Count()

	for _, tc := range typeCounts {
		percentage := float64(tc.Count) / float64(totalActivities) * 100
		summary.ActivityTypeDistribution = append(summary.ActivityTypeDistribution, struct {
			Type       string  `json:"type" example:"call"`
			Count      int64   `json:"count" example:"156"`
			Percentage float64 `json:"percentage" example:"35.2"`
		}{
			Type:       tc.Type,
			Count:      tc.Count,
			Percentage: percentage,
		})
	}

	return summary, nil
}

// GetQuickStats 获取快速统计数据
func (s *DashboardService) GetQuickStats(ctx context.Context) (*dto.QuickStatsResponse, error) {
	stats := &dto.QuickStatsResponse{}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 待处理订单
	pendingOrders, _ := s.q.Order.WithContext(ctx).
		Where(s.q.Order.Status.In("pending", "processing"),
			s.q.Order.DeletedAt.IsNull()).Count()

	// 今日收入
	var todayRevenue float64
	s.q.Order.WithContext(ctx).
		Where(s.q.Order.Status.Eq("completed"),
			s.q.Order.CreatedAt.Gte(today),
			s.q.Order.DeletedAt.IsNull()).
		Select(s.q.Order.TotalAmount.Sum()).
		Scan(&todayRevenue)

	// 活跃营销活动
	activeCampaigns, _ := s.q.MarketingCampaign.WithContext(ctx).
		Where(s.q.MarketingCampaign.Status.Eq("active")).Count()

	stats.OnlineUsers = 23 // 模拟数据，实际需要从会话或其他地方获取
	stats.PendingOrders = pendingOrders
	stats.TodayRevenue = todayRevenue
	stats.ActiveCampaigns = activeCampaigns
	stats.SystemAlerts = 2 // 模拟数据
	stats.LastUpdated = utils.FormatTime(now)

	return stats, nil
}

// calculateGrowthRate 计算增长率
func (s *DashboardService) calculateGrowthRate(current, previous float64) float64 {
	if previous == 0 {
		if current > 0 {
			return 100.0
		}
		return 0.0
	}
	return ((current - previous) / previous) * 100
}
