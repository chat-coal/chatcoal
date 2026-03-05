package services

import (
	"chatcoal/database"
	"chatcoal/models"

	"gorm.io/gorm"
)

func GetPinnedMessages(channelID models.Snowflake) ([]models.PinnedMessage, error) {
	var pins []models.PinnedMessage
	err := database.Database.
		Where("channel_id = ?", channelID).
		Order("id DESC").
		Preload("PinnedBy").
		Preload("Message", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Author").Preload("Reactions")
		}).
		Find(&pins).Error
	return pins, err
}

func PinMessage(messageID, channelID, pinnedByID models.Snowflake) (*models.PinnedMessage, error) {
	var msg models.Message
	if err := database.Database.Where("id = ? AND channel_id = ?", messageID, channelID).First(&msg).Error; err != nil {
		return nil, newValidationError("message not found in channel")
	}

	var existing models.PinnedMessage
	if err := database.Database.Where("channel_id = ? AND message_id = ?", channelID, messageID).First(&existing).Error; err == nil {
		return nil, newValidationError("message already pinned")
	}

	pin := models.PinnedMessage{
		ChannelID:  channelID,
		MessageID:  messageID,
		PinnedByID: pinnedByID,
	}
	if err := database.Database.Create(&pin).Error; err != nil {
		return nil, err
	}

	database.Database.
		Preload("PinnedBy").
		Preload("Message", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Author").Preload("Reactions")
		}).
		First(&pin, pin.ID)

	return &pin, nil
}

func UnpinMessage(messageID, channelID models.Snowflake) error {
	return database.Database.
		Where("channel_id = ? AND message_id = ?", channelID, messageID).
		Delete(&models.PinnedMessage{}).Error
}
