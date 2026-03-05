package models

import "time"

type Channel struct {
	ID        Snowflake `json:"id" gorm:"primaryKey;autoIncrement:false"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	ServerID  Snowflake `json:"server_id" gorm:"not null;index"`
	Server    Server    `json:"-" gorm:"foreignKey:ServerID"`
	Type      string    `json:"type" gorm:"size:10;default:text"`
	Topic     string    `json:"topic" gorm:"size:1024"`
	Position     int       `json:"position" gorm:"default:0"`
	FederationID *string   `json:"federation_id,omitempty" gorm:"size:64;uniqueIndex"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
