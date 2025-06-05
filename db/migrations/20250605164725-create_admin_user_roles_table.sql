
-- +migrate Up
CREATE TABLE IF NOT EXISTS admin_user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_user_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(admin_user_id, role_id)
);

-- 创建索引
CREATE INDEX idx_admin_user_roles_admin_user_id ON admin_user_roles(admin_user_id);
CREATE INDEX idx_admin_user_roles_role_id ON admin_user_roles(role_id);
-- +migrate Down

DROP TABLE IF EXISTS admin_user_roles;
