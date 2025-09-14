package impl

import "time"

// BilWallet 映射 bil_wallets（余额只读 by contract）
type BilWallet struct {
	ID         int64 `gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID int64 `gorm:"column:customer_id;uniqueIndex"`
	Balance    int64 `gorm:"column:balance;not null;default:0"` // cents
	Status     int8  `gorm:"column:status;not null;default:1"`
	UpdatedAt  int64 `gorm:"column:updated_at;not null"`
}

func (BilWallet) TableName() string { return "bil_wallets" }

// BilWalletTx 映射 bil_wallet_transactions（真相表）
type BilWalletTx struct {
	ID             int64  `gorm:"column:id;primaryKey;autoIncrement"`
	WalletID       int64  `gorm:"column:wallet_id;index:idx_wallet_time"`
	Direction      string `gorm:"column:direction;type:enum('credit','debit');not null"`
	Amount         int64  `gorm:"column:amount;not null"` // cents, positive
	Type           string `gorm:"column:type;type:enum('recharge','order_pay','order_refund','adjust_in','adjust_out');not null"`
	BizRefType     string `gorm:"column:biz_ref_type"`
	BizRefID       int64  `gorm:"column:biz_ref_id"`
	IdempotencyKey string `gorm:"column:idempotency_key;uniqueIndex:uk_idem;size:64;not null"`
	OperatorID     int64  `gorm:"column:operator_id"`
	ReasonCode     string `gorm:"column:reason_code"`
	Note           string `gorm:"column:note;size:255"`
	CreatedAt      int64  `gorm:"column:created_at;not null"`
	// shadow fields for convenience
	CreatedAtTime time.Time `gorm:"-"`
}

func (BilWalletTx) TableName() string { return "bil_wallet_transactions" }
