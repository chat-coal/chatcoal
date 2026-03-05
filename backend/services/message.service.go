package services

import (
	"chatcoal/database"
	"chatcoal/models"
	"chatcoal/storage"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
)

const MessagePageLimit = 50

func GetMessages(channelID models.Snowflake, before models.Snowflake, forumPostID *models.Snowflake) ([]models.Message, error) {
	var messages []models.Message
	query := database.Database.
		Preload("Author").
		Preload("Reactions").
		Preload("ReplyTo", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, content, author_id").Preload("Author", func(db *gorm.DB) *gorm.DB {
				return db.Select("id, display_name, avatar_url")
			})
		}).
		Where("channel_id = ?", channelID)

	if forumPostID != nil {
		query = query.Where("forum_post_id = ?", *forumPostID)
	} else {
		query = query.Where("forum_post_id IS NULL")
	}

	if before > 0 {
		query = query.Where("id < ?", before)
	}

	err := query.Order("id DESC").Limit(MessagePageLimit).Find(&messages).Error
	return messages, err
}

func CreateMessage(content string, channelID, serverID, authorID models.Snowflake, fileURL, fileName string, fileSize int64, imageWidth, imageHeight int, replyToID *models.Snowflake, forumPostID *models.Snowflake) (*models.Message, error) {
	if replyToID != nil {
		var replyMsg models.Message
		if err := database.Database.Select("id, channel_id").First(&replyMsg, *replyToID).Error; err != nil {
			return nil, newValidationError("reply_to message not found")
		}
		if replyMsg.ChannelID != channelID {
			return nil, newValidationError("reply_to message is in a different channel")
		}
	}

	message := models.Message{
		Content:     content,
		ChannelID:   channelID,
		AuthorID:    authorID,
		FileURL:     fileURL,
		FileName:    fileName,
		FileSize:    fileSize,
		ImageWidth:  imageWidth,
		ImageHeight: imageHeight,
		ReplyToID:   replyToID,
		ForumPostID: forumPostID,
	}
	if err := database.Database.Create(&message).Error; err != nil {
		return nil, err
	}

	// Update forum post reply_count and last_reply_at
	if forumPostID != nil {
		now := time.Now()
		database.Database.Model(&models.ForumPost{}).Where("id = ?", *forumPostID).
			Updates(map[string]interface{}{
				"reply_count":   gorm.Expr("reply_count + 1"),
				"last_reply_at": now,
			})
	}

	// Reload with author, reactions, and reply_to
	database.Database.
		Preload("Author").
		Preload("Reactions").
		Preload("ReplyTo", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, content, author_id").Preload("Author", func(db *gorm.DB) *gorm.DB {
				return db.Select("id, display_name, avatar_url")
			})
		}).
		First(&message, message.ID)

	IncrementUnreadForChannel(channelID, serverID, authorID)
	return &message, nil
}

func CreateSystemMessage(msgType, content string, channelID, serverID, authorID models.Snowflake) (*models.Message, error) {
	message := models.Message{
		Content:   content,
		Type:      msgType,
		ChannelID: channelID,
		AuthorID:  authorID,
	}
	if err := database.Database.Create(&message).Error; err != nil {
		return nil, err
	}

	database.Database.Preload("Author").First(&message, message.ID)

	IncrementUnreadForChannel(channelID, serverID, authorID)
	return &message, nil
}

func UpdateMessage(id models.Snowflake, authorID models.Snowflake, content string) (*models.Message, error) {
	var message models.Message
	if err := database.Database.First(&message, id).Error; err != nil {
		return nil, err
	}

	if message.Type != "user" {
		return nil, newValidationError("system messages cannot be edited")
	}

	if message.AuthorID != authorID {
		return nil, fiber_forbidden()
	}

	message.Content = content
	message.Edited = true
	if err := database.Database.Save(&message).Error; err != nil {
		return nil, err
	}
	database.Database.
		Preload("Author").
		Preload("ReplyTo", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, content, author_id").Preload("Author", func(db *gorm.DB) *gorm.DB {
				return db.Select("id, display_name, avatar_url")
			})
		}).
		First(&message, message.ID)
	return &message, nil
}

func DeleteMessage(id models.Snowflake, authorID models.Snowflake) error {
	var message models.Message
	if err := database.Database.First(&message, id).Error; err != nil {
		return err
	}

	if message.Type != "user" {
		return newValidationError("system messages cannot be deleted")
	}

	if message.AuthorID != authorID {
		return fiber_forbidden()
	}

	if err := storage.DeleteFileByURLErr(message.FileURL); err != nil {
		log.Errorf("[DeleteMessage] S3 delete failed for %q: %v — aborting DB delete", message.FileURL, err)
		return err
	}
	if err := database.Database.Delete(&message).Error; err != nil {
		return err
	}
	decrementForumReplyCount(message.ForumPostID)
	return nil
}

type validationError struct{ msg string }

func (e *validationError) Error() string { return e.msg }
func newValidationError(msg string) error { return &validationError{msg: msg} }
func IsValidationError(err error) bool    { _, ok := err.(*validationError); return ok }
func ValidationErrorMessage(err error) string {
	if e, ok := err.(*validationError); ok {
		return e.msg
	}
	return ""
}

type forbiddenError struct{}

func (e *forbiddenError) Error() string {
	return "forbidden"
}

func fiber_forbidden() error {
	return &forbiddenError{}
}

func IsForbiddenError(err error) bool {
	_, ok := err.(*forbiddenError)
	return ok
}

func GetMessageForDelete(id models.Snowflake, dest interface{}) {
	database.Database.Model(&models.Message{}).Select("id, channel_id").Where("id = ?", id).Scan(dest)
}

func GetMessageByID(id models.Snowflake) (*models.Message, error) {
	var message models.Message
	if err := database.Database.First(&message, id).Error; err != nil {
		return nil, err
	}
	return &message, nil
}

func ForceDeleteMessage(id models.Snowflake) error {
	var message models.Message
	if err := database.Database.First(&message, id).Error; err != nil {
		return err
	}
	if err := storage.DeleteFileByURLErr(message.FileURL); err != nil {
		log.Errorf("[ForceDeleteMessage] S3 delete failed for %q: %v — aborting DB delete", message.FileURL, err)
		return err
	}
	if err := database.Database.Delete(&message).Error; err != nil {
		return err
	}
	decrementForumReplyCount(message.ForumPostID)
	return nil
}

func decrementForumReplyCount(forumPostID *models.Snowflake) {
	if forumPostID != nil {
		database.Database.Model(&models.ForumPost{}).Where("id = ?", *forumPostID).
			Update("reply_count", gorm.Expr("GREATEST(reply_count - 1, 0)"))
	}
}
