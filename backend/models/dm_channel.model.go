package models

import "time"

type DMChannel struct {
	ID        Snowflake `json:"id" gorm:"primaryKey;autoIncrement:false"`
	User1ID   Snowflake `json:"user1_id" gorm:"not null;uniqueIndex:idx_dm_users"`
	User1     User      `json:"user1,omitempty" gorm:"foreignKey:User1ID"`
	User2ID   Snowflake `json:"user2_id" gorm:"not null;uniqueIndex:idx_dm_users"`
	User2     User      `json:"user2,omitempty" gorm:"foreignKey:User2ID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
