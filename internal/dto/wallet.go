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
	Type        string  `json:"type" binding:"required,oneof=recharge consume refund"` // 交易类型: recharge (充值), consume (消费), refund (退款)
	Amount      float64 `json:"amount" binding:"required,gt=0"`                        // 交易金额，必须为正数
	Source      string  `json:"source" binding:"required"`                             // 交易来源: manual, order, refund, system 等
	Remark      string  `json:"remark"`                                                // 备注
	RelatedID   int64   `json:"related_id"`                                            // 关联ID（如订单ID等）
	PhoneLast   string  `json:"phone_last"`                                            // 手机号后四位
	BonusAmount float64 `json:"bonus_amount"`                                          // 赠送金额
}

type WalletTransactionResponse struct {
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

type ListWalletTransactionsRequest struct {
	Page  int `form:"page" binding:"min=1"`
	Limit int `form:"limit" binding:"min=1,max=100"`
}

type ListWalletTransactionsResponse struct {
	Transactions []*WalletTransactionResponse `json:"transactions"`
	Total        int64                        `json:"total"`
}

// WalletRefundRequest 退款请求
type WalletRefundRequest struct {
	Amount  float64 `json:"amount" binding:"required,gt=0"` // 退款金额，必须为正数
	OrderID int64   `json:"order_id" binding:"required"`    // 关联订单ID
	Reason  string  `json:"reason" binding:"required"`      // 退款原因
	Remark  string  `json:"remark"`                         // 备注
}
