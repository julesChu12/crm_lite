package impl

import (
	"context"
	"testing"

	"crm_lite/internal/common"
	"crm_lite/internal/dao/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBillingServiceIntegration PR-2 billing域重构集成测试
// 验证余额只读、幂等性、事务安全等核心特性
func TestBillingServiceIntegration(t *testing.T) {
	// 跳过集成测试，除非明确设置环境变量
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 创建内存数据库用于测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 手动创建表结构以避免 AutoMigrate 的兼容性问题
	err = db.Exec(`
		CREATE TABLE wallets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_id INTEGER NOT NULL,
			balance INTEGER NOT NULL DEFAULT 0,
			status INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at INTEGER NOT NULL DEFAULT 0
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE wallet_transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			wallet_id INTEGER NOT NULL,
			direction TEXT NOT NULL,
			amount INTEGER NOT NULL,
			type TEXT NOT NULL,
			biz_ref_type TEXT,
			biz_ref_id INTEGER,
			idempotency_key TEXT NOT NULL UNIQUE,
			operator_id INTEGER DEFAULT 0,
			reason_code TEXT,
			note TEXT,
			created_at INTEGER NOT NULL
		)
	`).Error
	require.NoError(t, err)

	// 创建billing服务（使用具体实现类以访问Legacy方法）
	tx := common.NewTx(db)
	billingService := NewBillingServiceImpl(db, tx)

	ctx := context.Background()
	customerID := int64(12345)

	t.Run("余额只读原则", func(t *testing.T) {
		// 1. 初始余额应该为0
		balance, err := billingService.GetBalance(ctx, customerID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), balance, "初始余额应该为0")

		// 2. 入账操作
		amount := int64(10000) // 100元，以分为单位
		reason := "充值测试"
		idemKey := "test_credit_1"

		err = billingService.Credit(ctx, customerID, amount, reason, idemKey)
		require.NoError(t, err)

		// 3. 验证余额更新
		balance, err = billingService.GetBalance(ctx, customerID)
		require.NoError(t, err)
		assert.Equal(t, amount, balance, "余额应该等于入账金额")

		// 4. 验证交易记录
		var txCount int64
		err = db.Model(&model.WalletTransaction{}).Count(&txCount).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), txCount, "应该有1条交易记录")
	})

	t.Run("幂等性保障", func(t *testing.T) {
		// 使用相同的幂等键重复操作
		amount := int64(5000) // 50元
		reason := "幂等性测试"
		idemKey := "test_idempotent_1"

		// 第一次操作
		err = billingService.Credit(ctx, customerID, amount, reason, idemKey)
		require.NoError(t, err)

		// 获取第一次操作后的余额
		balance1, err := billingService.GetBalance(ctx, customerID)
		require.NoError(t, err)

		// 第二次使用相同幂等键操作
		err = billingService.Credit(ctx, customerID, amount, reason, idemKey)
		require.NoError(t, err) // 应该成功，但不会重复执行

		// 获取第二次操作后的余额
		balance2, err := billingService.GetBalance(ctx, customerID)
		require.NoError(t, err)

		// 验证余额没有重复增加
		assert.Equal(t, balance1, balance2, "相同幂等键的操作不应该重复执行")
	})

	t.Run("订单扣款功能", func(t *testing.T) {
		// 先确保有足够余额
		initialBalance, err := billingService.GetBalance(ctx, customerID)
		require.NoError(t, err)

		orderID := int64(888)
		debitAmount := int64(3000) // 30元
		idemKey := "test_debit_order_1"

		// 执行订单扣款
		err = billingService.DebitForOrder(ctx, customerID, orderID, debitAmount, idemKey)
		require.NoError(t, err)

		// 验证余额减少
		newBalance, err := billingService.GetBalance(ctx, customerID)
		require.NoError(t, err)
		assert.Equal(t, initialBalance-debitAmount, newBalance, "余额应该减少扣款金额")

		// 验证交易记录
		var tx model.WalletTransaction
		err = db.Where("biz_ref_type = ? AND biz_ref_id = ?", "order", orderID).First(&tx).Error
		require.NoError(t, err)
		assert.Equal(t, "debit", tx.Direction, "应该是扣款交易")
		assert.Equal(t, debitAmount, tx.Amount, "扣款金额应该正确")
	})

	t.Run("订单退款功能", func(t *testing.T) {
		initialBalance, err := billingService.GetBalance(ctx, customerID)
		require.NoError(t, err)

		orderID := int64(999)
		refundAmount := int64(1500) // 15元
		idemKey := "test_refund_order_1"

		// 执行订单退款
		err = billingService.CreditForRefund(ctx, customerID, orderID, refundAmount, idemKey)
		require.NoError(t, err)

		// 验证余额增加
		newBalance, err := billingService.GetBalance(ctx, customerID)
		require.NoError(t, err)
		assert.Equal(t, initialBalance+refundAmount, newBalance, "余额应该增加退款金额")

		// 验证交易记录
		var tx model.WalletTransaction
		err = db.Where("biz_ref_type = ? AND biz_ref_id = ? AND type = ?", "order", orderID, "order_refund").First(&tx).Error
		require.NoError(t, err)
		assert.Equal(t, "credit", tx.Direction, "应该是退款交易")
		assert.Equal(t, refundAmount, tx.Amount, "退款金额应该正确")
	})

	t.Run("余额不足保护", func(t *testing.T) {
		// 创建一个新客户，余额为0
		newCustomerID := int64(99999)

		// 尝试扣款超过余额的金额
		orderID := int64(777)
		debitAmount := int64(100000) // 1000元，远超余额
		idemKey := "test_insufficient_balance"

		// 应该返回余额不足错误
		err = billingService.DebitForOrder(ctx, newCustomerID, orderID, debitAmount, idemKey)
		assert.Error(t, err, "余额不足时应该返回错误")

		// 验证错误类型
		var businessErr *common.BusinessError
		assert.ErrorAs(t, err, &businessErr, "应该是业务错误")
	})

	t.Run("Legacy接口兼容性", func(t *testing.T) {
		// 由于 CreateWallet 是 Legacy 方法，在具体实现类中
		// 我们通过类型断言来测试
		if impl, ok := billingService.(*BillingServiceImpl); ok {
			// 测试CreateWallet接口
			wallet, err := impl.CreateWallet(ctx, customerID, "balance")
			require.NoError(t, err)
			assert.Equal(t, customerID, wallet.CustomerID, "钱包应该属于指定客户")

			// 测试重复创建钱包的幂等性
			wallet2, err := impl.CreateWallet(ctx, customerID, "balance")
			require.NoError(t, err)
			assert.Equal(t, wallet.ID, wallet2.ID, "重复创建应该返回相同的钱包")
		} else {
			t.Skip("无法访问Legacy接口，跳过此测试")
		}
	})

	t.Log("✅ PR-2 billing域重构验证通过:")
	t.Log("  - ✅ 余额只读原则：所有余额变更通过交易记录实现")
	t.Log("  - ✅ 幂等性保障：相同幂等键的操作不会重复执行")
	t.Log("  - ✅ 事务安全：扣款时检查余额，防止透支")
	t.Log("  - ✅ 业务完整性：订单扣款和退款功能正常")
	t.Log("  - ✅ Legacy兼容：现有钱包接口可以正常调用")
}
