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

func UpsertNotificationSetting(userID models.Snowflake, targetType string, targetID models.Snowflake, muted bool, notifyMode string) (*models.NotificationSetting, error) {
	// Sync muted and notify_mode
	if notifyMode == "" {
		if muted {
			notifyMode = "nothing"
		} else {
			notifyMode = "all"
		}
	}
	if notifyMode == "nothing" {
		muted = true
	} else if muted && notifyMode == "all" {
		notifyMode = "nothing"
	}

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
			NotifyMode: notifyMode,
		}
		if err := database.Database.Create(&setting).Error; err != nil {
			return nil, err
		}
		return &setting, nil
	}

	// Update existing
	setting.Muted = muted
	setting.NotifyMode = notifyMode
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

// GetNotifyModesForChannel returns the notify_mode for each user that has a setting
// for the given channel or its server. Users not in the map use the default ("all").
func GetNotifyModesForChannel(channelID, serverID models.Snowflake) map[models.Snowflake]string {
	var settings []models.NotificationSetting
	database.Database.
		Where("(target_type = ? AND target_id = ?) OR (target_type = ? AND target_id = ?)",
			"channel", channelID, "server", serverID).
		Find(&settings)

	result := make(map[models.Snowflake]string, len(settings))
	for _, s := range settings {
		// Channel-level setting overrides server-level
		if s.TargetType == "channel" {
			result[s.UserID] = s.NotifyMode
		} else if _, exists := result[s.UserID]; !exists {
			result[s.UserID] = s.NotifyMode
		}
	}
	return result
}
