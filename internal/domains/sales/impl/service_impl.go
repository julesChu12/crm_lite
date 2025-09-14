package impl

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"crm_lite/internal/common"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/domains/billing"
	"crm_lite/internal/domains/catalog"
	"crm_lite/internal/domains/sales"
	"crm_lite/pkg/utils"

	"gorm.io/gorm"
)

// SalesServiceImpl Sales 域完整实现
// 统一订单事务收口，集成产品快照、钱包扣减、outbox 事件
type SalesServiceImpl struct {
	db         *gorm.DB
	q          *query.Query
	tx         common.Tx
	catalogSvc catalog.Service
	billingSvc billing.Service
	outboxSvc  common.OutboxService
}

// NewSalesServiceImpl 创建 Sales 服务完整实现
func NewSalesServiceImpl(db *gorm.DB, tx common.Tx, catalogSvc catalog.Service, billingSvc billing.Service, outboxSvc common.OutboxService) *SalesServiceImpl {
	return &SalesServiceImpl{
		db:         db,
		q:          query.Use(db),
		tx:         tx,
		catalogSvc: catalogSvc,
		billingSvc: billingSvc,
		outboxSvc:  outboxSvc,
	}
}

// PlaceOrder 统一下单事务收口
// 在单一事务中完成：产品快照 + 订单创建 + 钱包扣减 + outbox 事件
func (s *SalesServiceImpl) PlaceOrder(ctx context.Context, req sales.PlaceOrderReq) (sales.Order, error) {
	var result sales.Order

	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		txDB := s.tx.GetDB(ctx)
		txQuery := query.Use(txDB)

		// 1. 校验客户是否存在
		_, err := txQuery.Customer.WithContext(ctx).Where(txQuery.Customer.ID.Eq(req.CustomerID)).First()
		if err != nil {
			return common.NewBusinessError(common.ErrCodeCustomerNotFound, "客户不存在")
		}

		// 2. 获取产品信息并校验可售性
		productIDs := make([]int64, len(req.Items))
		for i, item := range req.Items {
			productIDs[i] = item.ProductID
		}

		products, err := s.catalogSvc.BatchGet(ctx, productIDs)
		if err != nil {
			return fmt.Errorf("获取产品信息失败: %w", err)
		}
		if len(products) != len(req.Items) {
			return common.NewBusinessError(common.ErrCodeProductNotFound, "部分产品不存在")
		}

		// 构建产品映射
		productMap := make(map[int64]catalog.Product)
		for _, product := range products {
			productMap[product.ID] = product
			// 校验产品可售性
			if err := s.catalogSvc.EnsureSellable(ctx, product.ID); err != nil {
				return fmt.Errorf("产品 %s 不可售: %w", product.Name, err)
			}
		}

		// 3. 计算订单总金额并创建订单项快照
		var totalAmount int64 = 0
		orderItems := make([]*model.OrderItem, len(req.Items))

		for i, item := range req.Items {
			product := productMap[item.ProductID]
			itemAmount := product.Price * int64(item.Qty)
			totalAmount += itemAmount

			// 创建订单项（包含产品快照）
			orderItems[i] = &model.OrderItem{
				ProductID:           item.ProductID,
				ProductName:         product.Name, // 保持兼容性
				ProductNameSnapshot: product.Name, // 新增快照字段
				Quantity:            item.Qty,
				UnitPrice:           float64(product.Price) / 100, // 转换为元（兼容旧字段）
				UnitPriceSnapshot:   product.Price,                // 快照以分为单位
				DurationMinSnapshot: product.DurationMin,          // 服务时长快照
				FinalPrice:          float64(itemAmount) / 100,    // 转换为元（兼容旧字段）
			}
		}

		// 4. 应用折扣
		finalAmount := totalAmount - req.Discount

		// 5. 创建订单
		order := &model.Order{
			OrderNo:     utils.GenerateOrderNo(),
			CustomerID:  req.CustomerID,
			OrderDate:   time.Now(),
			Status:      "pending",                  // 待支付状态
			TotalAmount: float64(totalAmount) / 100, // 转换为元（兼容旧字段）
			FinalAmount: float64(finalAmount) / 100, // 转换为元（兼容旧字段）
			Remark:      req.Remark,
		}

		// 创建订单记录
		if err := txQuery.Order.WithContext(ctx).Create(order); err != nil {
			return fmt.Errorf("创建订单失败: %w", err)
		}

		// 关联订单项到订单
		for _, item := range orderItems {
			item.OrderID = order.ID
		}

		// 批量创建订单项
		if err := txQuery.OrderItem.WithContext(ctx).Create(orderItems...); err != nil {
			return fmt.Errorf("创建订单项失败: %w", err)
		}

		// 6. 如果是钱包支付，执行扣款
		if req.PayMethod == "wallet" && finalAmount > 0 {
			idemKey := fmt.Sprintf("order_pay_%d_%s", order.ID, req.IdemKey)
			err := s.billingSvc.DebitForOrder(ctx, req.CustomerID, order.ID, finalAmount, idemKey)
			if err != nil {
				return fmt.Errorf("钱包扣款失败: %w", err)
			}

			// 更新订单支付状态
			_, err = txQuery.Order.WithContext(ctx).Where(txQuery.Order.ID.Eq(order.ID)).Update(txQuery.Order.Status, "paid")
			if err != nil {
				return fmt.Errorf("更新订单状态失败: %w", err)
			}
			order.Status = "paid"
		}

		// 7. 写入 Outbox 事件
		orderEvent := common.OrderPlacedEvent{
			OrderID:     order.ID,
			OrderNo:     order.OrderNo,
			CustomerID:  req.CustomerID,
			TotalAmount: finalAmount,
			PayMethod:   req.PayMethod,
			CreatedAt:   time.Now().Unix(),
		}

		if err := s.outboxSvc.PublishEvent(ctx, common.EventTypeOrderPlaced, orderEvent); err != nil {
			// 日志记录但不阻断业务流程
			// TODO: 添加日志
		}

		// 如果是钱包支付且已扣款成功，发送支付事件
		if req.PayMethod == "wallet" && finalAmount > 0 {
			paidEvent := common.OrderPaidEvent{
				OrderID:    order.ID,
				OrderNo:    order.OrderNo,
				CustomerID: req.CustomerID,
				PaidAmount: finalAmount,
				PayMethod:  req.PayMethod,
				PaidAt:     time.Now().Unix(),
			}
			_ = s.outboxSvc.PublishEvent(ctx, common.EventTypeOrderPaid, paidEvent)
		}

		// 8. 构建返回结果
		result = sales.Order{
			ID:             order.ID,
			OrderNo:        order.OrderNo,
			CustomerID:     order.CustomerID,
			TotalAmount:    totalAmount,
			DiscountAmount: req.Discount,
			FinalAmount:    finalAmount,
			Status:         order.Status,
			PayMethod:      req.PayMethod,
			CreatedAt:      time.Now().Unix(),
		}

		return nil
	})

	if err != nil {
		return sales.Order{}, err
	}

	return result, nil
}

