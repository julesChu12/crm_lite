-- +migrate Up
-- 创建管理员用户角色关联表
CREATE TABLE IF NOT EXISTS admin_user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_user_id UUID NOT NULL, -- 逻辑外键 -> admin_users.id
    role_id UUID NOT NULL, -- 逻辑外键 -> roles.id
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(admin_user_id, role_id) -- 确保用户和角色的唯一性
);

-- 创建索引
CREATE INDEX idx_admin_user_roles_admin_user_id ON admin_user_roles(admin_user_id);
CREATE INDEX idx_admin_user_roles_role_id ON admin_user_roles(role_id);

-- +migrate Down
DROP TABLE IF EXISTS admin_user_roles; 