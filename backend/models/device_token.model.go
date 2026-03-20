package models

import "time"

type DeviceToken struct {
	ID        Snowflake `json:"id" gorm:"primaryKey;autoIncrement:false"`
	UserID    Snowflake `json:"user_id" gorm:"not null;index"`
	Token     string    `json:"token" gorm:"size:500;not null;uniqueIndex"`
	Platform  string    `json:"platform" gorm:"size:10;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
