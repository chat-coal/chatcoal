-- +goose Up
ALTER TABLE `users` ADD COLUMN `is_anonymous` tinyint(1) NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE `users` DROP COLUMN `is_anonymous`;
