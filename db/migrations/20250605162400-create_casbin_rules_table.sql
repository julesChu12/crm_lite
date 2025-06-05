-- +migrate Up
CREATE TABLE IF NOT EXISTS casbin_rules (
    id SERIAL PRIMARY KEY,
    ptype VARCHAR(100) NOT NULL,
    v0 VARCHAR(100),
    v1 VARCHAR(100),
    v2 VARCHAR(100),
    v3 VARCHAR(100),
    v4 VARCHAR(100),
    v5 VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引以提高查询性能
CREATE INDEX idx_casbin_rules_ptype ON casbin_rules(ptype);
CREATE INDEX idx_casbin_rules_v0 ON casbin_rules(v0);
CREATE INDEX idx_casbin_rules_v1 ON casbin_rules(v1);

-- +migrate Down
DROP TABLE IF EXISTS casbin_rules;
