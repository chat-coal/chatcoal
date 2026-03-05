package controllers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"chatcoal/cache"
	"chatcoal/models"
	"chatcoal/services"
	"chatcoal/ws"

	"github.com/gofiber/fiber/v2"
)

// parseSnowflakeParam extracts a route parameter as a Snowflake.
func parseSnowflakeParam(c *fiber.Ctx, name string) (models.Snowflake, error) {
	s := c.Params(name)
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s", name)
	}
	return models.Snowflake(v), nil
}

// parseSnowflakeQuery extracts an optional query parameter as a Snowflake.
func parseSnowflakeQuery(c *fiber.Ctx, name string) models.Snowflake {
	s := c.Query(name)
	if s == "" {
		return 0
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return models.Snowflake(v)
}

// parseSnowflakeString parses a raw string as a Snowflake.
func parseSnowflakeString(s string) (models.Snowflake, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return models.Snowflake(v), nil
}

func GetMessages(c *fiber.Ctx) error {
	channelID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid channel ID"})
	}

	ch, err := services.GetChannelByID(channelID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Channel not found"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, ch.ServerID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	before := parseSnowflakeQuery(c, "before")

	messages, err := services.GetMessages(channelID, before, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch messages"})
	}

	return c.JSON(messages)
}

var allowedFileExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".pdf": true, ".txt": true, ".zip": true,
}

