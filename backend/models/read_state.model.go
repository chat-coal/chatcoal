package models

import "time"

type ReadState struct {
	ID                Snowflake `json:"id" gorm:"primaryKey;autoIncrement:false"`
	UserID            Snowflake `json:"user_id" gorm:"not null;uniqueIndex:idx_read_state"`
	ChannelType       string    `json:"channel_type" gorm:"size:10;not null;uniqueIndex:idx_read_state"` // "server" or "dm"
	ChannelRefID      Snowflake `json:"channel_ref_id" gorm:"not null;uniqueIndex:idx_read_state"`
	ServerID          Snowflake `json:"server_id,omitempty" gorm:"not null;default:0"`
	LastReadMessageID Snowflake `json:"last_read_message_id" gorm:"not null"`
	UnreadCount       int       `json:"unread_count" gorm:"not null;default:0"`
	UpdatedAt         time.Time `json:"updated_at"`
}
