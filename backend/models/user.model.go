package models

import "time"

type User struct {
	ID           Snowflake `json:"id" gorm:"primaryKey;autoIncrement:false"`
	FirebaseUID  string    `json:"firebase_uid" gorm:"uniqueIndex;size:128;not null"`
	Username     *string   `json:"username" gorm:"size:32;uniqueIndex"`
	DisplayName  string    `json:"display_name" gorm:"size:100;not null"`
	AvatarURL    string    `json:"avatar_url" gorm:"size:500"`
	Status       string    `json:"status" gorm:"size:20;default:online"`
	IsAnonymous   bool      `json:"is_anonymous" gorm:"default:false"`
	EmailVerified bool      `json:"email_verified" gorm:"default:true"`
	HomeInstance  *string   `json:"home_instance" gorm:"size:255"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// IsRestricted returns true if the user should have limited capabilities
// (anonymous users and email/password users who haven't verified their email).
func (u *User) IsRestricted() bool {
	return u.IsAnonymous || !u.EmailVerified
}
