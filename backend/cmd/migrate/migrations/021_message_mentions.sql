-- +goose Up
ALTER TABLE messages ADD COLUMN mentions JSON DEFAULT NULL;

-- +goose Down
ALTER TABLE messages DROP COLUMN mentions;
