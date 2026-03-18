package models

import "time"

type UserBlock struct {
	ID        Snowflake `json:"id" gorm:"primaryKey;autoIncrement:false"`
	BlockerID Snowflake `json:"blocker_id" gorm:"not null"`
	BlockedID Snowflake `json:"blocked_id" gorm:"not null"`
	Blocked   User      `json:"blocked,omitempty" gorm:"foreignKey:BlockedID"`
	CreatedAt time.Time `json:"created_at"`
}
