-- +migrate Up
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL UNIQUE,
    email VARCHAR(100),
    gender VARCHAR(10) CHECK (gender IN ('male', 'female', 'unknown')) DEFAULT 'unknown',
    birthday DATE,
    level VARCHAR(20) DEFAULT '普通',
    tags JSONB DEFAULT '[]',
    note TEXT,
    source VARCHAR(50) DEFAULT 'manual', -- 客户来源：manual, referral, marketing, etc.
    assigned_to UUID, -- 分配给哪个员工 (逻辑外键 -> admin_users.id)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 创建索引
CREATE INDEX idx_customers_phone ON customers(phone);
CREATE INDEX idx_customers_email ON customers(email);
CREATE INDEX idx_customers_level ON customers(level);
CREATE INDEX idx_customers_assigned_to ON customers(assigned_to);
CREATE INDEX idx_customers_deleted_at ON customers(deleted_at);
CREATE INDEX idx_customers_tags ON customers USING GIN(tags);
CREATE INDEX idx_customers_created_at ON customers(created_at);

-- +migrate Down
DROP TABLE IF EXISTS customers;
