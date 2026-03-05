-- +goose Up
CREATE TABLE pinned_messages (
  id bigint PRIMARY KEY,
  channel_id bigint NOT NULL,
  message_id bigint NOT NULL,
  pinned_by_id bigint NOT NULL,
  created_at datetime(3),
  UNIQUE KEY idx_pinned_channel_message (channel_id, message_id),
  FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
  FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
  FOREIGN KEY (pinned_by_id) REFERENCES users(id),
  INDEX idx_pinned_channel (channel_id)
);

-- +goose Down
DROP TABLE IF EXISTS pinned_messages;
