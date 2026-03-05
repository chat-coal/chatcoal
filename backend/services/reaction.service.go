package services

import (
	"errors"

	"github.com/go-sql-driver/mysql"

	"chatcoal/database"
	"chatcoal/models"
)

func isDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

type ReactionGroup struct {
	Emoji   string             `json:"emoji"`
	Count   int                `json:"count"`
	UserIDs []models.Snowflake `json:"user_ids"`
}

// ToggleReaction atomically toggles a reaction on a server message.
// Returns true if added, false if removed.
// Uses try-insert-then-delete to avoid the check-then-act race condition.
func ToggleReaction(messageID, userID models.Snowflake, emoji string) (bool, error) {
	reaction := models.MessageReaction{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
	}
	err := database.Database.Create(&reaction).Error
	if err == nil {
		return true, nil
	}
	if isDuplicateKeyError(err) {
		// Already existed — delete it
		delErr := database.Database.
			Where("message_id = ? AND user_id = ? AND emoji = ?", messageID, userID, emoji).
			Delete(&models.MessageReaction{}).Error
		return false, delErr
	}
	return false, err
}

const ReactionLimit = 500

// GetReactions returns grouped reactions for a server message.
func GetReactions(messageID models.Snowflake) []ReactionGroup {
	var reactions []models.MessageReaction
	database.Database.Where("message_id = ?", messageID).Order("id ASC").Limit(ReactionLimit).Find(&reactions)

	return groupReactions(reactions)
}

// ToggleDMReaction atomically toggles a reaction on a DM message.
// Returns true if added, false if removed.
func ToggleDMReaction(dmMessageID, userID models.Snowflake, emoji string) (bool, error) {
	reaction := models.DMMessageReaction{
		DMMessageID: dmMessageID,
		UserID:      userID,
		Emoji:       emoji,
	}
	err := database.Database.Create(&reaction).Error
	if err == nil {
		return true, nil
	}
	if isDuplicateKeyError(err) {
		delErr := database.Database.
			Where("dm_message_id = ? AND user_id = ? AND emoji = ?", dmMessageID, userID, emoji).
			Delete(&models.DMMessageReaction{}).Error
		return false, delErr
	}
	return false, err
}

// GetDMReactions returns grouped reactions for a DM message.
func GetDMReactions(dmMessageID models.Snowflake) []ReactionGroup {
	var reactions []models.DMMessageReaction
	database.Database.Where("dm_message_id = ?", dmMessageID).Order("id ASC").Limit(ReactionLimit).Find(&reactions)

	return groupDMReactions(reactions)
}

func groupReactions(reactions []models.MessageReaction) []ReactionGroup {
	emojiMap := make(map[string]*ReactionGroup)
	var order []string

	for _, r := range reactions {
		if _, ok := emojiMap[r.Emoji]; !ok {
			emojiMap[r.Emoji] = &ReactionGroup{Emoji: r.Emoji}
			order = append(order, r.Emoji)
		}
		g := emojiMap[r.Emoji]
		g.Count++
		g.UserIDs = append(g.UserIDs, r.UserID)
	}

	result := make([]ReactionGroup, 0, len(order))
	for _, emoji := range order {
		result = append(result, *emojiMap[emoji])
	}
	return result
}

func groupDMReactions(reactions []models.DMMessageReaction) []ReactionGroup {
	emojiMap := make(map[string]*ReactionGroup)
	var order []string

	for _, r := range reactions {
		if _, ok := emojiMap[r.Emoji]; !ok {
			emojiMap[r.Emoji] = &ReactionGroup{Emoji: r.Emoji}
			order = append(order, r.Emoji)
		}
		g := emojiMap[r.Emoji]
		g.Count++
		g.UserIDs = append(g.UserIDs, r.UserID)
	}

	result := make([]ReactionGroup, 0, len(order))
	for _, emoji := range order {
		result = append(result, *emojiMap[emoji])
	}
	return result
}
