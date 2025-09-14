-- +migrate Up
-- 创建系统事件发布表（Outbox模式）
-- 用于实现业务操作和事件发布的原子性，支持最终一致性
CREATE TABLE IF NOT EXISTS sys_outbox (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  event_type VARCHAR(64) NOT NULL COMMENT '事件类型：order.placed, wallet.credited等',
  payload JSON NOT NULL COMMENT '事件载荷数据',
  created_at BIGINT NOT NULL COMMENT '创建时间（Unix时间戳）',
  processed_at BIGINT NULL COMMENT '处理时间（Unix时间戳），NULL表示未处理',
  INDEX idx_type_time (event_type, created_at),
  INDEX idx_processed (processed_at)
) ENGINE=InnoDB COMMENT='系统事件发布表，支持Outbox模式';

-- +migrate Down
DROP TABLE IF EXISTS sys_outbox;
