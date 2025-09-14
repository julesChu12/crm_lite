// Package common 通用组件包 - Outbox事件发布
// Outbox模式实现：确保业务操作和事件发布的原子性
package common

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// OutboxEvent Outbox事件模型
// 用于在业务事务中记录需要发布的事件，实现最终一致性
type OutboxEvent struct {
	ID          int64           `json:"id"`           // 事件ID
	EventType   string          `json:"event_type"`   // 事件类型
	Payload     json.RawMessage `json:"payload"`      // 事件载荷（JSON格式）
	CreatedAt   int64           `json:"created_at"`   // 创建时间（Unix时间戳）
	ProcessedAt *int64          `json:"processed_at"` // 处理时间（Unix时间戳）
}

// OutboxService Outbox事件服务接口
// 提供事件记录和处理的核心功能
type OutboxService interface {
	// PublishEvent 发布事件到Outbox
	// 在业务事务中调用，确保事件记录和业务操作的原子性
	PublishEvent(ctx context.Context, eventType string, payload interface{}) error

	// ProcessPendingEvents 处理待发布事件
	// 定期调用，将未处理的事件发送到消息队列或其他系统
	ProcessPendingEvents(ctx context.Context, limit int) error

	// MarkEventProcessed 标记事件为已处理
	// 在事件成功发布后调用，避免重复处理
	MarkEventProcessed(ctx context.Context, eventID int64) error
}

// OutboxRepository Outbox事件存储接口
// 定义事件的持久化操作
type OutboxRepository interface {
	// CreateEvent 创建事件记录
	CreateEvent(ctx context.Context, event *OutboxEvent) error

	// GetPendingEvents 获取未处理的事件
	GetPendingEvents(ctx context.Context, limit int) ([]*OutboxEvent, error)

	// UpdateEventProcessed 更新事件为已处理状态
	UpdateEventProcessed(ctx context.Context, eventID int64, processedAt int64) error
}

// 预定义的事件类型常量
// 用于标准化系统中的事件类型，便于统一处理
const (
	// 订单相关事件
	EventTypeOrderPlaced    = "order.placed"    // 订单已下单
	EventTypeOrderPaid      = "order.paid"      // 订单已支付
	EventTypeOrderRefunded  = "order.refunded"  // 订单已退款
	EventTypeOrderCancelled = "order.cancelled" // 订单已取消

	// 钱包相关事件
	EventTypeWalletCredited = "wallet.credited" // 钱包入账
	EventTypeWalletDebited  = "wallet.debited"  // 钱包出账

	// 客户相关事件
	EventTypeCustomerCreated = "customer.created" // 客户已创建
	EventTypeCustomerUpdated = "customer.updated" // 客户已更新
)

// OrderPlacedEvent 订单下单事件载荷
type OrderPlacedEvent struct {
	OrderID     int64  `json:"order_id"`
	OrderNo     string `json:"order_no"`
	CustomerID  int64  `json:"customer_id"`
	TotalAmount int64  `json:"total_amount"`
	PayMethod   string `json:"pay_method"`
	CreatedAt   int64  `json:"created_at"`
}

// OrderPaidEvent 订单支付事件载荷
type OrderPaidEvent struct {
	OrderID    int64  `json:"order_id"`
	OrderNo    string `json:"order_no"`
	CustomerID int64  `json:"customer_id"`
	PaidAmount int64  `json:"paid_amount"`
	PayMethod  string `json:"pay_method"`
	PaidAt     int64  `json:"paid_at"`
}

// OrderRefundedEvent 订单退款事件载荷
type OrderRefundedEvent struct {
	OrderID      int64  `json:"order_id"`
	OrderNo      string `json:"order_no"`
	CustomerID   int64  `json:"customer_id"`
	RefundAmount int64  `json:"refund_amount"`
	Reason       string `json:"reason"`
	RefundedAt   int64  `json:"refunded_at"`
}

// WalletCreditedEvent 钱包入账事件载荷
type WalletCreditedEvent struct {
	CustomerID      int64  `json:"customer_id"`
	Amount          int64  `json:"amount"`
	TransactionType string `json:"transaction_type"`
	BizRefType      string `json:"biz_ref_type"`
	BizRefID        int64  `json:"biz_ref_id"`
	CreditedAt      int64  `json:"credited_at"`
}