func SendMessage(c *fiber.Ctx) error {
	channelID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid channel ID"})
	}

	ch, err := services.GetChannelByID(channelID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Channel not found"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, ch.ServerID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	if user.IsRestricted() && !cache.AnonMessageRateLimitOK(int64(user.ID)) {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "Sending too fast. Verify your email to remove this limit"})
	}

	var content string
	var fileURL, fileName string
	var fileSize int64
	var imageWidth, imageHeight int
	var replyToID *models.Snowflake

	file, fileErr := c.FormFile("file")
	if fileErr == nil && file != nil && user.IsRestricted() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Verify your email to unlock attachments"})
	}
	if fileErr == nil && file != nil {
		// Multipart upload
		content = c.FormValue("content")
		if r := c.FormValue("reply_to_id"); r != "" {
			if val, err := strconv.ParseInt(r, 10, 64); err == nil {
				rid := models.Snowflake(val)
				replyToID = &rid
			}
		}
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !allowedFileExts[ext] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File type not allowed"})
		}
		if err := checkMagicBytes(file); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File type not allowed"})
		}
		savedName := fmt.Sprintf("%d_%s%s", user.ID, stamp(), ext)
		url, err := uploadFile(c, file, "", savedName)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
		}
		fileURL = url
		fileName = file.Filename
		fileSize = file.Size
		imageWidth, imageHeight = getImageDimensions(file)
	} else {
		// JSON body
		var body struct {
			Content   string            `json:"content" validate:"required,max=4000"`
			ReplyToID *models.Snowflake `json:"reply_to_id"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
		}
		if msg := validateBody(&body); msg != "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
		}
		content = body.Content
		replyToID = body.ReplyToID
	}

	if len([]rune(content)) > 4000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Content is too long (max 4000 characters)"})
	}

	message, err := services.CreateMessage(content, channelID, ch.ServerID, user.ID, fileURL, fileName, fileSize, imageWidth, imageHeight, replyToID, nil)
	if err != nil {
		if services.IsValidationError(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": services.ValidationErrorMessage(err)})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send message"})
	}

	// Broadcast via WebSocket
	broadcastEvent(ch.ServerID, "message", message)

	// Relay to federated channels if federation is enabled on this channel.
	if ch.FederationID != nil {
		go services.RelayFederatedMessage(ch, message, user)
	}

	// Fetch link embeds in the background
	if content != "" {
		msgID := message.ID
		srvID := ch.ServerID
		chID := message.ChannelID
		go services.FetchAndStoreEmbeds(msgID, content, "messages", func(embeds json.RawMessage) {
			broadcastEvent(srvID, "message_embed_update", fiber.Map{
				"id":         msgID,
				"channel_id": chID,
				"embeds":     embeds,
			})
		})
	}

	return c.Status(fiber.StatusCreated).JSON(message)
}

func EditMessage(c *fiber.Ctx) error {
	messageID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid message ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
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

	message, err := services.UpdateMessage(messageID, user.ID, body.Content)
	if err != nil {
		if services.IsForbiddenError(err) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to edit message"})
	}

	ch, _ := services.GetChannelByID(message.ChannelID)
	if ch != nil {
		if message.ForumPostID != nil {
			broadcastEvent(ch.ServerID, "forum_message_edit", message)
		} else {
			broadcastEvent(ch.ServerID, "message_edit", message)
		}

		// Re-fetch embeds on content change
		if body.Content != "" {
			msgID := message.ID
			srvID := ch.ServerID
			chID := message.ChannelID
			go services.FetchAndStoreEmbeds(msgID, body.Content, "messages", func(embeds json.RawMessage) {
				broadcastEvent(srvID, "message_embed_update", fiber.Map{
					"id":         msgID,
					"channel_id": chID,
					"embeds":     embeds,
				})
			})
		}
	}

	return c.JSON(message)
}

func DeleteMessageHandler(c *fiber.Ctx) error {
	messageID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid message ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	message, err := services.GetMessageByID(messageID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Message not found"})
	}

	ch, _ := services.GetChannelByID(message.ChannelID)

	if message.AuthorID != user.ID {
		// Not the author — check if user has manage_messages permission
		if ch == nil || !services.HasPermission(user.ID, ch.ServerID, services.PermManageMessages) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
		}
		// Has permission — force delete
		if err := services.ForceDeleteMessage(messageID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete message"})
		}
	} else {
		if err := services.DeleteMessage(messageID, user.ID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete message"})
		}
	}

	if ch != nil {
		if message.ForumPostID != nil {
			broadcastEvent(ch.ServerID, "forum_message_delete", fiber.Map{"id": messageID, "channel_id": message.ChannelID, "forum_post_id": message.ForumPostID})
		} else {
			broadcastEvent(ch.ServerID, "message_delete", fiber.Map{"id": messageID, "channel_id": message.ChannelID})
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func SearchMessages(c *fiber.Ctx) error {
	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, serverID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Query is required"})
	}
	if len(query) > 200 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Query too long (max 200 characters)"})
	}

	before := parseSnowflakeQuery(c, "before")

	results, err := services.SearchMessages(serverID, query, before)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Search failed"})
	}

	return c.JSON(results)
}

func GetPinnedMessages(c *fiber.Ctx) error {
	channelID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid channel ID"})
	}

	ch, err := services.GetChannelByID(channelID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Channel not found"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, ch.ServerID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	pins, err := services.GetPinnedMessages(channelID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch pinned messages"})
	}

	return c.JSON(pins)
}

func PinMessageHandler(c *fiber.Ctx) error {
	messageID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid message ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	message, err := services.GetMessageByID(messageID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Message not found"})
	}

	ch, err := services.GetChannelByID(message.ChannelID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Channel not found"})
	}

	if !services.HasPermission(user.ID, ch.ServerID, services.PermManageMessages) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Missing permission"})
	}

	pin, err := services.PinMessage(messageID, message.ChannelID, user.ID)
	if err != nil {
		if services.IsValidationError(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": services.ValidationErrorMessage(err)})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to pin message"})
	}

	broadcastEvent(ch.ServerID, "message_pin", fiber.Map{
		"channel_id": message.ChannelID,
		"pin":        pin,
	})

	return c.Status(fiber.StatusCreated).JSON(pin)
}

func UnpinMessageHandler(c *fiber.Ctx) error {
	messageID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid message ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	message, err := services.GetMessageByID(messageID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Message not found"})
	}

	ch, err := services.GetChannelByID(message.ChannelID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Channel not found"})
	}

	if !services.HasPermission(user.ID, ch.ServerID, services.PermManageMessages) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Missing permission"})
	}

	if err := services.UnpinMessage(messageID, message.ChannelID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to unpin message"})
	}

	broadcastEvent(ch.ServerID, "message_unpin", fiber.Map{
		"message_id": messageID,
		"channel_id": message.ChannelID,
	})

	return c.SendStatus(fiber.StatusNoContent)
}

func GetUserProfile(c *fiber.Ctx) error {
	targetID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	user, err := services.GetUserByID(targetID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(fiber.Map{
		"id":           user.ID,
		"display_name": user.DisplayName,
		"avatar_url":   user.AvatarURL,
		"status":       user.Status,
		"created_at":   user.CreatedAt,
		"deleted":      strings.HasPrefix(user.FirebaseUID, "deleted:"),
	})
}

func broadcastEvent(serverID models.Snowflake, eventType string, data interface{}) {
	payload, err := json.Marshal(map[string]interface{}{
		"type":      eventType,
		"server_id": serverID,
		"data":      data,
	})
	if err != nil {
		return
	}
	ws.GetHub().Broadcast(serverID, payload)
}
