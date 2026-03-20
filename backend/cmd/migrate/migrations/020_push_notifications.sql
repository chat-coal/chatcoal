-- +goose Up
CREATE TABLE device_tokens (
    id BIGINT NOT NULL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    token VARCHAR(500) NOT NULL,
    platform VARCHAR(10) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY idx_device_tokens_token (token),
    INDEX idx_device_tokens_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

ALTER TABLE notification_settings ADD COLUMN notify_mode VARCHAR(20) NOT NULL DEFAULT 'all';

-- +goose Down
ALTER TABLE notification_settings DROP COLUMN notify_mode;
DROP TABLE IF EXISTS device_tokens;
