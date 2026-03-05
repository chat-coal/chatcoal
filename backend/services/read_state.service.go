package services

import (
	"chatcoal/database"
	"chatcoal/models"
)

// UpdateReadState upserts a read_state row for a user+channel, resetting the
// unread counter to 0. server_id is not updated here (it is set by
// IncrementUnreadForChannel and stays correct across future increments).
func UpdateReadState(userID models.Snowflake, channelType string, channelRefID models.Snowflake, lastReadMessageID models.Snowflake) error {
	id := models.GenerateID()
	return database.Database.Exec(`
		INSERT INTO read_states (id, user_id, channel_type, channel_ref_id, server_id, last_read_message_id, unread_count, updated_at)
		VALUES (?, ?, ?, ?, 0, ?, 0, NOW())
		ON DUPLICATE KEY UPDATE last_read_message_id = VALUES(last_read_message_id), unread_count = 0, updated_at = NOW()
	`, id, userID, channelType, channelRefID, lastReadMessageID).Error
}

// UnreadCount holds the unread count for a channel.
type UnreadCount struct {
	ChannelType  string           `json:"channel_type"`
	ChannelRefID models.Snowflake `json:"channel_ref_id"`
	ServerID     models.Snowflake `json:"server_id,omitempty"`
	Count        int              `json:"count"`
}

// GetAllUnreadCounts returns unread counts for all channels where the user has
// pending messages. This is now a simple indexed scan on read_states instead of
// the previous expensive multi-table UNION query.
func GetAllUnreadCounts(userID models.Snowflake) ([]UnreadCount, error) {
	var rows []struct {
		ChannelType  string           `gorm:"column:channel_type"`
		ChannelRefID models.Snowflake `gorm:"column:channel_ref_id"`
		ServerID     models.Snowflake `gorm:"column:server_id"`
		Count        int              `gorm:"column:count"`
	}

	err := database.Database.Raw(`
		SELECT channel_type, channel_ref_id, server_id, unread_count AS count
		FROM read_states
		WHERE user_id = ? AND unread_count > 0
	`, userID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	counts := make([]UnreadCount, len(rows))
	for i, r := range rows {
		counts[i] = UnreadCount{
			ChannelType:  r.ChannelType,
			ChannelRefID: r.ChannelRefID,
			ServerID:     r.ServerID,
			Count:        r.Count,
		}
	}
	return counts, nil
}

// IncrementUnreadForChannel enqueues an unread increment for every server
// member except the author. The actual DB write is batched by the unread
// batcher (see unread_batcher.go) and flushed every ~2 seconds.
func IncrementUnreadForChannel(channelID, serverID, authorID models.Snowflake) {
	EnqueueUnread(channelID, serverID, authorID)
}

// IncrementUnreadForDM atomically increments the unread_count for the
// recipient of a DM message. Called in a goroutine after a DM is saved.
func IncrementUnreadForDM(dmChannelID, authorID models.Snowflake) {
	var ch models.DMChannel
	if err := database.Database.Select("user1_id, user2_id").First(&ch, dmChannelID).Error; err != nil {
		return
	}

	recipientID := ch.User1ID
	if recipientID == authorID {
		recipientID = ch.User2ID
	}

	database.Database.Exec(`
		INSERT INTO read_states (id, user_id, channel_type, channel_ref_id, server_id, last_read_message_id, unread_count, updated_at)
		VALUES (?, ?, 'dm', ?, 0, 0, 1, NOW())
		ON DUPLICATE KEY UPDATE unread_count = unread_count + 1, updated_at = NOW()
	`, models.GenerateID(), recipientID, dmChannelID)
}
