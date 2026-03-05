-- +goose Up
CREATE TABLE server_bans (
  id BIGINT NOT NULL,
  server_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  banned_by BIGINT NOT NULL,
  reason VARCHAR(512) DEFAULT NULL,
  created_at DATETIME(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY idx_server_bans_server_user (server_id, user_id),
  KEY fk_server_bans_server (server_id),
  KEY fk_server_bans_user (user_id),
  KEY fk_server_bans_banned_by (banned_by),
  CONSTRAINT fk_server_bans_server FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
  CONSTRAINT fk_server_bans_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_server_bans_banned_by FOREIGN KEY (banned_by) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- +goose Down
DROP TABLE IF EXISTS server_bans;