// RefundOrder 统一退款事务收口
// 在单一事务中完成：订单状态更新 + 钱包退款 + outbox 事件
func (s *SalesServiceImpl) RefundOrder(ctx context.Context, orderID int64, reason string) error {
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		txDB := s.tx.GetDB(ctx)
		txQuery := query.Use(txDB)

		// 1. 获取订单信息
		order, err := txQuery.Order.WithContext(ctx).Where(txQuery.Order.ID.Eq(orderID)).First()
		if err != nil {
			return common.NewBusinessError(common.ErrCodeOrderNotFound, "订单不存在")
		}

		// 2. 校验订单状态
		if order.Status == "refunded" {
			return common.NewBusinessError(common.ErrCodeOrderStatusInvalid, "订单已退款")
		}
		if order.Status != "paid" && order.Status != "completed" {
			return common.NewBusinessError(common.ErrCodeOrderCannotRefund, "订单状态不支持退款")
		}

		// 3. 更新订单状态为退款
		_, err = txQuery.Order.WithContext(ctx).Where(txQuery.Order.ID.Eq(orderID)).Update(txQuery.Order.Status, "refunded")
		if err != nil {
			return fmt.Errorf("更新订单状态失败: %w", err)
		}

		// 4. 如果原订单是钱包支付，执行退款
		finalAmount := int64(order.FinalAmount * 100) // 转换为分
		if finalAmount > 0 {
			// 这里假设原支付方式信息存储在订单中，实际可能需要从其他地方获取
			// 为了演示，我们假设支付金额需要退回钱包
			idemKey := fmt.Sprintf("order_refund_%d_%d", orderID, time.Now().UnixNano())
			err := s.billingSvc.CreditForRefund(ctx, order.CustomerID, orderID, finalAmount, idemKey)
			if err != nil {
				return fmt.Errorf("钱包退款失败: %w", err)
			}
		}

		// 5. 写入 Outbox 事件
		refundEvent := common.OrderRefundedEvent{
			OrderID:      orderID,
			OrderNo:      order.OrderNo,
			CustomerID:   order.CustomerID,
			RefundAmount: finalAmount,
			Reason:       reason,
			RefundedAt:   time.Now().Unix(),
		}

		if err := s.outboxSvc.PublishEvent(ctx, common.EventTypeOrderRefunded, refundEvent); err != nil {
			// 日志记录但不阻断业务流程
			// TODO: 添加日志
		}

		return nil
	})
}

