package impl

import (
	"context"

	"crm_lite/internal/common"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/domains/billing"
	"crm_lite/internal/domains/sales"
	"crm_lite/pkg/utils"

	"gorm.io/gorm"
)

// Sample wiring for future transaction fold-in (not used by controller yet)
type PlaceOrderDeps struct {
	RM      *resource.Manager
	Billing billing.Service
}

func (s *ServiceImpl) PlaceOrderWithDeps(ctx context.Context, req sales.PlaceOrderReq, deps PlaceOrderDeps) (sales.Order, error) {
	dbRes, _ := resource.Get[*resource.DBResource](deps.RM, resource.DBServiceKey)
	outbox := common.NewOutboxWriter(dbRes.DB)

	var created model.Order
	err := dbRes.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		// 1) 拉产品、计算快照与金额
		var total int64
		for _, it := range req.Items {
			p, err := qtx.Product.WithContext(ctx).Where(qtx.Product.ID.Eq(it.ProductID)).First()
			if err != nil {
				return err
			}
			total += int64(it.Qty) * int64(p.Price*100)
		}
		created = model.Order{OrderNo: utils.GenerateOrderNo(), CustomerID: req.CustomerID, Status: "draft", PaymentMethod: req.PayMethod, TotalAmount: float64(total) / 100.0, FinalAmount: float64(total) / 100.0}
		if err := qtx.Order.WithContext(ctx).Create(&created); err != nil {
			return err
		}
		// 2) 明细 + 快照（先占位，PR-3 正式完善）
		// 3) 钱包扣减（pay_method=wallet）
		if req.PayMethod == "wallet" && deps.Billing != nil {
			if err := deps.Billing.DebitForOrder(ctx, req.CustomerID, created.ID, total, req.IdemKey); err != nil {
				return err
			}
		}
		// 4) 写 outbox（示例）
		_ = outbox.WriteOutbox(ctx, "OrderPlaced", map[string]any{"order_id": created.ID})
		return nil
	})
	if err != nil {
		return sales.Order{}, err
	}
	return sales.Order{ID: created.ID, TotalAmount: int64(created.FinalAmount * 100), Status: created.Status, PayMethod: created.PaymentMethod}, nil
}
