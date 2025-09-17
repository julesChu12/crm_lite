// Package common 通用组件包
// 提供跨域的通用功能，如事务管理、错误处理、审计等
package common

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Tx 事务接口
// 提供统一的事务管理能力，确保跨域操作的原子性
// 设计原则：
// 1. 支持嵌套事务：如果已在事务中，直接复用当前事务
// 2. 透明性：业务代码无需关心是否已在事务中
// 3. 资源安全：自动处理事务的开启、提交和回滚
type Tx interface {
	// InTx 检查当前上下文是否在事务中
	// 用于避免嵌套事务，提高性能
	InTx(ctx context.Context) bool

	// WithTx 在事务中执行函数
	// 如果已在事务中，直接执行fn；否则开启新事务
	// fn中如果返回error，会自动回滚事务
	// fn执行成功，会自动提交事务
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error

	// GetDB 获取数据库连接
	// 如果在事务中，返回事务连接；否则返回普通连接
	GetDB(ctx context.Context) *gorm.DB
}

// gormTx GORM事务实现
// 基于GORM提供具体的事务管理功能
type gormTx struct {
	db *gorm.DB // 数据库连接实例
}

// NewTx 创建GORM事务管理器
// db: GORM数据库连接实例
func NewTx(db *gorm.DB) Tx {
	return &gormTx{db: db}
}


// InTx 实现Tx接口
// 检查当前上下文是否包含事务连接
func (t *gormTx) InTx(ctx context.Context) bool {
	_, ok := ctx.Value(txKey{}).(*gorm.DB)
	return ok
}

// WithTx 实现Tx接口
// 核心事务管理逻辑：
// 1. 如果已在事务中，直接执行业务函数
// 2. 如果不在事务中，开启新事务并执行业务函数
// 3. 根据业务函数的返回值决定提交或回滚
func (t *gormTx) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	// 如果已经在事务中，直接执行
	if t.InTx(ctx) {
		return fn(ctx)
	}

	// 开启新事务并执行
	return t.db.Transaction(func(tx *gorm.DB) error {
		// 将事务连接放入上下文
		ctx2 := context.WithValue(ctx, txKey{}, tx)
		return fn(ctx2)
	})
}

// GetDB 实现Tx接口
// 根据当前上下文返回合适的数据库连接
func (t *gormTx) GetDB(ctx context.Context) *gorm.DB {
	if txDB, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		// 如果在事务中，返回事务连接
		return txDB
	}
	// 否则返回普通连接
	return t.db
}

// TxFunc 事务函数类型定义
// 用于简化事务函数的声明
type TxFunc func(ctx context.Context) error

// BusinessError 业务错误
// 用于区分业务逻辑错误和系统错误
// 业务错误不会导致事务回滚，系统错误会导致事务回滚
type BusinessError struct {
	Code    string `json:"code"`    // 错误代码
	Message string `json:"message"` // 错误信息
	Details string `json:"details"` // 错误详情
}

// Error 实现error接口
func (e *BusinessError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewBusinessError 创建业务错误
func NewBusinessError(code, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}

// NewBusinessErrorWithDetails 创建带详情的业务错误
func NewBusinessErrorWithDetails(code, message, details string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// IsBusinessError 判断是否为业务错误
func IsBusinessError(err error) bool {
	_, ok := err.(*BusinessError)
	return ok
}

// 常用业务错误代码定义
const (
	// 通用错误
	ErrCodeInvalidParam      = "INVALID_PARAM"      // 参数无效
	ErrCodeResourceNotFound  = "RESOURCE_NOT_FOUND" // 资源不存在
	ErrCodePermissionDenied  = "PERMISSION_DENIED"  // 权限不足
	ErrCodeDuplicateResource = "DUPLICATE_RESOURCE" // 资源重复

	// 订单相关错误
	ErrCodeOrderNotFound      = "ORDER_NOT_FOUND"      // 订单不存在
	ErrCodeOrderStatusInvalid = "ORDER_STATUS_INVALID" // 订单状态无效
	ErrCodeOrderCannotRefund  = "ORDER_CANNOT_REFUND"  // 订单不能退款

	// 产品相关错误
	ErrCodeProductNotFound    = "PRODUCT_NOT_FOUND"    // 产品不存在
	ErrCodeProductNotSellable = "PRODUCT_NOT_SELLABLE" // 产品不可售
	ErrCodeInsufficientStock  = "INSUFFICIENT_STOCK"   // 库存不足

	// 钱包相关错误
	ErrCodeWalletNotFound       = "WALLET_NOT_FOUND"      // 钱包不存在
	ErrCodeInsufficientBalance  = "INSUFFICIENT_BALANCE"  // 余额不足
	ErrCodeWalletFrozen         = "WALLET_FROZEN"         // 钱包已冻结
	ErrCodeDuplicateTransaction = "DUPLICATE_TRANSACTION" // 重复交易

	// 客户相关错误
	ErrCodeCustomerNotFound = "CUSTOMER_NOT_FOUND" // 客户不存在
	ErrCodePhoneDuplicate   = "PHONE_DUPLICATE"    // 手机号重复

	// 认证相关错误
	ErrCodeUnauthorized   = "UNAUTHORIZED"    // 未授权
	ErrCodeResourceExists = "RESOURCE_EXISTS" // 资源已存在
)
