-- +migrate Up
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL, -- 逻辑外键 -> customers.id
    type VARCHAR(20) CHECK (type IN ('balance', 'points', 'coupon', 'deposit')) DEFAULT 'balance',
    balance DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    frozen_balance DECIMAL(10,2) DEFAULT 0.00, -- 冻结金额
    total_recharged DECIMAL(10,2) DEFAULT 0.00, -- 累计充值
    total_consumed DECIMAL(10,2) DEFAULT 0.00, -- 累计消费
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(customer_id, type) -- 每个客户每种类型只能有一个钱包
);

-- 创建索引
CREATE INDEX idx_wallets_customer_id ON wallets(customer_id);
CREATE INDEX idx_wallets_type ON wallets(type);
CREATE INDEX idx_wallets_balance ON wallets(balance);

-- +migrate Down
DROP TABLE IF EXISTS wallets;
