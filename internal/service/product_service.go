package service

import (
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

var (
	// ErrSKUAlreadyExists 表示 SKU 已存在。
	ErrSKUAlreadyExists = errors.New("sku already exists")
)

// ProductService 封装了与产品相关的业务逻辑。
type ProductService struct {
	q *query.Query
}

// NewProductService 创建一个新的 ProductService 实例。
func NewProductService(rm *resource.Manager) *ProductService {
	db, err := resource.Get[*resource.DBResource](rm, resource.DBServiceKey)
	if err != nil {
		panic("初始化 ProductService 失败，无法获取数据库资源: " + err.Error())
	}
	return &ProductService{q: query.Use(db.DB)}
}

// toProductResponse 将 model.Product 转换为 dto.ProductResponse。
// 注意：此函数处理了数据库模型和 DTO 之间的字段映射（例如 Category -> SKU, StockQuantity -> Stock）。
func (s *ProductService) toProductResponse(p *model.Product) *dto.ProductResponse {
	return &dto.ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		SKU:         p.Category,           // 将数据库中的 Category 字段映射到 DTO 的 SKU 字段
		Stock:       int(p.StockQuantity), // 将数据库中的 StockQuantity 字段映射到 DTO 的 Stock 字段
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// CreateProduct 创建一个新产品。
func (s *ProductService) CreateProduct(ctx context.Context, req *dto.ProductCreateRequest) (*dto.ProductResponse, error) {
	// 通过 Category 字段检查 SKU 的唯一性
	existing, _ := s.q.Product.WithContext(ctx).Where(s.q.Product.Category.Eq(req.SKU)).First()
	if existing != nil {
		return nil, ErrSKUAlreadyExists
	}

	product := &model.Product{
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		Category:      req.SKU, // 将请求中的 SKU 存入数据库的 Category 字段
		StockQuantity: int32(req.Stock),
	}

	if err := s.q.Product.WithContext(ctx).Create(product); err != nil {
		return nil, fmt.Errorf("创建产品失败: %w", err)
	}
	return s.toProductResponse(product), nil
}

// GetProductByID 根据 ID 获取单个产品。
func (s *ProductService) GetProductByID(ctx context.Context, idStr string) (*dto.ProductResponse, error) {
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

// ListProducts 获取产品列表，支持分页、筛选和排序。
func (s *ProductService) ListProducts(ctx context.Context, req *dto.ProductListRequest) (*dto.ProductListResponse, error) {
	q := s.q.Product.WithContext(ctx)

	// 如果提供了 ID 列表，则优先按 ID 查询，忽略其他筛选条件
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
			// 验证排序字段是否存在于模型中
			if col, ok := s.q.Product.GetFieldByName(parts[0]); ok {
				if parts[1] == "desc" {
					q = q.Order(col.Desc())
				} else {
					q = q.Order(col)
				}
			}
		}
	} else {
		// 默认按创建时间降序排序
		q = q.Order(s.q.Product.CreatedAt.Desc())
	}

	total, err := q.Count()
	if err != nil {
		return nil, err
	}

	var products []*model.Product
	// 如果是按 ID 列表查询，则不应用分页
	if len(req.IDs) > 0 {
		products, err = q.Find()
	} else {
		products, err = q.Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).Find()
	}

	if err != nil {
		return nil, err
	}

	items := make([]*dto.ProductResponse, len(products))
	for i, p := range products {
		items[i] = s.toProductResponse(p)
	}

	return &dto.ProductListResponse{
		Total:    total,
		Products: items,
	}, nil
}

// UpdateProduct 更新一个现有的产品。
func (s *ProductService) UpdateProduct(ctx context.Context, idStr string, req *dto.ProductUpdateRequest) (*dto.ProductResponse, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, ErrProductNotFound
	}

	// 验证产品是否存在
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
		updates["price"] = req.Price
	}
	if req.Stock >= 0 {
		updates["stock_quantity"] = int32(req.Stock)
	}

	if len(updates) > 0 {
		if _, err := s.q.Product.WithContext(ctx).Where(s.q.Product.ID.Eq(id)).Updates(updates); err != nil {
			return nil, err
		}
	}

	// 重新获取更新后的数据并返回
	return s.GetProductByID(ctx, idStr)
}

// DeleteProduct 删除一个产品。
func (s *ProductService) DeleteProduct(ctx context.Context, idStr string) error {
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

// ErrProductNotFound 表示未找到产品。
var ErrProductNotFound = errors.New("product not found")
