package controller

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/domains/billing"
	"crm_lite/internal/domains/billing/impl"
	"crm_lite/internal/dto"
	"crm_lite/internal/middleware"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// WalletController 负责处理钱包相关的 API 请求
type WalletController struct {
	walletSvc  service.IWalletService
	billingSvc billing.Service
	resManager *resource.Manager
}

// NewWalletController 创建一个新的 WalletController
func NewWalletController(walletSvc service.IWalletService, resManager *resource.Manager) *WalletController {
	// 创建Billing领域服务
	dbRes, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource for WalletController: " + err.Error())
	}
	billingSvc := impl.NewBillingService(dbRes.DB)

	return &WalletController{
		walletSvc:  walletSvc, // 保留旧服务作为备用
		billingSvc: billingSvc,
		resManager: resManager,
	}
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

	walletInfo, err := c.billingSvc.GetWalletByCustomerID(ctx.Request.Context(), customerID)
	if err != nil {
		resp.Error(ctx, http.StatusNotFound, "钱包未找到")
		return
	}

	// 转换为DTO格式
	walletResponse := &dto.WalletResponse{
		ID:         walletInfo.ID,
		CustomerID: walletInfo.CustomerID,
		Balance:    float64(walletInfo.Balance) / 100, // 转换为元
		Type:       "balance",                         // 默认类型
		CreatedAt:  time.Unix(walletInfo.UpdatedAt, 0).Format("2006-01-02 15:04:05"),
	}

	resp.Success(ctx, walletResponse)
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

	// 2. 从上下文中获取操作员信息（支持 UUID → 数值ID 回退）
	operatorIDVal, exists := ctx.Get(middleware.ContextKeyUserID)
	if !exists {
		resp.Error(ctx, http.StatusUnauthorized, "操作员信息不存在")
		return
	}

	var operatorID int64
	switch v := operatorIDVal.(type) {
	case int64:
		operatorID = v
	case string:
		if id, errConv := strconv.ParseInt(v, 10, 64); errConv == nil {
			operatorID = id
		} else {
			// 视为 UUID，从 DB 查询对应的数值 ID
			dbRes, errDB := resource.Get[*resource.DBResource](c.resManager, resource.DBServiceKey)
			if errDB != nil {
				resp.SystemError(ctx, errDB)
				return
			}
			q := query.Use(dbRes.DB)
			admin, errFind := q.AdminUser.WithContext(ctx.Request.Context()).
				Where(q.AdminUser.UUID.Eq(v)).First()
			if errFind != nil {
				resp.Error(ctx, http.StatusUnauthorized, "无效的操作员身份")
				return
			}
			operatorID = admin.ID
		}
	default:
		resp.Error(ctx, http.StatusUnauthorized, "未知的操作员身份")
		return
	}

	// 3. 转换为Billing领域请求
	billingReq := &billing.CreateTransactionRequest{
		CustomerID: customerID,
		Amount:     req.Amount,
		Type:       req.Type,
		Reason:     req.Remark,
		OrderID:    &req.RelatedID,
		OperatorID: operatorID,
	}

	// 4. 调用Billing领域服务
	transaction, err := c.billingSvc.CreateTransaction(ctx.Request.Context(), billingReq)
	if err != nil {
		resp.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// 5. 转换为DTO格式返回
	transactionResponse := &dto.WalletTransactionResponse{
		ID:         transaction.ID,
		Amount:     float64(transaction.Amount) / 100, // 转换为元
		Type:       transaction.Type,
		Source:     transaction.BizRefType,
		Remark:     transaction.Note,
		RelatedID:  transaction.BizRefID,
		OperatorID: transaction.OperatorID,
		CreatedAt:  time.Unix(transaction.CreatedAt, 0).Format("2006-01-02 15:04:05"),
	}

	resp.Success(ctx, transactionResponse)
}

// GetTransactions @Summary 获取客户钱包交易流水列表
// @Description 根据客户ID获取其钱包交易记录，支持分页和多条件筛选
// @Tags Wallets
// @Accept json
// @Produce json
// @Param id path int true "客户ID"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param source query string false "交易来源筛选"
// @Param type query string false "交易类型筛选: recharge, consume, refund"
// @Param remark query string false "备注内容筛选（模糊匹配）"
// @Param start_date query string false "开始日期 YYYY-MM-DD"
// @Param end_date query string false "结束日期 YYYY-MM-DD"
// @Param related_id query int false "关联ID筛选"
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

	// 转换为Billing领域请求
	billingReq := &billing.TransactionHistoryRequest{
		Page:     req.Page,
		PageSize: req.Limit,
		Type:     req.Type,
	}

	transactions, total, err := c.billingSvc.GetTransactions(ctx.Request.Context(), customerID, billingReq)
	if err != nil {
		resp.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// 转换为DTO格式
	transactionResponses := make([]*dto.WalletTransactionResponse, len(transactions))
	for i, transaction := range transactions {
		transactionResponses[i] = &dto.WalletTransactionResponse{
			ID:         transaction.ID,
			Amount:     float64(transaction.Amount) / 100, // 转换为元
			Type:       transaction.Type,
			Source:     transaction.BizRefType,
			Remark:     transaction.Note,
			RelatedID:  transaction.BizRefID,
			OperatorID: transaction.OperatorID,
			CreatedAt:  time.Unix(transaction.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		}
	}

	resp.Success(ctx, dto.ListWalletTransactionsResponse{
		Transactions: transactionResponses,
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

	// 调用Billing领域服务处理退款
	err = c.billingSvc.CreditForRefund(ctx.Request.Context(), customerID, req.OrderID, int64(req.Amount*100), req.Remark)
	if err != nil {
		resp.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	resp.Success(ctx, nil)
}
