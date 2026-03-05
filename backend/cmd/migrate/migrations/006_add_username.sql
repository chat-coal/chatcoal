-- +goose Up
ALTER TABLE users ADD COLUMN username VARCHAR(32) DEFAULT NULL;
CREATE UNIQUE INDEX idx_users_username ON users (username);

-- +goose Down
DROP INDEX idx_users_username ON users;
ALTER TABLE users DROP COLUMN username;
