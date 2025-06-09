-- +migrate Up
CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active TINYINT(1) DEFAULT 1,
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME(6) NULL,
    UNIQUE KEY (name)
);

-- 创建索引
CREATE INDEX idx_roles_deleted_at ON roles(deleted_at);

-- 插入默认角色
INSERT INTO roles (id, name, display_name, description) VALUES 
('c175f134-4536-4b8c-8514-8025234a1b9b', 'super_admin', '超级管理员', '系统超级管理员，拥有所有权限'),
('f8e3f49b-7c3e-4b5a-9a9b-1b7e6b8a2c1d', 'admin', '管理员', '系统管理员，拥有大部分管理权限'),
('a3d2e1c3-4b5a-6c7d-8e9f-0a1b2c3d4e5f', 'manager', '经理', '业务经理，可管理客户和订单'),
('b4e5f6a7-8b9c-0d1e-2f3a-4b5c6d7e8f9a', 'staff', '员工', '普通员工，基础操作权限');

-- +migrate Down
DROP TABLE IF EXISTS roles;
