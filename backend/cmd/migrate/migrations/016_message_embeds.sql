-- +goose Up
ALTER TABLE messages ADD COLUMN embeds JSON DEFAULT NULL;
ALTER TABLE dm_messages ADD COLUMN embeds JSON DEFAULT NULL;

-- +goose Down
ALTER TABLE dm_messages DROP COLUMN embeds;
ALTER TABLE messages DROP COLUMN embeds;
