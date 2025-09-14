package impl

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/domains/catalog"

	"gorm.io/gorm"
)

var (
	ErrProductNotFound  = errors.New("product not found")
	ErrProductOff       = errors.New("product is not sellable")
	ErrSKUAlreadyExists = errors.New("sku already exists")
)

// ServiceImpl 通过 gorm-gen query 访问产品数据
type ServiceImpl struct {
	q *query.Query
}

func New(q *query.Query) *ServiceImpl { return &ServiceImpl{q: q} }

// toDomain 将 DAO 模型转换为领域对象
func toDomain(p *model.Product) catalog.Product {
	// Price decimal 元 → 分
	priceCents := int64(math.Round(p.Price * 100))
	status := "off"
	if p.IsActive {
		status = "on"
	}
	return catalog.Product{
		ID:          p.ID,
		Name:        p.Name,
		Price:       priceCents,
		DurationMin: 0,          // 现有模型暂无时长字段，占位 0
		Status:      status,     // on/off 派生自 is_active
		Type:        "product",  // 默认类型
		Category:    p.Category, // 使用现有的 Category 字段
	}
}

// toProductResponse 将 model.Product 转换为 catalog.ProductResponse
func (s *ServiceImpl) toProductResponse(p *model.Product) *catalog.ProductResponse {
	if p == nil {
		return nil
	}
	return &catalog.ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       int64(math.Round(p.Price * 100)), // 元转分
		SKU:         p.Category,
		Stock:       p.StockQuantity,
		CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// Get 根据 ID 获取产品
func (s *ServiceImpl) Get(ctx context.Context, id int64) (catalog.Product, error) {
	p, err := s.q.Product.WithContext(ctx).Where(s.q.Product.ID.Eq(id)).First()
	if err != nil {
		return catalog.Product{}, ErrProductNotFound
	}
	return toDomain(p), nil
}

// EnsureSellable 校验产品可售
func (s *ServiceImpl) EnsureSellable(ctx context.Context, id int64) error {
	p, err := s.q.Product.WithContext(ctx).Where(s.q.Product.ID.Eq(id)).First()
	if err != nil {
		return ErrProductNotFound
	}
	if !p.IsActive {
		return ErrProductOff
	}
	return nil
}

// BatchGet 批量获取产品信息
func (s *ServiceImpl) BatchGet(ctx context.Context, ids []int64) ([]catalog.Product, error) {
	products, err := s.q.Product.WithContext(ctx).Where(s.q.Product.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}

	result := make([]catalog.Product, len(products))
	for i, p := range products {
		result[i] = toDomain(p)
	}
	return result, nil
}

// CreateProduct 创建一个新产品
func (s *ServiceImpl) CreateProduct(ctx context.Context, req *catalog.CreateProductRequest) (*catalog.ProductResponse, error) {
	existing, err := s.q.Product.WithContext(ctx).Where(s.q.Product.Category.Eq(req.SKU)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("检查SKU失败: %w", err)
	}
	if existing != nil {
		return nil, ErrSKUAlreadyExists
	}

	product := &model.Product{
		Name:          req.Name,
		Description:   req.Description,
		Price:         float64(req.Price) / 100, // 分转元
		Category:      req.SKU,
		StockQuantity: req.Stock,
		IsActive:      true, // 新产品默认激活
	}

	if err := s.q.Product.WithContext(ctx).Create(product); err != nil {
		return nil, fmt.Errorf("创建产品失败: %w", err)
	}
	return s.toProductResponse(product), nil
}

// GetProductByID 根据 ID 获取单个产品
func (s *ServiceImpl) GetProductByID(ctx context.Context, idStr string) (*catalog.ProductResponse, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, ErrProductNotFound
	}

	product, err := s.q.Product.WithContext(ctx).Where(s.q.Product.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return s.toProductResponse(product), nil
}

// ListProducts 获取产品列表
func (s *ServiceImpl) ListProducts(ctx context.Context, req *catalog.ProductListRequest) (*catalog.ProductListResponse, error) {
	q := s.q.Product.WithContext(ctx)

	if len(req.IDs) > 0 {
		q = q.Where(s.q.Product.ID.In(req.IDs...))
	} else {
		if req.Name != "" {
			q = q.Where(s.q.Product.Name.Like("%" + req.Name + "%"))
		}
		if req.SKU != "" {
			q = q.Where(s.q.Product.Category.Eq(req.SKU))
		}
	}

	if req.OrderBy != "" {
		parts := strings.Split(req.OrderBy, "_")
		if len(parts) == 2 {
			if col, ok := s.q.Product.GetFieldByName(parts[0]); ok {
				if parts[1] == "desc" {
					q = q.Order(col.Desc())
				} else {
					q = q.Order(col)
				}
			}
		}
	} else {
		q = q.Order(s.q.Product.CreatedAt.Desc())
	}

	total, err := q.Count()
	if err != nil {
		return nil, err
	}

	var products []*model.Product
	if len(req.IDs) > 0 {
		products, err = q.Find()
	} else if req.PageSize > 0 {
		products, err = q.Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).Find()
	} else {
		products, err = q.Find()
	}

	if err != nil {
		return nil, err
	}

	items := make([]*catalog.ProductResponse, len(products))
	for i, p := range products {
		items[i] = s.toProductResponse(p)
	}

	return &catalog.ProductListResponse{
		Total:    total,
		Products: items,
	}, nil
}

// UpdateProduct 更新一个现有的产品
func (s *ServiceImpl) UpdateProduct(ctx context.Context, idStr string, req *catalog.UpdateProductRequest) (*catalog.ProductResponse, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, ErrProductNotFound
	}

	if _, err := s.q.Product.WithContext(ctx).Where(s.q.Product.ID.Eq(id)).First(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Price > 0 {
		updates["price"] = float64(req.Price) / 100 // 分转元
	}
	if req.Stock >= 0 {
		updates["stock_quantity"] = req.Stock
	}

	if len(updates) > 0 {
		if _, err := s.q.Product.WithContext(ctx).Where(s.q.Product.ID.Eq(id)).Updates(updates); err != nil {
			return nil, err
		}
	}

	updatedProduct, err := s.q.Product.WithContext(ctx).Where(s.q.Product.ID.Eq(id)).First()
	if err != nil {
		return nil, ErrProductNotFound
	}
	return s.toProductResponse(updatedProduct), nil
}

// DeleteProduct 删除一个产品
func (s *ServiceImpl) DeleteProduct(ctx context.Context, idStr string) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ErrProductNotFound
	}

	result, err := s.q.Product.WithContext(ctx).Where(s.q.Product.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrProductNotFound
	}
	return nil
}

// 断言接口实现
var _ catalog.Service = (*ServiceImpl)(nil)
