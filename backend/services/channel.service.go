package services

import (
	"chatcoal/cache"
	"chatcoal/database"
	"chatcoal/models"
)

func GetChannelsByServerID(serverID models.Snowflake) ([]models.Channel, error) {
	if cached := cache.GetServerChannels(serverID); cached != nil {
		return cached, nil
	}
	var channels []models.Channel
	err := database.Database.Where("server_id = ?", serverID).Order("position ASC, id ASC").Find(&channels).Error
	if err != nil {
		return nil, err
	}
	cache.SetServerChannels(serverID, channels)
	return channels, nil
}

func GetChannelByID(id models.Snowflake) (*models.Channel, error) {
	var channel models.Channel
	if err := database.Database.First(&channel, id).Error; err != nil {
		return nil, err
	}
	return &channel, nil
}

func CreateChannel(name string, serverID models.Snowflake, channelType string, topic string) (*models.Channel, error) {
	var maxPos int
	database.Database.Model(&models.Channel{}).Where("server_id = ?", serverID).Select("COALESCE(MAX(position), -1)").Scan(&maxPos)

	channel := models.Channel{
		Name:     name,
		ServerID: serverID,
		Type:     channelType,
		Topic:    topic,
		Position: maxPos + 1,
	}
	if err := database.Database.Create(&channel).Error; err != nil {
		return nil, err
	}
	cache.InvalidateServerChannels(serverID)
	return &channel, nil
}

func UpdateChannel(id models.Snowflake, name string, topic *string) (*models.Channel, error) {
	var channel models.Channel
	if err := database.Database.First(&channel, id).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	if topic != nil {
		updates["topic"] = *topic
	}

	if err := database.Database.Model(&channel).Updates(updates).Error; err != nil {
		return nil, err
	}
	cache.InvalidateServerChannels(channel.ServerID)
	return &channel, nil
}

func DeleteChannel(id models.Snowflake, serverID models.Snowflake) error {
	if err := database.Database.Delete(&models.Channel{}, id).Error; err != nil {
		return err
	}
	// Clear system_channel_id if it pointed to the deleted channel
	database.Database.Model(&models.Server{}).
		Where("id = ? AND system_channel_id = ?", serverID, id).
		Update("system_channel_id", nil)
	cache.InvalidateServerChannels(serverID)
	return nil
}
