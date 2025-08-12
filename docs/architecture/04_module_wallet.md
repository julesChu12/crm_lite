# 💰 钱包模块设计文档（Wallet Module） - CRM Lite

## 📌 模块概述

钱包模块用于为客户账户提供余额记录、充值、消费、退款等基本交易能力。适用于会员卡余额、积分账户、押金账户等多种资金模式的实现。

---

## 🧱 数据模型设计

### 钱包主表（wallet）

| 字段名        | 类型     | 描述                    |
|---------------|----------|-------------------------|
| `id`          | UUID     | 钱包唯一标识            |
| `customer_id` | UUID     | 所属客户ID              |
| `balance`     | float    | 当前余额 (注意：使用 float 类型处理金额需注意精度问题，或考虑使用定点数/整数存储分)                |
| `type`        | string   | 钱包类型（如 `balance`现金余额, `points`积分）|
| `created_at`  | datetime | 创建时间                |
| `updated_at`  | datetime | 更新时间                |

### 钱包流水表（wallet_transaction）

| 字段名          | 类型     | 描述                      |
|------------------|----------|---------------------------|
| `id`             | UUID     | 交易唯一标识              |
| `wallet_id`      | UUID     | 所属钱包ID (`wallet.id`)   |
| `type`           | string   | 交易类型（`recharge`充值 / `consume`消费 / `refund`退款 / `correction`调整）|
| `amount`         | float    | 金额（正数表示增加, 负数表示减少）(注意：使用 float 类型处理金额需注意精度问题，或考虑使用定点数/整数存储分)         |
| `source`         | string   | 交易来源（如 `order:uuid`, `manual`, `system_init`）  |
| `related_id`     | string   | 关联ID (例如订单ID, 活动ID等，可选) |
| `remark`         | string   | 备注                      |
| `created_at`     | datetime | 创建时间                  |

---

## 🔌 接口设计（RESTful）

### 钱包接口

#### GET `/api/v1/customers/:id/wallet`

- 描述：获取指定客户的全部钱包账户信息（一个客户可能拥有多种类型的钱包，如现金钱包、积分钱包）。
- 请求参数：无
- 返回：

```json
{
  "wallets": [
    {
      "id": "wallet-uuid-balance",
      "customer_id": "customer-uuid",
      "type": "balance",
      "balance": 150.75,
      "updated_at": "2024-01-01T10:00:00Z"
    },
    {
      "id": "wallet-uuid-points",
      "customer_id": "customer-uuid",
      "type": "points",
      "balance": 2000,
      "updated_at": "2024-01-01T11:00:00Z"
    }
  ]
}
```

#### POST `/api/v1/customers/:id/wallet/transactions`  （充值、消费、退款等统一使用该接口）

- 描述：为客户指定类型的钱包创建一笔交易（充值/消费/退款）。
- 请求体 (JSON)：

```json
{
  "type": "recharge",            // recharge | consume | refund
  "amount": 100.00,               // 金额为正数，方向由 type 决定
  "source": "manual_recharge",  // 来源，如 promotion:FULL_100_GET_20
  "remark": "线下充值100元现金",
  "bonus_amount": 20.00,          // 可选：当 type=recharge 时传入，自动追加一条 correction 赠送流水
  "phone_last4": "6789"          // 可选：当 type=consume 时必填，用于手机号后四位校验
}
```

- 返回：成功时返回操作成功。

```json
{
  "transaction": {
    "id": "trans-uuid",
    "wallet_id": "wallet-uuid-balance",
    "type": "recharge",
    "amount": 100.00,
    "source": "manual_recharge",
    "remark": "线下充值100元现金",
    "created_at": "2024-01-02T14:30:00Z"
  },
  "wallet": {
      "id": "wallet-uuid-balance",
      "type": "balance",
      "balance": 250.75,
      "updated_at": "2024-01-02T14:30:00Z"
  }
}
```

不再区分 `recharge`、`consume` 单独端点，全部通过上述 `POST /transactions` 接口并在请求体中指定 `type` 字段 (`recharge`, `consume`, `refund`, `correction`)。

#### GET `/api/v1/customers/:id/wallet/transactions`

- 描述：获取某客户特定类型钱包的交易流水，或所有类型钱包的交易流水（如果 `wallet_type` 未指定）。
- 请求参数：
  - `wallet_type` (string, 可选)：钱包类型（如 `balance`, `points`）进行筛选。
  - `type` (string, 可选): 交易类型 (`recharge`, `consume`, `refund`, `correction`) 筛选。
  - `start_date` (date, 可选)：起始日期。
  - `end_date` (date, 可选)：结束日期。
  - `page` (int, 可选)：页码。
  - `page_size` (int, 可选)：每页数量。
- 返回：

```json
{
  "transactions": [
    {
      "id": 1,
      "wallet_id": 10,
      "type": "recharge",
      "amount": 100.00,
      "source": "manual_recharge",
      "remark": "首次充值",
      "balance_before": 0,
      "balance_after": 100.00,
      "operator_id": 1001,
      "created_at": "2024-01-01T10:00:00Z"
    },
    {
      "id": 2,
      "wallet_id": 10,
      "type": "consume",
      "amount": -20.50,
      "source": "manual_consume",
      "remark": "剪发",
      "balance_before": 100.00,
      "balance_after": 79.50,
      "operator_id": 1001,
      "created_at": "2024-01-01T12:00:00Z"
    }
  ],
  "pagination": {
    "total": 20,
    "page": 1,
    "page_size": 10
  }
}
```

---

### 🔐 权限与校验

- 所有资金操作（充值、消费等）均需用户具备如 `wallet:recharge`, `wallet:consume` 等精细化权限，或统一的 `wallet:write` 权限。
- 消费金额不可超出当前钱包类型的可用余额。
- 消费必须同时提供客户手机号后四位 `phone_last4`，与客户档案手机号做一致性校验，用于现场核对人物。
- 充值支持携带 `bonus_amount`，系统会自动追加一条 `correction` 赠送流水；累计 `total_recharged` 仅统计实付，不包含赠送。
- 所有交易操作均记录在 `wallet_transactions` 表中，作为资金流转的审计日志。
- 关键操作（如修改余额）应考虑并发控制，例如使用数据库事务和乐观锁/悲观锁。

---

### 🧪 单元测试建议

- 充值、消费、退款等操作的边界值测试（例如，0金额，超额消费）。
- 不同钱包类型操作的隔离性与正确性。
- 负数交易逻辑（如消费、退款部分冲正）和异常处理。
- 模拟并发消费或充值，验证数据一致性和事务的正确性。
- 交易流水查询的准确性（基于类型、日期等）。

---

### 💡 扩展建议

- 支持多种钱包类型，如现金余额 (`balance`)、积分 (`points`)、优惠券余额 (`coupon_balance`)、押金 (`deposit`) 等。
- 为特定钱包类型（如积分）增加自动过期机制和提醒。
- 对接第三方支付平台（如支付宝、微信支付）实现用户自助在线充值。
- 增加钱包冻结/解冻功能，用于处理争议交易或风险控制。
- 考虑资金密码或二次验证，增强高风险操作的安全性。
