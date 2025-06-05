-- +migrate Up
-- 创建订单主表
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_no VARCHAR(50) NOT NULL UNIQUE, -- 订单编号
    customer_id UUID NOT NULL, -- 逻辑外键 -> customers.id
    contact_id UUID, -- 关联的联系人 (逻辑外键 -> contacts.id)
    order_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) CHECK (status IN ('draft', 'pending', 'confirmed', 'processing', 'completed', 'cancelled', 'refunded')) DEFAULT 'draft',
    total_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    discount_amount DECIMAL(10,2) DEFAULT 0.00,
    final_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00, -- 最终金额 = total_amount - discount_amount
    payment_method VARCHAR(20), -- 支付方式：cash, card, wallet, etc.
    payment_status VARCHAR(20) CHECK (payment_status IN ('unpaid', 'partial', 'paid', 'refunded')) DEFAULT 'unpaid',
    remark TEXT,
    assigned_to UUID, -- 处理人员 (逻辑外键 -> admin_users.id)
    created_by UUID, -- 创建人 (逻辑外键 -> admin_users.id)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 创建订单明细表
CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL, -- 逻辑外键 -> orders.id
    product_id UUID NOT NULL, -- 逻辑外键 -> products.id
    product_name VARCHAR(100) NOT NULL, -- 冗余存储，防止产品信息变更影响历史订单
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL, -- quantity * unit_price
    discount_amount DECIMAL(10,2) DEFAULT 0.00,
    final_price DECIMAL(10,2) NOT NULL, -- total_price - discount_amount
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_orders_order_no ON orders(order_no);
CREATE INDEX idx_orders_customer_id ON orders(customer_id);
CREATE INDEX idx_orders_contact_id ON orders(contact_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_payment_status ON orders(payment_status);
CREATE INDEX idx_orders_order_date ON orders(order_date);
CREATE INDEX idx_orders_assigned_to ON orders(assigned_to);
CREATE INDEX idx_orders_created_by ON orders(created_by);
CREATE INDEX idx_orders_deleted_at ON orders(deleted_at);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);

-- 创建触发器函数，用于生成订单编号
CREATE OR REPLACE FUNCTION generate_order_no() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.order_no IS NULL OR NEW.order_no = '' THEN
        NEW.order_no := 'ORD' || TO_CHAR(CURRENT_DATE, 'YYYYMMDD') || LPAD(NEXTVAL('order_no_seq')::TEXT, 6, '0');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建序列
CREATE SEQUENCE IF NOT EXISTS order_no_seq START 1;

-- 创建触发器
CREATE TRIGGER trigger_generate_order_no
    BEFORE INSERT ON orders
    FOR EACH ROW
    EXECUTE FUNCTION generate_order_no();

-- +migrate Down
DROP TRIGGER IF EXISTS trigger_generate_order_no ON orders;
DROP FUNCTION IF EXISTS generate_order_no();
DROP SEQUENCE IF EXISTS order_no_seq;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
