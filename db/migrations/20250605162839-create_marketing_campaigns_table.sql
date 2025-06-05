-- +migrate Up
CREATE TABLE IF NOT EXISTS marketing_campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    type VARCHAR(20) CHECK (type IN ('sms', 'email', 'push_notification', 'wechat', 'call')) NOT NULL,
    status VARCHAR(20) CHECK (status IN ('draft', 'scheduled', 'active', 'paused', 'completed', 'archived')) DEFAULT 'draft',
    target_tags JSONB DEFAULT '[]', -- 目标客户标签
    target_segment_id UUID, -- 目标客户分群ID（如果有客户分群功能）
    content_template_id UUID, -- 内容模板ID（如果有模板功能）
    content TEXT NOT NULL, -- 活动具体内容或模板变量的JSON数据
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    actual_start_time TIMESTAMP WITH TIME ZONE,
    actual_end_time TIMESTAMP WITH TIME ZONE,
    target_count INTEGER DEFAULT 0, -- 目标客户数量
    sent_count INTEGER DEFAULT 0, -- 已发送数量
    success_count INTEGER DEFAULT 0, -- 成功数量
    click_count INTEGER DEFAULT 0, -- 点击数量
    created_by UUID, -- 创建人 (逻辑外键 -> admin_users.id)
    updated_by UUID, -- 更新人 (逻辑外键 -> admin_users.id)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 创建索引
CREATE INDEX idx_marketing_campaigns_name ON marketing_campaigns(name);
CREATE INDEX idx_marketing_campaigns_type ON marketing_campaigns(type);
CREATE INDEX idx_marketing_campaigns_status ON marketing_campaigns(status);
CREATE INDEX idx_marketing_campaigns_start_time ON marketing_campaigns(start_time);
CREATE INDEX idx_marketing_campaigns_end_time ON marketing_campaigns(end_time);
CREATE INDEX idx_marketing_campaigns_created_by ON marketing_campaigns(created_by);
CREATE INDEX idx_marketing_campaigns_deleted_at ON marketing_campaigns(deleted_at);
CREATE INDEX idx_marketing_campaigns_target_tags ON marketing_campaigns USING GIN(target_tags);

-- +migrate Down
DROP TABLE IF EXISTS marketing_campaigns;
