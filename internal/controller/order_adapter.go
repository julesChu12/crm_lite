// Package controller 订单服务适配器
// 将新的Sales域服务适配到旧的OrderService接口
package controller

import (
	"context"
	"crm_lite/internal/domains/sales"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"strconv"
	"time"
)

// OrderServiceAdapter 适配器：将Sales域服务适配到OrderService接口
type OrderServiceAdapter struct {
	salesService  sales.Service
	legacyService *service.OrderService // 保留作为回退选项
}

// CreateOrder 实现旧接口：创建订单
// 将DTO转换为Sales域服务的请求格式
func (a *OrderServiceAdapter) CreateOrder(ctx context.Context, req *dto.OrderCreateRequest) (*dto.OrderResponse, error) {
	// 转换DTO为Sales域请求
	salesReq := sales.PlaceOrderReq{
		CustomerID: req.CustomerID,
		ContactID:  req.CustomerID, // 暂时使用客户ID作为联系人ID
		Items:      make([]sales.OrderItemReq, len(req.Items)),
		Discount:   0,        // 目前DTO中没有折扣字段，设为0
		PayMethod:  "wallet", // 默认钱包支付
		Remark:     req.Remark,
		AssignedTo: 1, // 默认分配给管理员1
		IdemKey:    "order_" + strconv.FormatInt(time.Now().UnixNano(), 10),
	}

	// 转换订单项
	for i, item := range req.Items {
		salesReq.Items[i] = sales.OrderItemReq{
			ProductID: item.ProductID,
			Qty:       int32(item.Quantity),
		}
	}

	// 调用Sales域服务
	order, err := a.salesService.PlaceOrder(ctx, salesReq)
	if err != nil {
		// 如果新服务失败，回退到旧服务
		return a.legacyService.CreateOrder(ctx, req)
	}

	// 转换结果为DTO
	response := &dto.OrderResponse{
		ID:          order.ID,
		OrderNo:     order.OrderNo,
		CustomerID:  order.CustomerID,
		OrderDate:   time.Unix(order.CreatedAt, 0),
		Status:      order.Status,
		TotalAmount: float64(order.TotalAmount) / 100,
		FinalAmount: float64(order.FinalAmount) / 100,
		Remark:      "", // 从订单获取
		CreatedAt:   time.Unix(order.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		Items:       []*dto.OrderItemResponse{}, // 暂时为空，如需可补充
	}

	return response, nil
}

// GetOrderByID 实现旧接口：获取订单详情
func (a *OrderServiceAdapter) GetOrderByID(ctx context.Context, idStr string) (*dto.OrderResponse, error) {
	orderID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, err
	}

	// 调用Sales域服务
	order, orderItems, err := a.salesService.GetOrder(ctx, orderID)
	if err != nil {
		// 如果新服务失败，回退到旧服务
		return a.legacyService.GetOrderByID(ctx, idStr)
	}

	if order == nil {
		return nil, service.ErrOrderNotFound
	}

	// 转换订单项
	items := make([]*dto.OrderItemResponse, len(orderItems))
	for i, item := range orderItems {
		items[i] = &dto.OrderItemResponse{
			ID:         item.ID,
			ProductID:  item.ProductID,
			Quantity:   int(item.Quantity),
			UnitPrice:  float64(item.UnitPriceSnapshot) / 100,
			FinalPrice: float64(item.FinalPrice) / 100,
		}
	}

	// 转换为DTO响应
	response := &dto.OrderResponse{
		ID:          order.ID,
		OrderNo:     order.OrderNo,
		CustomerID:  order.CustomerID,
		OrderDate:   time.Unix(order.CreatedAt, 0),
		Status:      order.Status,
		TotalAmount: float64(order.TotalAmount) / 100,
		FinalAmount: float64(order.FinalAmount) / 100,
		CreatedAt:   time.Unix(order.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		Items:       items,
	}

	return response, nil
}

// ListOrders 实现旧接口：订单列表
func (a *OrderServiceAdapter) ListOrders(ctx context.Context, req *dto.OrderListRequest) (*dto.OrderListResponse, error) {
	// 调用Sales域服务，使用默认分页参数
	orders, err := a.salesService.ListOrders(ctx, req.CustomerID, req.Status, 1, 20)
	if err != nil {
		// 如果新服务失败，回退到旧服务
		return a.legacyService.ListOrders(ctx, req)
	}

	// 转换为DTO响应
	orderResponses := make([]*dto.OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = &dto.OrderResponse{
			ID:          order.ID,
			OrderNo:     order.OrderNo,
			CustomerID:  order.CustomerID,
			OrderDate:   time.Unix(order.CreatedAt, 0),
			Status:      order.Status,
			TotalAmount: float64(order.TotalAmount) / 100,
			FinalAmount: float64(order.FinalAmount) / 100,
			CreatedAt:   time.Unix(order.CreatedAt, 0).Format("2006-01-02 15:04:05"),
			Items:       []*dto.OrderItemResponse{}, // 列表不包含详细项目
		}
	}

	return &dto.OrderListResponse{
		Orders: orderResponses,
		Total:  int64(len(orders)),
	}, nil
}
