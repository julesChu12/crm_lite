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

type DashboardService struct {
    q  *query.Query
    db *gorm.DB
}

func NewDashboardService(resManager *resource.Manager) *DashboardService {
    dbRes, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to init DashboardService: " + err.Error())
	}
    return &DashboardService{q: query.Use(dbRes.DB), db: dbRes.DB}
}

// GetOverview 计算基础总览指标（MVP 版：不依赖订单与产品）
func (s *DashboardService) GetOverview(ctx context.Context, _ *dto.DashboardRequest) (*dto.DashboardOverviewResponse, error) {
	// 总客户数
	totalCustomers, _ := s.q.Customer.WithContext(ctx).Where(s.q.Customer.DeletedAt.IsNull()).Count()

	// 本月新增客户
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthlyNewCustomers, _ := s.q.Customer.WithContext(ctx).
		Where(s.q.Customer.DeletedAt.IsNull(), s.q.Customer.CreatedAt.Gte(monthStart)).
		Count()

	// 今日新增客户
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayNewCustomers, _ := s.q.Customer.WithContext(ctx).
		Where(s.q.Customer.DeletedAt.IsNull(), s.q.Customer.CreatedAt.Gte(todayStart)).
		Count()

    // 钱包指标（以 wallet 与 wallet_transactions 汇总）
    totalWallets, _ := s.q.Wallet.WithContext(ctx).Count()
    // 余额汇总
    var totalBalance float64
    s.db.WithContext(ctx).Model(&model.Wallet{}).Select("COALESCE(SUM(balance),0)").Scan(&totalBalance)
    // 汇总充值与消费（总口径，来自 wallets 汇总字段，避免复杂查询）
    var totalRecharge, totalConsume float64
    s.db.WithContext(ctx).Model(&model.Wallet{}).Select("COALESCE(SUM(total_recharged),0)").Scan(&totalRecharge)
    s.db.WithContext(ctx).Model(&model.Wallet{}).Select("COALESCE(SUM(total_consumed),0)").Scan(&totalConsume)
    // 本月消费（以交易表 type=consume，消费金额为负数，使用 -amount 求和）
    var monthConsume float64
    s.db.WithContext(ctx).Model(&model.WalletTransaction{}).
        Where("type = ? AND created_at >= ?", "consume", monthStart).
        Select("COALESCE(SUM(-amount),0)").Scan(&monthConsume)

    // 今日收入（以消费为收入）
    var todayRevenue float64
    s.db.WithContext(ctx).Model(&model.WalletTransaction{}).
        Where("type = ? AND created_at >= ?", "consume", todayStart).
        Select("COALESCE(SUM(-amount),0)").Scan(&todayRevenue)

    // 环比/同比（以月为周期，来源：消费）
    prevMonthStart := monthStart.AddDate(0, -1, 0)
    prevMonthEnd := monthStart.Add(-time.Nanosecond)
    var prevMonthConsume float64
    s.db.WithContext(ctx).Model(&model.WalletTransaction{}).
        Where("type = ? AND created_at >= ? AND created_at <= ?", "consume", prevMonthStart, prevMonthEnd).
        Select("COALESCE(SUM(-amount),0)").Scan(&prevMonthConsume)

    lastYearMonthStart := monthStart.AddDate(-1, 0, 0)
    lastYearMonthEnd := lastYearMonthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
    var lastYearMonthConsume float64
    s.db.WithContext(ctx).Model(&model.WalletTransaction{}).
        Where("type = ? AND created_at >= ? AND created_at <= ?", "consume", lastYearMonthStart, lastYearMonthEnd).
        Select("COALESCE(SUM(-amount),0)").Scan(&lastYearMonthConsume)

    // 计算同比/环比
    calcRate := func(current, base float64) float64 {
        if base <= 0 {
            if current > 0 {
                return 100
            }
            return 0
        }
        return (current - base) / base * 100
    }

    // 填充响应（订单/产品相关指标置 0）
	resp := &dto.DashboardOverviewResponse{
		TotalCustomers:           totalCustomers,
		TotalOrders:              0,
        TotalRevenue:             monthConsume, // 本月“收入”口径取消费额
		TotalProducts:            0,
		MonthlyNewCustomers:      monthlyNewCustomers,
        MonthlyOrders:            0,
        MonthlyRevenue:           monthConsume,
		CustomerGrowthRate:       0,
		OrderGrowthRate:          0,
        RevenueGrowthRate:        calcRate(monthConsume, prevMonthConsume),
		TodayNewCustomers:        todayNewCustomers,
		TodayOrders:              0,
        TodayRevenue:             todayRevenue,
		ActiveMarketingCampaigns: 0,
		PendingActivities:        0,
		LowStockProducts:         0,
        TotalWallets:             totalWallets,
        TotalBalance:             totalBalance,
        TotalRecharge:            totalRecharge,
        TotalConsumption:         totalConsume,
        RevenueMoMRate:           calcRate(monthConsume, prevMonthConsume),
        RevenueYoYRate:           calcRate(monthConsume, lastYearMonthConsume),
	}
	return resp, nil
}
