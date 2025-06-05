-- +migrate Up
CREATE TABLE IF NOT EXISTS marketing_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL, -- 逻辑外键 -> marketing_campaigns.id
    customer_id UUID NOT NULL, -- 逻辑外键 -> customers.id
    contact_id UUID, -- 具体联系人 (逻辑外键 -> contacts.id)
    channel VARCHAR(20) NOT NULL, -- 触达渠道
    status VARCHAR(20) CHECK (status IN ('pending', 'sent', 'delivered', 'failed', 'opened', 'clicked', 'replied', 'unsubscribed')) DEFAULT 'pending',
    error_message TEXT, -- 发送失败时的错误信息
    response JSONB, -- 客户反馈或互动数据
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    opened_at TIMESTAMP WITH TIME ZONE, -- 打开时间（邮件/推送）
    clicked_at TIMESTAMP WITH TIME ZONE, -- 点击时间
    replied_at TIMESTAMP WITH TIME ZONE, -- 回复时间
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
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
CREATE INDEX idx_marketing_records_response ON marketing_records USING GIN(response);

-- +migrate Down
DROP TABLE IF EXISTS marketing_records;
