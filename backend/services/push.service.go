package services

import (
	"context"

	"chatcoal/cache"
	"chatcoal/database"
	"chatcoal/models"

	"firebase.google.com/go/v4/messaging"
	"github.com/gofiber/fiber/v2/log"
)

var fcmClient *messaging.Client

// InitFCM sets the Firebase Cloud Messaging client.
// The caller (server.go) creates the client from the Firebase app.
func InitFCM(client *messaging.Client) {
	fcmClient = client
}

// RegisterDeviceToken upserts a device token for a user.
func RegisterDeviceToken(userID models.Snowflake, token, platform string) error {
	var existing models.DeviceToken
	err := database.Database.Where("token = ?", token).First(&existing).Error
	if err == nil {
		// Token exists — update user association if needed
		if existing.UserID != userID || existing.Platform != platform {
			existing.UserID = userID
			existing.Platform = platform
			return database.Database.Save(&existing).Error
		}
		return nil
	}
	dt := models.DeviceToken{
		UserID:   userID,
		Token:    token,
		Platform: platform,
	}
	return database.Database.Create(&dt).Error
}

// UnregisterDeviceToken removes a device token.
func UnregisterDeviceToken(token string) error {
	return database.Database.Where("token = ?", token).Delete(&models.DeviceToken{}).Error
}

// UnregisterUserDeviceTokens removes all device tokens for a user.
func UnregisterUserDeviceTokens(userID models.Snowflake) error {
	return database.Database.Where("user_id = ?", userID).Delete(&models.DeviceToken{}).Error
}

// GetUserDeviceTokens fetches all device tokens for a user.
func GetUserDeviceTokens(userID models.Snowflake) ([]models.DeviceToken, error) {
	var tokens []models.DeviceToken
	err := database.Database.Where("user_id = ?", userID).Find(&tokens).Error
	return tokens, err
}

// isUserOnline checks if a user is currently connected via WebSocket.
// Uses Redis presence cache to avoid import cycle with ws package.
func isUserOnline(userID models.Snowflake) bool {
	online := cache.GetOnlineUserIDs()
	if online == nil {
		return false // No Redis, assume offline (will send push)
	}
	return online[userID]
}

// SendPushToUser sends a push notification to a user if they are offline.
func SendPushToUser(userID models.Snowflake, title, body string, data map[string]string) {
	if fcmClient == nil {
		return
	}

	if isUserOnline(userID) {
		return
	}

	tokens, err := GetUserDeviceTokens(userID)
	if err != nil || len(tokens) == 0 {
		return
	}

	for _, dt := range tokens {
		msg := &messaging.Message{
			Token: dt.Token,
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: data,
		}
		if _, err := fcmClient.Send(context.Background(), msg); err != nil {
			if messaging.IsUnregistered(err) || messaging.IsInvalidArgument(err) {
				database.Database.Where("token = ?", dt.Token).Delete(&models.DeviceToken{})
			} else {
				log.Warnf("[push] send to user %d failed: %v", userID, err)
			}
		}
	}
}

// SendPushToUsers sends push notifications to multiple users, skipping online users.
func SendPushToUsers(userIDs []models.Snowflake, title, body string, data map[string]string) {
	if fcmClient == nil || len(userIDs) == 0 {
		return
	}

	online := cache.GetOnlineUserIDs()

	for _, uid := range userIDs {
		if online != nil && online[uid] {
			continue
		}
		tokens, err := GetUserDeviceTokens(uid)
		if err != nil || len(tokens) == 0 {
			continue
		}
		for _, dt := range tokens {
			msg := &messaging.Message{
				Token: dt.Token,
				Notification: &messaging.Notification{
					Title: title,
					Body:  body,
				},
				Data: data,
			}
			if _, err := fcmClient.Send(context.Background(), msg); err != nil {
				if messaging.IsUnregistered(err) || messaging.IsInvalidArgument(err) {
					database.Database.Where("token = ?", dt.Token).Delete(&models.DeviceToken{})
				}
			}
		}
	}
}

// truncateBody truncates a message body for push notification display.
func truncateBody(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// userDisplayTitle returns a user's display name or username for notification title.
func userDisplayTitle(user *models.User) string {
	if user.DisplayName != "" {
		return user.DisplayName
	}
	if user.Username != nil {
		return *user.Username
	}
	return "Someone"
}

// SendDMPush sends a push notification for a new DM message.
func SendDMPush(recipientID, authorID models.Snowflake, dmChannelID models.Snowflake, content string) {
	if fcmClient == nil {
		return
	}

	author, err := GetUserByID(authorID)
	if err != nil || author == nil {
		return
	}

	SendPushToUser(recipientID, userDisplayTitle(author), truncateBody(content, 100), map[string]string{
		"type":          "dm",
		"dm_channel_id": dmChannelID.String(),
	})
}

// SendChannelPush sends push notifications for a new channel message.
// mentionedIDs is the set of user IDs mentioned in the message (for "mentions" mode filtering).
func SendChannelPush(channelID, serverID, authorID models.Snowflake, content string, memberUserIDs []models.Snowflake, notifyModes map[models.Snowflake]string, mentionedIDs []models.Snowflake) {
	if fcmClient == nil || len(memberUserIDs) == 0 {
		return
	}

	author, err := GetUserByID(authorID)
	if err != nil || author == nil {
		return
	}

	title := userDisplayTitle(author)
	body := truncateBody(content, 100)

	// Build set for O(1) mention lookup
	mentionSet := make(map[models.Snowflake]bool, len(mentionedIDs))
	for _, id := range mentionedIDs {
		mentionSet[id] = true
	}

	var recipients []models.Snowflake
	for _, uid := range memberUserIDs {
		if uid == authorID {
			continue
		}
		mode := notifyModes[uid]
		switch mode {
		case "nothing":
			continue
		case "mentions":
			if !mentionSet[uid] {
				continue
			}
		}
		recipients = append(recipients, uid)
	}

	data := map[string]string{
		"type":       "channel",
		"channel_id": channelID.String(),
		"server_id":  serverID.String(),
	}

	SendPushToUsers(recipients, title, body, data)
}

// SendMentionPush sends push notifications to mentioned users directly.
// Called from message creation so that actual content is available for the push body.
func SendMentionPush(message *models.Message, mentionedIDs []models.Snowflake, serverID models.Snowflake) {
	if fcmClient == nil || len(mentionedIDs) == 0 {
		return
	}

	author, err := GetUserByID(message.AuthorID)
	if err != nil || author == nil {
		return
	}

	// Resolve <@id> to @name in push body
	body := truncateBody(ResolveMentionsForDisplay(message.Content, mentionedIDs), 100)
	title := userDisplayTitle(author) + " mentioned you"

	// Get notify modes for all mentioned users to skip muted/nothing users
	notifyModes := GetNotifyModesForChannel(message.ChannelID, serverID)

	var recipients []models.Snowflake
	for _, uid := range mentionedIDs {
		if uid == message.AuthorID {
			continue
		}
		mode := notifyModes[uid]
		if mode == "nothing" {
			continue
		}
		recipients = append(recipients, uid)
	}

	data := map[string]string{
		"type":       "mention",
		"channel_id": message.ChannelID.String(),
		"server_id":  serverID.String(),
		"message_id": message.ID.String(),
	}

	SendPushToUsers(recipients, title, body, data)
}
