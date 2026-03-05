package models

import "time"

type NotificationSetting struct {
	ID         Snowflake `json:"id" gorm:"primaryKey;autoIncrement:false"`
	UserID     Snowflake `json:"user_id" gorm:"not null"`
	TargetType string    `json:"target_type" gorm:"size:10;not null"`
	TargetID   Snowflake `json:"target_id" gorm:"not null"`
	Muted      bool      `json:"muted" gorm:"not null;default:false"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
