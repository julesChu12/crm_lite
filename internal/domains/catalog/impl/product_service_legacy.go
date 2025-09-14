package impl

import (
	"context"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// 纯搬家：旧 ProductService 的等价实现，暂不接入 controller，保留行为

// IProductRepo 定义了产品数据仓库的接口
type IProductRepo interface {
	FindBySKU(ctx context.Context, sku string) (*model.Product, error)
	Create(ctx context.Context, product *model.Product) error
	FindByID(ctx context.Context, id int64) (*model.Product, error)
	List(ctx context.Context, req *dto.ProductListRequest) ([]*model.Product, int64, error)
	Updates(ctx context.Context, id int64, updates map[string]interface{}) (int64, error)
	Delete(ctx context.Context, id int64) (int64, error)
}

// productRepo 实现了 IProductRepo 接口
type productRepo struct {
	q *query.Query
}

// NewProductRepo 创建一个新的产品仓库实例
func NewProductRepo(db *gorm.DB) IProductRepo {
	return &productRepo{q: query.Use(db)}
}

func (r *productRepo) FindBySKU(ctx context.Context, sku string) (*model.Product, error) {
	return r.q.Product.WithContext(ctx).Where(r.q.Product.Category.Eq(sku)).First()
}

func (r *productRepo) Create(ctx context.Context, product *model.Product) error {
	return r.q.Product.WithContext(ctx).Create(product)
}

func (r *productRepo) FindByID(ctx context.Context, id int64) (*model.Product, error) {
	return r.q.Product.WithContext(ctx).Where(r.q.Product.ID.Eq(id)).First()
}

func (r *productRepo) List(ctx context.Context, req *dto.ProductListRequest) ([]*model.Product, int64, error) {
	q := r.q.Product.WithContext(ctx)

	if len(req.IDs) > 0 {
		q = q.Where(r.q.Product.ID.In(req.IDs...))
	} else {
		if req.Name != "" {
			q = q.Where(r.q.Product.Name.Like("%" + req.Name + "%"))
		}
		if req.SKU != "" {
			q = q.Where(r.q.Product.Category.Eq(req.SKU))
		}
	}

	if req.OrderBy != "" {
		parts := strings.Split(req.OrderBy, "_")
		if len(parts) == 2 {
			if col, ok := r.q.Product.GetFieldByName(parts[0]); ok {
				if parts[1] == "desc" {
					q = q.Order(col.Desc())
				} else {
					q = q.Order(col)
				}
			}
		}
	} else {
		q = q.Order(r.q.Product.CreatedAt.Desc())
	}

	total, err := q.Count()
	if err != nil {
		return nil, 0, err
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
		return nil, 0, err
	}

	return products, total, err
}

func (r *productRepo) Updates(ctx context.Context, id int64, updates map[string]interface{}) (int64, error) {
	result, err := r.q.Product.WithContext(ctx).Where(r.q.Product.ID.Eq(id)).Updates(updates)
	return result.RowsAffected, err
}

func (r *productRepo) Delete(ctx context.Context, id int64) (int64, error) {
	result, err := r.q.Product.WithContext(ctx).Where(r.q.Product.ID.Eq(id)).Delete()
	return result.RowsAffected, err
}

// ProductService 封装了与产品相关的业务逻辑（纯搬家）
type ProductService struct {
	repo IProductRepo
}

// NewProductService 创建一个新的 ProductService 实例。
func NewProductService(repo IProductRepo) *ProductService {
	return &ProductService{repo: repo}
}

// toProductResponse 将 model.Product 转换为 dto.ProductResponse。
func (s *ProductService) toProductResponse(p *model.Product) *dto.ProductResponse {
	if p == nil {
		return nil
	}
	return &dto.ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		SKU:         p.Category,
		Stock:       int(p.StockQuantity),
		CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// CreateProduct 创建一个新产品。
func (s *ProductService) CreateProduct(ctx context.Context, req *dto.ProductCreateRequest) (*dto.ProductResponse, error) {
	existing, err := s.repo.FindBySKU(ctx, req.SKU)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("检查SKU失败: %w", err)
	}
	if existing != nil {
		return nil, service.ErrSKUAlreadyExists
	}

	product := &model.Product{
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		Category:      req.SKU,
		StockQuantity: int32(req.Stock),
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("创建产品失败: %w", err)
	}
	return s.toProductResponse(product), nil
}

// GetProductByID 根据 ID 获取单个产品。
func (s *ProductService) GetProductByID(ctx context.Context, idStr string) (*dto.ProductResponse, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, service.ErrProductNotFound
	}

	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrProductNotFound
		}
		return nil, err
	}
	return s.toProductResponse(product), nil
}

// ListProducts 获取产品列表。
func (s *ProductService) ListProducts(ctx context.Context, req *dto.ProductListRequest) (*dto.ProductListResponse, error) {
	products, total, err := s.repo.List(ctx, req)
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
		return nil, service.ErrProductNotFound
	}

	if _, err := s.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrProductNotFound
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
		if _, err := s.repo.Updates(ctx, id, updates); err != nil {
			return nil, err
		}
	}

	updatedProduct, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, service.ErrProductNotFound
	}
	return s.toProductResponse(updatedProduct), nil
}

// DeleteProduct 删除一个产品。
func (s *ProductService) DeleteProduct(ctx context.Context, idStr string) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return service.ErrProductNotFound
	}

	rowsAffected, err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return service.ErrProductNotFound
	}
	return nil
}
