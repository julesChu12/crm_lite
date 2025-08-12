package dto

import "time"

// WalletResponse 钱包信息的响应
type WalletResponse struct {
	ID             int64     `json:"id"`
	CustomerID     int64     `json:"customer_id"`
	Type           string    `json:"type"`
	Balance        float64   `json:"balance"`
	FrozenBalance  float64   `json:"frozen_balance"`
	TotalRecharged float64   `json:"total_recharged"`
	TotalConsumed  float64   `json:"total_consumed"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// WalletTransactionRequest 创建钱包交易的请求
type WalletTransactionRequest struct {
    Type        string  `json:"type" binding:"required,oneof=recharge consume refund"` // 交易类型: recharge(充值) / consume(消费) / refund(退款)
    Amount      float64 `json:"amount" binding:"required,gt=0"`                        // 金额为正数，方向由类型决定
    Source      string  `json:"source" binding:"required"`                             // 交易来源: manual_recharge, manual_consume, promotion 等
    Remark      string  `json:"remark"`                                                // 备注
    RelatedID   int64   `json:"related_id"`                                            // 关联ID（如订单ID等）
    BonusAmount float64 `json:"bonus_amount"`                                          // 充值满赠的赠送金额（可选）；>0 时将追加一条 correction 流水
    PhoneLast4  string  `json:"phone_last4"`                                          // 消费校验：客户手机号后四位（consume 时必填，用于现场核对）
}

// WalletTransactionItem 钱包流水项
type WalletTransactionItem struct {
    ID            int64     `json:"id"`
    WalletID      int64     `json:"wallet_id"`
    Type          string    `json:"type"`
    Amount        float64   `json:"amount"`
    BalanceBefore float64   `json:"balance_before"`
    BalanceAfter  float64   `json:"balance_after"`
    Source        string    `json:"source"`
    RelatedID     int64     `json:"related_id"`
    Remark        string    `json:"remark"`
    OperatorID    int64     `json:"operator_id"`
    CreatedAt     time.Time `json:"created_at"`
}

// WalletTransactionListRequest 查询钱包流水的请求
type WalletTransactionListRequest struct {
    WalletType string `form:"wallet_type"`              // 预留多钱包类型，默认 balance
    Type       string `form:"type"`                     // 交易类型筛选
    Source     string `form:"source"`                   // 来源筛选
    StartDate  string `form:"start_date"`               // YYYY-MM-DD
    EndDate    string `form:"end_date"`                 // YYYY-MM-DD
    Page       int    `form:"page"`                     // 页码
    PageSize   int    `form:"page_size"`                // 每页数量
}

// WalletTransactionListResponse 钱包流水查询响应
type WalletTransactionListResponse struct {
    Total        int64                     `json:"total"`
    Page         int                       `json:"page"`
    PageSize     int                       `json:"page_size"`
    Transactions []*WalletTransactionItem  `json:"transactions"`
}
