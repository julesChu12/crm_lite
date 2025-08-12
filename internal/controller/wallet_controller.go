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
// @Description 为客户的钱包创建一笔交易（充值/消费/退款）。\n- 规则：\n- 1) 充值 (type=recharge)：支持可选字段 bonus_amount>0，将自动追加一条 correction 赠送流水；\n- 2) 消费 (type=consume)：必须提供客户手机号后四位 phone_last4 以做现场核对；\n- 3) 退款 (type=refund)：正向增加余额；\n- 4) 累计统计：total_recharged 仅统计实付充值金额（不含赠送）。
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

// ListTransactions @Summary 获取客户钱包流水
// @Description 查询客户钱包的交易流水
// @Tags Wallets
// @Produce json
// @Param id path int true "客户ID"
// @Param type query string false "交易类型"
// @Param source query string false "来源"
// @Param start_date query string false "开始日期(YYYY-MM-DD)"
// @Param end_date query string false "结束日期(YYYY-MM-DD)"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} resp.Response{data=dto.WalletTransactionListResponse} "成功"
// @Failure 404 {object} resp.Response "钱包未找到"
// @Router /v1/customers/{id}/wallet/transactions [get]
func (c *WalletController) ListTransactions(ctx *gin.Context) {
    customerIDStr := ctx.Param("id")
    customerID, err := strconv.ParseInt(customerIDStr, 10, 64)
    if err != nil {
        resp.Error(ctx, http.StatusBadRequest, "无效的客户ID")
        return
    }

    var req dto.WalletTransactionListRequest
    if err := ctx.ShouldBindQuery(&req); err != nil {
        resp.Error(ctx, http.StatusBadRequest, err.Error())
        return
    }

    result, err := c.walletSvc.ListTransactions(ctx.Request.Context(), customerID, &req)
    if err != nil {
        if errors.Is(err, service.ErrWalletNotFound) {
            resp.Error(ctx, http.StatusNotFound, err.Error())
            return
        }
        resp.Error(ctx, http.StatusInternalServerError, err.Error())
        return
    }

    resp.Success(ctx, result)
}
