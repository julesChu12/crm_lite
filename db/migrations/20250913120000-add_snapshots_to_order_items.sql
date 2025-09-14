-- +migrate Up
-- 为 order_items 表添加产品快照字段
-- 确保订单历史数据的完整性和可追溯性，支持产品信息变更后的历史查询
ALTER TABLE order_items
  ADD COLUMN IF NOT EXISTS product_name_snapshot VARCHAR(255) NULL COMMENT '产品名称快照，下单时的产品名称',
  ADD COLUMN IF NOT EXISTS unit_price_snapshot BIGINT NULL COMMENT '单价快照（分），下单时的产品价格',
  ADD COLUMN IF NOT EXISTS duration_min_snapshot INT NULL COMMENT '服务时长快照（分钟），下单时的服务时长';

-- 添加索引以优化查询性能
CREATE INDEX IF NOT EXISTS idx_order_items_snapshots ON order_items(product_name_snapshot, unit_price_snapshot);

-- +migrate Down
-- 回滚时删除快照字段和相关索引
DROP INDEX IF EXISTS idx_order_items_snapshots ON order_items;
ALTER TABLE order_items
  DROP COLUMN IF EXISTS product_name_snapshot,
  DROP COLUMN IF EXISTS unit_price_snapshot,
  DROP COLUMN IF EXISTS duration_min_snapshot;
