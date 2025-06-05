-- +migrate Up
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL, -- 逻辑外键 -> wallets.id
    type VARCHAR(20) CHECK (type IN ('recharge', 'consume', 'refund', 'freeze', 'unfreeze', 'correction')) NOT NULL,
    amount DECIMAL(10,2) NOT NULL, -- 正数表示增加，负数表示减少
    balance_before DECIMAL(10,2) NOT NULL, -- 交易前余额
    balance_after DECIMAL(10,2) NOT NULL, -- 交易后余额
    source VARCHAR(50) NOT NULL, -- 交易来源：manual, order, refund, system等
    related_id UUID, -- 关联ID（如订单ID、退款ID等）
    remark TEXT,
    operator_id UUID, -- 操作人员 (逻辑外键 -> admin_users.id)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
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
