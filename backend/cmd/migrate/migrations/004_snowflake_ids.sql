-- +goose Up
-- Drop all tables and recreate with snowflake IDs (bigint, no auto_increment).
-- Dev-only migration: no data to preserve.

SET FOREIGN_KEY_CHECKS = 0;
DROP TABLE IF EXISTS `dm_message_reactions`;
DROP TABLE IF EXISTS `message_reactions`;
DROP TABLE IF EXISTS `read_states`;
DROP TABLE IF EXISTS `dm_messages`;
DROP TABLE IF EXISTS `dm_channels`;
DROP TABLE IF EXISTS `messages`;
DROP TABLE IF EXISTS `forum_posts`;
DROP TABLE IF EXISTS `invites`;
DROP TABLE IF EXISTS `server_members`;
DROP TABLE IF EXISTS `channels`;
DROP TABLE IF EXISTS `servers`;
DROP TABLE IF EXISTS `users`;
SET FOREIGN_KEY_CHECKS = 1;

CREATE TABLE `users` (
  `id` bigint NOT NULL,
  `firebase_uid` varchar(128) NOT NULL,
  `display_name` varchar(100) NOT NULL,
  `avatar_url` varchar(500) DEFAULT NULL,
  `status` varchar(20) DEFAULT 'online',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_firebase_uid` (`firebase_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `servers` (
  `id` bigint NOT NULL,
  `name` varchar(100) NOT NULL,
  `icon_url` varchar(500) DEFAULT NULL,
  `owner_id` bigint NOT NULL,
  `invite_code` varchar(20) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_servers_invite_code` (`invite_code`),
  KEY `fk_servers_owner` (`owner_id`),
  CONSTRAINT `fk_servers_owner` FOREIGN KEY (`owner_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `channels` (
  `id` bigint NOT NULL,
  `name` varchar(100) NOT NULL,
  `server_id` bigint NOT NULL,
  `type` varchar(10) DEFAULT 'text',
  `topic` varchar(1024) DEFAULT NULL,
  `position` bigint DEFAULT '0',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_channels_server_id` (`server_id`),
  CONSTRAINT `fk_channels_server` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `server_members` (
  `id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `server_id` bigint NOT NULL,
  `role` varchar(20) DEFAULT 'member',
  `joined_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_server` (`user_id`,`server_id`),
  KEY `fk_server_members_server` (`server_id`),
  CONSTRAINT `fk_server_members_server` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`),
  CONSTRAINT `fk_server_members_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `invites` (
  `id` bigint NOT NULL,
  `code` varchar(20) NOT NULL,
  `server_id` bigint NOT NULL,
  `creator_id` bigint NOT NULL,
  `max_uses` bigint DEFAULT '0',
  `uses` bigint DEFAULT '0',
  `expires_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_invites_code` (`code`),
  KEY `fk_invites_creator` (`creator_id`),
  KEY `fk_invites_server` (`server_id`),
  CONSTRAINT `fk_invites_creator` FOREIGN KEY (`creator_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_invites_server` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `messages` (
  `id` bigint NOT NULL,
  `content` text NOT NULL,
  `channel_id` bigint NOT NULL,
  `author_id` bigint NOT NULL,
  `reply_to_id` bigint DEFAULT NULL,
  `forum_post_id` bigint DEFAULT NULL,
  `edited` tinyint(1) DEFAULT '0',
  `file_url` longtext,
  `file_name` longtext,
  `file_size` bigint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_messages_channel_id` (`channel_id`),
  KEY `idx_messages_author_id` (`author_id`),
  KEY `idx_messages_reply_to` (`reply_to_id`),
  KEY `idx_messages_forum_post` (`forum_post_id`),
  KEY `idx_messages_channel_id_id` (`channel_id`, `id`),
  FULLTEXT KEY `ft_messages_content` (`content`),
  CONSTRAINT `fk_messages_author` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_messages_channel` FOREIGN KEY (`channel_id`) REFERENCES `channels` (`id`),
  CONSTRAINT `fk_messages_reply_to` FOREIGN KEY (`reply_to_id`) REFERENCES `messages` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `forum_posts` (
  `id` bigint NOT NULL,
  `title` varchar(200) NOT NULL,
  `content` text NOT NULL,
  `channel_id` bigint NOT NULL,
  `author_id` bigint NOT NULL,
  `reply_count` int unsigned DEFAULT '0',
  `last_reply_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_forum_posts_channel` (`channel_id`),
  CONSTRAINT `fk_forum_posts_channel` FOREIGN KEY (`channel_id`) REFERENCES `channels` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_forum_posts_author` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Add FK for messages.forum_post_id after forum_posts exists
ALTER TABLE `messages` ADD CONSTRAINT `fk_messages_forum_post` FOREIGN KEY (`forum_post_id`) REFERENCES `forum_posts` (`id`) ON DELETE CASCADE;

CREATE TABLE `dm_channels` (
  `id` bigint NOT NULL,
  `user1_id` bigint NOT NULL,
  `user2_id` bigint NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_dm_users` (`user1_id`,`user2_id`),
  KEY `fk_dm_channels_user2` (`user2_id`),
  CONSTRAINT `fk_dm_channels_user1` FOREIGN KEY (`user1_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_dm_channels_user2` FOREIGN KEY (`user2_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `dm_messages` (
  `id` bigint NOT NULL,
  `content` text NOT NULL,
  `dm_channel_id` bigint NOT NULL,
  `author_id` bigint NOT NULL,
  `edited` tinyint(1) DEFAULT '0',
  `file_url` longtext,
  `file_name` longtext,
  `file_size` bigint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_dm_messages_dm_channel_id` (`dm_channel_id`),
  KEY `idx_dm_messages_author_id` (`author_id`),
  KEY `idx_dm_messages_dm_channel_id_id` (`dm_channel_id`, `id`),
  CONSTRAINT `fk_dm_messages_author` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_dm_messages_dm_channel` FOREIGN KEY (`dm_channel_id`) REFERENCES `dm_channels` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `read_states` (
  `id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `channel_type` varchar(10) NOT NULL,
  `channel_ref_id` bigint NOT NULL,
  `last_read_message_id` bigint NOT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_read_state` (`user_id`,`channel_type`,`channel_ref_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `message_reactions` (
  `id` bigint NOT NULL,
  `message_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `emoji` varchar(32) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_msg_user_emoji` (`message_id`,`user_id`,`emoji`),
  KEY `idx_msg_reactions_message_user` (`message_id`, `user_id`),
  CONSTRAINT `fk_messages_reactions` FOREIGN KEY (`message_id`) REFERENCES `messages` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `dm_message_reactions` (
  `id` bigint NOT NULL,
  `dm_message_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `emoji` varchar(32) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_dm_msg_user_emoji` (`dm_message_id`,`user_id`,`emoji`),
  KEY `idx_dm_msg_reactions_dm_message_user` (`dm_message_id`, `user_id`),
  CONSTRAINT `fk_dm_messages_reactions` FOREIGN KEY (`dm_message_id`) REFERENCES `dm_messages` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- +goose Down
SET FOREIGN_KEY_CHECKS = 0;
DROP TABLE IF EXISTS `dm_message_reactions`;
DROP TABLE IF EXISTS `message_reactions`;
DROP TABLE IF EXISTS `read_states`;
DROP TABLE IF EXISTS `dm_messages`;
DROP TABLE IF EXISTS `dm_channels`;
DROP TABLE IF EXISTS `forum_posts`;
ALTER TABLE `messages` DROP FOREIGN KEY `fk_messages_forum_post`;
DROP TABLE IF EXISTS `messages`;
DROP TABLE IF EXISTS `invites`;
DROP TABLE IF EXISTS `server_members`;
DROP TABLE IF EXISTS `channels`;
DROP TABLE IF EXISTS `servers`;
DROP TABLE IF EXISTS `users`;
SET FOREIGN_KEY_CHECKS = 1;

-- Restore original auto-increment tables (from migrations 001-003)
-- This would require re-running migrations 001, 002, 003 after rolling back.
