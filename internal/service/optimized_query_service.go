package service

import (
	"context"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// OptimizedQueryService 优化的查询服务
// 实现了各种查询性能优化技巧
type OptimizedQueryService struct {
	db *gorm.DB
	q  *query.Query
}

// NewOptimizedQueryService 创建优化查询服务
func NewOptimizedQueryService(db *gorm.DB) *OptimizedQueryService {
	return &OptimizedQueryService{
		db: db,
		q:  query.Use(db),
	}
}

// OrderWithItems 订单及其订单项的聚合结构
type OrderWithItems struct {
	Order *model.Order       `json:"order"`
	Items []*model.OrderItem `json:"items"`
}

// GetOrdersWithItems 获取订单及其订单项（解决N+1问题）
func (s *OptimizedQueryService) GetOrdersWithItems(ctx context.Context, customerID int64, limit int) ([]*OrderWithItems, error) {
	// 使用单个查询获取所有数据，避免N+1问题
	type OrderItemJoin struct {
		// Order fields
		OrderID             int64   `json:"order_id"`
		OrderNo             string  `json:"order_no"`
		CustomerID          int64   `json:"customer_id"`
		OrderStatus         string  `json:"order_status"`
		TotalAmount         float64 `json:"total_amount"`
		FinalAmount         float64 `json:"final_amount"`
		OrderCreatedAt      string  `json:"order_created_at"`

		// OrderItem fields (nullable)
		ItemID              *int64  `json:"item_id"`
		ProductID           *int64  `json:"product_id"`
		ProductName         *string `json:"product_name"`
		Quantity            *int32  `json:"quantity"`
		UnitPrice           *float64 `json:"unit_price"`
		FinalPrice          *float64 `json:"final_price"`
		ItemCreatedAt       *string  `json:"item_created_at"`
	}

	var joinResults []OrderItemJoin
	err := s.db.WithContext(ctx).
		Table("orders o").
		Select(`
			o.id as order_id,
			o.order_no,
			o.customer_id,
			o.status as order_status,
			o.total_amount,
			o.final_amount,
			o.created_at as order_created_at,
			oi.id as item_id,
			oi.product_id,
			oi.product_name,
			oi.quantity,
			oi.unit_price,
			oi.final_price,
			oi.created_at as item_created_at
		`).
		Joins("LEFT JOIN order_items oi ON o.id = oi.order_id").
		Where("o.customer_id = ?", customerID).
		Order("o.created_at DESC").
		Limit(limit).
		Scan(&joinResults).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get orders with items: %w", err)
	}

	// 组织数据结构
	orderMap := make(map[int64]*OrderWithItems)
	for _, result := range joinResults {
		if _, exists := orderMap[result.OrderID]; !exists {
			orderMap[result.OrderID] = &OrderWithItems{
				Order: &model.Order{
					ID:          result.OrderID,
					OrderNo:     result.OrderNo,
					CustomerID:  result.CustomerID,
					Status:      result.OrderStatus,
					TotalAmount: result.TotalAmount,
					FinalAmount: result.FinalAmount,
					// 其他字段根据需要添加
				},
				Items: []*model.OrderItem{},
			}
		}

		// 如果有订单项数据，添加到列表
		if result.ItemID != nil {
			orderMap[result.OrderID].Items = append(orderMap[result.OrderID].Items, &model.OrderItem{
				ID:          *result.ItemID,
				OrderID:     result.OrderID,
				ProductID:   *result.ProductID,
				ProductName: *result.ProductName,
				Quantity:    *result.Quantity,
				UnitPrice:   *result.UnitPrice,
				FinalPrice:  *result.FinalPrice,
				// 其他字段根据需要添加
			})
		}
	}

	// 转换为slice
	results := make([]*OrderWithItems, 0, len(orderMap))
	for _, orderWithItems := range orderMap {
		results = append(results, orderWithItems)
	}

	return results, nil
}

// BatchCreateOrderItems 批量创建订单项（优化批量插入）
func (s *OptimizedQueryService) BatchCreateOrderItems(ctx context.Context, items []*model.OrderItem) error {
	if len(items) == 0 {
		return nil
	}

	// 使用批量插入，每批100条记录
	batchSize := 100
	return s.q.OrderItem.WithContext(ctx).CreateInBatches(items, batchSize)
}

// GetCustomersByCursorPagination 基于游标的分页查询（优化大offset问题）
func (s *OptimizedQueryService) GetCustomersByCursorPagination(ctx context.Context, lastID int64, pageSize int) ([]*model.Customer, error) {
	query := s.q.Customer.WithContext(ctx)

	if lastID > 0 {
		query = query.Where(s.q.Customer.ID.Gt(lastID))
	}

	return query.Order(s.q.Customer.ID).Limit(pageSize).Find()
}

// SearchCustomersOptimized 优化的客户搜索（避免全表扫描）
func (s *OptimizedQueryService) SearchCustomersOptimized(ctx context.Context, keyword string) ([]*model.Customer, error) {
	if keyword == "" {
		return []*model.Customer{}, nil
	}

	query := s.q.Customer.WithContext(ctx)

	// 如果是手机号格式，使用手机号索引
	if len(keyword) >= 11 && isPhoneNumber(keyword) {
		return query.Where(s.q.Customer.Phone.Like(keyword + "%")).Limit(50).Find()
	}

	// 否则搜索姓名，限制结果数量避免性能问题
	return query.Where(s.q.Customer.Name.Like("%" + keyword + "%")).Limit(100).Find()
}

// GetWalletTransactionsSummary 获取钱包交易汇总（使用聚合查询）
func (s *OptimizedQueryService) GetWalletTransactionsSummary(ctx context.Context, walletID int64, days int) (*WalletTransactionSummary, error) {
	type SummaryResult struct {
		Direction    string  `json:"direction"`
		TotalAmount  int64   `json:"total_amount"`
		Count        int64   `json:"count"`
	}

	var summaries []SummaryResult
	err := s.db.WithContext(ctx).
		Table("wallet_transactions").
		Select("direction, SUM(amount) as total_amount, COUNT(*) as count").
		Where("wallet_id = ? AND created_at >= ?", walletID, getTimestamp(days)).
		Group("direction").
		Scan(&summaries).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get wallet transaction summary: %w", err)
	}

	summary := &WalletTransactionSummary{
		WalletID: walletID,
		Days:     days,
	}

	for _, s := range summaries {
		if s.Direction == "credit" {
			summary.TotalCreditAmount = s.TotalAmount
			summary.CreditCount = s.Count
		} else if s.Direction == "debit" {
			summary.TotalDebitAmount = s.TotalAmount
			summary.DebitCount = s.Count
		}
	}

	return summary, nil
}

// WalletTransactionSummary 钱包交易汇总
type WalletTransactionSummary struct {
	WalletID          int64 `json:"wallet_id"`
	Days              int   `json:"days"`
	TotalCreditAmount int64 `json:"total_credit_amount"`
	TotalDebitAmount  int64 `json:"total_debit_amount"`
	CreditCount       int64 `json:"credit_count"`
	DebitCount        int64 `json:"debit_count"`
}

// isPhoneNumber 检查是否为手机号格式
func isPhoneNumber(s string) bool {
	if len(s) != 11 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s[0] == '1' // 简单的中国手机号检查
}

// getTimestamp 获取N天前的时间戳
func getTimestamp(days int) int64 {
	return time.Now().AddDate(0, 0, -days).Unix()
}