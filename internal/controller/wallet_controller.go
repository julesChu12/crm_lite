package controller

import (
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"
	"errors"
	"net/http"
	"strconv"

	"crm_lite/internal/dto"

	"github.com/gin-gonic/gin"
)

// WalletController 负责处理钱包相关的 API 请求
type WalletController struct {
	walletSvc *service.WalletService
}

// NewWalletController 创建一个新的 WalletController
func NewWalletController(walletSvc *service.WalletService) *WalletController {
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
// @Description 为客户的钱包创建一笔交易（充值或消费）
// @Tags Wallets
// @Accept json
// @Produce json
// @Param id path int true "客户ID"
// @Param transaction body dto.WalletTransactionRequest true "交易信息"
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
	operatorIDVal, _ := ctx.Get("user_id")
	operatorID := operatorIDVal.(int64)

	// 3. 调用服务
	err = c.walletSvc.CreateTransaction(ctx.Request.Context(), customerID, operatorID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCustomerNotFound), errors.Is(err, service.ErrWalletNotFound):
			resp.Error(ctx, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrInsufficientBalance):
			resp.Error(ctx, http.StatusUnprocessableEntity, err.Error())
		default:
			resp.Error(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	resp.Success(ctx, nil)
}
