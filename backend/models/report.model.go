package models

import "time"

type Report struct {
	ID         Snowflake `json:"id" gorm:"primaryKey;autoIncrement:false"`
	ReporterID Snowflake `json:"reporter_id" gorm:"not null"`
	TargetType string    `json:"target_type" gorm:"size:20;not null"`
	TargetID   Snowflake `json:"target_id" gorm:"not null"`
	Reason     string    `json:"reason" gorm:"size:1000;not null"`
	Status     string    `json:"status" gorm:"size:20;not null;default:pending"`
	CreatedAt  time.Time `json:"created_at"`
}