// GetOrder 获取订单详情
func (s *SalesServiceImpl) GetOrder(ctx context.Context, orderID int64) (*sales.Order, []sales.OrderItem, error) {
	// 获取订单基本信息
	order, err := s.q.Order.WithContext(ctx).Where(s.q.Order.ID.Eq(orderID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, common.NewBusinessError(common.ErrCodeOrderNotFound, "订单不存在")
		}
		return nil, nil, err
	}

	// 获取订单项
	orderItems, err := s.q.OrderItem.WithContext(ctx).Where(s.q.OrderItem.OrderID.Eq(orderID)).Find()
	if err != nil {
		return nil, nil, fmt.Errorf("获取订单项失败: %w", err)
	}

	// 转换为域模型
	salesOrder := &sales.Order{
		ID:          order.ID,
		OrderNo:     order.OrderNo,
		CustomerID:  order.CustomerID,
		TotalAmount: int64(order.TotalAmount * 100), // 转换为分
		FinalAmount: int64(order.FinalAmount * 100), // 转换为分
		Status:      order.Status,
		CreatedAt:   order.CreatedAt.Unix(),
	}

	salesItems := make([]sales.OrderItem, len(orderItems))
	for i, item := range orderItems {
		salesItems[i] = sales.OrderItem{
			ID:                  item.ID,
			OrderID:             item.OrderID,
			ProductID:           item.ProductID,
			ProductNameSnapshot: item.ProductNameSnapshot,
			UnitPriceSnapshot:   item.UnitPriceSnapshot,
			DurationMinSnapshot: item.DurationMinSnapshot,
			Quantity:            item.Quantity,
			FinalPrice:          int64(item.FinalPrice * 100), // 转换为分
		}
	}

	return salesOrder, salesItems, nil
}

// ListOrders 分页查询订单
func (s *SalesServiceImpl) ListOrders(ctx context.Context, customerID int64, status string, page, pageSize int) ([]sales.Order, error) {
	q := s.q.Order.WithContext(ctx)

	if customerID > 0 {
		q = q.Where(s.q.Order.CustomerID.Eq(customerID))
	}
	if status != "" {
		q = q.Where(s.q.Order.Status.Eq(status))
	}

	offset := (page - 1) * pageSize
	orders, err := q.Order(s.q.Order.CreatedAt.Desc()).Offset(offset).Limit(pageSize).Find()
	if err != nil {
		return nil, fmt.Errorf("查询订单列表失败: %w", err)
	}

	result := make([]sales.Order, len(orders))
	for i, order := range orders {
		result[i] = sales.Order{
			ID:          order.ID,
			OrderNo:     order.OrderNo,
			CustomerID:  order.CustomerID,
			TotalAmount: int64(order.TotalAmount * 100), // 转换为分
			FinalAmount: int64(order.FinalAmount * 100), // 转换为分
			Status:      order.Status,
			CreatedAt:   order.CreatedAt.Unix(),
		}
	}

	return result, nil
}

// CreateOrder 创建订单（控制器接口）
func (s *SalesServiceImpl) CreateOrder(ctx context.Context, req *sales.CreateOrderRequest) (*sales.OrderResponse, error) {
	// 转换 DTO 为领域模型
	placeOrderReq := sales.PlaceOrderReq{
		CustomerID: req.CustomerID,
		ContactID:  req.CustomerID, // 暂时使用客户ID作为联系人ID
		Channel:    "web",          // 默认渠道
		PayMethod:  "wallet",       // 默认钱包支付
		Items:      make([]sales.OrderItemReq, len(req.Items)),
		Discount:   0, // 目前DTO中没有折扣字段，设为0
		Remark:     req.Remark,
		AssignedTo: 1, // 默认分配给管理员1
		IdemKey:    fmt.Sprintf("order_%d_%d", req.CustomerID, time.Now().UnixNano()),
	}

	// 转换订单项
	for i, item := range req.Items {
		placeOrderReq.Items[i] = sales.OrderItemReq{
			ProductID: item.ProductID,
			Qty:       int32(item.Quantity),
		}
	}

	// 调用领域方法
	order, err := s.PlaceOrder(ctx, placeOrderReq)
	if err != nil {
		return nil, err
	}

	// 转换结果为响应格式
	response := &sales.OrderResponse{
		ID:          order.ID,
		OrderNo:     order.OrderNo,
		CustomerID:  order.CustomerID,
		TotalAmount: float64(order.TotalAmount) / 100.0,
		Status:      order.Status,
		Items:       []sales.OrderItemResponse{}, // TODO: 填充订单项
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.CreatedAt, // 暂时使用创建时间
	}

	return response, nil
}

