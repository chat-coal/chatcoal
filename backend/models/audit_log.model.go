package models

import "time"

type AuditLog struct {
	ID        Snowflake  `json:"id"        gorm:"primaryKey;autoIncrement:false"`
	ServerID  Snowflake  `json:"server_id" gorm:"not null"`
	ActorID   Snowflake  `json:"actor_id"  gorm:"not null"`
	Action    string     `json:"action"    gorm:"size:50;not null"`
	TargetID  *Snowflake `json:"target_id"`
	Metadata  *string    `json:"metadata"  gorm:"type:json"`
	CreatedAt time.Time  `json:"created_at"`
}
