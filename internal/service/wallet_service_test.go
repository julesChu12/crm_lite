package service

import (
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalletService_CreateTransaction_Recharge(t *testing.T) {
	// 跳过集成测试，除非设置了环境变量
	if os.Getenv("RUN_DB_TESTS") != "1" {
		t.Skip("Skipping integration test - set RUN_DB_TESTS=1 to run")
	}

	// 初始化服务
	walletSvc := NewWalletService(testResManager)
	dbResource, err := resource.Get[*resource.DBResource](testResManager, resource.DBServiceKey)
	require.NoError(t, err)
	q := query.Use(dbResource.DB)
	ctx := context.Background()

	// 清理测试数据
	defer func() {
		q.WalletTransaction.WithContext(ctx).Where(q.WalletTransaction.RelatedID.Eq(3990)).Delete()
		q.Wallet.WithContext(ctx).Where(q.Wallet.CustomerID.Eq(999)).Delete()
		q.Customer.WithContext(ctx).Where(q.Customer.ID.Eq(999)).Delete()
	}()

	// 1. 创建测试客户
	testCustomer := &model.Customer{
		ID:     999,
		Name:   "测试客户",
		Phone:  "13800138001",
		Gender: "unknown",
		Tags:   "[]",
	}
	err = q.Customer.WithContext(ctx).Create(testCustomer)
	require.NoError(t, err)

	// 2. 创建测试钱包
	testWallet, err := walletSvc.CreateWallet(ctx, 999, "balance")
	require.NoError(t, err)
	assert.Equal(t, int64(999), testWallet.CustomerID)
	assert.Equal(t, "balance", testWallet.Type)
	assert.Equal(t, 0.0, testWallet.Balance)

	// 3. 测试充值交易 - 使用用户提供的参数
	req := &dto.WalletTransactionRequest{
		Type:      "recharge",
		Amount:    300.0,
		Source:    "promotion:FULL_100_GET_20",
		Remark:    "会员充值奖励",
		RelatedID: 3990,
	}

	operatorID := int64(1) // 测试操作员ID
	err = walletSvc.CreateTransaction(ctx, 999, operatorID, req)
	require.NoError(t, err)

	// 4. 验证钱包余额更新
	updatedWallet, err := walletSvc.GetWalletByCustomerID(ctx, 999)
	require.NoError(t, err)
	assert.Equal(t, 300.0, updatedWallet.Balance)
	assert.Equal(t, 300.0, updatedWallet.TotalRecharged)
	assert.Equal(t, 0.0, updatedWallet.TotalConsumed)

	// 5. 验证交易记录
	transactions, total, err := walletSvc.GetTransactions(ctx, 999, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, transactions, 1)

	txn := transactions[0]
	assert.Equal(t, "recharge", txn.Type)
	assert.Equal(t, 300.0, txn.Amount)
	assert.Equal(t, 0.0, txn.BalanceBefore)
	assert.Equal(t, 300.0, txn.BalanceAfter)
	assert.Equal(t, "promotion:FULL_100_GET_20", txn.Source)
	assert.Equal(t, int64(3990), txn.RelatedID)
	assert.Equal(t, "会员充值奖励", txn.Remark)
	assert.Equal(t, operatorID, txn.OperatorID)
}

func TestWalletService_CreateTransaction_RechargeWithBonus(t *testing.T) {
	// 跳过集成测试，除非设置了环境变量
	if os.Getenv("RUN_DB_TESTS") != "1" {
		t.Skip("Skipping integration test - set RUN_DB_TESTS=1 to run")
	}

	// 初始化服务
	walletSvc := NewWalletService(testResManager)
	dbResource, err := resource.Get[*resource.DBResource](testResManager, resource.DBServiceKey)
	require.NoError(t, err)
	q := query.Use(dbResource.DB)
	ctx := context.Background()

	// 清理测试数据
	defer func() {
		q.WalletTransaction.WithContext(ctx).Where(q.WalletTransaction.RelatedID.Eq(3991)).Delete()
		q.Wallet.WithContext(ctx).Where(q.Wallet.CustomerID.Eq(998)).Delete()
		q.Customer.WithContext(ctx).Where(q.Customer.ID.Eq(998)).Delete()
	}()

	// 1. 创建测试客户
	testCustomer := &model.Customer{
		ID:     998,
		Name:   "测试客户2",
		Phone:  "13800138002",
		Gender: "unknown",
		Tags:   "[]",
	}
	err = q.Customer.WithContext(ctx).Create(testCustomer)
	require.NoError(t, err)

	// 2. 创建测试钱包
	_, err = walletSvc.CreateWallet(ctx, 998, "balance")
	require.NoError(t, err)

	// 3. 测试充值交易（含赠送金额）
	req := &dto.WalletTransactionRequest{
		Type:        "recharge",
		Amount:      300.0,
		Source:      "promotion:FULL_100_GET_20",
		Remark:      "会员充值奖励",
		RelatedID:   3991,
		BonusAmount: 50.0, // 赠送50元
	}

	operatorID := int64(1)
	err = walletSvc.CreateTransaction(ctx, 998, operatorID, req)
	require.NoError(t, err)

	// 4. 验证钱包余额更新（包含赠送金额）
	updatedWallet, err := walletSvc.GetWalletByCustomerID(ctx, 998)
	require.NoError(t, err)
	assert.Equal(t, 350.0, updatedWallet.Balance)        // 300 + 50
	assert.Equal(t, 300.0, updatedWallet.TotalRecharged) // 赠送金额不计入总充值
	assert.Equal(t, 0.0, updatedWallet.TotalConsumed)

	// 5. 验证交易记录（应该有两条：充值 + 赠送）
	transactions, total, err := walletSvc.GetTransactions(ctx, 998, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, transactions, 2)

	// 验证充值记录
	var rechargeTransaction, bonusTransaction *dto.WalletTransactionResponse
	for _, txn := range transactions {
		if txn.Type == "recharge" {
			rechargeTransaction = txn
		} else if txn.Type == "correction" {
			bonusTransaction = txn
		}
	}

	require.NotNil(t, rechargeTransaction)
	assert.Equal(t, 300.0, rechargeTransaction.Amount)
	assert.Equal(t, 0.0, rechargeTransaction.BalanceBefore)
	assert.Equal(t, 300.0, rechargeTransaction.BalanceAfter)

	require.NotNil(t, bonusTransaction)
	assert.Equal(t, 50.0, bonusTransaction.Amount)
	assert.Equal(t, 300.0, bonusTransaction.BalanceBefore)
	assert.Equal(t, 350.0, bonusTransaction.BalanceAfter)
	assert.Equal(t, "promotion", bonusTransaction.Source)
}

func TestWalletService_CreateTransaction_InvalidType(t *testing.T) {
	// 跳过集成测试，除非设置了环境变量
	if os.Getenv("RUN_DB_TESTS") != "1" {
		t.Skip("Skipping integration test - set RUN_DB_TESTS=1 to run")
	}

	walletSvc := NewWalletService(testResManager)
	ctx := context.Background()

	req := &dto.WalletTransactionRequest{
		Type:      "invalid_type",
		Amount:    100.0,
		Source:    "test",
		Remark:    "测试无效类型",
		RelatedID: 1000,
	}

	err := walletSvc.CreateTransaction(ctx, 999, 1, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported transaction type")
}

func TestWalletService_CreateTransaction_CustomerNotFound(t *testing.T) {
	// 跳过集成测试，除非设置了环境变量
	if os.Getenv("RUN_DB_TESTS") != "1" {
		t.Skip("Skipping integration test - set RUN_DB_TESTS=1 to run")
	}

	walletSvc := NewWalletService(testResManager)
	ctx := context.Background()

	req := &dto.WalletTransactionRequest{
		Type:      "recharge",
		Amount:    100.0,
		Source:    "test",
		Remark:    "测试客户不存在",
		RelatedID: 1001,
	}

	// 使用不存在的客户ID
	err := walletSvc.CreateTransaction(ctx, 99999, 1, req)
	assert.Error(t, err)
	assert.Equal(t, ErrCustomerNotFound, err)
}
