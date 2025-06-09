-- +migrate Up
CREATE TABLE IF NOT EXISTS marketing_records (
    id VARCHAR(36) PRIMARY KEY,
    campaign_id VARCHAR(36) NOT NULL,
    customer_id VARCHAR(36) NOT NULL,
    contact_id VARCHAR(36),
    channel VARCHAR(20) NOT NULL COMMENT '触达渠道',
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending, sent, delivered, failed, opened, clicked, replied, unsubscribed',
    error_message TEXT COMMENT '发送失败时的错误信息',
    response JSON COMMENT '客户反馈或互动数据',
    sent_at DATETIME(6) NULL,
    delivered_at DATETIME(6) NULL,
    opened_at DATETIME(6) NULL COMMENT '打开时间（邮件/推送）',
    clicked_at DATETIME(6) NULL COMMENT '点击时间',
    replied_at DATETIME(6) NULL COMMENT '回复时间',
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_marketing_records_campaign_id ON marketing_records(campaign_id);
CREATE INDEX idx_marketing_records_customer_id ON marketing_records(customer_id);
CREATE INDEX idx_marketing_records_contact_id ON marketing_records(contact_id);
CREATE INDEX idx_marketing_records_channel ON marketing_records(channel);
CREATE INDEX idx_marketing_records_status ON marketing_records(status);
CREATE INDEX idx_marketing_records_sent_at ON marketing_records(sent_at);
CREATE INDEX idx_marketing_records_delivered_at ON marketing_records(delivered_at);
CREATE INDEX idx_marketing_records_opened_at ON marketing_records(opened_at);
CREATE INDEX idx_marketing_records_clicked_at ON marketing_records(clicked_at);

-- +migrate Down
DROP TABLE IF EXISTS marketing_records;
