-- +migrate Up
CREATE TABLE IF NOT EXISTS orders (
  id               VARCHAR(36) PRIMARY KEY,
  order_no         VARCHAR(50) NOT NULL,
  customer_id      VARCHAR(36) NOT NULL,
  contact_id       VARCHAR(36),
  order_date       DATETIME(6) NOT NULL,
  status           VARCHAR(20) NOT NULL DEFAULT 'draft' COMMENT '订单状态: draft, pending, confirmed, processing, shipped, completed, cancelled, refunded',
  payment_status   VARCHAR(20) NOT NULL DEFAULT 'unpaid' COMMENT '支付状态: unpaid, paid, partially_paid, refunded, pending',
  total_amount     DECIMAL(12,2)     NOT NULL DEFAULT 0,
  discount_amount  DECIMAL(12,2)     NOT NULL DEFAULT 0,
  final_amount     DECIMAL(12,2)     NOT NULL,
  payment_method   VARCHAR(20) COMMENT '支付方式: online, offline_transfer, cash_on_delivery, wallet_balance',
  remark           TEXT,
  assigned_to      VARCHAR(36),
  created_by       VARCHAR(36),
  created_at       TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at       TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at       DATETIME(6) NULL,
  UNIQUE KEY (order_no)
);

CREATE TABLE IF NOT EXISTS order_items (
  id             VARCHAR(36) PRIMARY KEY,
  order_id       VARCHAR(36) NOT NULL,
  product_id     VARCHAR(36) NOT NULL,
  product_name   VARCHAR(100) NOT NULL,
  quantity       INT            NOT NULL DEFAULT 1,
  unit_price     DECIMAL(12,2)  NOT NULL,
  discount_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
  final_price    DECIMAL(12,2)  NOT NULL,
  created_at     TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_orders_cust_date ON orders(customer_id, order_date DESC);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_pay_status ON orders(payment_status, payment_method);
CREATE INDEX idx_orders_assigned_open ON orders(assigned_to);
CREATE INDEX idx_items_order ON order_items(order_id);

-- +migrate Down
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;