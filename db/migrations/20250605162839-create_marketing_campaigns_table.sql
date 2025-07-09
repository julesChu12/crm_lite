-- +migrate Up
CREATE TABLE IF NOT EXISTS marketing_campaigns (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    type VARCHAR(20) NOT NULL COMMENT '营销类型: sms, email, push_notification, wechat, call',
    status VARCHAR(20) DEFAULT 'draft' COMMENT '状态: draft, scheduled, active, paused, completed, archived',
    target_tags JSON COMMENT '目标客户标签',
    target_segment_id BIGINT COMMENT '目标客户分群ID（如果有客户分群功能）',
    content_template_id BIGINT COMMENT '内容模板ID（如果有模板功能）',
    content TEXT NOT NULL COMMENT '活动具体内容或模板变量的JSON数据',
    start_time DATETIME(6) NULL,
    end_time DATETIME(6) NULL,
    actual_start_time DATETIME(6) NULL,
    actual_end_time DATETIME(6) NULL,
    target_count INTEGER DEFAULT 0 COMMENT '目标客户数量',
    sent_count INTEGER DEFAULT 0 COMMENT '已发送数量',
    success_count INTEGER DEFAULT 0 COMMENT '成功数量',
    click_count INTEGER DEFAULT 0 COMMENT '点击数量',
    created_by BIGINT COMMENT '创建人',
    updated_by BIGINT COMMENT '更新人',
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME(6) NULL
);

-- 创建索引
CREATE INDEX idx_marketing_campaigns_name ON marketing_campaigns(name);
CREATE INDEX idx_marketing_campaigns_type ON marketing_campaigns(type);
CREATE INDEX idx_marketing_campaigns_status ON marketing_campaigns(status);
CREATE INDEX idx_marketing_campaigns_start_time ON marketing_campaigns(start_time);
CREATE INDEX idx_marketing_campaigns_end_time ON marketing_campaigns(end_time);
CREATE INDEX idx_marketing_campaigns_created_by ON marketing_campaigns(created_by);
CREATE INDEX idx_marketing_campaigns_deleted_at ON marketing_campaigns(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS marketing_campaigns;
