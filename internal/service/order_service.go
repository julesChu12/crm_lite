package service

import (
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"crm_lite/pkg/utils"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

var (
	// ErrOrderNotFound 表示订单未找到。
	ErrOrderNotFound = errors.New("order not found")
	// ErrProductInfoChanged 表示产品信息已更改（例如，价格、库存），提示用户重新确认。
	ErrProductInfoChanged = errors.New("product information has changed, please re-confirm")
)

// OrderService 封装了与订单相关的业务逻辑。
type OrderService struct {
	db *gorm.DB
	q  *query.Query
}

// NewOrderService 创建一个新的 OrderService 实例。
func NewOrderService(rm *resource.Manager) *OrderService {
	dbRes, _ := resource.Get[*resource.DBResource](rm, resource.DBServiceKey)
	return &OrderService{
		db: dbRes.DB,
		q:  query.Use(dbRes.DB),
	}
}

// toOrderResponse 将 model.Order（及其关联的 OrderItems）转换为 dto.OrderResponse。
func (s *OrderService) toOrderResponse(o *model.Order) *dto.OrderResponse {
	items := make([]*dto.OrderItemResponse, len(o.OrderItems))
	for i, item := range o.OrderItems {
		items[i] = &dto.OrderItemResponse{
			ID:         item.ID,
			ProductID:  item.ProductID,
			Quantity:   int(item.Quantity),
			UnitPrice:  item.UnitPrice,
			FinalPrice: item.FinalPrice,
		}
	}
	return &dto.OrderResponse{
		ID:          o.ID,
		OrderNo:     o.OrderNo,
		CustomerID:  o.CustomerID,
		OrderDate:   o.OrderDate,
		Status:      o.Status,
		TotalAmount: o.TotalAmount,
		FinalAmount: o.FinalAmount,
		Remark:      o.Remark,
		CreatedAt:   o.CreatedAt,
		Items:       items,
	}
}

// CreateOrder 在单个事务中创建一个新订单及其订单项。
func (s *OrderService) CreateOrder(ctx context.Context, req *dto.OrderCreateRequest) (*dto.OrderResponse, error) {
	var createdOrder *model.Order
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 验证客户是否存在
		var customer model.Customer
		if err := tx.First(&customer, req.CustomerID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCustomerNotFound
			}
			return err
		}

		// 2. 批量验证商品是否存在
		var products []*model.Product
		productIDs := make([]int64, len(req.Items))
		for i, item := range req.Items {
			productIDs[i] = item.ProductID
		}
		if err := tx.Find(&products, productIDs).Error; err != nil || len(products) != len(req.Items) {
			return ErrProductNotFound
		}
		productMap := make(map[int64]*model.Product, len(products))
		for _, p := range products {
			productMap[p.ID] = p
		}

		// 3. 计算总价并构建订单项
		var totalAmount float64
		orderItems := make([]*model.OrderItem, len(req.Items))
		for i, item := range req.Items {
			product := productMap[item.ProductID]
			finalPrice := float64(item.Quantity) * item.UnitPrice
			totalAmount += finalPrice
			orderItems[i] = &model.OrderItem{
				ProductID:   item.ProductID,
				ProductName: product.Name,
				Quantity:    int32(item.Quantity),
				UnitPrice:   item.UnitPrice,
				FinalPrice:  finalPrice,
			}
		}

		// 4. 创建订单主体
		order := &model.Order{
			OrderNo:     utils.GenerateOrderNo(),
			CustomerID:  req.CustomerID,
			OrderDate:   req.OrderDate,
			Status:      "draft", // 默认状态为草稿
			TotalAmount: totalAmount,
			FinalAmount: totalAmount, // 假设没有折扣或运费等
			Remark:      req.Remark,
		}
		if req.Status != "" {
			order.Status = req.Status
		}

		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// 5. 创建订单项并关联到订单
		for _, item := range orderItems {
			item.OrderID = order.ID
		}
		if err := tx.Create(orderItems).Error; err != nil {
			return err
		}

		// 将新创建的订单项附加到模型中，以便 toOrderResponse 可以使用
		order.OrderItems = make([]model.OrderItem, len(orderItems))
		for i, item := range orderItems {
			order.OrderItems[i] = *item
		}
		createdOrder = order
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("在事务中创建订单失败: %w", err)
	}

	return s.toOrderResponse(createdOrder), nil
}

// GetOrderByID 根据 ID 获取单个订单，并预加载其订单项。
func (s *OrderService) GetOrderByID(ctx context.Context, idStr string) (*dto.OrderResponse, error) {
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if id == 0 {
		return nil, ErrOrderNotFound
	}
	var order model.Order
	// 使用原生 GORM 进行预加载
	err := s.db.WithContext(ctx).Preload("OrderItems").First(&order, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return s.toOrderResponse(&order), nil
}

// ListOrders 获取订单列表，支持分页、筛选、排序，并预加载订单项。
func (s *OrderService) ListOrders(ctx context.Context, req *dto.OrderListRequest) (*dto.OrderListResponse, error) {
	// 使用原生 GORM 以便利用 Preload
	db := s.db.WithContext(ctx).Model(&model.Order{})

	if len(req.IDs) > 0 {
		db = db.Where("id IN ?", req.IDs)
	} else {
		if req.CustomerID > 0 {
			db = db.Where("customer_id = ?", req.CustomerID)
		}
		if req.Status != "" {
			db = db.Where("status = ?", req.Status)
		}
	}

	if req.OrderBy != "" {
		parts := strings.Split(req.OrderBy, "_")
		if len(parts) == 2 {
			// 这里我们直接使用列名字符串，因为 GORM 支持这样做
			// 注意：为了安全，生产环境中应验证列名是否合法
			colName := parts[0]
			direction := parts[1]
			if direction == "desc" {
				db = db.Order(fmt.Sprintf("%s DESC", colName))
			} else {
				db = db.Order(colName)
			}
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	var orders []*model.Order
	if len(req.IDs) == 0 {
		db = db.Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize)
	}

	// 预加载关联的订单项
	if err := db.Preload("OrderItems").Find(&orders).Error; err != nil {
		return nil, err
	}

	items := make([]*dto.OrderResponse, len(orders))
	for i := range orders {
		items[i] = s.toOrderResponse(orders[i])
	}

	return &dto.OrderListResponse{
		Total:  total,
		Orders: items,
	}, nil
}
