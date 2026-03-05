package services

import (
	"crypto/rand"
	"chatcoal/database"
	"chatcoal/models"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
)

func generateInviteCode10() string {
	b := make([]byte, 5)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func CreateInvite(serverID, creatorID models.Snowflake, maxUses int, expiresIn int) (*models.Invite, error) {
	invite := models.Invite{
		Code:      generateInviteCode10(),
		ServerID:  serverID,
		CreatorID: creatorID,
		MaxUses:   maxUses,
	}
	if expiresIn > 0 {
		t := time.Now().Add(time.Duration(expiresIn) * time.Second)
		invite.ExpiresAt = &t
	}

	if err := database.Database.Create(&invite).Error; err != nil {
		return nil, err
	}

	// Preload creator
	database.Database.Preload("Creator").First(&invite, invite.ID)
	return &invite, nil
}

func CreateInviteInTx(tx *gorm.DB, serverID, creatorID models.Snowflake, code string) error {
	invite := models.Invite{
		Code:      code,
		ServerID:  serverID,
		CreatorID: creatorID,
		MaxUses:   0,
	}
	return tx.Create(&invite).Error
}

const InvitePageLimit = 50

func GetInvitesByServerID(serverID models.Snowflake, before models.Snowflake) ([]models.Invite, error) {
	var invites []models.Invite
	query := database.Database.Preload("Creator").
		Where("server_id = ?", serverID)

	if before > 0 {
		query = query.Where("id < ?", before)
	}

	err := query.Order("id DESC").Limit(InvitePageLimit).Find(&invites).Error
	return invites, err
}

func GetInviteByCode(code string) (*models.Invite, error) {
	var invite models.Invite
	if err := database.Database.Preload("Creator").Preload("Server").Where("code = ?", code).First(&invite).Error; err != nil {
		return nil, err
	}

	// Check expiration
	if invite.ExpiresAt != nil && invite.ExpiresAt.Before(time.Now()) {
		return nil, gorm.ErrRecordNotFound
	}

	// Check max uses
	if invite.MaxUses > 0 && invite.Uses >= invite.MaxUses {
		return nil, gorm.ErrRecordNotFound
	}

	return &invite, nil
}

func UseInvite(inviteID models.Snowflake) error {
	return database.Database.Model(&models.Invite{}).
		Where("id = ?", inviteID).
		UpdateColumn("uses", gorm.Expr("uses + 1")).Error
}

func DeleteInvite(inviteID, serverID models.Snowflake) error {
	return database.Database.Where("id = ? AND server_id = ?", inviteID, serverID).Delete(&models.Invite{}).Error
}
