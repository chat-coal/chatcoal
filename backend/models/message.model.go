package models

import (
	"encoding/json"
	"time"
)

type LinkEmbed struct {
	URL         string `json:"url"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Image       string `json:"image,omitempty"`
	SiteName    string `json:"site_name,omitempty"`
}

type Message struct {
	ID          Snowflake         `json:"id" gorm:"primaryKey;autoIncrement:false"`
	Content     string            `json:"content" gorm:"type:text;not null"`
	Type        string            `json:"type" gorm:"size:10;not null;default:'user'"`
	ChannelID   Snowflake         `json:"channel_id" gorm:"not null;index"`
	Channel     Channel           `json:"-" gorm:"foreignKey:ChannelID"`
	AuthorID    Snowflake         `json:"author_id" gorm:"not null;index"`
	Author      User              `json:"author,omitempty" gorm:"foreignKey:AuthorID"`
	ReplyToID   *Snowflake        `json:"reply_to_id,omitempty" gorm:"index"`
	ReplyTo     *Message          `json:"reply_to,omitempty" gorm:"foreignKey:ReplyToID"`
	ForumPostID *Snowflake        `json:"forum_post_id,omitempty" gorm:"index"`
	Edited      bool              `json:"edited" gorm:"default:false"`
	FileURL     string            `json:"file_url,omitempty"`
	FileName    string            `json:"file_name,omitempty"`
	FileSize    int64             `json:"file_size,omitempty"`
	ImageWidth  int               `json:"image_width,omitempty"`
	ImageHeight int               `json:"image_height,omitempty"`
	Embeds      json.RawMessage   `json:"embeds,omitempty" gorm:"type:json"`
	Reactions   []MessageReaction `json:"reactions" gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}
