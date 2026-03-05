-- +goose Up
ALTER TABLE messages ADD COLUMN image_width INT NULL, ADD COLUMN image_height INT NULL;
ALTER TABLE dm_messages ADD COLUMN image_width INT NULL, ADD COLUMN image_height INT NULL;

-- +goose Down
ALTER TABLE messages DROP COLUMN image_width, DROP COLUMN image_height;
ALTER TABLE dm_messages DROP COLUMN image_width, DROP COLUMN image_height;
