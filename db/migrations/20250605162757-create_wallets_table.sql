-- +migrate Up
CREATE TABLE IF NOT EXISTS wallets (
    id VARCHAR(36) PRIMARY KEY,
    customer_id VARCHAR(36) NOT NULL,
    type VARCHAR(20) DEFAULT 'balance' COMMENT '钱包类型: balance, points, coupon, deposit',
    balance DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    frozen_balance DECIMAL(10,2) DEFAULT 0.00 COMMENT '冻结金额',
    total_recharged DECIMAL(10,2) DEFAULT 0.00 COMMENT '累计充值',
    total_consumed DECIMAL(10,2) DEFAULT 0.00 COMMENT '累计消费',
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY (customer_id, type)
);

-- 创建索引
CREATE INDEX idx_wallets_balance ON wallets(balance);

-- +migrate Down
DROP TABLE IF EXISTS wallets;
