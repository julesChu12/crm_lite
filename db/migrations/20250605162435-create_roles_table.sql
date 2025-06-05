-- +migrate Up
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 创建索引
CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_is_active ON roles(is_active);
CREATE INDEX idx_roles_deleted_at ON roles(deleted_at);

-- 插入默认角色
INSERT INTO roles (name, display_name, description) VALUES 
('super_admin', '超级管理员', '系统超级管理员，拥有所有权限'),
('admin', '管理员', '系统管理员，拥有大部分管理权限'),
('manager', '经理', '业务经理，可管理客户和订单'),
('staff', '员工', '普通员工，基础操作权限');

-- +migrate Down
DROP TABLE IF EXISTS roles;
