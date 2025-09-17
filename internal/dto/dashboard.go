package dto

// ================ 工作台数据大盘相关DTO ================

// DashboardOverviewResponse 工作台总览数据响应
type DashboardOverviewResponse struct {
	// 总体统计
	TotalCustomers int64   `json:"total_customers" example:"1250"`
	TotalOrders    int64   `json:"total_orders" example:"3456"`
	TotalRevenue   float64 `json:"total_revenue" example:"2450000.00"`
	TotalProducts  int64   `json:"total_products" example:"89"`

	// 本月数据
	MonthlyNewCustomers int64   `json:"monthly_new_customers" example:"56"`
	MonthlyOrders       int64   `json:"monthly_orders" example:"234"`
	MonthlyRevenue      float64 `json:"monthly_revenue" example:"186500.00"`

	// 增长率（与上月对比）
	CustomerGrowthRate float64 `json:"customer_growth_rate" example:"12.5"` // 客户增长率
	OrderGrowthRate    float64 `json:"order_growth_rate" example:"8.3"`     // 订单增长率
	RevenueGrowthRate  float64 `json:"revenue_growth_rate" example:"15.2"`  // 收入增长率

	// 今日数据
    TodayNewCustomers int64   `json:"today_new_customers" example:"3"`
    TodayOrders       int64   `json:"today_orders" example:"12"`
    TodayRevenue      float64 `json:"today_revenue" example:"8500.00"`

    // 钱包相关统计（以钱包消费作为收入口径）
    TotalWallets     int64   `json:"total_wallets" example:"1180"`
    TotalBalance     float64 `json:"total_balance" example:"125600.00"`
    TotalRecharge    float64 `json:"total_recharge" example:"356000.00"`
    TotalConsumption float64 `json:"total_consumption" example:"230400.00"`

    // 收入同比/环比（以本月为周期，来源：钱包消费）
    RevenueMoMRate float64 `json:"revenue_mom_rate" example:"8.6"`  // 月环比
    RevenueYoYRate float64 `json:"revenue_yoy_rate" example:"12.3"` // 年同比

	// 活跃度指标
	ActiveMarketingCampaigns int64 `json:"active_marketing_campaigns" example:"5"`
	PendingActivities        int64 `json:"pending_activities" example:"23"`
	LowStockProducts         int64 `json:"low_stock_products" example:"7"`
}

// CustomerAnalyticsResponse 客户分析数据响应
type CustomerAnalyticsResponse struct {
	// 客户等级分布
	CustomerLevelDistribution []struct {
		Level      string  `json:"level" example:"VIP"`
		Count      int64   `json:"count" example:"125"`
		Percentage float64 `json:"percentage" example:"10.0"`
	} `json:"customer_level_distribution"`

	// 客户来源分布
	CustomerSourceDistribution []struct {
		Source     string  `json:"source" example:"微信"`
		Count      int64   `json:"count" example:"356"`
		Percentage float64 `json:"percentage" example:"28.5"`
	} `json:"customer_source_distribution"`

	// 客户性别分布
	CustomerGenderDistribution []struct {
		Gender     string  `json:"gender" example:"male"`
		Count      int64   `json:"count" example:"687"`
		Percentage float64 `json:"percentage" example:"55.0"`
	} `json:"customer_gender_distribution"`

	// 客户年龄分布
	CustomerAgeDistribution []struct {
		AgeRange   string  `json:"age_range" example:"18-25"`
		Count      int64   `json:"count" example:"234"`
		Percentage float64 `json:"percentage" example:"18.7"`
	} `json:"customer_age_distribution"`

	// 新客户趋势（最近12个月）
	CustomerTrend []struct {
		Month          string `json:"month" example:"2024-01"`
		NewCustomers   int64  `json:"new_customers" example:"45"`
		TotalCustomers int64  `json:"total_customers" example:"1205"`
	} `json:"customer_trend"`
}

// SalesAnalyticsResponse 销售分析数据响应
type SalesAnalyticsResponse struct {
	// 销售趋势（最近12个月）
	SalesTrend []struct {
		Month   string  `json:"month" example:"2024-01"`
		Orders  int64   `json:"orders" example:"289"`
		Revenue float64 `json:"revenue" example:"156800.00"`
	} `json:"sales_trend"`

	// 销售渠道分布
	SalesChannelDistribution []struct {
		Channel    string  `json:"channel" example:"线上"`
		Orders     int64   `json:"orders" example:"2134"`
		Revenue    float64 `json:"revenue" example:"1890000.00"`
		Percentage float64 `json:"percentage" example:"77.1"`
	} `json:"sales_channel_distribution"`

	// 订单状态分布
	OrderStatusDistribution []struct {
		Status     string  `json:"status" example:"completed"`
		Count      int64   `json:"count" example:"2896"`
		Percentage float64 `json:"percentage" example:"83.8"`
	} `json:"order_status_distribution"`

	// 客单价分析
	AverageOrderValue struct {
		Current    float64 `json:"current" example:"708.50"`
		Previous   float64 `json:"previous" example:"652.20"`
		GrowthRate float64 `json:"growth_rate" example:"8.6"`
	} `json:"average_order_value"`

	// 复购率分析
	RepeatPurchaseRate struct {
		Rate            float64 `json:"rate" example:"32.5"`
		RepeatCustomers int64   `json:"repeat_customers" example:"406"`
		TotalCustomers  int64   `json:"total_customers" example:"1250"`
	} `json:"repeat_purchase_rate"`
}

