package models

import "time"

type Server struct {
	ID              Snowflake  `json:"id" gorm:"primaryKey;autoIncrement:false"`
	Name            string     `json:"name" gorm:"size:100;not null"`
	IconURL         string     `json:"icon_url" gorm:"size:500"`
	OwnerID         Snowflake  `json:"owner_id" gorm:"not null"`
	Owner           User       `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	InviteCode      string     `json:"invite_code" gorm:"uniqueIndex;size:20;not null"`
	IsPublic        bool       `json:"is_public" gorm:"default:false"`
	ShowJoinLeave   bool       `json:"show_join_leave" gorm:"default:false"`
	SystemChannelID *Snowflake `json:"system_channel_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// PublicServer is a lightweight DTO returned by the Explore endpoint.
type PublicServer struct {
	ID          Snowflake `json:"id"`
	Name        string    `json:"name"`
	IconURL     string    `json:"icon_url"`
	MemberCount int64     `json:"member_count"`
}
