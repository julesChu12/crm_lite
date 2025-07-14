package dto

import "time"

// OrderItemRequest 代表创建订单请求中的单个订单项。
type OrderItemRequest struct {
	ProductID int64   `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,gt=0"`
	UnitPrice float64 `json:"unit_price" binding:"required,gte=0"` // 允许在下单时覆盖产品单价
}

// OrderCreateRequest 定义了创建新订单的请求体。
type OrderCreateRequest struct {
	CustomerID int64               `json:"customer_id" binding:"required"`
	OrderDate  time.Time           `json:"order_date" binding:"required"`
	Status     string              `json:"status" binding:"omitempty,oneof=draft pending confirmed"`
	Items      []*OrderItemRequest `json:"items" binding:"required,min=1"` // 订单项，至少要有一项
	Remark     string              `json:"remark"`
}

// OrderUpdateRequest 定义了更新订单的请求体。
type OrderUpdateRequest struct {
	Status string `json:"status" binding:"omitempty,oneof=draft pending confirmed processing shipped completed cancelled refunded"`
	Remark string `json:"remark"`
}

// OrderItemResponse 代表 API 响应中的单个订单项。
type OrderItemResponse struct {
	ID         int64   `json:"id"`
	ProductID  int64   `json:"product_id"`  // 产品ID
	Quantity   int     `json:"quantity"`    // 数量
	UnitPrice  float64 `json:"unit_price"`  // 成交单价
	FinalPrice float64 `json:"final_price"` // 最终价格 (数量 * 单价)
}

// OrderResponse 代表 API 响应中的单个订单。
type OrderResponse struct {
	ID           int64                `json:"id"`
	OrderNo      string               `json:"order_no"`                // 订单号
	CustomerID   int64                `json:"customer_id"`             // 客户ID
	CustomerName string               `json:"customer_name,omitempty"` // 客户名称（关联查询时填充）
	OrderDate    time.Time            `json:"order_date"`              // 下单日期
	Status       string               `json:"status"`                  // 订单状态
	TotalAmount  float64              `json:"total_amount"`            // 订单总金额
	FinalAmount  float64              `json:"final_amount"`            // 最终成交金额
	Remark       string               `json:"remark"`                  // 备注
	CreatedAt    time.Time            `json:"created_at"`              // 创建时间
	Items        []*OrderItemResponse `json:"items"`                   // 订单项列表
}

// OrderListRequest 定义了列出订单的查询参数。
type OrderListRequest struct {
	Page       int     `form:"page,default=1"`
	PageSize   int     `form:"page_size,default=10"`
	CustomerID int64   `form:"customer_id"`                      // 按客户ID筛选
	Status     string  `form:"status"`                           // 按状态筛选
	OrderBy    string  `form:"order_by,default=order_date_desc"` // 排序字段
	IDs        []int64 `form:"ids"`                              // 按 ID 列表过滤
}

// OrderListResponse 定义了订单列表的响应。
type OrderListResponse struct {
	Total  int64            `json:"total"`
	Orders []*OrderResponse `json:"orders"`
}