// WalletDebitedEvent 钱包出账事件载荷
type WalletDebitedEvent struct {
	CustomerID      int64  `json:"customer_id"`
	Amount          int64  `json:"amount"`
	TransactionType string `json:"transaction_type"`
	BizRefType      string `json:"biz_ref_type"`
	BizRefID        int64  `json:"biz_ref_id"`
	DebitedAt       int64  `json:"debited_at"`
}

// CustomerCreatedEvent 客户创建事件载荷
type CustomerCreatedEvent struct {
	CustomerID int64  `json:"customer_id"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Level      string `json:"level"`
	Source     string `json:"source"`
	AssignedTo int64  `json:"assigned_to"`
	CreatedAt  int64  `json:"created_at"`
}

// NewOutboxEvent 创建新的Outbox事件
func NewOutboxEvent(eventType string, payload interface{}) (*OutboxEvent, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &OutboxEvent{
		EventType: eventType,
		Payload:   payloadBytes,
		CreatedAt: time.Now().Unix(),
	}, nil
}

// OutboxServiceImpl Outbox事件服务实现
// 基于数据库的事件存储实现，确保事务一致性
type OutboxServiceImpl struct {
	db *gorm.DB
	tx Tx
}

// NewOutboxService 创建Outbox事件服务
func NewOutboxService(db *gorm.DB, tx Tx) OutboxService {
	return &OutboxServiceImpl{
		db: db,
		tx: tx,
	}
}

// PublishEvent 发布事件到Outbox
func (o *OutboxServiceImpl) PublishEvent(ctx context.Context, eventType string, payload interface{}) error {
	event, err := NewOutboxEvent(eventType, payload)
	if err != nil {
		return fmt.Errorf("创建outbox事件失败: %w", err)
	}

	// 在当前事务中插入事件记录
	txDB := o.tx.GetDB(ctx)
	result := txDB.WithContext(ctx).Exec(`
		INSERT INTO sys_outbox (event_type, payload, created_at)
		VALUES (?, ?, ?)
	`, event.EventType, event.Payload, event.CreatedAt)

	if result.Error != nil {
		return fmt.Errorf("插入outbox事件失败: %w", result.Error)
	}

	return nil
}

// ProcessPendingEvents 处理待发布事件
func (o *OutboxServiceImpl) ProcessPendingEvents(ctx context.Context, limit int) error {
	// 查询未处理的事件
	var events []*OutboxEvent
	err := o.db.WithContext(ctx).
		Raw(`SELECT * FROM sys_outbox WHERE processed_at IS NULL ORDER BY created_at LIMIT ?`, limit).
		Scan(&events).Error

	if err != nil {
		return fmt.Errorf("查询待处理事件失败: %w", err)
	}

	// 处理每个事件
	for _, event := range events {
		// 这里可以发送到消息队列或其他系统
		// 为了演示，我们只是标记为已处理
		if err := o.MarkEventProcessed(ctx, event.ID); err != nil {
			// 日志记录但继续处理其他事件
			continue
		}
	}

	return nil
}

// MarkEventProcessed 标记事件为已处理
func (o *OutboxServiceImpl) MarkEventProcessed(ctx context.Context, eventID int64) error {
	processedAt := time.Now().Unix()
	result := o.db.WithContext(ctx).Exec(`
		UPDATE sys_outbox SET processed_at = ? WHERE id = ?
	`, processedAt, eventID)

	if result.Error != nil {
		return fmt.Errorf("标记事件已处理失败: %w", result.Error)
	}

	return nil
}

// OutboxWriter Outbox事件写入器
// 简单的Outbox写入实现，用于在事务中记录事件
type OutboxWriter struct {
	// 暂时空实现，后续根据需要扩展
}

// NewOutboxWriter 创建Outbox写入器
// 临时实现，用于支持现有代码编译通过
func NewOutboxWriter(db interface{}) *OutboxWriter {
	return &OutboxWriter{}
}

// WriteOutbox 写入Outbox事件
// 临时实现，后续在PR-3中完善
func (w *OutboxWriter) WriteOutbox(ctx context.Context, eventType string, payload map[string]any) error {
	// TODO: 在PR-3中实现具体的事件写入逻辑
	return nil
}
