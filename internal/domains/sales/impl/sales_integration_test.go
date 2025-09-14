package impl

import (
	"context"
	"testing"

	"crm_lite/internal/common"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/domains/billing/impl"
	catalogImpl "crm_lite/internal/domains/catalog/impl"
	"crm_lite/internal/domains/sales"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSalesIntegration PR-3 Sales域统一事务收口集成测试
// 验证订单下单、快照、钱包扣减、退款等完整流程
func TestSalesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping sales integration test in short mode")
	}

	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 手动创建测试表结构
	err = db.Exec(`
		CREATE TABLE customers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			phone TEXT UNIQUE,
			email TEXT,
			gender TEXT DEFAULT 'unknown',
			birthday DATETIME,
			level TEXT DEFAULT 'normal',
			tags TEXT,
			note TEXT,
			source TEXT DEFAULT 'manual',
			assigned_to INTEGER DEFAULT 0,
			deleted_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			type TEXT DEFAULT 'product',
			category TEXT,
			price REAL NOT NULL DEFAULT 0,
			cost REAL DEFAULT 0,
			stock_quantity INTEGER DEFAULT 0,
			min_stock_level INTEGER DEFAULT 0,
			unit TEXT DEFAULT '个',
			is_active INTEGER DEFAULT 1,
			deleted_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_no TEXT UNIQUE NOT NULL,
			customer_id INTEGER NOT NULL,
			order_date DATETIME,
			status TEXT DEFAULT 'pending',
			total_amount REAL DEFAULT 0,
			final_amount REAL DEFAULT 0,
			remark TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE order_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			product_name TEXT NOT NULL,
			product_name_snapshot TEXT,
			quantity INTEGER NOT NULL,
			unit_price REAL NOT NULL,
			unit_price_snapshot INTEGER,
			duration_min_snapshot INTEGER DEFAULT 0,
			final_price REAL NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	require.NoError(t, err)

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

	err = db.Exec(`
		CREATE TABLE sys_outbox (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			event_type TEXT NOT NULL,
			payload TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			processed_at INTEGER DEFAULT 0
		)
	`).Error
	require.NoError(t, err)

	// 初始化测试数据
	q := query.Use(db)

	// 创建测试客户
	customer := &model.Customer{
		Name:  "测试客户",
		Phone: "13800138000",
		Email: "test@example.com",
		Level: "gold",
	}
	err = q.Customer.WithContext(context.Background()).Create(customer)
	require.NoError(t, err)

	// 创建测试产品
	product := &model.Product{
		Name:          "测试服务",
		Price:         100.0, // 100元
		Category:      "service001",
		StockQuantity: 10,
		IsActive:      true,
	}
	err = q.Product.WithContext(context.Background()).Create(product)
	require.NoError(t, err)

	// 创建服务实例
	tx := common.NewTx(db)
	catalogSvc := catalogImpl.New(q)
	billingSvc := impl.NewBillingServiceImpl(db, tx)
	outboxSvc := common.NewOutboxService(db, tx)

	salesSvc := NewSalesServiceImpl(db, tx, catalogSvc, billingSvc, outboxSvc)

	ctx := context.Background()

	t.Run("完整订单流程-现金支付", func(t *testing.T) {
		// 构建下单请求
		placeReq := sales.PlaceOrderReq{
			CustomerID: customer.ID,
			Channel:    "web",
			PayMethod:  "cash",
			Items: []sales.OrderItemReq{
				{
					ProductID: product.ID,
					Qty:       2,
				},
			},
			Discount:   1000, // 10元折扣
			IdemKey:    "test_cash_order_001",
			Remark:     "现金支付测试订单",
			AssignedTo: 1001,
		}

		// 执行下单
		order, err := salesSvc.PlaceOrder(ctx, placeReq)
		require.NoError(t, err)

		// 验证订单基本信息
		assert.NotZero(t, order.ID, "订单ID应该被赋值")
		assert.NotEmpty(t, order.OrderNo, "订单号应该被生成")
		assert.Equal(t, customer.ID, order.CustomerID, "客户ID应该正确")
		assert.Equal(t, int64(20000), order.TotalAmount, "订单总金额应该是200元(20000分)")
		assert.Equal(t, int64(1000), order.DiscountAmount, "折扣金额应该是10元(1000分)")
		assert.Equal(t, int64(19000), order.FinalAmount, "最终金额应该是190元(19000分)")
		assert.Equal(t, "pending", order.Status, "现金支付订单状态应该是pending")
		assert.Equal(t, "cash", order.PayMethod, "支付方式应该是现金")

		// 验证订单项包含产品快照
		_, orderItems, err := salesSvc.GetOrder(ctx, order.ID)
		require.NoError(t, err)
		require.Len(t, orderItems, 1, "应该有1个订单项")

		item := orderItems[0]
		assert.Equal(t, product.ID, item.ProductID, "产品ID应该正确")
		assert.Equal(t, "测试服务", item.ProductNameSnapshot, "产品名称快照应该保存")
		assert.Equal(t, int64(10000), item.UnitPriceSnapshot, "单价快照应该是100元(10000分)")
		assert.Equal(t, int32(0), item.DurationMinSnapshot, "服务时长快照应该保存")
		assert.Equal(t, int32(2), item.Quantity, "数量应该正确")

		// 验证Outbox事件
		var outboxCount int64
		err = db.Model(&model.SysOutbox{}).Where("event_type = ?", common.EventTypeOrderPlaced).Count(&outboxCount).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), outboxCount, "应该有1个订单下单事件")

		t.Log("✅ 现金支付订单流程验证通过")
	})

	t.Run("完整订单流程-钱包支付", func(t *testing.T) {
		// 先给客户充值
		err := billingSvc.Credit(ctx, customer.ID, 50000, "测试充值", "test_recharge_001")
		require.NoError(t, err)

		// 构建下单请求
		placeReq := sales.PlaceOrderReq{
			CustomerID: customer.ID,
			Channel:    "mobile",
			PayMethod:  "wallet",
			Items: []sales.OrderItemReq{
				{
					ProductID: product.ID,
					Qty:       1,
				},
			},
			Discount:   0, // 无折扣
			IdemKey:    "test_wallet_order_001",
			Remark:     "钱包支付测试订单",
			AssignedTo: 1002,
		}

		// 获取下单前的钱包余额
		balanceBefore, err := billingSvc.GetBalance(ctx, customer.ID)
		require.NoError(t, err)
		t.Logf("下单前钱包余额: %d分", balanceBefore)

		// 执行下单
		order, err := salesSvc.PlaceOrder(ctx, placeReq)
		require.NoError(t, err)

		// 验证订单基本信息
		assert.NotZero(t, order.ID, "订单ID应该被赋值")
		assert.Equal(t, int64(10000), order.TotalAmount, "订单总金额应该是100元(10000分)")
		assert.Equal(t, int64(0), order.DiscountAmount, "折扣金额应该是0")
		assert.Equal(t, int64(10000), order.FinalAmount, "最终金额应该是100元(10000分)")
		assert.Equal(t, "paid", order.Status, "钱包支付订单状态应该是paid")
		assert.Equal(t, "wallet", order.PayMethod, "支付方式应该是钱包")

		// 验证钱包扣款
		balanceAfter, err := billingSvc.GetBalance(ctx, customer.ID)
		require.NoError(t, err)
		t.Logf("下单后钱包余额: %d分", balanceAfter)
		assert.Equal(t, balanceBefore-10000, balanceAfter, "钱包余额应该减少100元")

		// 验证钱包交易记录
		var txCount int64
		err = db.Model(&model.WalletTransaction{}).
			Where("biz_ref_type = ? AND biz_ref_id = ? AND direction = ?", "order", order.ID, "debit").
			Count(&txCount).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), txCount, "应该有1条钱包扣款交易记录")

		// 验证Outbox事件（下单+支付）
		var orderPlacedCount, orderPaidCount int64
		err = db.Model(&model.SysOutbox{}).Where("event_type = ?", common.EventTypeOrderPlaced).Count(&orderPlacedCount).Error
		require.NoError(t, err)
		err = db.Model(&model.SysOutbox{}).Where("event_type = ?", common.EventTypeOrderPaid).Count(&orderPaidCount).Error
		require.NoError(t, err)
		assert.GreaterOrEqual(t, orderPlacedCount, int64(2), "应该有订单下单事件")
		assert.GreaterOrEqual(t, orderPaidCount, int64(1), "应该有订单支付事件")

		t.Log("✅ 钱包支付订单流程验证通过")

		// 测试订单退款
		t.Run("订单退款流程", func(t *testing.T) {
			// 获取退款前的钱包余额
			balanceBeforeRefund, err := billingSvc.GetBalance(ctx, customer.ID)
			require.NoError(t, err)

			// 执行退款
			err = salesSvc.RefundOrder(ctx, order.ID, "测试退款")
			require.NoError(t, err)

			// 验证订单状态更新为退款
			refundedOrder, _, err := salesSvc.GetOrder(ctx, order.ID)
			require.NoError(t, err)
			assert.Equal(t, "refunded", refundedOrder.Status, "订单状态应该更新为退款")

			// 验证钱包退款
			balanceAfterRefund, err := billingSvc.GetBalance(ctx, customer.ID)
			require.NoError(t, err)
			assert.Equal(t, balanceBeforeRefund+10000, balanceAfterRefund, "钱包余额应该增加退款金额")

			// 验证退款交易记录
			var refundTxCount int64
			err = db.Model(&model.WalletTransaction{}).
				Where("biz_ref_type = ? AND biz_ref_id = ? AND direction = ? AND type = ?",
					"order", order.ID, "credit", "order_refund").
				Count(&refundTxCount).Error
			require.NoError(t, err)
			assert.Equal(t, int64(1), refundTxCount, "应该有1条退款交易记录")

			// 验证退款Outbox事件
			var refundEventCount int64
			err = db.Model(&model.SysOutbox{}).Where("event_type = ?", common.EventTypeOrderRefunded).Count(&refundEventCount).Error
			require.NoError(t, err)
			assert.GreaterOrEqual(t, refundEventCount, int64(1), "应该有退款事件")

			t.Log("✅ 订单退款流程验证通过")
		})
	})

	t.Run("产品快照完整性验证", func(t *testing.T) {
		// 创建另一个产品用于测试快照
		product2 := &model.Product{
			Name:          "高级服务",
			Price:         200.0, // 200元
			Category:      "service002",
			StockQuantity: 5,
			IsActive:      true,
		}
		err = q.Product.WithContext(context.Background()).Create(product2)
		require.NoError(t, err)

		placeReq := sales.PlaceOrderReq{
			CustomerID: customer.ID,
			Channel:    "admin",
			PayMethod:  "cash",
			Items: []sales.OrderItemReq{
				{
					ProductID: product2.ID,
					Qty:       1,
				},
			},
			IdemKey: "test_snapshot_order_001",
			Remark:  "产品快照测试",
		}

		// 执行下单
		order, err := salesSvc.PlaceOrder(ctx, placeReq)
		require.NoError(t, err)

		// 获取订单项详情
		_, orderItems, err := salesSvc.GetOrder(ctx, order.ID)
		require.NoError(t, err)
		require.Len(t, orderItems, 1)

		item := orderItems[0]
		assert.Equal(t, "高级服务", item.ProductNameSnapshot, "产品名称快照应该正确保存")
		assert.Equal(t, int64(20000), item.UnitPriceSnapshot, "单价快照应该正确保存(200元=20000分)")

		// 模拟产品信息变更（在真实场景中产品价格可能会变化）
		// 但订单中的快照信息应该保持不变
		_, err = q.Product.WithContext(context.Background()).
			Where(q.Product.ID.Eq(product2.ID)).
			Update(q.Product.Price, 300.0)
		require.NoError(t, err)

		// 重新获取订单项，快照信息应该保持不变
		_, orderItemsAfter, err := salesSvc.GetOrder(ctx, order.ID)
		require.NoError(t, err)
		require.Len(t, orderItemsAfter, 1)

		itemAfter := orderItemsAfter[0]
		assert.Equal(t, "高级服务", itemAfter.ProductNameSnapshot, "快照不应该受产品变更影响")
		assert.Equal(t, int64(20000), itemAfter.UnitPriceSnapshot, "价格快照不应该受产品变更影响")

		t.Log("✅ 产品快照完整性验证通过")
	})

	t.Log("🎉 PR-3 Sales域统一事务收口验证完成:")
	t.Log("  - ✅ 统一下单事务收口：产品快照+订单创建+钱包扣减+outbox事件")
	t.Log("  - ✅ 统一退款事务收口：状态更新+钱包退款+outbox事件")
	t.Log("  - ✅ 产品快照功能：保存产品名称、价格、时长等关键信息")
	t.Log("  - ✅ 钱包域集成：无缝对接billing域进行扣减和退款")
	t.Log("  - ✅ 事件驱动架构：完整的outbox事件记录")
	t.Log("  - ✅ 业务完整性：订单状态与钱包操作保持一致")
}