// ProductAnalyticsResponse 产品分析数据响应
type ProductAnalyticsResponse struct {
	// 热销产品排行（Top 10）
	TopSellingProducts []struct {
		ID           int64   `json:"id" example:"1"`
		Name         string  `json:"name" example:"iPhone 15 Pro"`
		Category     string  `json:"category" example:"数码产品"`
		SoldQuantity int32   `json:"sold_quantity" example:"156"`
		Revenue      float64 `json:"revenue" example:"156000.00"`
	} `json:"top_selling_products"`

	// 产品类别销售分布
	CategorySalesDistribution []struct {
		Category   string  `json:"category" example:"数码产品"`
		Products   int64   `json:"products" example:"23"`
		Revenue    float64 `json:"revenue" example:"890000.00"`
		Percentage float64 `json:"percentage" example:"36.3"`
	} `json:"category_sales_distribution"`

	// 库存预警产品
	LowStockProducts []struct {
		ID            int64  `json:"id" example:"45"`
		Name          string `json:"name" example:"MacBook Air M2"`
		CurrentStock  int32  `json:"current_stock" example:"3"`
		MinStockLevel int32  `json:"min_stock_level" example:"10"`
		Category      string `json:"category" example:"数码产品"`
	} `json:"low_stock_products"`

	// 产品销售趋势
	ProductSalesTrend []struct {
		Month            string `json:"month" example:"2024-01"`
		ProductsSold     int64  `json:"products_sold" example:"1234"`
		CategoriesActive int64  `json:"categories_active" example:"12"`
	} `json:"product_sales_trend"`
}

// MarketingAnalyticsResponse 营销分析数据响应
type MarketingAnalyticsResponse struct {
	// 营销活动概览
	CampaignOverview struct {
		TotalCampaigns     int64 `json:"total_campaigns" example:"28"`
		ActiveCampaigns    int64 `json:"active_campaigns" example:"5"`
		CompletedCampaigns int64 `json:"completed_campaigns" example:"20"`
		TotalReach         int64 `json:"total_reach" example:"15600"`
	} `json:"campaign_overview"`

	// 营销渠道效果
	ChannelPerformance []struct {
		Channel      string  `json:"channel" example:"sms"`
		Campaigns    int64   `json:"campaigns" example:"12"`
		Reach        int64   `json:"reach" example:"5600"`
		DeliveryRate float64 `json:"delivery_rate" example:"96.5"`
		OpenRate     float64 `json:"open_rate" example:"32.1"`
		ClickRate    float64 `json:"click_rate" example:"8.9"`
	} `json:"channel_performance"`

	// 营销活动ROI排行
	CampaignROI []struct {
		ID      int64   `json:"id" example:"1"`
		Name    string  `json:"name" example:"春季促销活动"`
		Type    string  `json:"type" example:"sms"`
		Cost    float64 `json:"cost" example:"5000.00"`
		Revenue float64 `json:"revenue" example:"35000.00"`
		ROI     float64 `json:"roi" example:"600.0"`
		Reach   int32   `json:"reach" example:"1200"`
	} `json:"campaign_roi"`

	// 营销效果趋势
	MarketingTrend []struct {
		Month          string  `json:"month" example:"2024-01"`
		Campaigns      int64   `json:"campaigns" example:"5"`
		TotalReach     int64   `json:"total_reach" example:"3400"`
		Conversions    int64   `json:"conversions" example:"256"`
		ConversionRate float64 `json:"conversion_rate" example:"7.5"`
	} `json:"marketing_trend"`
}

