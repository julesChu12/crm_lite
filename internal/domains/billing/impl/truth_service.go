package impl

import (
	"context"
	"errors"
	"time"

	"crm_lite/internal/core/resource"
	"crm_lite/internal/domains/billing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TruthService 为后续替换 adapter 的真相实现骨架（暂不接入控制器）
type TruthService struct {
	db *gorm.DB
}

func NewTruthService(rm *resource.Manager) (*TruthService, error) {
	dbRes, err := resource.Get[*resource.DBResource](rm, resource.DBServiceKey)
	if err != nil {
		return nil, err
	}
	return &TruthService{db: dbRes.DB}, nil
}

func (s *TruthService) Credit(ctx context.Context, customerID int64, amount int64, reason, idem string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		wallet, err := s.lockWallet(tx, customerID)
		if err != nil {
			return err
		}
		if err := s.ensureIdem(tx, idem); err != nil {
			return err
		}
		now := time.Now().Unix()
		rec := &BilWalletTx{WalletID: wallet.ID, Direction: "credit", Amount: amount, Type: "recharge", IdempotencyKey: idem, ReasonCode: reason, CreatedAt: now}
		if err := tx.Create(rec).Error; err != nil {
			return err
		}
		// 派生更新余额（仍保持“余额只读”语义：余额=累计派生值）
		return tx.Model(&BilWallet{}).Where("id = ?", wallet.ID).Update("balance", gorm.Expr("balance + ?", amount)).Error
	})
}

func (s *TruthService) DebitForOrder(ctx context.Context, customerID, orderID int64, amount int64, idem string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		wallet, err := s.lockWallet(tx, customerID)
		if err != nil {
			return err
		}
		if err := s.ensureIdem(tx, idem); err != nil {
			return err
		}
		// 余额校验
		var bal int64
		if err := tx.Model(&BilWallet{}).Where("id = ?", wallet.ID).Select("balance").Scan(&bal).Error; err != nil {
			return err
		}
		if bal < amount {
			return errors.New("insufficient balance")
		}
		now := time.Now().Unix()
		rec := &BilWalletTx{WalletID: wallet.ID, Direction: "debit", Amount: amount, Type: "order_pay", BizRefType: "order", BizRefID: orderID, IdempotencyKey: idem, CreatedAt: now}
		if err := tx.Create(rec).Error; err != nil {
			return err
		}
		return tx.Model(&BilWallet{}).Where("id = ?", wallet.ID).Update("balance", gorm.Expr("balance - ?", amount)).Error
	})
}

func (s *TruthService) CreditForRefund(ctx context.Context, customerID, orderID int64, amount int64, idem string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		wallet, err := s.lockWallet(tx, customerID)
		if err != nil {
			return err
		}
		if err := s.ensureIdem(tx, idem); err != nil {
			return err
		}
		now := time.Now().Unix()
		rec := &BilWalletTx{WalletID: wallet.ID, Direction: "credit", Amount: amount, Type: "order_refund", BizRefType: "order", BizRefID: orderID, IdempotencyKey: idem, CreatedAt: now}
		if err := tx.Create(rec).Error; err != nil {
			return err
		}
		return tx.Model(&BilWallet{}).Where("id = ?", wallet.ID).Update("balance", gorm.Expr("balance + ?", amount)).Error
	})
}

// helpers
func (s *TruthService) lockWallet(tx *gorm.DB, customerID int64) (*BilWallet, error) {
	var w BilWallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("customer_id = ?", customerID).First(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *TruthService) ensureIdem(tx *gorm.DB, idem string) error {
	if idem == "" {
		return errors.New("idempotency key required")
	}
	var count int64
	if err := tx.Model(&BilWalletTx{}).Where("idempotency_key = ?", idem).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("duplicate idempotency key")
	}
	return nil
}

// GetBalance 获取客户钱包余额 - 临时实现
func (s *TruthService) GetBalance(ctx context.Context, customerID int64) (int64, error) {
	// TODO: 在PR-2中实现具体逻辑
	return 0, nil
}

// GetTransactionHistory 获取交易历史 - 临时实现
func (s *TruthService) GetTransactionHistory(ctx context.Context, customerID int64, page, pageSize int) ([]billing.Transaction, error) {
	// TODO: 在PR-2中实现具体逻辑
	return []billing.Transaction{}, nil
}

// GetWalletByCustomerID 获取客户钱包信息 - 临时实现
func (s *TruthService) GetWalletByCustomerID(ctx context.Context, customerID int64) (*billing.WalletInfo, error) {
	// TODO: 在PR-2中实现具体逻辑
	return &billing.WalletInfo{
		ID:         0,
		CustomerID: customerID,
		Balance:    0,
		Status:     1,
		UpdatedAt:  0,
	}, nil
}

// CreateTransaction 创建交易 - 临时实现
func (s *TruthService) CreateTransaction(ctx context.Context, req *billing.CreateTransactionRequest) (*billing.Transaction, error) {
	// TODO: 在PR-2中实现具体逻辑
	return &billing.Transaction{
		ID:             0,
		WalletID:       0,
		Direction:      "credit",
		Amount:         int64(req.Amount * 100),
		Type:           req.Type,
		BizRefType:     "manual",
		BizRefID:       0,
		IdempotencyKey: "",
		OperatorID:     req.OperatorID,
		ReasonCode:     req.Reason,
		Note:           req.Reason,
		CreatedAt:      0,
	}, nil
}

// GetTransactions 获取交易历史 - 临时实现
func (s *TruthService) GetTransactions(ctx context.Context, customerID int64, req *billing.TransactionHistoryRequest) ([]billing.Transaction, int64, error) {
	// TODO: 在PR-2中实现具体逻辑
	return []billing.Transaction{}, 0, nil
}

var _ billing.Service = (*TruthService)(nil)
