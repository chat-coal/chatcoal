package models

import "time"

type ServerMember struct {
	ID       Snowflake `json:"id" gorm:"primaryKey;autoIncrement:false"`
	UserID   Snowflake `json:"user_id" gorm:"not null;uniqueIndex:idx_user_server"`
	User     User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ServerID Snowflake `json:"server_id" gorm:"not null;uniqueIndex:idx_user_server"`
	Server   Server    `json:"-" gorm:"foreignKey:ServerID"`
	Role     string    `json:"role" gorm:"size:20;default:member"`
	JoinedAt time.Time `json:"joined_at"`
}
