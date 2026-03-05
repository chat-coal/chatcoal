-- +goose Up
-- Add counter columns so GetAllUnreadCounts is a simple indexed scan
-- instead of an expensive multi-table UNION query.
ALTER TABLE read_states
    ADD COLUMN unread_count INT NOT NULL DEFAULT 0,
    ADD COLUMN server_id    BIGINT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE read_states
    DROP COLUMN unread_count,
    DROP COLUMN server_id;
