package controller

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"
	"errors"

	"github.com/gin-gonic/gin"
)

// OrderController 负责处理与订单相关的 HTTP 请求。
type OrderController struct {
	orderService *service.OrderService
}

// NewOrderController 创建一个新的 OrderController 实例。
func NewOrderController(rm *resource.Manager) *OrderController {
	return &OrderController{
		orderService: service.NewOrderService(rm),
	}
}

// CreateOrder
// @Summary 创建订单
// @Description 创建一个新订单及其订单项
// @Tags Orders
// @Accept json
// @Produce json
// @Param order body dto.OrderCreateRequest true "订单信息"
// @Success 200 {object} resp.Response{data=dto.OrderResponse}
// @Router /orders [post]
func (cc *OrderController) CreateOrder(c *gin.Context) {
	var req dto.OrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	order, err := cc.orderService.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			resp.Error(c, resp.CodeInvalidParam, "客户不存在")
			return
		}
		if errors.Is(err, service.ErrProductNotFound) {
			resp.Error(c, resp.CodeInvalidParam, "一个或多个产品不存在")
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, order)
}

// GetOrder
// @Summary 获取单个订单
// @Description 根据 ID 获取单个订单的详细信息，包括订单项
// @Tags Orders
// @Produce json
// @Param id path string true "订单 ID"
// @Success 200 {object} resp.Response{data=dto.OrderResponse}
// @Router /orders/{id} [get]
func (cc *OrderController) GetOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := cc.orderService.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			resp.Error(c, resp.CodeNotFound, "订单未找到")
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, order)
}

// ListOrders
// @Summary 获取订单列表
// @Description 获取订单列表，支持分页和筛选
// @Tags Orders
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(10)
// @Param customer_id query int false "按客户 ID 筛选"
// @Param status query string false "按状态筛选"
// @Param order_by query string false "排序字段 (e.g., order_date_desc)"
// @Success 200 {object} resp.Response{data=dto.OrderListResponse}
// @Router /orders [get]
func (cc *OrderController) ListOrders(c *gin.Context) {
	var req dto.OrderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	result, err := cc.orderService.ListOrders(c.Request.Context(), &req)
	if err != nil {
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, result)
}
