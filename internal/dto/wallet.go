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
	Type      string  `json:"type" binding:"required,oneof=recharge consume"` // 交易类型: recharge (充值), consume (消费)
	Amount    float64 `json:"amount" binding:"required,gt=0"`                 // 交易金额，必须为正数
	Source    string  `json:"source" binding:"required"`                      // 交易来源: manual, system 等
	Remark    string  `json:"remark"`                                         // 备注
	RelatedID int64   `json:"related_id"`                                     // 关联ID（如订单ID等）
}
