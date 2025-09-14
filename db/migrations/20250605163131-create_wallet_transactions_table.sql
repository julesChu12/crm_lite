-- +migrate Up  
-- 钱包交易流水表：钱包余额变更的唯一真相来源
-- 核心设计理念：
-- 1. 所有钱包余额变更必须通过此表记录，禁止直接修改wallets.balance
-- 2. 每笔交易必须有幂等键，防止重复操作
-- 3. 业务引用字段确保交易的业务可追溯性
-- 4. direction字段明确资金流向，amount始终为正数
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    wallet_id BIGINT NOT NULL COMMENT '钱包ID',
    direction ENUM('credit','debit') NOT NULL COMMENT '资金方向：credit-入账，debit-出账',
    amount BIGINT NOT NULL COMMENT '交易金额（分），始终为正数',
    type ENUM('recharge','order_pay','order_refund','adjust_in','adjust_out') NOT NULL COMMENT '交易类型',
    biz_ref_type VARCHAR(32) NULL COMMENT '业务引用类型：order/refund/manual等',
    biz_ref_id BIGINT NULL COMMENT '业务引用ID，如订单ID',
    idempotency_key VARCHAR(64) NOT NULL COMMENT '幂等键，防止重复交易',
    operator_id BIGINT NULL COMMENT '操作员ID',
    reason_code VARCHAR(32) NULL COMMENT '交易原因代码',
    note VARCHAR(255) NULL COMMENT '备注信息',
    created_at BIGINT NOT NULL COMMENT '创建时间（Unix时间戳）',
    
    -- 幂等约束：确保相同幂等键不会重复执行
    UNIQUE KEY uk_idempotency (idempotency_key),
    -- 复合索引：按钱包和时间查询交易历史
    INDEX idx_wallet_time (wallet_id, created_at),
    -- 业务引用索引：用于查找特定订单的交易记录
    INDEX idx_biz_ref (biz_ref_type, biz_ref_id),
    -- 操作员索引：用于审计和查询
    INDEX idx_operator (operator_id),
    -- 交易类型索引：用于统计分析
    INDEX idx_type (type)
);

-- +migrate Down
DROP TABLE IF EXISTS wallet_transactions;
