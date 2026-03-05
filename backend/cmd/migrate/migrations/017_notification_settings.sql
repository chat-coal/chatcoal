-- +goose Up
CREATE TABLE notification_settings (
  id          BIGINT PRIMARY KEY,
  user_id     BIGINT NOT NULL,
  target_type VARCHAR(10) NOT NULL,
  target_id   BIGINT NOT NULL,
  muted       BOOLEAN NOT NULL DEFAULT FALSE,
  created_at  DATETIME(3),
  updated_at  DATETIME(3),
  UNIQUE KEY idx_notif_user_target (user_id, target_type, target_id),
  KEY idx_notif_user (user_id)
);

-- +goose Down
DROP TABLE IF EXISTS notification_settings;
