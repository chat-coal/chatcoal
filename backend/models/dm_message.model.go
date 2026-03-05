package models

import (
	"encoding/json"
	"time"
)

type DMMessage struct {
	ID          Snowflake             `json:"id" gorm:"primaryKey;autoIncrement:false"`
	Content     string                `json:"content" gorm:"type:text;not null"`
	DMChannelID Snowflake             `json:"dm_channel_id" gorm:"not null;index"`
	DMChannel   DMChannel             `json:"-" gorm:"foreignKey:DMChannelID"`
	AuthorID    Snowflake             `json:"author_id" gorm:"not null;index"`
	Author      User                  `json:"author,omitempty" gorm:"foreignKey:AuthorID"`
	Edited      bool                  `json:"edited" gorm:"default:false"`
	FileURL     string                `json:"file_url,omitempty"`
	FileName    string                `json:"file_name,omitempty"`
	FileSize    int64                 `json:"file_size,omitempty"`
	ImageWidth  int                   `json:"image_width,omitempty"`
	ImageHeight int                   `json:"image_height,omitempty"`
	Embeds      json.RawMessage       `json:"embeds,omitempty" gorm:"type:json"`
	Reactions   []DMMessageReaction   `json:"reactions" gorm:"foreignKey:DMMessageID;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}
