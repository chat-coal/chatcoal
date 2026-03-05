-- +goose Up
ALTER TABLE messages ADD COLUMN type VARCHAR(10) NOT NULL DEFAULT 'user';
ALTER TABLE servers ADD COLUMN show_join_leave BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE servers ADD COLUMN system_channel_id BIGINT NULL;

-- +goose Down
ALTER TABLE servers DROP COLUMN system_channel_id;
ALTER TABLE servers DROP COLUMN show_join_leave;
ALTER TABLE messages DROP COLUMN type;
