-- +goose Up
ALTER TABLE servers ADD COLUMN is_public BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE servers DROP COLUMN is_public;
