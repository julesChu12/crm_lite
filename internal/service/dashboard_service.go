package service

import (
    "context"
    "time"

    "crm_lite/internal/core/resource"
    "crm_lite/internal/dao/query"
    "crm_lite/internal/dto"
)

type DashboardService struct {
    q *query.Query
}

func NewDashboardService(resManager *resource.Manager) *DashboardService {
    dbRes, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
    if err != nil {
        panic("Failed to init DashboardService: " + err.Error())
    }
    return &DashboardService{q: query.Use(dbRes.DB)}
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

    // 填充响应（与订单/产品相关的指标置 0，后续如接入再补）
    resp := &dto.DashboardOverviewResponse{
        TotalCustomers:        totalCustomers,
        TotalOrders:           0,
        TotalRevenue:          0,
        TotalProducts:         0,
        MonthlyNewCustomers:   monthlyNewCustomers,
        MonthlyOrders:         0,
        MonthlyRevenue:        0,
        CustomerGrowthRate:    0,
        OrderGrowthRate:       0,
        RevenueGrowthRate:     0,
        TodayNewCustomers:     todayNewCustomers,
        TodayOrders:           0,
        TodayRevenue:          0,
        ActiveMarketingCampaigns: 0,
        PendingActivities:        0,
        LowStockProducts:         0,
    }
    return resp, nil
}


