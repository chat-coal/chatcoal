package controllers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"chatcoal/cache"
	"chatcoal/models"
	"chatcoal/services"
	"chatcoal/ws"

	"github.com/gofiber/fiber/v2"
)

// GetDMChannels returns all DM channels for the current user, enriched with last message and unread count.
func GetDMChannels(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	channels, err := services.GetDMChannelsByUserID(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch DM channels"})
	}

	// Enrich with last message and unread counts
	unreadCounts, _ := services.GetAllUnreadCounts(user.ID)
	dmUnreadMap := make(map[models.Snowflake]int)
	for _, u := range unreadCounts {
		if u.ChannelType == "dm" {
			dmUnreadMap[u.ChannelRefID] = u.Count
		}
	}

	type enrichedChannel struct {
		ID          models.Snowflake    `json:"id"`
		User1ID     models.Snowflake    `json:"user1_id"`
		User1       interface{}         `json:"user1"`
		User2ID     models.Snowflake    `json:"user2_id"`
		User2       interface{}         `json:"user2"`
		CreatedAt   interface{}         `json:"created_at"`
		UpdatedAt   interface{}         `json:"updated_at"`
		LastMessage *services.DMLastMsg `json:"last_message"`
		UnreadCount int                 `json:"unread_count"`
	}

	// Bulk fetch last messages (single query instead of N+1)
	channelIDs := make([]models.Snowflake, len(channels))
	for i, ch := range channels {
		channelIDs[i] = ch.ID
	}
	lastMessages, _ := services.GetLastDMMessagesBulk(channelIDs)

	result := make([]enrichedChannel, 0, len(channels))
	for _, ch := range channels {
		result = append(result, enrichedChannel{
			ID:          ch.ID,
			User1ID:     ch.User1ID,
			User1:       ch.User1,
			User2ID:     ch.User2ID,
			User2:       ch.User2,
			CreatedAt:   ch.CreatedAt,
			UpdatedAt:   ch.UpdatedAt,
			LastMessage: lastMessages[ch.ID],
			UnreadCount: dmUnreadMap[ch.ID],
		})
	}

	return c.JSON(result)
}

// CreateOrGetDMChannel creates or returns an existing DM channel with the target user.
func CreateOrGetDMChannel(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var body struct {
		UserID models.Snowflake `json:"user_id"`
	}
	if err := c.BodyParser(&body); err != nil || body.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id is required"})
	}

	if body.UserID == user.ID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot DM yourself"})
	}

	if services.IsUserDeleted(body.UserID) {
		return c.Status(fiber.StatusGone).JSON(fiber.Map{"error": "This user no longer exists"})
	}

	channel, err := services.GetOrCreateDMChannel(user.ID, body.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create DM channel"})
	}

	return c.JSON(channel)
}

// GetDMMessages returns paginated messages for a DM channel.
func GetDMMessages(c *fiber.Ctx) error {
	dmChannelID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid DM channel ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsDMParticipant(user.ID, dmChannelID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	before := parseSnowflakeQuery(c, "before")

	messages, err := services.GetDMMessages(dmChannelID, before)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch messages"})
	}

	return c.JSON(messages)
}

