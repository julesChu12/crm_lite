package dto

import "time"

// ProductResponse 用于 API 响应的单个产品数据。
type ProductResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`        // 产品名称
	Description string    `json:"description"` // 产品描述
	Price       float64   `json:"price"`       // 价格
	SKU         string    `json:"sku"`         // 库存单位
	Stock       int       `json:"stock"`       // 库存数量
	CreatedAt   time.Time `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`  // 更新时间
}

// ProductCreateRequest 定义了创建新产品的请求体。
type ProductCreateRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	SKU         string  `json:"sku" binding:"required,alphanum"`
	Stock       int     `json:"stock" binding:"gte=0"`
}

// ProductUpdateRequest 定义了更新现有产品的请求体。
type ProductUpdateRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"omitempty,gt=0"`
	Stock       int     `json:"stock" binding:"omitempty,gte=0"`
}

// ProductListRequest 定义了列出产品的查询参数。
type ProductListRequest struct {
	Page     int     `form:"page,default=1"`
	PageSize int     `form:"page_size,default=10"`
	Name     string  `form:"name"`     // 按名称模糊搜索
	SKU      string  `form:"sku"`      // 按 SKU 精确搜索
	OrderBy  string  `form:"order_by"` // 例如, created_at_desc
	IDs      []int64 `form:"ids"`      // 按 ID 列表过滤
}

// ProductBatchGetRequest 定义了批量获取产品的请求体。
type ProductBatchGetRequest struct {
	IDs []int64 `json:"ids" binding:"required"`
}

// ProductListResponse 定义了产品列表的响应。
type ProductListResponse struct {
	Total    int64              `json:"total"`
	Products []*ProductResponse `json:"products"`
}
