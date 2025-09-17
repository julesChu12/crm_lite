package common

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// TransactionManager 事务管理器接口
type TransactionManager interface {
	// WithTx 在事务中执行操作
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error

	// WithTxIsolation 在指定隔离级别的事务中执行操作
	WithTxIsolation(ctx context.Context, isolation sql.IsolationLevel, fn func(ctx context.Context) error) error

	// WithTxTimeout 在有超时限制的事务中执行操作
	WithTxTimeout(ctx context.Context, timeout time.Duration, fn func(ctx context.Context) error) error

	// GetDB 从上下文获取数据库连接（事务或普通连接）
	GetDB(ctx context.Context) *gorm.DB
}

// transactionManager 事务管理器实现
type transactionManager struct {
	db *gorm.DB
}

// NewTransactionManager 创建新的事务管理器
func NewTransactionManager(db *gorm.DB) TransactionManager {
	return &transactionManager{db: db}
}

// txKey 事务上下文键
type txKey struct{}

// WithTx 在事务中执行操作
func (tm *transactionManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.WithTxIsolation(ctx, sql.LevelDefault, fn)
}

// WithTxIsolation 在指定隔离级别的事务中执行操作
func (tm *transactionManager) WithTxIsolation(ctx context.Context, isolation sql.IsolationLevel, fn func(ctx context.Context) error) error {
	// 检查是否已经在事务中
	if tx := tm.getTxFromContext(ctx); tx != nil {
		// 已经在事务中，直接执行
		return fn(ctx)
	}

	// 开始新事务
	var tx *gorm.DB
	var err error

	if isolation == sql.LevelDefault {
		tx = tm.db.Begin()
	} else {
		tx = tm.db.Begin(&sql.TxOptions{Isolation: isolation})
	}

	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// 将事务存储到上下文
	txCtx := context.WithValue(ctx, txKey{}, tx)

	// 设置事务完成标志
	completed := false
	defer func() {
		if !completed {
			// 如果事务未正常完成，执行回滚
			if rbErr := tx.Rollback().Error; rbErr != nil {
				// 记录回滚失败（在实际项目中应该记录到日志）
				fmt.Printf("Failed to rollback transaction: %v\n", rbErr)
			}
		}
	}()

	// 执行业务逻辑
	err = fn(txCtx)
	if err != nil {
		// 业务逻辑出错，回滚事务
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return fmt.Errorf("business error: %w, rollback error: %v", err, rbErr)
		}
		completed = true
		return err
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	completed = true
	return nil
}

// WithTxTimeout 在有超时限制的事务中执行操作
func (tm *transactionManager) WithTxTimeout(ctx context.Context, timeout time.Duration, fn func(ctx context.Context) error) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- tm.WithTx(timeoutCtx, fn)
	}()

	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		return fmt.Errorf("transaction timeout after %v", timeout)
	}
}

// GetDB 从上下文获取数据库连接
func (tm *transactionManager) GetDB(ctx context.Context) *gorm.DB {
	if tx := tm.getTxFromContext(ctx); tx != nil {
		return tx
	}
	return tm.db
}

// getTxFromContext 从上下文获取事务
func (tm *transactionManager) getTxFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return nil
}

// TransactionRetryOptions 事务重试选项
type TransactionRetryOptions struct {
	MaxRetries int           // 最大重试次数
	RetryDelay time.Duration // 重试间隔
	Backoff    float64       // 退避系数
}

// WithRetryableTx 可重试的事务执行
func (tm *transactionManager) WithRetryableTx(ctx context.Context, opts TransactionRetryOptions, fn func(ctx context.Context) error) error {
	var lastErr error
	delay := opts.RetryDelay

	for attempt := 0; attempt <= opts.MaxRetries; attempt++ {
		if attempt > 0 {
			// 等待重试
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// 指数退避
				delay = time.Duration(float64(delay) * opts.Backoff)
			}
		}

		err := tm.WithTx(ctx, fn)
		if err == nil {
			return nil // 成功
		}

		lastErr = err

		// 检查是否应该重试
		if !shouldRetryTransaction(err) {
			break
		}
	}

	return fmt.Errorf("transaction failed after %d attempts: %w", opts.MaxRetries+1, lastErr)
}

// shouldRetryTransaction 判断事务是否应该重试
func shouldRetryTransaction(err error) bool {
	// 这里可以根据具体的错误类型决定是否重试
	// 例如：死锁、连接失败等可以重试
	// 业务逻辑错误、约束冲突等不应该重试

	// 简单实现：只重试超时和连接相关的错误
	errStr := err.Error()
	return containsAny(errStr, []string{
		"timeout",
		"connection",
		"deadlock",
		"lock wait timeout",
	})
}

// containsAny 检查字符串是否包含任一子字符串
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				match := true
				for j, r := range substr {
					if s[i+j] != byte(r) {
						match = false
						break
					}
				}
				if match {
					return true
				}
			}
		}
	}
	return false
}

// NestedTransactionManager 支持嵌套事务的管理器
type NestedTransactionManager struct {
	*transactionManager
}

// NewNestedTransactionManager 创建支持嵌套事务的管理器
func NewNestedTransactionManager(db *gorm.DB) *NestedTransactionManager {
	return &NestedTransactionManager{
		transactionManager: &transactionManager{db: db},
	}
}

// WithNestedTx 嵌套事务执行（使用Savepoint）
func (ntm *NestedTransactionManager) WithNestedTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx := ntm.getTxFromContext(ctx)
	if tx == nil {
		// 不在事务中，开始新事务
		return ntm.WithTx(ctx, fn)
	}

	// 已经在事务中，创建保存点
	savepoint := fmt.Sprintf("sp_%d", time.Now().UnixNano())

	if err := tx.Exec("SAVEPOINT " + savepoint).Error; err != nil {
		return fmt.Errorf("failed to create savepoint: %w", err)
	}

	// 执行业务逻辑
	err := fn(ctx)
	if err != nil {
		// 回滚到保存点
		if rbErr := tx.Exec("ROLLBACK TO SAVEPOINT " + savepoint).Error; rbErr != nil {
			return fmt.Errorf("business error: %w, savepoint rollback error: %v", err, rbErr)
		}
		return err
	}

	// 释放保存点
	if err := tx.Exec("RELEASE SAVEPOINT " + savepoint).Error; err != nil {
		return fmt.Errorf("failed to release savepoint: %w", err)
	}

	return nil
}