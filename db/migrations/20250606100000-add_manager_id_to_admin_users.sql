-- +migrate Up
ALTER TABLE `admin_users`
ADD COLUMN `manager_id` BIGINT NULL COMMENT '上级经理ID' AFTER `avatar`,
ADD INDEX `idx_manager_id` (`manager_id`);

-- +migrate Down
ALTER TABLE `admin_users`
DROP COLUMN `manager_id`,
DROP INDEX `idx_manager_id`; 