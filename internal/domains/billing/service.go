// Package billing 钱包域服务接口
// 职责：钱包余额管理、交易记录、幂等性保障
// 核心原则：余额只读，一切变更通过交易流水实现
package billing

import "context"

// Service 钱包域服务接口
// 提供钱包相关的核心业务操作，确保资金安全和幂等性
type Service interface {
	// Credit 入账操作
	// 增加客户钱包余额，如充值、退款等场景
	// reason: 入账原因描述
	// idem: 幂等键，确保相同操作不会重复执行
	Credit(ctx context.Context, customerID int64, amount int64, reason, idem string) error

	// DebitForOrder 为订单扣款
	// 专用于订单支付场景的扣款操作
	// 必须关联具体订单ID，确保业务可追溯
	DebitForOrder(ctx context.Context, customerID, orderID int64, amount int64, idem string) error

	// CreditForRefund 订单退款入账
	// 专用于订单退款场景的入账操作
	// 必须关联原订单ID，确保退款可追溯
	CreditForRefund(ctx context.Context, customerID, orderID int64, amount int64, idem string) error

	// GetBalance 获取客户钱包余额
	// 只读操作，返回当前可用余额
	GetBalance(ctx context.Context, customerID int64) (int64, error)

	// GetTransactionHistory 获取交易历史
	// 分页查询客户的钱包交易记录
	GetTransactionHistory(ctx context.Context, customerID int64, page, pageSize int) ([]Transaction, error)

	// 控制器接口 - 兼容现有控制器
	GetWalletByCustomerID(ctx context.Context, customerID int64) (*WalletInfo, error)
	CreateTransaction(ctx context.Context, req *CreateTransactionRequest) (*Transaction, error)
	GetTransactions(ctx context.Context, customerID int64, req *TransactionHistoryRequest) ([]Transaction, int64, error)
}

// CreateTransactionRequest 创建交易请求
type CreateTransactionRequest struct {
	CustomerID int64   `json:"customer_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required"`
	Type       string  `json:"type" binding:"required"`
	Reason     string  `json:"reason"`
	OrderID    *int64  `json:"order_id,omitempty"`
	OperatorID int64   `json:"operator_id"`
}

// TransactionHistoryRequest 交易历史请求
type TransactionHistoryRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Type     string `json:"type,omitempty"`
}

// Transaction 钱包交易记录
// 记录所有钱包变更的详细信息，用于审计和对账
type Transaction struct {
	ID             int64  `json:"id"`              // 交易ID
	WalletID       int64  `json:"wallet_id"`       // 钱包ID
	Direction      string `json:"direction"`       // 方向：credit/debit
	Amount         int64  `json:"amount"`          // 金额（分）
	Type           string `json:"type"`            // 类型：recharge/order_pay/order_refund/adjust_in/adjust_out
	BizRefType     string `json:"biz_ref_type"`    // 业务引用类型：order/refund/manual
	BizRefID       int64  `json:"biz_ref_id"`      // 业务引用ID
	IdempotencyKey string `json:"idempotency_key"` // 幂等键
	OperatorID     int64  `json:"operator_id"`     // 操作员ID
	ReasonCode     string `json:"reason_code"`     // 原因代码
	Note           string `json:"note"`            // 备注
	CreatedAt      int64  `json:"created_at"`      // 创建时间（Unix时间戳）
}

// WalletInfo 钱包信息
// 包含钱包的基本状态信息
type WalletInfo struct {
	ID         int64 `json:"id"`          // 钱包ID
	CustomerID int64 `json:"customer_id"` // 客户ID
	Balance    int64 `json:"balance"`     // 当前余额（分）
	Status     int32 `json:"status"`      // 状态：1-正常，0-冻结
	UpdatedAt  int64 `json:"updated_at"`  // 最后更新时间
}

// Repository 钱包域数据访问接口
// 定义钱包相关的数据持久化操作
type Repository interface {
	// CreateWallet 创建钱包
	CreateWallet(ctx context.Context, customerID int64) (*WalletInfo, error)

	// GetWalletByCustomerID 根据客户ID获取钱包
	GetWalletByCustomerID(ctx context.Context, customerID int64) (*WalletInfo, error)

	// CreateTransaction 创建交易记录
	// 同时原子性更新钱包余额
	CreateTransaction(ctx context.Context, tx *Transaction) error

	// GetTransactionByIdempotencyKey 根据幂等键查找交易
	// 用于幂等性检查
	GetTransactionByIdempotencyKey(ctx context.Context, key string) (*Transaction, error)
}