// SendDMMessage sends a message in a DM channel.
func SendDMMessage(c *fiber.Ctx) error {
	dmChannelID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid DM channel ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsDMParticipant(user.ID, dmChannelID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	if user.IsRestricted() && !cache.AnonMessageRateLimitOK(int64(user.ID)) {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "Sending too fast. Verify your email to remove this limit"})
	}

	// Block sending to deleted users
	if dmChannel, _ := services.GetDMChannelByID(dmChannelID); dmChannel != nil {
		otherID := dmChannel.User1ID
		if otherID == user.ID {
			otherID = dmChannel.User2ID
		}
		if services.IsUserDeleted(otherID) {
			return c.Status(fiber.StatusGone).JSON(fiber.Map{"error": "This user no longer exists"})
		}
	}

	var content string
	var fileURL, fileName string
	var fileSize int64
	var imageWidth, imageHeight int

	file, fileErr := c.FormFile("file")
	if fileErr == nil && file != nil && user.IsRestricted() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Verify your email to unlock attachments"})
	}
	if fileErr == nil && file != nil {
		content = c.FormValue("content")
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !allowedFileExts[ext] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File type not allowed"})
		}
		if err := checkMagicBytes(file); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File type not allowed"})
		}
		savedName := fmt.Sprintf("dm_%d_%s%s", user.ID, stamp(), ext)
		url, err := uploadFile(c, file, "", savedName)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
		}
		fileURL = url
		fileName = file.Filename
		fileSize = file.Size
		imageWidth, imageHeight = getImageDimensions(file)
	} else {
		var body struct {
			Content string `json:"content" validate:"required,max=4000"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
		}
		if msg := validateBody(&body); msg != "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
		}
		content = body.Content
	}

	if len([]rune(content)) > 4000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Content is too long (max 4000 characters)"})
	}

	message, err := services.CreateDMMessage(content, dmChannelID, user.ID, fileURL, fileName, fileSize, imageWidth, imageHeight)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send message"})
	}

	// Get the DM channel to find both participants
	channel, _ := services.GetDMChannelByID(dmChannelID)
	if channel != nil {
		broadcastDMEvent("dm_message", message, channel.User1ID, channel.User2ID)

		// Fetch link embeds in the background
		if content != "" {
			msgID := message.ID
			u1 := channel.User1ID
			u2 := channel.User2ID
			dmChID := dmChannelID
			go services.FetchAndStoreEmbeds(msgID, content, "dm_messages", func(embeds json.RawMessage) {
				broadcastDMEvent("dm_embed_update", fiber.Map{
					"id":            msgID,
					"dm_channel_id": dmChID,
					"embeds":        embeds,
				}, u1, u2)
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(message)
}

// EditDMMessage edits a DM message.
func EditDMMessage(c *fiber.Ctx) error {
	messageID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid message ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var body struct {
		Content string `json:"content" validate:"required,max=4000"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if msg := validateBody(&body); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	message, err := services.UpdateDMMessage(messageID, user.ID, body.Content)
	if err != nil {
		if services.IsForbiddenError(err) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to edit message"})
	}

	// Get DM channel to broadcast
	channel, _ := services.GetDMChannelByID(message.DMChannelID)
	if channel != nil {
		broadcastDMEvent("dm_message_edit", message, channel.User1ID, channel.User2ID)

		// Re-fetch embeds on content change
		if body.Content != "" {
			msgID := message.ID
			u1 := channel.User1ID
			u2 := channel.User2ID
			dmChID := message.DMChannelID
			go services.FetchAndStoreEmbeds(msgID, body.Content, "dm_messages", func(embeds json.RawMessage) {
				broadcastDMEvent("dm_embed_update", fiber.Map{
					"id":            msgID,
					"dm_channel_id": dmChID,
					"embeds":        embeds,
				}, u1, u2)
			})
		}
	}

	return c.JSON(message)
}

// DeleteDMMessage deletes a DM message.
func DeleteDMMessage(c *fiber.Ctx) error {
	messageID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid message ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Get message info before deleting
	var msgInfo struct {
		ID          models.Snowflake
		DMChannelID models.Snowflake `gorm:"column:dm_channel_id"`
	}
	services.GetDMMessageForDelete(messageID, &msgInfo)

	if err := services.DeleteDMMessage(messageID, user.ID); err != nil {
		if services.IsForbiddenError(err) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete message"})
	}

	// Broadcast deletion
	if msgInfo.DMChannelID > 0 {
		channel, _ := services.GetDMChannelByID(msgInfo.DMChannelID)
		if channel != nil {
			broadcastDMEvent("dm_message_delete", fiber.Map{
				"id":            messageID,
				"dm_channel_id": msgInfo.DMChannelID,
			}, channel.User1ID, channel.User2ID)
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// broadcastDMEvent sends a WebSocket event to both DM participants.
func broadcastDMEvent(eventType string, data interface{}, user1ID, user2ID models.Snowflake) {
	payload, err := json.Marshal(map[string]interface{}{
		"type": eventType,
		"data": data,
	})
	if err != nil {
		return
	}
	hub := ws.GetHub()
	hub.SendToUserGlobal(user1ID, payload)
	hub.SendToUserGlobal(user2ID, payload)
}
