package service

import (
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// IWalletService 定义了钱包服务的接口
type IWalletService interface {
	CreateWallet(ctx context.Context, customerID int64, walletType string) (*model.Wallet, error)
	CreateTransaction(ctx context.Context, customerID int64, operatorID int64, req *dto.WalletTransactionRequest) error
	GetWalletByCustomerID(ctx context.Context, customerID int64) (*dto.WalletResponse, error)
	GetTransactions(ctx context.Context, customerID int64, page, limit int) ([]*dto.WalletTransactionResponse, int64, error)
	ProcessRefund(ctx context.Context, customerID int64, operatorID int64, req *dto.WalletRefundRequest) error
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

	// 2. consume 时校验手机号后四位
	if req.Type == "consume" {
		// 获取客户手机号做后四位校验
		customerPhone := ""
		cust, err := s.q.Customer.WithContext(ctx).Select(s.q.Customer.Phone).Where(s.q.Customer.ID.Eq(customerID)).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCustomerNotFound
			}
			return err
		}
		customerPhone = cust.Phone
		if len(req.PhoneLast) > 0 {
			if len(customerPhone) < 4 || req.PhoneLast != customerPhone[len(customerPhone)-4:] {
				return errors.New("phone last4 verification failed")
			}
		} else {
			return errors.New("phone last4 is required for consume")
		}
	}

	// 3. 在事务中执行交易
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
		case "refund":
			transactionAmount = req.Amount
			newBalance = balanceBefore + req.Amount
			newTotalRecharged += req.Amount // 退款计入总充值金额
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
		if err != nil {
			return err
		}

		// 满赠：bonus_amount > 0 时，追加一条 correction 正向流水，不计入 total_recharged
		if req.Type == "recharge" && req.BonusAmount > 0 {
			bonusBefore := newBalance
			bonusAfter := bonusBefore + req.BonusAmount
			bonusTx := &model.WalletTransaction{
				WalletID:      wallet.ID,
				Type:          "correction",
				Amount:        req.BonusAmount,
				BalanceBefore: bonusBefore,
				BalanceAfter:  bonusAfter,
				Source:        "promotion",
				RelatedID:     req.RelatedID,
				Remark:        req.Remark,
				OperatorID:    operatorID,
			}
			if err := tx.WalletTransaction.WithContext(ctx).Create(bonusTx); err != nil {
				return err
			}
			// 更新余额
			_, err = tx.Wallet.WithContext(ctx).Where(tx.Wallet.ID.Eq(wallet.ID)).
				Updates(map[string]interface{}{
					"balance": bonusAfter,
				})
			if err != nil {
				return err
			}
		}

		return nil
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

// GetTransactions 获取客户的交易流水列表
func (s *WalletService) GetTransactions(ctx context.Context, customerID int64, page, limit int) ([]*dto.WalletTransactionResponse, int64, error) {
	// 1. 获取钱包ID
	wallet, err := s.q.Wallet.WithContext(ctx).
		Where(s.q.Wallet.CustomerID.Eq(customerID), s.q.Wallet.Type.Eq("balance")).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, ErrWalletNotFound
		}
		return nil, 0, err
	}

	// 2. 计算偏移量
	offset := (page - 1) * limit

	// 3. 查询交易记录
	transactions, err := s.q.WalletTransaction.WithContext(ctx).
		Where(s.q.WalletTransaction.WalletID.Eq(wallet.ID)).
		Order(s.q.WalletTransaction.CreatedAt.Desc()).
		Limit(limit).
		Offset(offset).
		Find()
	if err != nil {
		return nil, 0, err
	}

	// 4. 获取总数
	total, err := s.q.WalletTransaction.WithContext(ctx).
		Where(s.q.WalletTransaction.WalletID.Eq(wallet.ID)).
		Count()
	if err != nil {
		return nil, 0, err
	}

	// 5. 转换为DTO
	var resp []*dto.WalletTransactionResponse
	for _, t := range transactions {
		resp = append(resp, &dto.WalletTransactionResponse{
			ID:            t.ID,
			WalletID:      t.WalletID,
			Type:          t.Type,
			Amount:        t.Amount,
			BalanceBefore: t.BalanceBefore,
			BalanceAfter:  t.BalanceAfter,
			Source:        t.Source,
			RelatedID:     t.RelatedID,
			Remark:        t.Remark,
			OperatorID:    t.OperatorID,
			CreatedAt:     t.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return resp, total, nil
}

// ProcessRefund 处理退款
func (s *WalletService) ProcessRefund(ctx context.Context, customerID int64, operatorID int64, req *dto.WalletRefundRequest) error {
	// 1. 检查订单是否存在且属于该客户
	order, err := s.q.Order.WithContext(ctx).
		Where(s.q.Order.ID.Eq(req.OrderID), s.q.Order.CustomerID.Eq(customerID)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("订单不存在或不属于该客户")
		}
		return err
	}

	// 2. 检查订单状态是否允许退款
	if order.Status != "completed" && order.Status != "shipped" {
		return errors.New("订单状态不允许退款")
	}

	// 3. 创建退款交易请求
	refundReq := &dto.WalletTransactionRequest{
		Type:      "refund",
		Amount:    req.Amount,
		Source:    "refund",
		RelatedID: req.OrderID,
		Remark:    fmt.Sprintf("订单退款 - 原因: %s", req.Reason),
	}

	// 添加用户备注
	if req.Remark != "" {
		refundReq.Remark += fmt.Sprintf(" - 备注: %s", req.Remark)
	}

	// 4. 调用通用交易创建方法
	return s.CreateTransaction(ctx, customerID, operatorID, refundReq)
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
		CreatedAt:      wallet.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      wallet.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