// FinancialAnalyticsResponse 财务分析数据响应
type FinancialAnalyticsResponse struct {
	// 收入概览
	RevenueOverview struct {
		TotalRevenue   float64 `json:"total_revenue" example:"2450000.00"`
		MonthlyRevenue float64 `json:"monthly_revenue" example:"186500.00"`
		DailyRevenue   float64 `json:"daily_revenue" example:"8500.00"`
		GrowthRate     float64 `json:"growth_rate" example:"15.2"`
	} `json:"revenue_overview"`

	// 钱包统计
	WalletStats struct {
		TotalWallets     int64   `json:"total_wallets" example:"1180"`
		ActiveWallets    int64   `json:"active_wallets" example:"856"`
		TotalBalance     float64 `json:"total_balance" example:"125600.00"`
		TotalRecharge    float64 `json:"total_recharge" example:"356000.00"`
		TotalConsumption float64 `json:"total_consumption" example:"230400.00"`
	} `json:"wallet_stats"`

	// 收入来源分布
	RevenueSourceDistribution []struct {
		Source     string  `json:"source" example:"产品销售"`
		Revenue    float64 `json:"revenue" example:"1890000.00"`
		Percentage float64 `json:"percentage" example:"77.1"`
	} `json:"revenue_source_distribution"`

	// 财务趋势（最近12个月）
	FinancialTrend []struct {
		Month      string  `json:"month" example:"2024-01"`
		Revenue    float64 `json:"revenue" example:"156800.00"`
		Cost       float64 `json:"cost" example:"89600.00"`
		Profit     float64 `json:"profit" example:"67200.00"`
		ProfitRate float64 `json:"profit_rate" example:"42.9"`
	} `json:"financial_trend"`
}

// ActivitySummaryResponse 活动摘要数据响应
type ActivitySummaryResponse struct {
	// 待办活动统计
	PendingActivities struct {
		Total        int64 `json:"total" example:"23"`
		HighPriority int64 `json:"high_priority" example:"5"`
		Overdue      int64 `json:"overdue" example:"3"`
		DueToday     int64 `json:"due_today" example:"8"`
	} `json:"pending_activities"`

	// 活动类型分布
	ActivityTypeDistribution []struct {
		Type       string  `json:"type" example:"call"`
		Count      int64   `json:"count" example:"156"`
		Percentage float64 `json:"percentage" example:"35.2"`
	} `json:"activity_type_distribution"`

	// 最近活动列表（最近10条）
	RecentActivities []struct {
		ID             int64  `json:"id" example:"1"`
		Type           string `json:"type" example:"call"`
		Title          string `json:"title" example:"客户回访电话"`
		CustomerName   string `json:"customer_name" example:"张三"`
		Status         string `json:"status" example:"completed"`
		Priority       string `json:"priority" example:"high"`
		ScheduledAt    string `json:"scheduled_at" example:"2024-06-10T09:00:00Z"`
		AssignedToName string `json:"assigned_to_name" example:"李销售"`
	} `json:"recent_activities"`

	// 员工活动统计（Top 5）
	StaffActivityStats []struct {
		StaffID             int64   `json:"staff_id" example:"1"`
		StaffName           string  `json:"staff_name" example:"李销售"`
		CompletedActivities int64   `json:"completed_activities" example:"45"`
		PendingActivities   int64   `json:"pending_activities" example:"8"`
		CompletionRate      float64 `json:"completion_rate" example:"84.9"`
	} `json:"staff_activity_stats"`
}

// DashboardRequest 工作台数据请求参数
type DashboardRequest struct {
	DateRange string `form:"date_range" binding:"omitempty,date_range" example:"month"`
	TimeZone  string `form:"timezone" example:"Asia/Shanghai"`
}

// QuickStatsResponse 快速统计数据响应（用于实时刷新）
type QuickStatsResponse struct {
	OnlineUsers     int64   `json:"online_users" example:"23"`
	PendingOrders   int64   `json:"pending_orders" example:"12"`
	TodayRevenue    float64 `json:"today_revenue" example:"8500.00"`
	ActiveCampaigns int64   `json:"active_campaigns" example:"5"`
	SystemAlerts    int64   `json:"system_alerts" example:"2"`
	LastUpdated     string  `json:"last_updated" example:"2024-06-10T14:30:00Z"`
}

// TrendAnalysisRequest 趋势分析请求参数
type TrendAnalysisRequest struct {
	Metrics     []string `form:"metrics" binding:"required" example:"revenue,customers,orders"`
	StartDate   string   `form:"start_date" binding:"required" example:"2024-01-01"`
	EndDate     string   `form:"end_date" binding:"required" example:"2024-06-30"`
	Granularity string   `form:"granularity" binding:"omitempty,granularity" example:"month"`
}

// TrendAnalysisResponse 趋势分析响应
type TrendAnalysisResponse struct {
	DateRange struct {
		StartDate string `json:"start_date" example:"2024-01-01"`
		EndDate   string `json:"end_date" example:"2024-06-30"`
	} `json:"date_range"`

	Trends []struct {
		Date    string                 `json:"date" example:"2024-01"`
		Metrics map[string]interface{} `json:"metrics"` // 动态指标数据
	} `json:"trends"`

	Summary struct {
		TotalDataPoints int64                  `json:"total_data_points" example:"6"`
		Metrics         map[string]interface{} `json:"metrics"` // 汇总统计
	} `json:"summary"`
}
