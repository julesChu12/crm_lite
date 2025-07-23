package controller

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"
	"errors"

	"github.com/gin-gonic/gin"
)

// ProductController 负责处理与产品相关的 HTTP 请求。
type ProductController struct {
	productService *service.ProductService
}

// NewProductController 创建一个新的 ProductController 实例。
func NewProductController(rm *resource.Manager) *ProductController {
	// 1. 从资源管理器获取数据库资源
	db, err := resource.Get[*resource.DBResource](rm, resource.DBServiceKey)
	if err != nil {
		panic("初始化 ProductController 失败，无法获取数据库资源: " + err.Error())
	}
	// 2. 创建 repo
	productRepo := service.NewProductRepo(db.DB)
	// 3. 注入 repo 来创建 service
	return &ProductController{
		productService: service.NewProductService(productRepo),
	}
}

// CreateProduct
// @Summary 创建产品
// @Description 创建一个新产品
// @Tags Products
// @Accept json
// @Produce json
// @Param product body dto.ProductCreateRequest true "产品信息"
// @Success 200 {object} resp.Response{data=dto.ProductResponse}
// @Router /products [post]
func (cc *ProductController) CreateProduct(c *gin.Context) {
	var req dto.ProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	product, err := cc.productService.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrSKUAlreadyExists) {
			resp.Error(c, resp.CodeConflict, "SKU 已存在")
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, product)
}

// GetProduct
// @Summary 获取单个产品
// @Description 根据 ID 获取单个产品的详细信息
// @Tags Products
// @Produce json
// @Param id path string true "产品 ID"
// @Success 200 {object} resp.Response{data=dto.ProductResponse}
// @Router /products/{id} [get]
func (cc *ProductController) GetProduct(c *gin.Context) {
	id := c.Param("id")
	product, err := cc.productService.GetProductByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			resp.Error(c, resp.CodeNotFound, "产品未找到")
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, product)
}

// ListProducts
// @Summary 获取产品列表
// @Description 获取产品列表，支持分页和筛选
// @Tags Products
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(10)
// @Param name query string false "按名称模糊搜索"
// @Param sku query string false "按 SKU 精确搜索"
// @Param order_by query string false "排序字段 (e.g., created_at_desc)"
// @Success 200 {object} resp.Response{data=dto.ProductListResponse}
// @Router /products [get]
func (cc *ProductController) ListProducts(c *gin.Context) {
	var req dto.ProductListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	result, err := cc.productService.ListProducts(c.Request.Context(), &req)
	if err != nil {
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, result)
}

// BatchGetProducts
// @Summary 批量获取产品
// @Description 根据 ID 列表批量获取产品信息
// @Tags Products
// @Accept json
// @Produce json
// @Param body body dto.ProductBatchGetRequest true "产品 ID 列表"
// @Success 200 {object} resp.Response{data=dto.ProductListResponse}
// @Router /products/batch-get [post]
func (cc *ProductController) BatchGetProducts(c *gin.Context) {
	var req dto.ProductBatchGetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	serviceReq := &dto.ProductListRequest{IDs: req.IDs}
	result, err := cc.productService.ListProducts(c.Request.Context(), serviceReq)
	if err != nil {
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, result)
}

// UpdateProduct
// @Summary 更新产品
// @Description 更新一个现有产品
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "产品 ID"
// @Param product body dto.ProductUpdateRequest true "要更新的产品信息"
// @Success 200 {object} resp.Response{data=dto.ProductResponse}
// @Router /products/{id} [put]
func (cc *ProductController) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var req dto.ProductUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	product, err := cc.productService.UpdateProduct(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			resp.Error(c, resp.CodeNotFound, "产品未找到")
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, product)
}

// DeleteProduct
// @Summary 删除产品
// @Description 根据 ID 删除一个产品
// @Tags Products
// @Produce json
// @Param id path string true "产品 ID"
// @Success 204
// @Router /products/{id} [delete]
func (cc *ProductController) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	if err := cc.productService.DeleteProduct(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			resp.Error(c, resp.CodeNotFound, "产品未找到")
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.SuccessWithCode(c, resp.CodeNoContent, nil)
}
