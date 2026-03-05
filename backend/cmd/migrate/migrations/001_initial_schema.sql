-- +goose Up
CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `firebase_uid` varchar(128) NOT NULL,
  `display_name` varchar(100) NOT NULL,
  `avatar_url` varchar(500) DEFAULT NULL,
  `status` varchar(20) DEFAULT 'online',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_firebase_uid` (`firebase_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `servers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `icon_url` varchar(500) DEFAULT NULL,
  `owner_id` bigint unsigned NOT NULL,
  `invite_code` varchar(20) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_servers_invite_code` (`invite_code`),
  KEY `fk_servers_owner` (`owner_id`),
  CONSTRAINT `fk_servers_owner` FOREIGN KEY (`owner_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `channels` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `server_id` bigint unsigned NOT NULL,
  `type` varchar(10) DEFAULT 'text',
  `topic` varchar(1024) DEFAULT NULL,
  `position` bigint DEFAULT '0',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_channels_server_id` (`server_id`),
  CONSTRAINT `fk_channels_server` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `server_members` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `server_id` bigint unsigned NOT NULL,
  `role` varchar(20) DEFAULT 'member',
  `joined_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_server` (`user_id`,`server_id`),
  KEY `fk_server_members_server` (`server_id`),
  CONSTRAINT `fk_server_members_server` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`),
  CONSTRAINT `fk_server_members_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `invites` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(20) NOT NULL,
  `server_id` bigint unsigned NOT NULL,
  `creator_id` bigint unsigned NOT NULL,
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

CREATE TABLE IF NOT EXISTS `messages` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `content` text NOT NULL,
  `channel_id` bigint unsigned NOT NULL,
  `author_id` bigint unsigned NOT NULL,
  `edited` tinyint(1) DEFAULT '0',
  `file_url` longtext,
  `file_name` longtext,
  `file_size` bigint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_messages_channel_id` (`channel_id`),
  KEY `idx_messages_author_id` (`author_id`),
  CONSTRAINT `fk_messages_author` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_messages_channel` FOREIGN KEY (`channel_id`) REFERENCES `channels` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `dm_channels` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user1_id` bigint unsigned NOT NULL,
  `user2_id` bigint unsigned NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_dm_users` (`user1_id`,`user2_id`),
  KEY `fk_dm_channels_user2` (`user2_id`),
  CONSTRAINT `fk_dm_channels_user1` FOREIGN KEY (`user1_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_dm_channels_user2` FOREIGN KEY (`user2_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `dm_messages` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `content` text NOT NULL,
  `dm_channel_id` bigint unsigned NOT NULL,
  `author_id` bigint unsigned NOT NULL,
  `edited` tinyint(1) DEFAULT '0',
  `file_url` longtext,
  `file_name` longtext,
  `file_size` bigint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_dm_messages_dm_channel_id` (`dm_channel_id`),
  KEY `idx_dm_messages_author_id` (`author_id`),
  CONSTRAINT `fk_dm_messages_author` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_dm_messages_dm_channel` FOREIGN KEY (`dm_channel_id`) REFERENCES `dm_channels` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `read_states` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `channel_type` varchar(10) NOT NULL,
  `channel_ref_id` bigint unsigned NOT NULL,
  `last_read_message_id` bigint unsigned NOT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_read_state` (`user_id`,`channel_type`,`channel_ref_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `message_reactions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `message_id` bigint unsigned NOT NULL,
  `user_id` bigint unsigned NOT NULL,
  `emoji` varchar(32) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_msg_user_emoji` (`message_id`,`user_id`,`emoji`),
  CONSTRAINT `fk_messages_reactions` FOREIGN KEY (`message_id`) REFERENCES `messages` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `dm_message_reactions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `dm_message_id` bigint unsigned NOT NULL,
  `user_id` bigint unsigned NOT NULL,
  `emoji` varchar(32) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_dm_msg_user_emoji` (`dm_message_id`,`user_id`,`emoji`),
  CONSTRAINT `fk_dm_messages_reactions` FOREIGN KEY (`dm_message_id`) REFERENCES `dm_messages` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- +goose Down
SET FOREIGN_KEY_CHECKS = 0;
DROP TABLE IF EXISTS `dm_message_reactions`;
DROP TABLE IF EXISTS `message_reactions`;
DROP TABLE IF EXISTS `read_states`;
DROP TABLE IF EXISTS `dm_messages`;
DROP TABLE IF EXISTS `dm_channels`;
DROP TABLE IF EXISTS `messages`;
DROP TABLE IF EXISTS `invites`;
DROP TABLE IF EXISTS `server_members`;
DROP TABLE IF EXISTS `channels`;
DROP TABLE IF EXISTS `servers`;
DROP TABLE IF EXISTS `users`;
SET FOREIGN_KEY_CHECKS = 1;
