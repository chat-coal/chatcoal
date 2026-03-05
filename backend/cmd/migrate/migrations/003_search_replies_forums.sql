-- +goose Up

-- Message search: fulltext index on content
ALTER TABLE messages ADD FULLTEXT INDEX ft_messages_content (content);

-- Inline replies: reply_to_id column
ALTER TABLE messages ADD COLUMN reply_to_id BIGINT UNSIGNED NULL AFTER author_id;
ALTER TABLE messages ADD INDEX idx_messages_reply_to (reply_to_id);
ALTER TABLE messages ADD CONSTRAINT fk_messages_reply_to FOREIGN KEY (reply_to_id) REFERENCES messages(id) ON DELETE SET NULL;

-- Forum posts table
CREATE TABLE forum_posts (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  title VARCHAR(200) NOT NULL,
  content TEXT NOT NULL,
  channel_id BIGINT UNSIGNED NOT NULL,
  author_id BIGINT UNSIGNED NOT NULL,
  reply_count INT UNSIGNED DEFAULT 0,
  last_reply_at DATETIME(3) NULL,
  created_at DATETIME(3),
  updated_at DATETIME(3),
  CONSTRAINT fk_forum_posts_channel FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
  CONSTRAINT fk_forum_posts_author FOREIGN KEY (author_id) REFERENCES users(id),
  INDEX idx_forum_posts_channel (channel_id)
);

-- Forum post messages: link messages to forum posts
ALTER TABLE messages ADD COLUMN forum_post_id BIGINT UNSIGNED NULL;
ALTER TABLE messages ADD INDEX idx_messages_forum_post (forum_post_id);
ALTER TABLE messages ADD CONSTRAINT fk_messages_forum_post FOREIGN KEY (forum_post_id) REFERENCES forum_posts(id) ON DELETE CASCADE;

-- +goose Down

ALTER TABLE messages DROP FOREIGN KEY fk_messages_forum_post;
ALTER TABLE messages DROP INDEX idx_messages_forum_post;
ALTER TABLE messages DROP COLUMN forum_post_id;

DROP TABLE IF EXISTS forum_posts;

ALTER TABLE messages DROP FOREIGN KEY fk_messages_reply_to;
ALTER TABLE messages DROP INDEX idx_messages_reply_to;
ALTER TABLE messages DROP COLUMN reply_to_id;

ALTER TABLE messages DROP INDEX ft_messages_content;
