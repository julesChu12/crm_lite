package impl

import (
	"context"

	"crm_lite/internal/dao/model"
	"crm_lite/internal/domains/billing"
)

// billingAdapter 适配器，将 billing 域的接口适配为 CRM 域所需的 WalletPort
type billingAdapter struct {
	billingSvc billing.Service
}

// newBillingAdapter 创建 billing 适配器
func newBillingAdapter(billingSvc billing.Service) WalletPort {
	return &billingAdapter{
		billingSvc: billingSvc,
	}
}

// CreateWallet 创建钱包
// 实现 WalletPort 接口，适配 billing 域的钱包创建功能
func (a *billingAdapter) CreateWallet(ctx context.Context, customerID int64, walletType string) (*model.Wallet, error) {
	// 由于 billing.Service 接口没有直接的创建钱包方法，
	// 我们通过获取余额来触发钱包的自动创建
	_, err := a.billingSvc.GetBalance(ctx, customerID)
	if err != nil {
		return nil, err
	}

	// 返回一个基本的钱包信息，实际的钱包管理由 billing 域负责
	// 这里主要是为了兼容现有的接口
	return &model.Wallet{
		CustomerID: customerID,
		Balance:    0, // 余额由 billing 域管理
		Status:     1, // 1-正常状态
	}, nil
}

// GetWalletByCustomerID 获取客户钱包余额
// 实现 WalletPort 接口，通过 billing 域获取客户余额
func (a *billingAdapter) GetWalletByCustomerID(ctx context.Context, customerID int64) (balance int64, err error) {
	return a.billingSvc.GetBalance(ctx, customerID)
}