// GetOrderByID 根据ID获取订单（控制器接口）
func (s *SalesServiceImpl) GetOrderByID(ctx context.Context, idStr string) (*sales.OrderResponse, error) {
	// 解析订单ID
	orderID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %w", err)
	}

	// 调用领域方法
	order, items, err := s.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// 转换订单项
	itemResponses := make([]sales.OrderItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = sales.OrderItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  int(item.Quantity),
			Price:     float64(item.UnitPriceSnapshot) / 100.0,
			Amount:    float64(item.FinalPrice) / 100.0,
		}
	}

	// 转换结果为响应格式
	response := &sales.OrderResponse{
		ID:          order.ID,
		OrderNo:     order.OrderNo,
		CustomerID:  order.CustomerID,
		TotalAmount: float64(order.TotalAmount) / 100.0,
		Status:      order.Status,
		Items:       itemResponses,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.CreatedAt, // 暂时使用创建时间
	}

	return response, nil
}

// ListOrdersForController 获取订单列表（控制器接口）
func (s *SalesServiceImpl) ListOrdersForController(ctx context.Context, req *sales.ListOrdersRequest) (*sales.ListOrdersResponse, error) {
	// 调用领域方法
	orders, err := s.ListOrders(ctx, req.CustomerID, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	// 转换订单列表
	orderResponses := make([]sales.OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = sales.OrderResponse{
			ID:          order.ID,
			OrderNo:     order.OrderNo,
			CustomerID:  order.CustomerID,
			TotalAmount: float64(order.TotalAmount) / 100.0,
			Status:      order.Status,
			Items:       []sales.OrderItemResponse{}, // TODO: 填充订单项
			CreatedAt:   order.CreatedAt,
			UpdatedAt:   order.CreatedAt, // 暂时使用创建时间
		}
	}

	// 返回响应
	response := &sales.ListOrdersResponse{
		Orders:   orderResponses,
		Total:    int64(len(orderResponses)), // TODO: 获取真实总数
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	return response, nil
}

// 占位实现，保持向后兼容
type ServiceImpl struct {
	fullImpl *SalesServiceImpl
}

func New() *ServiceImpl {
	return &ServiceImpl{}
}

func (s *ServiceImpl) PlaceOrder(ctx context.Context, req sales.PlaceOrderReq) (sales.Order, error) {
	if s.fullImpl != nil {
		return s.fullImpl.PlaceOrder(ctx, req)
	}
	return sales.Order{}, fmt.Errorf("销售域服务未完全初始化，请使用 NewSalesServiceImpl")
}

func (s *ServiceImpl) RefundOrder(ctx context.Context, orderID int64, reason string) error {
	if s.fullImpl != nil {
		return s.fullImpl.RefundOrder(ctx, orderID, reason)
	}
	return fmt.Errorf("销售域服务未完全初始化，请使用 NewSalesServiceImpl")
}

func (s *ServiceImpl) GetOrder(ctx context.Context, orderID int64) (*sales.Order, []sales.OrderItem, error) {
	if s.fullImpl != nil {
		return s.fullImpl.GetOrder(ctx, orderID)
	}
	return nil, nil, fmt.Errorf("销售域服务未完全初始化，请使用 NewSalesServiceImpl")
}

func (s *ServiceImpl) ListOrders(ctx context.Context, customerID int64, status string, page, pageSize int) ([]sales.Order, error) {
	if s.fullImpl != nil {
		return s.fullImpl.ListOrders(ctx, customerID, status, page, pageSize)
	}
	return nil, fmt.Errorf("销售域服务未完全初始化，请使用 NewSalesServiceImpl")
}

// CreateOrder 创建订单（控制器接口）
func (s *ServiceImpl) CreateOrder(ctx context.Context, req *sales.CreateOrderRequest) (*sales.OrderResponse, error) {
	if s.fullImpl != nil {
		return s.fullImpl.CreateOrder(ctx, req)
	}
	return nil, fmt.Errorf("销售域服务未完全初始化，请使用 NewSalesServiceImpl")
}

// GetOrderByID 根据ID获取订单（控制器接口）
func (s *ServiceImpl) GetOrderByID(ctx context.Context, idStr string) (*sales.OrderResponse, error) {
	if s.fullImpl != nil {
		return s.fullImpl.GetOrderByID(ctx, idStr)
	}
	return nil, fmt.Errorf("销售域服务未完全初始化，请使用 NewSalesServiceImpl")
}

// ListOrdersForController 获取订单列表（控制器接口）
func (s *ServiceImpl) ListOrdersForController(ctx context.Context, req *sales.ListOrdersRequest) (*sales.ListOrdersResponse, error) {
	if s.fullImpl != nil {
		return s.fullImpl.ListOrdersForController(ctx, req)
	}
	return nil, fmt.Errorf("销售域服务未完全初始化，请使用 NewSalesServiceImpl")
}

var _ sales.Service = (*ServiceImpl)(nil)
var _ sales.Service = (*SalesServiceImpl)(nil)
