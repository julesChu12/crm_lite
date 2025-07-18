// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"
)

const TableNameWalletTransaction = "wallet_transactions"

// WalletTransaction mapped from table <wallet_transactions>
type WalletTransaction struct {
	ID            int64     `gorm:"column:id;type:bigint(20);primaryKey;autoIncrement:true" json:"id"`
	WalletID      int64     `gorm:"column:wallet_id;type:bigint(20);not null;index:idx_wallet_transactions_wallet_id,priority:1" json:"wallet_id"`
	Type          string    `gorm:"column:type;type:varchar(20);not null;index:idx_wallet_transactions_type,priority:1;comment:交易类型: recharge, consume, refund, freeze, unfreeze, correction" json:"type"` // 交易类型: recharge, consume, refund, freeze, unfreeze, correction
	Amount        float64   `gorm:"column:amount;type:decimal(10,2);not null;comment:正数表示增加，负数表示减少" json:"amount"`                                                                                         // 正数表示增加，负数表示减少
	BalanceBefore float64   `gorm:"column:balance_before;type:decimal(10,2);not null;comment:交易前余额" json:"balance_before"`                                                                                 // 交易前余额
	BalanceAfter  float64   `gorm:"column:balance_after;type:decimal(10,2);not null;comment:交易后余额" json:"balance_after"`                                                                                   // 交易后余额
	Source        string    `gorm:"column:source;type:varchar(50);not null;index:idx_wallet_transactions_source,priority:1;comment:交易来源：manual, order, refund, system等" json:"source"`                     // 交易来源：manual, order, refund, system等
	RelatedID     int64     `gorm:"column:related_id;type:bigint(20);index:idx_wallet_transactions_related_id,priority:1;comment:关联ID（如订单ID、退款ID等）" json:"related_id"`                                     // 关联ID（如订单ID、退款ID等）
	Remark        string    `gorm:"column:remark;type:text" json:"remark"`
	OperatorID    int64     `gorm:"column:operator_id;type:bigint(20);index:idx_wallet_transactions_operator_id,priority:1;comment:操作人员" json:"operator_id"` // 操作人员
	CreatedAt     time.Time `gorm:"column:created_at;type:timestamp;index:idx_wallet_transactions_created_at,priority:1;default:current_timestamp()" json:"created_at"`
}

// TableName WalletTransaction's table name
func (*WalletTransaction) TableName() string {
	return TableNameWalletTransaction
}
