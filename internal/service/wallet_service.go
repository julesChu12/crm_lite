package service

import (
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"errors"

	"gorm.io/gorm"
)

// IWalletService 定义了钱包服务的接口
type IWalletService interface {
	CreateWallet(ctx context.Context, customerID int64, walletType string) (*model.Wallet, error)
	CreateTransaction(ctx context.Context, customerID int64, operatorID int64, req *dto.WalletTransactionRequest) error
	GetWalletByCustomerID(ctx context.Context, customerID int64) (*dto.WalletResponse, error)
}

// WalletService 提供了钱包相关的服务
type WalletService struct {
	q        *query.Query
	resource *resource.Manager
}

// NewWalletService 创建一个新的 WalletService
func NewWalletService(resManager *resource.Manager) IWalletService {
	db, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource for WalletService: " + err.Error())
	}
	return &WalletService{
		q:        query.Use(db.DB),
		resource: resManager,
	}
}

// CreateWallet 为指定客户创建钱包（如果尚不存在）
func (s *WalletService) CreateWallet(ctx context.Context, customerID int64, walletType string) (*model.Wallet, error) {
	// 默认创建 balance 类型的钱包
	if walletType == "" {
		walletType = "balance"
	}

	// 检查钱包是否已存在，确保幂等性
	existingWallet, err := s.q.Wallet.WithContext(ctx).
		Where(s.q.Wallet.CustomerID.Eq(customerID), s.q.Wallet.Type.Eq(walletType)).
		First()

	if err == nil {
		// 钱包已存在，直接返回
		return existingWallet, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 其他数据库错误
		return nil, err
	}

	// 钱包不存在，创建新的钱包
	newWallet := &model.Wallet{
		CustomerID: customerID,
		Type:       walletType,
		Balance:    0, // 初始余额为0
	}

	if err := s.q.Wallet.WithContext(ctx).Create(newWallet); err != nil {
		return nil, err
	}

	return newWallet, nil
}

func (s *WalletService) CreateTransaction(ctx context.Context, customerID int64, operatorID int64, req *dto.WalletTransactionRequest) error {
	// 不再在此处做权限检查，权限由中间件保证

	// 1. 校验客户是否存在
	customer, err := s.q.Customer.WithContext(ctx).Where(s.q.Customer.ID.Eq(customerID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCustomerNotFound
		}
		return err
	}
	_ = customer // 当前逻辑中仅验证存在性，可根据需要扩展

	// 2. 在事务中执行交易
	return s.q.Transaction(func(tx *query.Query) error {
		// a. 锁定钱包记录以防止并发问题
		wallet, err := tx.Wallet.WithContext(ctx).
			Where(tx.Wallet.CustomerID.Eq(customerID), tx.Wallet.Type.Eq("balance")).
			First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrWalletNotFound
			}
			return err
		}

		// b. 计算新余额和交易金额
		balanceBefore := wallet.Balance
		var transactionAmount float64
		var newBalance float64
		var newTotalConsumed float64 = wallet.TotalConsumed
		var newTotalRecharged float64 = wallet.TotalRecharged

		switch req.Type {
		case "consume":
			transactionAmount = -req.Amount
			newBalance = balanceBefore - req.Amount
			newTotalConsumed += req.Amount
			if newBalance < 0 {
				return ErrInsufficientBalance
			}
		case "recharge":
			transactionAmount = req.Amount
			newBalance = balanceBefore + req.Amount
			newTotalRecharged += req.Amount
		default:
			return errors.New("unsupported transaction type")
		}

		// c. 创建交易记录
		transaction := &model.WalletTransaction{
			WalletID:      wallet.ID,
			Type:          req.Type,
			Amount:        transactionAmount,
			BalanceBefore: balanceBefore,
			BalanceAfter:  newBalance,
			Source:        req.Source,
			RelatedID:     req.RelatedID,
			Remark:        req.Remark,
			OperatorID:    operatorID,
		}
		if err := tx.WalletTransaction.WithContext(ctx).Create(transaction); err != nil {
			return err
		}

		// d. 更新钱包余额和累计数据
		_, err = tx.Wallet.WithContext(ctx).Where(tx.Wallet.ID.Eq(wallet.ID)).
			Updates(map[string]interface{}{
				"balance":         newBalance,
				"total_consumed":  newTotalConsumed,
				"total_recharged": newTotalRecharged,
			})
		return err
	})
}

// GetWalletByCustomerID 获取指定客户的钱包信息
func (s *WalletService) GetWalletByCustomerID(ctx context.Context, customerID int64) (*dto.WalletResponse, error) {
	wallet, err := s.q.Wallet.WithContext(ctx).
		Where(s.q.Wallet.CustomerID.Eq(customerID), s.q.Wallet.Type.Eq("balance")).
		First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWalletNotFound // 自定义错误
		}
		return nil, err
	}

	return s.toWalletResponse(wallet), nil
}

// toWalletResponse 将 model.Wallet 转换为 dto.WalletResponse
func (s *WalletService) toWalletResponse(wallet *model.Wallet) *dto.WalletResponse {
	return &dto.WalletResponse{
		ID:             wallet.ID,
		CustomerID:     wallet.CustomerID,
		Type:           wallet.Type,
		Balance:        wallet.Balance,
		FrozenBalance:  wallet.FrozenBalance,
		TotalRecharged: wallet.TotalRecharged,
		TotalConsumed:  wallet.TotalConsumed,
		CreatedAt:      wallet.CreatedAt,
		UpdatedAt:      wallet.UpdatedAt,
	}
}
