package services

import (
	"time"

	"chatcoal/database"
	"chatcoal/models"
	"chatcoal/storage"

	"github.com/gofiber/fiber/v2/log"
)

const DMMessagePageLimit = 50

// DMLastMsg is a lightweight struct for DM channel enrichment.
type DMLastMsg struct {
	ID        models.Snowflake `json:"id"`
	Content   string           `json:"content"`
	AuthorID  models.Snowflake `json:"author_id"`
	CreatedAt time.Time        `json:"created_at"`
}

// GetOrCreateDMChannel finds or creates a DM channel between two users.
// Always stores User1ID < User2ID for uniqueness.
func GetOrCreateDMChannel(userID1, userID2 models.Snowflake) (*models.DMChannel, error) {
	low, high := userID1, userID2
	if low > high {
		low, high = high, low
	}

	var channel models.DMChannel
	err := database.Database.
		Where("user1_id = ? AND user2_id = ?", low, high).
		First(&channel).Error

	if err == nil {
		database.Database.Preload("User1").Preload("User2").First(&channel, channel.ID)
		return &channel, nil
	}

	channel = models.DMChannel{
		User1ID: low,
		User2ID: high,
	}
	if err := database.Database.Create(&channel).Error; err != nil {
		return nil, err
	}
	database.Database.Preload("User1").Preload("User2").First(&channel, channel.ID)
	return &channel, nil
}

// GetDMChannelsByUserID returns all DM channels for a user.
func GetDMChannelsByUserID(userID models.Snowflake) ([]models.DMChannel, error) {
	var channels []models.DMChannel
	err := database.Database.
		Preload("User1").Preload("User2").
		Where("user1_id = ? OR user2_id = ?", userID, userID).
		Order("updated_at DESC").
		Find(&channels).Error
	return channels, err
}

// GetDMChannelByID returns a DM channel by ID with users preloaded.
func GetDMChannelByID(id models.Snowflake) (*models.DMChannel, error) {
	var channel models.DMChannel
	err := database.Database.Preload("User1").Preload("User2").First(&channel, id).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// IsDMParticipant checks if a user is part of a DM channel.
// Uses two queries instead of OR to allow composite index usage.
func IsDMParticipant(userID, channelID models.Snowflake) bool {
	var ch models.DMChannel
	if database.Database.Select("id").Where("id = ? AND user1_id = ?", channelID, userID).First(&ch).Error == nil {
		return true
	}
	return database.Database.Select("id").Where("id = ? AND user2_id = ?", channelID, userID).First(&ch).Error == nil
}

// GetOtherUserID returns the other user's ID in a DM channel.
func GetOtherUserID(channel *models.DMChannel, userID models.Snowflake) models.Snowflake {
	if channel.User1ID == userID {
		return channel.User2ID
	}
	return channel.User1ID
}

// GetDMMessages returns paginated messages for a DM channel.
func GetDMMessages(dmChannelID models.Snowflake, before models.Snowflake) ([]models.DMMessage, error) {
	var messages []models.DMMessage
	query := database.Database.Preload("Author").Preload("Reactions").Where("dm_channel_id = ?", dmChannelID)

	if before > 0 {
		query = query.Where("id < ?", before)
	}

	err := query.Order("id DESC").Limit(DMMessagePageLimit).Find(&messages).Error
	return messages, err
}

// CreateDMMessage creates a new DM message and returns it with author preloaded.
func CreateDMMessage(content string, dmChannelID models.Snowflake, authorID models.Snowflake, fileURL, fileName string, fileSize int64, imageWidth, imageHeight int) (*models.DMMessage, error) {
	message := models.DMMessage{
		Content:     content,
		DMChannelID: dmChannelID,
		AuthorID:    authorID,
		FileURL:     fileURL,
		FileName:    fileName,
		FileSize:    fileSize,
		ImageWidth:  imageWidth,
		ImageHeight: imageHeight,
	}
	if err := database.Database.Create(&message).Error; err != nil {
		return nil, err
	}
	// Update the DM channel's updated_at so it sorts to top
	database.Database.Model(&models.DMChannel{}).Where("id = ?", dmChannelID).Update("updated_at", message.CreatedAt)
	// Reload with author and reactions
	database.Database.Preload("Author").Preload("Reactions").First(&message, message.ID)

	go IncrementUnreadForDM(dmChannelID, authorID)
	return &message, nil
}

// UpdateDMMessage edits a DM message (author-only).
func UpdateDMMessage(id models.Snowflake, authorID models.Snowflake, content string) (*models.DMMessage, error) {
	var message models.DMMessage
	if err := database.Database.First(&message, id).Error; err != nil {
		return nil, err
	}

	if message.AuthorID != authorID {
		return nil, fiber_forbidden()
	}

	message.Content = content
	message.Edited = true
	if err := database.Database.Save(&message).Error; err != nil {
		return nil, err
	}
	database.Database.Preload("Author").First(&message, message.ID)
	return &message, nil
}

// DeleteDMMessage deletes a DM message (author-only).
func DeleteDMMessage(id models.Snowflake, authorID models.Snowflake) error {
	var message models.DMMessage
	if err := database.Database.First(&message, id).Error; err != nil {
		return err
	}

	if message.AuthorID != authorID {
		return fiber_forbidden()
	}

	if err := storage.DeleteFileByURLErr(message.FileURL); err != nil {
		log.Errorf("[DeleteDMMessage] S3 delete failed for %q: %v — aborting DB delete", message.FileURL, err)
		return err
	}
	if err := database.Database.Delete(&message).Error; err != nil {
		return err
	}
	return nil
}

// GetLastDMMessage returns the most recent message in a DM channel.
func GetLastDMMessage(dmChannelID models.Snowflake) (*models.DMMessage, error) {
	var message models.DMMessage
	err := database.Database.
		Preload("Author").
		Where("dm_channel_id = ?", dmChannelID).
		Order("id DESC").
		First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// GetLastDMMessagesBulk returns the most recent message for each given channel in a single query.
func GetLastDMMessagesBulk(channelIDs []models.Snowflake) (map[models.Snowflake]*DMLastMsg, error) {
	if len(channelIDs) == 0 {
		return map[models.Snowflake]*DMLastMsg{}, nil
	}

	// Subquery: get max message ID per channel
	var rows []struct {
		ID          models.Snowflake `gorm:"column:id"`
		DMChannelID models.Snowflake `gorm:"column:dm_channel_id"`
		Content     string           `gorm:"column:content"`
		AuthorID    models.Snowflake `gorm:"column:author_id"`
		CreatedAt   time.Time        `gorm:"column:created_at"`
	}

	err := database.Database.Raw(`
		SELECT m.id, m.dm_channel_id, m.content, m.author_id, m.created_at
		FROM dm_messages m
		INNER JOIN (
			SELECT dm_channel_id, MAX(id) AS max_id
			FROM dm_messages
			WHERE dm_channel_id IN ?
			GROUP BY dm_channel_id
		) latest ON m.id = latest.max_id
	`, channelIDs).Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	result := make(map[models.Snowflake]*DMLastMsg, len(rows))
	for _, r := range rows {
		result[r.DMChannelID] = &DMLastMsg{
			ID:        r.ID,
			Content:   r.Content,
			AuthorID:  r.AuthorID,
			CreatedAt: r.CreatedAt,
		}
	}
	return result, nil
}

// GetDMMessageForDelete returns lightweight info for a DM message.
func GetDMMessageForDelete(id models.Snowflake, dest interface{}) {
	database.Database.Model(&models.DMMessage{}).Select("id, dm_channel_id").Where("id = ?", id).Scan(dest)
}
