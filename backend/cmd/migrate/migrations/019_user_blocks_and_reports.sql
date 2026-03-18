-- +goose Up
CREATE TABLE user_blocks (
    id BIGINT UNSIGNED NOT NULL,
    blocker_id BIGINT UNSIGNED NOT NULL,
    blocked_id BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE INDEX idx_blocker_blocked (blocker_id, blocked_id),
    INDEX idx_blocked_user (blocked_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE reports (
    id BIGINT UNSIGNED NOT NULL,
    reporter_id BIGINT UNSIGNED NOT NULL,
    target_type VARCHAR(20) NOT NULL,
    target_id BIGINT UNSIGNED NOT NULL,
    reason VARCHAR(1000) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    INDEX idx_reporter (reporter_id),
    INDEX idx_target (target_type, target_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +goose Down
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS user_blocks;
