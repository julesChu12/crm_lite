// Package catalog 产品域服务接口
// 职责：产品查询、可售性校验、库存管理
package catalog

import "context"

// Product 产品领域模型
// 统一产品的核心属性，用于跨域交互
type Product struct {
	ID          int64  `json:"id"`           // 产品ID
	Name        string `json:"name"`         // 产品名称
	Price       int64  `json:"price"`        // 价格（分为单位，避免浮点精度问题）
	DurationMin int32  `json:"duration_min"` // 服务时长（分钟），用于预约类产品
	Status      string `json:"status"`       // 状态：active/inactive/deleted
	Type        string `json:"type"`         // 类型：product/service
	Category    string `json:"category"`     // 分类
}

// CreateProductRequest 创建产品请求
type CreateProductRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Price       int64  `json:"price" binding:"required,min=1"`
	SKU         string `json:"sku" binding:"required"`
	Stock       int32  `json:"stock" binding:"min=0"`
}

// UpdateProductRequest 更新产品请求
type UpdateProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price" binding:"min=1"`
	Stock       int32  `json:"stock" binding:"min=0"`
}

// ProductListRequest 产品列表请求
type ProductListRequest struct {
	Page     int     `json:"page" binding:"min=1"`
	PageSize int     `json:"page_size" binding:"min=1,max=100"`
	IDs      []int64 `json:"ids"`
	Name     string  `json:"name"`
	SKU      string  `json:"sku"`
	OrderBy  string  `json:"order_by"`
}

// ProductResponse 产品响应
type ProductResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	SKU         string `json:"sku"`
	Stock       int32  `json:"stock"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ProductListResponse 产品列表响应
type ProductListResponse struct {
	Total    int64              `json:"total"`
	Products []*ProductResponse `json:"products"`
}

// Service 产品域服务接口
// 提供产品相关的核心业务操作，供其他域调用
type Service interface {
	// Get 根据ID获取产品信息
	// 用于订单下单时获取产品详情和快照
	Get(ctx context.Context, id int64) (Product, error)

	// EnsureSellable 校验产品是否可售
	// 检查产品状态、库存等，确保可以下单
	// 返回error表示不可售及原因
	EnsureSellable(ctx context.Context, id int64) error

	// BatchGet 批量获取产品信息
	// 用于订单下单时一次性获取多个产品信息，提高性能
	BatchGet(ctx context.Context, ids []int64) ([]Product, error)

	// Legacy CRUD 接口 - 兼容现有控制器
	CreateProduct(ctx context.Context, req *CreateProductRequest) (*ProductResponse, error)
	GetProductByID(ctx context.Context, idStr string) (*ProductResponse, error)
	ListProducts(ctx context.Context, req *ProductListRequest) (*ProductListResponse, error)
	UpdateProduct(ctx context.Context, idStr string, req *UpdateProductRequest) (*ProductResponse, error)
	DeleteProduct(ctx context.Context, idStr string) error
}

// Repository 产品域数据访问接口
// 定义产品数据的持久化操作，由具体实现决定使用何种数据源
type Repository interface {
	// FindByID 根据ID查找产品
	FindByID(ctx context.Context, id int64) (*Product, error)

	// FindByIDs 批量根据ID查找产品
	FindByIDs(ctx context.Context, ids []int64) ([]*Product, error)

	// CheckStock 检查库存是否足够
	CheckStock(ctx context.Context, productID int64, quantity int32) (bool, error)
}
