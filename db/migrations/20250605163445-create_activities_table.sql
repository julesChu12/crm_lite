-- +migrate Up
CREATE TABLE IF NOT EXISTS activities (
    id VARCHAR(36) PRIMARY KEY,
    customer_id VARCHAR(36) NOT NULL,
    contact_id VARCHAR(36) COMMENT '具体联系人',
    type VARCHAR(20) NOT NULL COMMENT '活动类型: call, meeting, email, visit, follow_up, complaint, feedback',
    title VARCHAR(200) NOT NULL,
    content TEXT,
    status VARCHAR(20) DEFAULT 'planned' COMMENT '活动状态: planned, in_progress, completed, cancelled',
    priority VARCHAR(10) DEFAULT 'medium' COMMENT '优先级: low, medium, high, urgent',
    scheduled_at DATETIME(6) NULL,
    completed_at DATETIME(6) NULL,
    assigned_to VARCHAR(36) COMMENT '负责人',
    created_by VARCHAR(36) COMMENT '创建人',
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME(6) NULL
);

-- 创建索引
CREATE INDEX idx_activities_customer_id ON activities(customer_id);
CREATE INDEX idx_activities_contact_id ON activities(contact_id);
CREATE INDEX idx_activities_type ON activities(type);
CREATE INDEX idx_activities_status ON activities(status);
CREATE INDEX idx_activities_priority ON activities(priority);
CREATE INDEX idx_activities_scheduled_at ON activities(scheduled_at);
CREATE INDEX idx_activities_assigned_to ON activities(assigned_to);
CREATE INDEX idx_activities_created_by ON activities(created_by);
CREATE INDEX idx_activities_deleted_at ON activities(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS activities;
