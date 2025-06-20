-- +migrate Up
CREATE TABLE IF NOT EXISTS admin_users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uuid VARCHAR(36) NOT NULL,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    real_name VARCHAR(50),
    phone VARCHAR(20),
    avatar VARCHAR(255),
    is_active TINYINT(1) DEFAULT 1,
    last_login_at DATETIME(6) NULL,
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME(6) NULL,
    UNIQUE KEY (uuid),
    UNIQUE KEY (username),
    UNIQUE KEY (email)
);

-- 创建索引
CREATE INDEX idx_admin_users_deleted_at ON admin_users(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS admin_users;
