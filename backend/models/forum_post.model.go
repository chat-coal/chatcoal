package models

import "time"

type ForumPost struct {
	ID          Snowflake  `json:"id" gorm:"primaryKey;autoIncrement:false"`
	Title       string     `json:"title" gorm:"size:200;not null"`
	Content     string     `json:"content" gorm:"type:text;not null"`
	ChannelID   Snowflake  `json:"channel_id" gorm:"not null;index"`
	Channel     Channel    `json:"-" gorm:"foreignKey:ChannelID"`
	AuthorID    Snowflake  `json:"author_id" gorm:"not null"`
	Author      User       `json:"author,omitempty" gorm:"foreignKey:AuthorID"`
	ReplyCount  int        `json:"reply_count" gorm:"default:0"`
	LastReplyAt *time.Time `json:"last_reply_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
