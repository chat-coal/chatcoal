package models

import "time"

type ServerBan struct {
	ID        Snowflake `json:"id" gorm:"primaryKey;autoIncrement:false"`
	ServerID  Snowflake `json:"server_id" gorm:"not null"`
	UserID    Snowflake `json:"user_id" gorm:"not null"`
	User      User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	BannedBy  Snowflake `json:"banned_by" gorm:"not null"`
	Reason    *string   `json:"reason" gorm:"size:512"`
	CreatedAt time.Time `json:"created_at"`
}
