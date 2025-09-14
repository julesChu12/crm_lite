-- +migrate Up
-- 钱包表：实现余额只读原则
-- 核心设计理念：balance字段由wallet_transactions表聚合计算，不可直接修改
-- 所有余额变更必须通过创建交易记录实现，确保资金安全和可追溯性
CREATE TABLE IF NOT EXISTS wallets (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    customer_id BIGINT NOT NULL UNIQUE COMMENT '客户ID，每个客户只有一个钱包',
    balance BIGINT NOT NULL DEFAULT 0 COMMENT '当前余额（分），只读字段，由交易聚合计算',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '钱包状态：1-正常，0-冻结',
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at BIGINT NOT NULL DEFAULT 0 COMMENT 'Unix时间戳，用于乐观锁'
);

-- 创建索引
CREATE INDEX idx_wallets_customer ON wallets(customer_id);
CREATE INDEX idx_wallets_status ON wallets(status);

-- +migrate Down
DROP TABLE IF EXISTS wallets;
