-- +goose Up
CREATE TABLE audit_logs (
  id         BIGINT PRIMARY KEY,
  server_id  BIGINT NOT NULL,
  actor_id   BIGINT NOT NULL,
  action     VARCHAR(50) NOT NULL,
  target_id  BIGINT DEFAULT NULL,
  metadata   JSON DEFAULT NULL,
  created_at DATETIME(3),
  KEY idx_audit_server (server_id),
  KEY idx_audit_actor  (actor_id)
);

-- +goose Down
DROP TABLE IF EXISTS audit_logs;
