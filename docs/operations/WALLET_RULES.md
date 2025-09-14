## 钱包规则与审计

### 不变量
- 余额只读：`balance = Σ(credit) - Σ(debit)`，禁止任何直接 `UPDATE wallets SET balance = ...`。
- 幂等：每笔交易必须携带 `idempotency_key`，全局唯一。
- 业务引用：
  - 扣减（debit）必须绑定 `biz_ref_type='order'` 与 `biz_ref_id`。
  - 退款（credit）需绑定原 `order_id`。
- 审计：记录 actor, action, resource, id, redacted_payload。

### 交易类型
- recharge（credit）
- order_pay（debit）
- order_refund（credit）
- adjust_in（credit）/ adjust_out（debit）

### 并发与锁
- 扣减前 `SELECT ... FOR UPDATE` 行锁钱包或使用数据库原子更新。
- 幂等冲突返回首单结果或冲突错误码（409）。

### 反洗钱与风控（预留）
- 金额阈值、频次控制、黑名单源等逻辑位于 billing 层实现。


