-- +migrate Up
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id VARCHAR(36) PRIMARY KEY,
    wallet_id VARCHAR(36) NOT NULL,
    type VARCHAR(20) NOT NULL COMMENT '交易类型: recharge, consume, refund, freeze, unfreeze, correction',
    amount DECIMAL(10,2) NOT NULL COMMENT '正数表示增加，负数表示减少',
    balance_before DECIMAL(10,2) NOT NULL COMMENT '交易前余额',
    balance_after DECIMAL(10,2) NOT NULL COMMENT '交易后余额',
    source VARCHAR(50) NOT NULL COMMENT '交易来源：manual, order, refund, system等',
    related_id VARCHAR(36) COMMENT '关联ID（如订单ID、退款ID等）',
    remark TEXT,
    operator_id VARCHAR(36) COMMENT '操作人员',
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX idx_wallet_transactions_type ON wallet_transactions(type);
CREATE INDEX idx_wallet_transactions_source ON wallet_transactions(source);
CREATE INDEX idx_wallet_transactions_related_id ON wallet_transactions(related_id);
CREATE INDEX idx_wallet_transactions_created_at ON wallet_transactions(created_at);
CREATE INDEX idx_wallet_transactions_operator_id ON wallet_transactions(operator_id);

-- +migrate Down
DROP TABLE IF EXISTS wallet_transactions;
