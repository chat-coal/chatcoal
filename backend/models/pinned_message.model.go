package models

import "time"

type PinnedMessage struct {
	ID         Snowflake `json:"id" gorm:"primaryKey;autoIncrement:false"`
	ChannelID  Snowflake `json:"channel_id" gorm:"not null;index"`
	MessageID  Snowflake `json:"message_id" gorm:"not null;uniqueIndex:idx_pinned_channel_message"`
	PinnedByID Snowflake `json:"pinned_by_id" gorm:"not null"`
	PinnedBy   User      `json:"pinned_by,omitempty" gorm:"foreignKey:PinnedByID"`
	Message    *Message  `json:"message,omitempty" gorm:"foreignKey:MessageID"`
	CreatedAt  time.Time `json:"created_at"`
}
