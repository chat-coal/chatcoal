package models

import "time"

type MessageReaction struct {
	ID        Snowflake `json:"id" gorm:"primaryKey;autoIncrement:false"`
	MessageID Snowflake `json:"message_id" gorm:"not null;uniqueIndex:idx_msg_user_emoji"`
	UserID    Snowflake `json:"user_id" gorm:"not null;uniqueIndex:idx_msg_user_emoji"`
	Emoji     string    `json:"emoji" gorm:"type:varchar(32);not null;uniqueIndex:idx_msg_user_emoji"`
	CreatedAt time.Time `json:"created_at"`
}

type DMMessageReaction struct {
	ID          Snowflake `json:"id" gorm:"primaryKey;autoIncrement:false"`
	DMMessageID Snowflake `json:"dm_message_id" gorm:"not null;uniqueIndex:idx_dm_msg_user_emoji"`
	UserID      Snowflake `json:"user_id" gorm:"not null;uniqueIndex:idx_dm_msg_user_emoji"`
	Emoji       string    `json:"emoji" gorm:"type:varchar(32);not null;uniqueIndex:idx_dm_msg_user_emoji"`
	CreatedAt   time.Time `json:"created_at"`
}
