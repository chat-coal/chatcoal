package services

import (
	"chatcoal/database"
	"chatcoal/models"
	"errors"

	"gorm.io/gorm"
)

func BlockUser(blockerID, blockedID models.Snowflake) (*models.UserBlock, error) {
	if blockerID == blockedID {
		return nil, errors.New("cannot block yourself")
	}

	block := models.UserBlock{
		BlockerID: blockerID,
		BlockedID: blockedID,
	}
	if err := database.Database.Create(&block).Error; err != nil {
		return nil, err
	}
	return &block, nil
}

func UnblockUser(blockerID, blockedID models.Snowflake) error {
	result := database.Database.
		Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
		Delete(&models.UserBlock{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("not blocked")
	}
	return nil
}

func GetBlockedUsers(blockerID models.Snowflake) ([]models.UserBlock, error) {
	var blocks []models.UserBlock
	err := database.Database.
		Where("blocker_id = ?", blockerID).
		Preload("Blocked", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, firebase_uid, display_name, username, avatar_url, status, home_instance")
		}).
		Order("created_at DESC").
		Find(&blocks).Error
	return blocks, err
}

func GetBlockedUserIDs(blockerID models.Snowflake) ([]models.Snowflake, error) {
	var ids []models.Snowflake
	err := database.Database.
		Model(&models.UserBlock{}).
		Where("blocker_id = ?", blockerID).
		Pluck("blocked_id", &ids).Error
	return ids, err
}

func IsBlocked(blockerID, blockedID models.Snowflake) bool {
	var count int64
	database.Database.
		Model(&models.UserBlock{}).
		Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
		Count(&count)
	return count > 0
}

// IsEitherBlocked returns true if either user has blocked the other.
func IsEitherBlocked(userA, userB models.Snowflake) bool {
	var count int64
	database.Database.
		Model(&models.UserBlock{}).
		Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)",
			userA, userB, userB, userA).
		Count(&count)
	return count > 0
}
