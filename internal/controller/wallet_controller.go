package controller

import (
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"crm_lite/internal/dto"
	"crm_lite/internal/middleware"

	"github.com/gin-gonic/gin"
)

// WalletController 负责处理钱包相关的 API 请求
type WalletController struct {
	walletSvc service.IWalletService
}

// NewWalletController 创建一个新的 WalletController
func NewWalletController(walletSvc service.IWalletService) *WalletController {
	return &WalletController{walletSvc: walletSvc}
}

// GetWalletByCustomerID @Summary 获取客户钱包信息
// @Description 根据客户ID获取其默认的余额钱包信息
// @Tags Wallets
// @Accept json
// @Produce json
// @Param id path int true "客户ID"
// @Success 200 {object} resp.Response{data=dto.WalletResponse} "成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 404 {object} resp.Response "钱包未找到"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /v1/customers/{id}/wallet [get]
func (c *WalletController) GetWalletByCustomerID(ctx *gin.Context) {
	customerIDStr := ctx.Param("id")
	customerID, err := strconv.ParseInt(customerIDStr, 10, 64)
	if err != nil {
		resp.Error(ctx, http.StatusBadRequest, "无效的客户ID")
		return
	}

	wallet, err := c.walletSvc.GetWalletByCustomerID(ctx.Request.Context(), customerID)
	if err != nil {
		if errors.Is(err, service.ErrWalletNotFound) {
			resp.Error(ctx, http.StatusNotFound, err.Error())
		} else {
			resp.Error(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	resp.Success(ctx, wallet)
}

// CreateTransaction @Summary 创建钱包交易
// @Description 为客户的钱包创建一笔交易（充值、消费或退款）
// @Tags Wallets
// @Accept json
// @Produce json
// @Param id path int true "客户ID"
// @Param transaction body dto.WalletTransactionRequest true "交易信息 (type: recharge|consume|refund; bonus_amount 可选; consume 时需 phone_last4)"
// @Success 200 {object} resp.Response "操作成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 403 {object} resp.Response "无权操作"
// @Failure 404 {object} resp.Response "客户或钱包未找到"
// @Failure 422 {object} resp.Response "业务逻辑错误（如余额不足）"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /v1/customers/{id}/wallet/transactions [post]
func (c *WalletController) CreateTransaction(ctx *gin.Context) {
	// 1. 解析参数和请求体
	customerIDStr := ctx.Param("id")
	customerID, err := strconv.ParseInt(customerIDStr, 10, 64)
	if err != nil {
		resp.Error(ctx, http.StatusBadRequest, "无效的客户ID")
		return
	}

	var req dto.WalletTransactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// 2. 从上下文中获取操作员信息
	operatorIDVal, _ := ctx.Get(middleware.ContextKeyUserID)
	operatorID := operatorIDVal.(int64)

	// 3. 调用服务
	err = c.walletSvc.CreateTransaction(ctx.Request.Context(), customerID, operatorID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCustomerNotFound), errors.Is(err, service.ErrWalletNotFound):
			resp.Error(ctx, http.StatusNotFound, err.Error())
		case strings.Contains(err.Error(), "phone last4"):
			resp.Error(ctx, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInsufficientBalance):
			resp.Error(ctx, http.StatusUnprocessableEntity, err.Error())
		default:
			resp.Error(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	resp.Success(ctx, nil)
}

// GetTransactions @Summary 获取客户钱包交易流水列表
// @Description 根据客户ID获取其钱包交易记录，支持分页
// @Tags Wallets
// @Accept json
// @Produce json
// @Param id path int true "客户ID"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} resp.Response{data=dto.ListWalletTransactionsResponse} "成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 404 {object} resp.Response "钱包未找到"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /v1/customers/{id}/wallet/transactions [get]
func (c *WalletController) GetTransactions(ctx *gin.Context) {
	customerIDStr := ctx.Param("id")
	customerID, err := strconv.ParseInt(customerIDStr, 10, 64)
	if err != nil {
		resp.Error(ctx, http.StatusBadRequest, "无效的客户ID")
		return
	}

	var req dto.ListWalletTransactionsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		resp.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	transactions, total, err := c.walletSvc.GetTransactions(ctx.Request.Context(), customerID, req.Page, req.Limit)
	if err != nil {
		if errors.Is(err, service.ErrWalletNotFound) {
			resp.Error(ctx, http.StatusNotFound, err.Error())
		} else {
			resp.Error(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	resp.Success(ctx, dto.ListWalletTransactionsResponse{
		Transactions: transactions,
		Total:        total,
	})
}

// ProcessRefund @Summary 处理钱包退款
// @Description 为客户处理订单退款，将金额退回到钱包
// @Tags Wallets
// @Accept json
// @Produce json
// @Param id path int true "客户ID"
// @Param refund body dto.WalletRefundRequest true "退款信息"
// @Success 200 {object} resp.Response "退款成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 403 {object} resp.Response "无权操作"
// @Failure 404 {object} resp.Response "客户或订单未找到"
// @Failure 422 {object} resp.Response "业务逻辑错误（如订单状态不允许退款）"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /v1/customers/{id}/wallet/refund [post]
func (c *WalletController) ProcessRefund(ctx *gin.Context) {
	customerIDStr := ctx.Param("id")
	customerID, err := strconv.ParseInt(customerIDStr, 10, 64)
	if err != nil {
		resp.Error(ctx, http.StatusBadRequest, "无效的客户ID")
		return
	}

	var req dto.WalletRefundRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// 从上下文中获取操作员信息
	operatorIDVal, _ := ctx.Get(middleware.ContextKeyUserID)
	operatorID := operatorIDVal.(int64)

	// 调用服务处理退款
	err = c.walletSvc.ProcessRefund(ctx.Request.Context(), customerID, operatorID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCustomerNotFound), errors.Is(err, service.ErrWalletNotFound):
			resp.Error(ctx, http.StatusNotFound, err.Error())
		case err.Error() == "订单不存在或不属于该客户" || err.Error() == "订单状态不允许退款":
			resp.Error(ctx, http.StatusUnprocessableEntity, err.Error())
		default:
			resp.Error(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	resp.Success(ctx, nil)
}
