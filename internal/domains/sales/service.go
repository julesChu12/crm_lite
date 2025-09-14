// Package sales 订单域服务接口
// 职责：订单生命周期管理、产品快照、事务收口
// 核心原则：统一下单和退款的事务边界，确保订单状态与钱包操作的一致性
package sales

import "context"

// OrderItemReq 下单商品项请求
// 用于接收前端传递的下单商品信息
type OrderItemReq struct {
	ProductID int64 `json:"product_id"` // 产品ID
	Qty       int32 `json:"qty"`        // 数量
}

// SourceRef 来源引用
// 用于标识订单的业务来源，如营销活动、推荐等
type SourceRef struct {
	Type string `json:"type"` // 来源类型：marketing/referral/direct
	ID   int64  `json:"id"`   // 来源ID
}

// PlaceOrderReq 下单请求
// 封装下单所需的所有信息，确保事务原子性
type PlaceOrderReq struct {
	CustomerID int64          `json:"customer_id"` // 客户ID
	ContactID  int64          `json:"contact_id"`  // 联系人ID（可选）
	Channel    string         `json:"channel"`     // 下单渠道：web/mobile/admin
	PayMethod  string         `json:"pay_method"`  // 支付方式：wallet/cash/online
	Items      []OrderItemReq `json:"items"`       // 订单项列表
	Discount   int64          `json:"discount"`    // 折扣金额（分）
	SourceRef  *SourceRef     `json:"source_ref"`  // 来源引用（可选）
	IdemKey    string         `json:"idem_key"`    // 幂等键，防止重复下单
	Remark     string         `json:"remark"`      // 备注
	AssignedTo int64          `json:"assigned_to"` // 分配给的员工ID
}

// Order 订单领域模型
// 表示订单的核心状态信息
type Order struct {
	ID             int64  `json:"id"`              // 订单ID
	OrderNo        string `json:"order_no"`        // 订单号
	CustomerID     int64  `json:"customer_id"`     // 客户ID
	ContactID      int64  `json:"contact_id"`      // 联系人ID
	TotalAmount    int64  `json:"total_amount"`    // 订单总金额（分）
	DiscountAmount int64  `json:"discount_amount"` // 折扣金额（分）
	FinalAmount    int64  `json:"final_amount"`    // 最终金额（分）
	Status         string `json:"status"`          // 订单状态
	PaymentStatus  string `json:"payment_status"`  // 支付状态
	PayMethod      string `json:"pay_method"`      // 支付方式
	CreatedAt      int64  `json:"created_at"`      // 创建时间
}

// OrderItem 订单项领域模型
// 包含产品快照信息，确保历史数据的完整性
type OrderItem struct {
	ID                  int64  `json:"id"`                    // 订单项ID
	OrderID             int64  `json:"order_id"`              // 订单ID
	ProductID           int64  `json:"product_id"`            // 产品ID
	ProductNameSnapshot string `json:"product_name_snapshot"` // 产品名称快照
	UnitPriceSnapshot   int64  `json:"unit_price_snapshot"`   // 单价快照（分）
	DurationMinSnapshot int32  `json:"duration_min_snapshot"` // 服务时长快照（分钟）
	Quantity            int32  `json:"quantity"`              // 数量
	DiscountAmount      int64  `json:"discount_amount"`       // 该项折扣金额（分）
	FinalPrice          int64  `json:"final_price"`           // 该项最终价格（分）
}

// Service 订单域服务接口
// 提供订单相关的核心业务操作，统一事务边界
type Service interface {
	// PlaceOrder 下单
	// 统一事务内完成：
	// 1. 获取产品信息并创建快照
	// 2. 创建订单和订单项
	// 3. 如果是钱包支付，调用billing域扣款
	// 4. 写入outbox事件
	// 5. 提交事务
	PlaceOrder(ctx context.Context, req PlaceOrderReq) (Order, error)

	// RefundOrder 订单退款
	// 统一事务内完成：
	// 1. 校验订单状态
	// 2. 更新订单状态为已退款
	// 3. 如果原为钱包支付，调用billing域退款
	// 4. 写入outbox事件
	// 5. 提交事务
	RefundOrder(ctx context.Context, orderID int64, reason string) error

	// GetOrder 获取订单详情
	// 包含订单基本信息和订单项列表
	GetOrder(ctx context.Context, orderID int64) (*Order, []OrderItem, error)

	// ListOrders 分页查询订单
	// 支持按客户、状态、时间等条件筛选
	ListOrders(ctx context.Context, customerID int64, status string, page, pageSize int) ([]Order, error)

	// 控制器接口 - 兼容现有控制器
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*OrderResponse, error)
	GetOrderByID(ctx context.Context, idStr string) (*OrderResponse, error)
	ListOrdersForController(ctx context.Context, req *ListOrdersRequest) (*ListOrdersResponse, error)
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	CustomerID int64                    `json:"customer_id" binding:"required"`
	Items      []CreateOrderItemRequest `json:"items" binding:"required"`
	Remark     string                   `json:"remark"`
}

// CreateOrderItemRequest 创建订单项请求
type CreateOrderItemRequest struct {
	ProductID int64   `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required"`
	Price     float64 `json:"price" binding:"required"`
}

// OrderResponse 订单响应
type OrderResponse struct {
	ID          int64               `json:"id"`
	OrderNo     string              `json:"order_no"`
	CustomerID  int64               `json:"customer_id"`
	TotalAmount float64             `json:"total_amount"`
	Status      string              `json:"status"`
	Items       []OrderItemResponse `json:"items"`
	CreatedAt   int64               `json:"created_at"`
	UpdatedAt   int64               `json:"updated_at"`
}

// OrderItemResponse 订单项响应
type OrderItemResponse struct {
	ID        int64   `json:"id"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Amount    float64 `json:"amount"`
}

// ListOrdersRequest 订单列表请求
type ListOrdersRequest struct {
	CustomerID int64  `json:"customer_id,omitempty"`
	Status     string `json:"status,omitempty"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
}

// ListOrdersResponse 订单列表响应
type ListOrdersResponse struct {
	Orders   []OrderResponse `json:"orders"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// Repository 订单域数据访问接口
// 定义订单相关的数据持久化操作
type Repository interface {
	// CreateOrder 创建订单
	// 在事务内同时创建订单和订单项
	CreateOrder(ctx context.Context, order *Order, items []OrderItem) error

	// UpdateOrderStatus 更新订单状态
	UpdateOrderStatus(ctx context.Context, orderID int64, status string) error

	// UpdatePaymentStatus 更新支付状态
	UpdatePaymentStatus(ctx context.Context, orderID int64, paymentStatus string) error

	// GetOrderByID 根据ID获取订单
	GetOrderByID(ctx context.Context, orderID int64) (*Order, error)

	// GetOrderItemsByOrderID 根据订单ID获取订单项
	GetOrderItemsByOrderID(ctx context.Context, orderID int64) ([]OrderItem, error)

	// GenerateOrderNo 生成订单号
	// 确保订单号的唯一性和业务意义
	GenerateOrderNo(ctx context.Context) string
}
