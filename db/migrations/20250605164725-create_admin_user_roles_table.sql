-- +migrate Up
CREATE TABLE IF NOT EXISTS admin_user_roles (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    admin_user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY (admin_user_id, role_id)
);

-- 创建索引
CREATE INDEX idx_admin_user_roles_admin_user_id ON admin_user_roles(admin_user_id);
CREATE INDEX idx_admin_user_roles_role_id ON admin_user_roles(role_id);

-- +migrate Down
DROP TABLE IF EXISTS admin_user_roles;
