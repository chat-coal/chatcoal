package models

import "time"

type Invite struct {
	ID        Snowflake  `json:"id" gorm:"primaryKey;autoIncrement:false"`
	Code      string     `json:"code" gorm:"uniqueIndex;size:20;not null"`
	ServerID  Snowflake  `json:"server_id" gorm:"not null"`
	Server    Server     `json:"server,omitempty" gorm:"foreignKey:ServerID"`
	CreatorID Snowflake  `json:"creator_id" gorm:"not null"`
	Creator   User       `json:"creator,omitempty" gorm:"foreignKey:CreatorID"`
	MaxUses   int        `json:"max_uses" gorm:"default:0"`
	Uses      int        `json:"uses" gorm:"default:0"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}
