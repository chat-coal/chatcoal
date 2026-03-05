package models

import "time"

// FederatedInstance caches public keys for remote chatcoal instances.
type FederatedInstance struct {
	ID        Snowflake `json:"id" gorm:"primaryKey;autoIncrement:false"`
	Domain    string    `json:"domain" gorm:"uniqueIndex;size:255;not null"`
	PublicKey string    `json:"public_key" gorm:"type:text;not null"`
	Name      string    `json:"name" gorm:"size:100"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// InstanceConfig holds this instance's Ed25519 keypair (single row, id=1).
type InstanceConfig struct {
	ID            int    `gorm:"primaryKey;autoIncrement:false"`
	PrivateKey    string `gorm:"type:text;not null"`
	PublicKey     string `gorm:"type:text;not null"`
	Domain        string `gorm:"size:255;not null"`
	DefaultPolicy string `gorm:"size:10;not null;default:open"`
}

// InstancePolicy represents an allow/block rule for a remote instance domain.
type InstancePolicy struct {
	ID        Snowflake `json:"id" gorm:"primaryKey;autoIncrement:false"`
	Domain    string    `json:"domain" gorm:"uniqueIndex;size:255;not null"`
	Policy    string    `json:"policy" gorm:"size:10;not null"`
	Note      string    `json:"note" gorm:"size:500;default:''"`
	CreatedBy Snowflake `json:"created_by" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FederatedChannelLink maps a local channel to a remote federated channel.
type FederatedChannelLink struct {
	ID                 Snowflake `json:"id" gorm:"primaryKey;autoIncrement:false"`
	ChannelID          Snowflake `json:"channel_id" gorm:"not null"`
	RemoteDomain       string    `json:"remote_domain" gorm:"size:255;not null"`
	RemoteFederationID string    `json:"remote_federation_id" gorm:"size:64;not null;index:idx_remote_fed"`
	Active             bool      `json:"active" gorm:"not null;default:true"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
