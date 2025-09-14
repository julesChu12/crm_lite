package controller

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/domains/sales"
	"crm_lite/internal/domains/sales/impl"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"
	"time"

	"github.com/gin-gonic/gin"
)

// OrderController 负责处理与订单相关的 HTTP 请求。
type OrderController struct {
	orderService *service.OrderService
	salesService sales.Service
}

// NewOrderController 创建一个新的 OrderController 实例。
// 现在使用新的Sales域服务
func NewOrderController(rm *resource.Manager) *OrderController {
	// 创建Sales领域服务
	salesService := impl.ProvideSales(rm)

	return &OrderController{
		orderService: service.NewOrderService(rm), // 保留旧服务作为备用
		salesService: salesService,
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

	// 转换为Sales领域请求
	salesReq := &sales.CreateOrderRequest{
		CustomerID: req.CustomerID,
		Items:      make([]sales.CreateOrderItemRequest, len(req.Items)),
		Remark:     req.Remark,
	}

	// 转换订单项
	for i, item := range req.Items {
		salesReq.Items[i] = sales.CreateOrderItemRequest{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.UnitPrice,
		}
	}

	order, err := cc.salesService.CreateOrder(c.Request.Context(), salesReq)
	if err != nil {
		resp.SystemError(c, err)
		return
	}

	// 转换为DTO格式
	orderResponse := &dto.OrderResponse{
		ID:          order.ID,
		OrderNo:     order.OrderNo,
		CustomerID:  order.CustomerID,
		OrderDate:   time.Unix(order.CreatedAt, 0),
		Status:      order.Status,
		TotalAmount: order.TotalAmount, // Sales领域已经是float64
		FinalAmount: order.TotalAmount, // 使用TotalAmount作为FinalAmount
		Items:       make([]*dto.OrderItemResponse, len(order.Items)),
		CreatedAt:   time.Unix(order.CreatedAt, 0).Format("2006-01-02 15:04:05"),
	}

	// 转换订单项
	for i, item := range order.Items {
		orderResponse.Items[i] = &dto.OrderItemResponse{
			ID:         item.ID,
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  item.Price,  // 使用Price字段
			FinalPrice: item.Amount, // 使用Amount字段
		}
	}

	resp.Success(c, orderResponse)
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
	order, err := cc.salesService.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		resp.Error(c, resp.CodeNotFound, "订单未找到")
		return
	}

	// 转换为DTO格式
	orderResponse := &dto.OrderResponse{
		ID:          order.ID,
		OrderNo:     order.OrderNo,
		CustomerID:  order.CustomerID,
		OrderDate:   time.Unix(order.CreatedAt, 0),
		Status:      order.Status,
		TotalAmount: order.TotalAmount, // Sales领域已经是float64
		FinalAmount: order.TotalAmount, // 使用TotalAmount作为FinalAmount
		Items:       make([]*dto.OrderItemResponse, len(order.Items)),
		CreatedAt:   time.Unix(order.CreatedAt, 0).Format("2006-01-02 15:04:05"),
	}

	// 转换订单项
	for i, item := range order.Items {
		orderResponse.Items[i] = &dto.OrderItemResponse{
			ID:         item.ID,
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  item.Price,  // 使用Price字段
			FinalPrice: item.Amount, // 使用Amount字段
		}
	}

	resp.Success(c, orderResponse)
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

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 转换为Sales领域请求
	salesReq := &sales.ListOrdersRequest{
		CustomerID: req.CustomerID,
		Status:     req.Status,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	result, err := cc.salesService.ListOrdersForController(c.Request.Context(), salesReq)
	if err != nil {
		resp.SystemError(c, err)
		return
	}

	// 转换为DTO格式
	orderListResponse := &dto.OrderListResponse{
		Orders: make([]*dto.OrderResponse, len(result.Orders)),
		Total:  result.Total,
	}

	for i, order := range result.Orders {
		orderListResponse.Orders[i] = &dto.OrderResponse{
			ID:          order.ID,
			OrderNo:     order.OrderNo,
			CustomerID:  order.CustomerID,
			OrderDate:   time.Unix(order.CreatedAt, 0),
			Status:      order.Status,
			TotalAmount: order.TotalAmount, // Sales领域已经是float64
			FinalAmount: order.TotalAmount, // 使用TotalAmount作为FinalAmount
			Items:       make([]*dto.OrderItemResponse, len(order.Items)),
			CreatedAt:   time.Unix(order.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		}

		// 转换订单项
		for j, item := range order.Items {
			orderListResponse.Orders[i].Items[j] = &dto.OrderItemResponse{
				ID:         item.ID,
				ProductID:  item.ProductID,
				Quantity:   item.Quantity,
				UnitPrice:  item.Price,  // 使用Price字段
				FinalPrice: item.Amount, // 使用Amount字段
			}
		}
	}

	resp.Success(c, orderListResponse)
}
