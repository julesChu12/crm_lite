-- +migrate Up
CREATE TABLE IF NOT EXISTS customers (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(100),
    gender VARCHAR(10) DEFAULT 'unknown' COMMENT '性别: male, female, unknown',
    birthday DATE,
    level VARCHAR(20) DEFAULT '普通' COMMENT '客户等级: 普通, 银牌, 金牌, 铂金',
    tags JSON,
    note TEXT,
    source VARCHAR(50) DEFAULT 'manual' COMMENT '客户来源：manual, referral, marketing, etc.',
    assigned_to VARCHAR(36) COMMENT '分配给哪个员工',
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME(6) NULL,
    UNIQUE KEY (phone)
);

-- 创建索引
CREATE INDEX idx_customers_email ON customers(email);
CREATE INDEX idx_customers_level ON customers(level);
CREATE INDEX idx_customers_assigned_to ON customers(assigned_to);
CREATE INDEX idx_customers_deleted_at ON customers(deleted_at);
CREATE INDEX idx_customers_created_at ON customers(created_at);

-- +migrate Down
DROP TABLE IF EXISTS customers;
