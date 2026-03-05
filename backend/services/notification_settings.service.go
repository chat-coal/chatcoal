package services

import (
	"chatcoal/database"
	"chatcoal/models"
)

func GetNotificationSettings(userID models.Snowflake) ([]models.NotificationSetting, error) {
	var settings []models.NotificationSetting
	err := database.Database.Where("user_id = ?", userID).Find(&settings).Error
	return settings, err
}

func UpsertNotificationSetting(userID models.Snowflake, targetType string, targetID models.Snowflake, muted bool) (*models.NotificationSetting, error) {
	var setting models.NotificationSetting
	err := database.Database.
		Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		First(&setting).Error

	if err != nil {
		// Create new
		setting = models.NotificationSetting{
			UserID:     userID,
			TargetType: targetType,
			TargetID:   targetID,
			Muted:      muted,
		}
		if err := database.Database.Create(&setting).Error; err != nil {
			return nil, err
		}
		return &setting, nil
	}

	// Update existing
	setting.Muted = muted
	if err := database.Database.Save(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

// GetMutedUserIDsForChannel returns user IDs that have muted either the channel or its server.
func GetMutedUserIDsForChannel(channelID, serverID models.Snowflake) map[models.Snowflake]bool {
	var settings []models.NotificationSetting
	database.Database.
		Where("muted = ? AND ((target_type = ? AND target_id = ?) OR (target_type = ? AND target_id = ?))",
			true, "channel", channelID, "server", serverID).
		Find(&settings)

	result := make(map[models.Snowflake]bool, len(settings))
	for _, s := range settings {
		result[s.UserID] = true
	}
	return result
}
